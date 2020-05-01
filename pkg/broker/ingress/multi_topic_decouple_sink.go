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

package ingress

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cepubsub "github.com/cloudevents/sdk-go/v2/protocol/pubsub"
	"knative.dev/eventing/pkg/logging"

	"github.com/google/knative-gcp/pkg/broker/config"
	"github.com/google/knative-gcp/pkg/broker/config/volume"
	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"
	"github.com/google/knative-gcp/pkg/utils"
)

const projectEnvKey = "PROJECT_ID"

// NewMultiTopicDecoupleSink creates a new multiTopicDecoupleSink.
func NewMultiTopicDecoupleSink(ctx context.Context, options ...MultiTopicDecoupleSinkOption) (*multiTopicDecoupleSink, error) {
	var err error
	opts := new(multiTopicDecoupleSinkOptions)
	for _, opt := range options {
		opt(opts)
	}

	// Apply defaults
	if opts.client == nil {
		if opts.pubsub == nil {
			var projectID string
			if projectID, err = utils.ProjectID(os.Getenv(projectEnvKey), metadataClient.NewDefaultMetadataClient()); err != nil {
				return nil, err
			}
			if opts.pubsub, err = pubsub.NewClient(ctx, projectID); err != nil {
				return nil, err
			}
		}
		if opts.client, err = newPubSubClient(ctx, opts.pubsub); err != nil {
			return nil, err
		}
	}

	if opts.brokerConfig == nil {
		if opts.brokerConfig, err = volume.NewTargetsFromFile(); err != nil {
			return nil, fmt.Errorf("creating broker config for default multi topic decouple sink")
		}
	}

	sink := &multiTopicDecoupleSink{
		logger:       logging.FromContext(ctx),
		client:       opts.client,
		brokerConfig: opts.brokerConfig,
	}
	return sink, nil
}

// multiTopicDecoupleSink implements DecoupleSink and routes events to pubsub topics corresponding
// to the broker to which the events are sent.
type multiTopicDecoupleSink struct {
	// client talks to pubsub.
	client cev2.Client
	// brokerConfig holds configurations for all brokers. It's a view of a configmap populated by
	// the broker controller.
	brokerConfig config.ReadonlyTargets
	logger       *zap.Logger
}

// Send sends incoming event to its corresponding pubsub topic based on which broker it belongs to.
func (m *multiTopicDecoupleSink) Send(ctx context.Context, ns, broker string, event cev2.Event) protocol.Result {
	topic, err := m.getTopicForBroker(ns, broker)
	if err != nil {
		return err
	}
	ctx = cecontext.WithTopic(ctx, topic)
	return m.client.Send(ctx, event)
}

// getTopicForBroker finds the corresponding decouple topic for the broker from the mounted broker configmap volume.
func (m *multiTopicDecoupleSink) getTopicForBroker(ns, broker string) (string, error) {
	brokerConfig, ok := m.brokerConfig.GetBroker(ns, broker)
	if !ok {
		// There is an propagation delay between the controller reconciles the broker config and
		// the config being pushed to the configmap volume in the ingress pod. So sometimes we return
		// an error even if the request is valid.
		m.logger.Warn("config is not found for", zap.Any("ns", ns), zap.Any("broker", broker))
		return "", fmt.Errorf("%q/%q: %w", ns, broker, ErrNotFound)
	}
	if brokerConfig.DecoupleQueue == nil || brokerConfig.DecoupleQueue.Topic == "" {
		m.logger.Error("DecoupleQueue or topic missing for broker, this should NOT happen.", zap.Any("brokerConfig", brokerConfig))
		return "", fmt.Errorf("decouple queue of %q/%q: %w", ns, broker, ErrIncomplete)
	}
	return brokerConfig.DecoupleQueue.Topic, nil
}

// newPubSubClient creates a pubsub client using the given project ID.
func newPubSubClient(ctx context.Context, client *pubsub.Client) (cev2.Client, error) {
	// Make a pubsub protocol for the CloudEvents client.
	p, err := cepubsub.New(ctx, cepubsub.WithClient(client))
	if err != nil {
		return nil, err
	}

	// Use the pubsub prototol to make a new CloudEvents client.
	return cev2.NewClientObserved(p,
		cev2.WithUUIDs(),
		cev2.WithTimeNow(),
		cev2.WithTracePropagation,
	)
}
