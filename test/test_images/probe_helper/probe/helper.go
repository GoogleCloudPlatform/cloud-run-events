/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"

	"knative.dev/pkg/logging"

	schemasv1 "github.com/google/knative-gcp/pkg/schemas/v1"
	"github.com/google/knative-gcp/test/test_images/probe_helper/utils"
)

const (
	probeEventTimeoutExtension       = "timeout"
	probeEventTargetServiceExtension = "targetservice"
	probeEventTargetServiceHeader    = "Ce-TargetService"
)

var forwardProbeHandlers = map[string]Handler{
	BrokerE2EDeliveryProbeEventType:                brokerE2EDeliveryProbe,
	CloudPubSubSourceProbeEventType:                cloudPubSubSourceProbe,
	CloudStorageSourceCreateProbeEventType:         cloudStorageSourceCreateProbe,
	CloudStorageSourceUpdateMetadataProbeEventType: cloudStorageSourceUpdateMetadataProbe,
	CloudStorageSourceArchiveProbeEventType:        cloudStorageSourceArchiveProbe,
	CloudStorageSourceDeleteProbeEventType:         cloudStorageSourceDeleteProbe,
	CloudAuditLogsSourceProbeEventType:             cloudAuditLogsSourceProbe,
	CloudSchedulerSourceProbeEventType:             cloudSchedulerSourceProbe,
}

var receiveProbeHandlers = map[string]Handler{
	BrokerE2EDeliveryProbeEventType:                      brokerE2EDeliveryProbe,
	schemasv1.CloudPubSubMessagePublishedEventType:       cloudPubSubSourceProbe,
	schemasv1.CloudStorageObjectFinalizedEventType:       cloudStorageSourceCreateProbe,
	schemasv1.CloudStorageObjectMetadataUpdatedEventType: cloudStorageSourceUpdateMetadataProbe,
	schemasv1.CloudStorageObjectArchivedEventType:        cloudStorageSourceArchiveProbe,
	schemasv1.CloudStorageObjectDeletedEventType:         cloudStorageSourceDeleteProbe,
	schemasv1.CloudAuditLogsLogWrittenEventType:          cloudAuditLogsSourceProbe,
	schemasv1.CloudSchedulerJobExecutedEventType:         cloudSchedulerSourceProbe,
}

// WithProbeTimeout returns a context with a timeout specified from the 'timeout'
// extension of a given CloudEvent, defaulting to a certain value if not specified,
// and capped to a maximum.
func WithProbeTimeout(ctx context.Context, event cloudevents.Event, defaultTimeout time.Duration, maxTimeout time.Duration) (context.Context, context.CancelFunc) {
	timeout := defaultTimeout
	if _, ok := event.Extensions()[probeEventTimeoutExtension]; ok {
		customTimeoutExtension := fmt.Sprint(event.Extensions()[probeEventTimeoutExtension])
		if customTimeout, err := time.ParseDuration(customTimeoutExtension); err != nil {
			logging.FromContext(ctx).Warnw("Failed to parse custom timeout extension duration", zap.String("timeout", customTimeoutExtension), zap.Error(err))
		} else {
			timeout = customTimeout
		}
	}
	if timeout.Nanoseconds() > maxTimeout.Nanoseconds() {
		logging.FromContext(ctx).Warnw("Desired timeout exceeds the maximum, clamping to maximum value", zap.Duration("timeout", timeout), zap.Duration("maximumTimeout", maxTimeout))
		timeout = maxTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

// WithProbeEventLoggingContext attaches a logger to the context which contains
// useful information about probe requests.
func WithProbeEventLoggingContext(ctx context.Context, event cloudevents.Event) context.Context {
	logger := logging.FromContext(ctx)
	logger = logger.With(zap.Any("event", map[string]interface{}{
		"id":          event.ID(),
		"source":      event.Source(),
		"specversion": event.SpecVersion(),
		"type":        event.Type(),
		"extensions":  event.Extensions(),
	}))
	return logging.WithLogger(ctx, logger)
}

type cloudEventsFunc func(cloudevents.Event) cloudevents.Result

// forwardFromProbe is the base forward probe request handler which is called
// whenever the probe helper receives a CloudEvent through port PROBE_PORT or
// through the specified probe port listener.
func (ph *Helper) forwardFromProbe(ctx context.Context) cloudEventsFunc {
	return func(event cloudevents.Event) cloudevents.Result {
		// Attach important metadata about the event to the logging context.
		ctx := WithProbeEventLoggingContext(ctx, event)
		logging.FromContext(ctx).Infow("Received probe request")

		// Refresh the forward probe liveness time
		ph.ProbeChecker.LastForwardEventTime.SetNow()

		// Retrieve the probe handler based on the event type
		pr, ok := forwardProbeHandlers[event.Type()]
		if !ok {
			logging.FromContext(ctx).Warnw("Probe forwarding failed, unrecognized forward probe type")
			return cloudevents.ResultNACK
		}

		// Ensure there is a targetservice CloudEvent extension
		if _, ok := event.Extensions()[probeEventTargetServiceExtension]; !ok {
			logging.FromContext(ctx).Warnf("Probe forwarding failed, forward probe event missing '%s' extension", probeEventTargetServiceExtension)
			return cloudevents.ResultNACK
		}

		// Add timeout to the context
		ctx, cancel := WithProbeTimeout(ctx, event, ph.DefaultTimeoutDuration, ph.MaxTimeoutDuration)
		defer cancel()

		// Forward the probe event. This call is likely to be blocking.
		if err := pr.Forward(ctx, ph, event); err != nil {
			logging.FromContext(ctx).Warnw("Probe forwarding failed", zap.Error(err))
			return cloudevents.ResultNACK
		}
		return cloudevents.ResultACK
	}
}

// receiveEvent is the base receiver probe request handler which is called
// whenever the probe helper receives a CloudEvent through port RECEIVER_PORT or
// through the specified receiver port listener.
func (ph *Helper) receiveEvent(ctx context.Context) cloudEventsFunc {
	return func(event cloudevents.Event) cloudevents.Result {
		// Attach important metadata about the event to the logging context.
		ctx := WithProbeEventLoggingContext(ctx, event)
		logging.FromContext(ctx).Infow("Received event")

		// Refresh the receiver probe liveness time
		ph.ProbeChecker.LastReceiverEventTime.SetNow()

		// Retrieve the probe handler based on the event type
		pr, ok := receiveProbeHandlers[event.Type()]
		if !ok {
			logging.FromContext(ctx).Warnw("Probe receiver failed, unrecognized receive probe type")
			return cloudevents.ResultNACK
		}

		// Ensure there is a targetservice CloudEvent extension
		if _, ok := event.Extensions()[probeEventTargetServiceExtension]; !ok {
			logging.FromContext(ctx).Warnf("Probe receiver failed, receiver probe event missing '%s' extension", probeEventTargetServiceExtension)
			return cloudevents.ResultNACK
		}

		// Receive the probe event
		if err := pr.Receive(ctx, ph, event); err != nil {
			logging.FromContext(ctx).Warnw("Probe receiver failed", zap.Error(err))
			return cloudevents.ResultNACK
		}
		return cloudevents.ResultACK
	}
}

// Run initializes the probe helper and starts the probe forwarder and receiver.
func (ph *Helper) Run(ctx context.Context) {
	var err error
	logger := logging.FromContext(ctx)

	// initialize the probe checker
	if ph.ProbeChecker == nil {
		logger.Fatalw("Unspecified probe checker")
	}
	ph.ProbeChecker.LastForwardEventTime.SetNow()
	ph.ProbeChecker.LastReceiverEventTime.SetNow()

	// create pubsub client
	if ph.PubsubClient == nil {
		ph.PubsubClient, err = pubsub.NewClient(ctx, ph.ProjectID)
		if err != nil {
			logger.Fatalw("Failed to create cloud pubsub client", zap.Error(err))
		}
	}
	if ph.CePubsubClient == nil {
		pst, err := cepubsub.New(ctx,
			cepubsub.WithClient(ph.PubsubClient),
			cepubsub.WithProjectID(ph.ProjectID))
		if err != nil {
			logger.Fatalw("Failed to create pubsub transport", zap.Error(err))
		}
		ph.CePubsubClient, err = cloudevents.NewClient(pst)
		if err != nil {
			logger.Fatalw("Failed to create CloudEvents pubsub client", zap.Error(err))
		}
	}

	// create cloud storage client
	if ph.StorageClient == nil {
		ph.StorageClient, err = storage.NewClient(ctx)
		if err != nil {
			logger.Fatalw("Failed to create cloud storage client", zap.Error(err))
		}
	}

	// create sender client
	if ph.CeForwardClient == nil {
		spOpts := []cehttp.Option{}
		if ph.ProbeListener != nil {
			spOpts = append(spOpts, cloudevents.WithListener(ph.ProbeListener))
		} else {
			spOpts = append(spOpts, cloudevents.WithPort(ph.ProbePort))
		}
		sp, err := cloudevents.NewHTTP(spOpts...)
		if err != nil {
			logger.Fatalw("Failed to create sender transport", zap.Error(err))
		}
		ph.CeForwardClient, err = cloudevents.NewClient(sp)
		if err != nil {
			logger.Fatalw("Failed to create sender client", zap.Error(err))
		}
	}

	// create receiver client
	if ph.CeReceiveClient == nil {
		rpOpts := []cehttp.Option{
			cloudevents.WithGetHandlerFunc(ph.ProbeChecker.StalenessHandlerFunc(ctx)),
			cloudevents.WithMiddleware(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					if req.Header.Get(probeEventTargetServiceHeader) == "" {
						req.Header.Set(probeEventTargetServiceHeader, req.Host)
					}
					next.ServeHTTP(rw, req)
				})
			}),
		}
		if ph.ProbeListener != nil {
			rpOpts = append(rpOpts, cloudevents.WithListener(ph.ReceiverListener))
		} else {
			rpOpts = append(rpOpts, cloudevents.WithPort(ph.ReceiverPort))
		}
		rp, err := cloudevents.NewHTTP(rpOpts...)
		if err != nil {
			logger.Fatalw("Failed to create receiver transport", zap.Error(err))
		}
		ph.CeReceiveClient, err = cloudevents.NewClient(rp)
		if err != nil {
			logger.Fatalw("Failed to create receiver client", zap.Error(err))
		}
	}

	// make a map to store the channel for each event
	if ph.ReceivedEvents == nil {
		ph.ReceivedEvents = &utils.ReceivedEventsMap{
			Channels: make(map[string]chan bool),
		}
	}

	// Start a goroutine to receive the probe request event and forward it appropriately
	logger.Infow("Starting event forwarder client...")
	go ph.CeForwardClient.StartReceiver(ctx, ph.forwardFromProbe(ctx))

	// Receive the event and return the result back to the probe
	logger.Infow("Starting event receiver client...")
	ph.CeReceiveClient.StartReceiver(ctx, ph.receiveEvent(ctx))
}

// Helper is the main probe helper object which contains the metadata and clients
// shared between all probe handlers.
type Helper struct {
	// The project ID
	ProjectID string

	// The base URL for the BrokerCell Ingress used in the broker delivery probe
	BrokerCellIngressBaseURL string

	// The client responsible for handling forward probe requests
	CeForwardClient cloudevents.Client

	// The client responsible for receiving events
	CeReceiveClient cloudevents.Client

	// The pubsub client wrapped by a CloudEvents client for the CloudPubSubSource
	// probe and used for the CloudAuditLogsSource probe
	PubsubClient *pubsub.Client

	// The CloudEvents client responsible for forwarding events as messages to a
	// topic for the CloudPubSubSource probe.
	CePubsubClient cloudevents.Client

	// The storage client used in the CloudStorageSource
	StorageClient *storage.Client

	// Handle for the bucket used in the CloudStorageSource probe
	Bucket *storage.BucketHandle

	// The port through which the probe helper receives probe requests
	ProbePort int
	// If a listener is specified instead, the port is ignored
	ProbeListener net.Listener

	// The port through which the probe helper receives source events
	ReceiverPort int
	// If a listener is specified instead, the port is ignored
	ReceiverListener net.Listener

	// The default duration after which the probe helper times out after forwarding an event, if no custom timeout duration is specified
	DefaultTimeoutDuration time.Duration

	// The maximum duration after which the probe helper times out after forwarding an event
	MaxTimeoutDuration time.Duration

	// The map of received events to be tracked by the probe and receiver clients
	ReceivedEvents *utils.ReceivedEventsMap

	// The probe checker invoked in the liveness probe
	ProbeChecker *utils.ProbeChecker
}
