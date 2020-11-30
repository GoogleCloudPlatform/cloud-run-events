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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	grpcstatus "google.golang.org/grpc/status"

	"knative.dev/pkg/logging"
	logtest "knative.dev/pkg/logging/testing"

	. "github.com/google/knative-gcp/pkg/pubsub/adapter/context"
	"github.com/google/knative-gcp/pkg/pubsub/adapter/converters"
	schemasv1 "github.com/google/knative-gcp/pkg/schemas/v1"
	"github.com/google/knative-gcp/test/test_images/probe_helper/utils"
)

const (
	// the fake namespace used in the Broker E2E delivery probe
	testNamespace = "test-namespace"
	// the fake project ID used by the test resources
	testProjectID = "test-project-id"
	// the fake pubsub topic ID used in the test CloudPubSubSource
	testTopicID = "cloudpubsubsource-topic"
	// the fake pubsub subscription ID used in the test CloudPubSubSource
	testSubscriptionID = "cre-src-test-subscription-id"
	// the fake Cloud Storage bucket ID used in the test CloudStorageSource
	testStorageBucket = "cloudstoragesource-bucket"
)

var (
	testStorageUploadRequest      = fmt.Sprintf("/upload/storage/v1/b/%s/o?alt=json&name=1234567890&prettyPrint=false&projection=full&uploadType=multipart", testStorageBucket)
	testStorageRequest            = fmt.Sprintf("/b/%s/o/1234567890?alt=json&prettyPrint=false&projection=full", testStorageBucket)
	testStorageGenerationRequest  = fmt.Sprintf("/b/%s/o/1234567890?alt=json&generation=0&prettyPrint=false", testStorageBucket)
	testStorageCreateBody         = fmt.Sprintf(`{"bucket":"%s","name":"1234567890"}`, testStorageBucket)
	testStorageUpdateMetadataBody = fmt.Sprintf(`{"bucket":"%s","metadata":{"some-key":"Metadata updated!"}}`, testStorageBucket)
	testStorageArchiveBody        = fmt.Sprintf(`{"bucket":"%s","name":"1234567890","storageClass":"ARCHIVE"}`, testStorageBucket)
	testRequestHost               = fmt.Sprintf("probe-helper-external-receiver.%s.svc.cluster.local", testNamespace)
)

// A helper function that starts a test Broker which receives events forwarded by
// the probe helper and delivers the events back to the probe helper receiver.
func runTestBroker(ctx context.Context, group *errgroup.Group, probeReceiverURL string) string {
	brokerListener, err := GetFreePortListener()
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to get free broker port listener: %v", err)
	}
	brokerPort := brokerListener.Addr().(*net.TCPAddr).Port
	bp, err := cloudevents.NewHTTP(
		cloudevents.WithListener(brokerListener),
		cloudevents.WithTarget(probeReceiverURL),
		cloudevents.WithPath(fmt.Sprintf("/%s/default", testNamespace)),
	)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create http protocol of the test Broker: %v", err)
	}
	bc, err := cloudevents.NewClient(bp)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create the test Broker client: %v", err)
	}
	group.Go(func() error {
		bc.StartReceiver(ctx, func(event cloudevents.Event) {
			event.SetExtension("requesthost", testRequestHost)
			if res := bc.Send(ctx, event); !cloudevents.IsACK(res) {
				logging.FromContext(ctx).Warnf("Failed to send CloudEvent from the test Broker: %v", res)
			}
		})
		return nil
	})
	return fmt.Sprintf("http://localhost:%d", brokerPort)
}

// A helper function that starts a test CloudPubSubSource which watches a pubsub
// Subscription for messages and delivers them as CloudEvents to the probe
// helper receiver.
func runTestCloudPubSubSource(ctx context.Context, group *errgroup.Group, sub *pubsub.Subscription, probeReceiverURL string) {
	converter := converters.NewPubSubConverter()
	cp, err := cloudevents.NewHTTP(cloudevents.WithTarget(probeReceiverURL))
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create http protocol of the test CloudPubSubSource, %v", err)
	}
	c, err := cloudevents.NewClient(cp)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create the test CloudPubSubSource client, %v", err)
	}
	msgHandler := func(ctx context.Context, msg *pubsub.Message) {
		event, err := converter.Convert(ctx, msg, converters.CloudPubSub)
		event.SetExtension("requesthost", testRequestHost)
		if err != nil {
			logging.FromContext(ctx).Warnf("Could not convert message to CloudEvent: %v", err)
		}
		if res := c.Send(ctx, *event); !cloudevents.IsACK(res) {
			logging.FromContext(ctx).Warnf("Failed to send CloudEvent from the test CloudPubSubSource: %v", err)
		}
	}
	group.Go(func() error {
		if err := sub.Receive(ctx, msgHandler); err != nil {
			if _, ok := grpcstatus.FromError(err); !ok {
				logging.FromContext(ctx).Warnf("Could not receive from subscription: %v", err)
			}
		}
		return nil
	})
}

// A helper function that starts a test CloudAuditLogsSource which watches
// periodically for a change of state in the existence of pubsub topics and
// forwards the appropriate events to the probe helper receiver.
func runTestCloudAuditLogsSource(ctx context.Context, group *errgroup.Group, pubsubClient *pubsub.Client, probeReceiverURL string) {
	cp, err := cloudevents.NewHTTP(cloudevents.WithTarget(probeReceiverURL))
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create http protocol of the test CloudAuditLogsSource, %v", err)
	}
	c, err := cloudevents.NewClient(cp)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create the test CloudAuditLogsSource client, %v", err)
	}
	topicCreated := false
	ticker := time.NewTicker(100 * time.Millisecond)
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				exists, err := pubsubClient.Topic("cloudauditlogssource-probe-1234567890").Exists(ctx)
				if err != nil {
					logging.FromContext(ctx).Warnf("Failed to determine existence of test pubsub topic: %v", err)
				}
				if exists && !topicCreated {
					createTopicEvent := cloudevents.NewEvent()
					createTopicEvent.SetID("1234567890")
					createTopicEvent.SetSubject(schemasv1.CloudAuditLogsEventSubject("pubsub.googleapis.com", "projects/test-project-id/topics/cloudauditlogssource-probe-1234567890"))
					createTopicEvent.SetType(schemasv1.CloudAuditLogsLogWrittenEventType)
					createTopicEvent.SetSource(schemasv1.CloudAuditLogsEventSource("projects/test-project-id", "activity"))
					createTopicEvent.SetExtension("methodname", "google.pubsub.v1.Publisher.CreateTopic")
					createTopicEvent.SetExtension("requesthost", testRequestHost)
					if res := c.Send(ctx, createTopicEvent); !cloudevents.IsACK(res) {
						logging.FromContext(ctx).Warnf("Failed to send topic created CloudEvent from the test CloudAuditLogsSource: %v", res)
					}
					topicCreated = true
				}
			}
		}
	})
}

// A helper function that starts a test CloudStorageSource which intercepts
// Cloud Storage HTTP requests and forwards the appropriate notifications as
// CloudEvents to the probe helper receiver.
func runTestCloudStorageSource(ctx context.Context, group *errgroup.Group, gotRequest chan *http.Request, probeReceiverURL string) {
	cp, err := cloudevents.NewHTTP(cloudevents.WithTarget(probeReceiverURL))
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create http protocol of the test CloudStorageSource, %v", err)
	}
	c, err := cloudevents.NewClient(cp)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create the test CloudStorageSource client, %v", err)
	}
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case req := <-gotRequest:
				bodyBytes, err := ioutil.ReadAll(req.Body)
				if err != nil {
					logging.FromContext(ctx).Warnf("Failed to read request body in test CloudStorageSource, %v", err)
				}
				body := string(bodyBytes)
				method := req.Method
				url := req.URL.String()
				logging.FromContext(ctx).Info(method, url, body)
				if method == "POST" && url == testStorageUploadRequest && strings.Contains(body, testStorageCreateBody) {
					// This request indicates the client's intent to create a new object.
					finalizeEvent := cloudevents.NewEvent()
					finalizeEvent.SetID("1234567890")
					finalizeEvent.SetSubject(schemasv1.CloudStorageEventSubject("1234567890"))
					finalizeEvent.SetType(schemasv1.CloudStorageObjectFinalizedEventType)
					finalizeEvent.SetSource(schemasv1.CloudStorageEventSource(testStorageBucket))
					finalizeEvent.SetExtension("requesthost", testRequestHost)
					if res := c.Send(ctx, finalizeEvent); !cloudevents.IsACK(res) {
						logging.FromContext(ctx).Warnf("Failed to send object finalized CloudEvent from the test CloudStorageSource: %v", res)
					}
				} else if method == "PATCH" && url == testStorageRequest && strings.Contains(body, testStorageUpdateMetadataBody) {
					// This request indicates the client's intent to update the object's metadata.
					updateMetadataEvent := cloudevents.NewEvent()
					updateMetadataEvent.SetID("1234567890")
					updateMetadataEvent.SetSubject(schemasv1.CloudStorageEventSubject("1234567890"))
					updateMetadataEvent.SetType(schemasv1.CloudStorageObjectMetadataUpdatedEventType)
					updateMetadataEvent.SetSource(schemasv1.CloudStorageEventSource(testStorageBucket))
					updateMetadataEvent.SetExtension("requesthost", testRequestHost)
					if res := c.Send(ctx, updateMetadataEvent); !cloudevents.IsACK(res) {
						logging.FromContext(ctx).Warnf("Failed to send object metadata updated CloudEvent from the test CloudStorageSource: %v", res)
					}
				} else if method == "POST" && url == testStorageUploadRequest && strings.Contains(body, testStorageArchiveBody) {
					// This request indicates the client's intent to archive the object.
					archivedEvent := cloudevents.NewEvent()
					archivedEvent.SetID("1234567890")
					archivedEvent.SetSubject(schemasv1.CloudStorageEventSubject("1234567890"))
					archivedEvent.SetType(schemasv1.CloudStorageObjectArchivedEventType)
					archivedEvent.SetSource(schemasv1.CloudStorageEventSource(testStorageBucket))
					archivedEvent.SetExtension("requesthost", testRequestHost)
					if res := c.Send(ctx, archivedEvent); !cloudevents.IsACK(res) {
						logging.FromContext(ctx).Warnf("Failed to send object archived CloudEvent from the test CloudStorageSource: %v", res)
					}
				} else if method == "DELETE" && url == testStorageGenerationRequest {
					// This request indicates the client's intent to delete the object.
					deletedEvent := cloudevents.NewEvent()
					deletedEvent.SetID("1234567890")
					deletedEvent.SetSubject(schemasv1.CloudStorageEventSubject("1234567890"))
					deletedEvent.SetType(schemasv1.CloudStorageObjectDeletedEventType)
					deletedEvent.SetSource(schemasv1.CloudStorageEventSource(testStorageBucket))
					deletedEvent.SetExtension("requesthost", testRequestHost)
					if res := c.Send(ctx, deletedEvent); !cloudevents.IsACK(res) {
						logging.FromContext(ctx).Warnf("Failed to send object deleted CloudEvent from the test CloudStorageSource: %v", res)
					}
				}
			}
		}
	})
}

// A helper function that starts a test CloudSchedulerSource which ticks
// periodically and sends the appropriate event notifications to the probe
// helper receiver.
func runTestCloudSchedulerSource(ctx context.Context, group *errgroup.Group, period time.Duration, probeReceiverURL string) {
	cp, err := cloudevents.NewHTTP(cloudevents.WithTarget(probeReceiverURL))
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create http protocol of the test CloudSchedulerSource, %v", err)
	}
	c, err := cloudevents.NewClient(cp)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Failed to create the test CloudSchedulerSource client, %v", err)
	}
	ticker := time.NewTicker(period)
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				executedEvent := cloudevents.NewEvent()
				executedEvent.SetID("1234567890")
				executedEvent.SetType(schemasv1.CloudSchedulerJobExecutedEventType)
				executedEvent.SetSource(schemasv1.CloudSchedulerEventSource("test-cloud-scheduler-source"))
				executedEvent.SetExtension("requesthost", testRequestHost)
				if res := c.Send(ctx, executedEvent); !cloudevents.IsACK(res) {
					logging.FromContext(ctx).Warnf("Failed to send job executed CloudEvent from the test CloudSchedulerSource: %v", res)
				}
			}
		}
	})
}

type probeEventOption func(*cloudevents.Event)

func withProbeExtension(key, value string) probeEventOption {
	return func(event *cloudevents.Event) {
		event.SetExtension(key, value)
	}
}

func withProbeTimeout(timeout time.Duration) probeEventOption {
	return withProbeExtension("timeout", timeout.String())
}

// Creates a new CloudEvent in the shape of probe events sent to the probe helper.
func probeEvent(name string, opts ...probeEventOption) *cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetID(name + "-1234567890")
	event.SetSource("probe")
	event.SetType(name)
	event.SetTime(time.Now())
	event.SetExtension("requesthost", testRequestHost)
	for _, opt := range opts {
		opt(&event)
	}
	return &event
}

func testPubsubClient(ctx context.Context, t *testing.T, projectID string) (*pubsub.Client, func()) {
	srv := pstest.NewServer()
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial test pubsub connection: %v", err)
	}
	close := func() {
		srv.Close()
		conn.Close()
	}
	c, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create test pubsub client: %v", err)
	}
	return c, close
}

func testStorageClient(ctx context.Context, t *testing.T) (*storage.Client, chan *http.Request, func()) {
	gotRequest := make(chan *http.Request, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The test Cloud Storage server forwards the client's generated HTTP requests.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logging.FromContext(ctx).Fatal("Test Cloud Storage server could not read request body.")
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		gotRequest <- r
		w.Write([]byte("{}"))
	}))
	c, err := storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatalf("Failed to create test storage client: %v", err)
	}
	return c, gotRequest, srv.Close
}

type eventAndResult struct {
	event      *cloudevents.Event
	wantResult protocol.Result
}

func TestProbeHelper(t *testing.T) {
	ctx := logtest.TestContextWithLogger(t)
	ctx = WithProjectKey(ctx, testProjectID)
	ctx = WithTopicKey(ctx, testTopicID)
	ctx = WithSubscriptionKey(ctx, testSubscriptionID)
	ctx = cloudevents.ContextWithRetriesConstantBackoff(ctx, 100*time.Millisecond, 30)
	group, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	phr := makeProbeHelper(ctx, t, group)
	go phr.probeHelper.Run(ctx)

	// Create a testing client from which to send probe events to the probe helper.
	p, err := cloudevents.NewHTTP(cloudevents.WithTarget(phr.probeURL))
	if err != nil {
		t.Fatalf("Failed to create HTTP protocol of the testing client: %s", err.Error())
	}
	c, err := cloudevents.NewClient(p)
	if err != nil {
		t.Fatalf("Failed to create testing client: %s", err.Error())
	}

	cases := []struct {
		name  string
		steps []eventAndResult
	}{{
		name: "Broker E2E delivery probe",
		steps: []eventAndResult{
			{
				event:      probeEvent("broker-e2e-delivery-probe", withProbeExtension("namespace", testNamespace)),
				wantResult: cloudevents.ResultACK,
			},
		},
	}, {
		name: "Broker E2E delivery probe missing namespace",
		steps: []eventAndResult{
			{
				event:      probeEvent("broker-e2e-delivery-probe"),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}, {
		name: "CloudPubSubSource probe",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudpubsubsource-probe", withProbeExtension("topic", "cloudpubsubsource-topic")),
				wantResult: cloudevents.ResultACK,
			},
		},
	}, {
		name: "CloudPubSubSource probe missing topic",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudpubsubsource-probe"),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}, {
		name: "CloudStorageSource probe",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudstoragesource-probe-create", withProbeExtension("bucket", testStorageBucket)),
				wantResult: cloudevents.ResultACK,
			},
			{
				event:      probeEvent("cloudstoragesource-probe-update-metadata", withProbeExtension("bucket", testStorageBucket)),
				wantResult: cloudevents.ResultACK,
			},
			{
				event:      probeEvent("cloudstoragesource-probe-archive", withProbeExtension("bucket", testStorageBucket)),
				wantResult: cloudevents.ResultACK,
			},
			{
				event:      probeEvent("cloudstoragesource-probe-delete", withProbeExtension("bucket", testStorageBucket)),
				wantResult: cloudevents.ResultACK,
			},
		},
	}, {
		name: "CloudStorageSource probe missing bucket",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudstoragesource-probe-create"),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}, {
		name: "CloudAuditLogsSource probe",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudauditlogssource-probe"),
				wantResult: cloudevents.ResultACK,
			},
		},
	}, {
		name: "CloudSchedulerSource probe",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudschedulersource-probe", withProbeExtension("timeout", "200ms")),
				wantResult: cloudevents.ResultACK,
			},
		},
	}, {
		name: "CloudSchedulerSource delay exceeds timeout",
		steps: []eventAndResult{
			{
				event:      probeEvent("cloudschedulersource-probe", withProbeExtension("timeout", "10ms")),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}, {
		name: "Unrecognized probe event type",
		steps: []eventAndResult{
			{
				event:      probeEvent("unrecognized-probe-type"),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}, {
		name: "Custom timeout",
		steps: []eventAndResult{
			{
				event:      probeEvent("broker-e2e-delivery-probe", withProbeExtension("namespace", "test-namespace"), withProbeTimeout(0)),
				wantResult: cloudevents.ResultNACK,
			},
		},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, step := range tc.steps {
				if result := c.Send(ctx, *step.event); !errors.Is(result, step.wantResult) {
					t.Fatalf("wanted result %+v, got %+v", step.wantResult, result)
				}
			}
		})
	}
	// Cancel gracefully to avoid logger panic if parent goroutine terminates.
	phr.cleanup()
	cancel()
	if err := group.Wait(); err != nil {
		t.Fatalf("Error in probe helper fake sources: %v", err)
	}
}

type makeProbeHelperReturn struct {
	probeHelper   *Helper
	probeURL      string
	probeCheckURL string
	cleanup       func()
}

func makeProbeHelper(ctx context.Context, t *testing.T, group *errgroup.Group) makeProbeHelperReturn {
	// Set up ports for testing the probe helper.
	receiverListener, err := GetFreePortListener()
	if err != nil {
		t.Fatalf("Failed to get free receiver port listener: %v", err)
	}
	receiverPort := receiverListener.Addr().(*net.TCPAddr).Port
	receiverURL := fmt.Sprintf("http://localhost:%d", receiverPort)
	probeListener, err := GetFreePortListener()
	if err != nil {
		t.Fatalf("Failed to get free probe port listener: %v", err)
	}
	probePort := probeListener.Addr().(*net.TCPAddr).Port
	probeURL := fmt.Sprintf("http://localhost:%d", probePort)
	probeCheckURL := fmt.Sprintf("http://localhost:%d/healthz", receiverPort)

	// Set up the resources for testing the CloudPubSubSource.
	pubsubClient, closePubsub := testPubsubClient(ctx, t, testProjectID)
	topic, err := pubsubClient.CreateTopic(ctx, testTopicID)
	if err != nil {
		t.Fatalf("Failed to create test topic: %v", err)
	}
	sub, err := pubsubClient.CreateSubscription(ctx, testSubscriptionID, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		t.Fatalf("Failed to create test subscription: %v", err)
	}
	// Run the test CloudPubSubSource.
	runTestCloudPubSubSource(ctx, group, sub, receiverURL)

	// Set up resources for testing the CloudStorageSource.
	storageClient, gotCloudStorageRequest, closeStorage := testStorageClient(ctx, t)
	// Run the test CloudStorageSource.
	runTestCloudStorageSource(ctx, group, gotCloudStorageRequest, receiverURL)

	// Run the test CloudSchedulerSource.
	runTestCloudSchedulerSource(ctx, group, 100*time.Millisecond, receiverURL)

	// Run the test CloudAuditLogsSource.
	runTestCloudAuditLogsSource(ctx, group, pubsubClient, receiverURL)

	// Run the test Broker for testing Broker E2E delivery.
	brokerCellIngressBaseURL := runTestBroker(ctx, group, receiverURL)
	// Create the probe helper and start a goroutine to run it.
	ph := &Helper{
		ProjectID:                testProjectID,
		BrokerCellIngressBaseURL: brokerCellIngressBaseURL,
		PubsubClient:             pubsubClient,
		StorageClient:            storageClient,
		ProbeListener:            probeListener,
		ReceiverListener:         receiverListener,
		DefaultTimeoutDuration:   2 * time.Minute,
		MaxTimeoutDuration:       30 * time.Minute,
		ProbeChecker: &utils.ProbeChecker{
			MaxStaleDuration: time.Second,
		},
	}
	return makeProbeHelperReturn{
		probeHelper:   ph,
		probeURL:      probeURL,
		probeCheckURL: probeCheckURL,
		cleanup: func() {
			closeStorage()
			closePubsub()
		},
	}
}

func assertProbeCheckResult(t *testing.T, url string, ok bool) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal("Failed to create probe check request:", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Log("Failed to execute probe check:", err)
		if ok {
			t.Errorf("probe check result ok got=%v, want=%v", !ok, ok)
		}
		return
	}
	if ok != (resp.StatusCode == http.StatusOK) {
		t.Log("Got probe check status code:", resp.StatusCode)
		t.Errorf("probe check result ok got=%v, want=%v", !ok, ok)
	}
}

// GetFreePortListener opens a listener on a free port.
func GetFreePortListener() (net.Listener, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", addr)
}

func TestProbeHelperHealth(t *testing.T) {
	ctx := logtest.TestContextWithLogger(t)
	group, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	phr := makeProbeHelper(ctx, t, group)
	go phr.probeHelper.Run(ctx)

	// Make sure the probe checker is up.
	time.Sleep(500 * time.Millisecond)
	assertProbeCheckResult(t, phr.probeCheckURL, true)

	// Guarantee that it has been long enough that the stale duration has been reached. This will cause
	// the probe checker's result to be unhealthy.
	time.Sleep(2 * phr.probeHelper.ProbeChecker.MaxStaleDuration)
	assertProbeCheckResult(t, phr.probeCheckURL, false)

	// Cancel gracefully to avoid logger panic if parent goroutine terminates.
	phr.cleanup()
	cancel()
	if err := group.Wait(); err != nil {
		t.Fatalf("Error in probe helper fake sources: %v", err)
	}
}
