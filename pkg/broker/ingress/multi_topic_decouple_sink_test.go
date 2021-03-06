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
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/uuid"

	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/go-cmp/cmp"
	"github.com/google/knative-gcp/pkg/broker/config"
	"github.com/google/knative-gcp/pkg/broker/config/memory"
	logtest "knative.dev/pkg/logging/testing"
)

func TestMultiTopicDecoupleSink(t *testing.T) {
	// TODO(#1804): remove this mock when enabling the feature by default.
	origEnableEventFilterFunc := enableEventFilterFunc
	defer func() { enableEventFilterFunc = origEnableEventFilterFunc }()
	enableEventFilterFunc = func() bool {
		return true
	}

	// If the broker has no targets, it will drop events at ingress without sending them
	// to pub/sub. So we add a target with no filter to the broker to ensure events are not
	// dropped due to ingress filtering.
	brokerTargets := map[string]*config.Target{"target": {
		CellTenantType: config.CellTenantType_BROKER,
	}}

	type brokerTestCase struct {
		broker  *config.CellTenantKey
		topic   string
		wantErr bool
	}
	tests := []struct {
		name         string
		brokerConfig *config.TargetsConfig
		cases        []brokerTestCase
	}{
		{
			name: "happy path single broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_READY,
						},
						Targets: brokerTargets},
				},
			},
			cases: []brokerTestCase{
				{
					broker: config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:  "test_topic_1",
				},
			},
		},
		{
			name: "happy path multiple brokers",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_READY,
						},
						Targets: brokerTargets},
					"test_ns_2/test_broker_2": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_2",
							State: config.State_READY,
						},
						Targets: brokerTargets},
				},
			},
			cases: []brokerTestCase{
				{
					broker: config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:  "test_topic_1",
				},
				{
					broker: config.TestOnlyBrokerKey("test_ns_2", "test_broker_2"),
					topic:  "test_topic_2",
				},
			},
		},
		{
			name: "broker doesn't exist in config",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "broker is not ready",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_UNKNOWN,
						},
						Targets: brokerTargets},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "decouple queue is nil for broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type:          config.CellTenantType_BROKER,
						DecoupleQueue: nil,
						Targets:       brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "empty topic for broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type:          config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{Topic: ""},
						Targets:       brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := logtest.TestContextWithLogger(t)
			psSrv := pstest.NewServer()
			defer psSrv.Close()
			psClient := createPubsubClient(ctx, t, psSrv)

			brokerConfig := memory.NewTargets(tt.brokerConfig)
			for i, testCase := range tt.cases {
				topic := psClient.Topic(testCase.topic)
				if exists, err := topic.Exists(ctx); err != nil {
					t.Fatal(err)
				} else if !exists {
					if topic, err = psClient.CreateTopic(ctx, testCase.topic); err != nil {
						t.Fatal(err)
					}
				}
				subscription, err := psClient.CreateSubscription(
					ctx, fmt.Sprintf("test-sub-%d", i), pubsub.SubscriptionConfig{Topic: topic})
				if err != nil {
					t.Fatal(err)
				}

				sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, pubsub.DefaultPublishSettings)
				// Send events
				event := createTestEvent(uuid.New().String())
				err = sink.Send(context.Background(), testCase.broker, *event)

				// Verify results.
				if testCase.wantErr && err == nil {
					t.Fatal("Want error but got nil")
				}
				if !testCase.wantErr && err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !testCase.wantErr {
					rctx, cancel := context.WithCancel(ctx)
					msgCh := make(chan *pubsub.Message, 1)
					subscription.Receive(rctx,
						func(ctx context.Context, m *pubsub.Message) {
							select {
							case msgCh <- m:
								cancel()
							case <-ctx.Done():
							}
							m.Ack()
						},
					)
					msg := <-msgCh
					if got, err := binding.ToEvent(ctx, cepubsub.NewMessage(msg)); err != nil {
						t.Error(err)
					} else if diff := cmp.Diff(event, got); diff != "" {
						t.Errorf("Output event doesn't match input, diff: %v", diff)
					}
				}
			}
		})
	}
}

// Temoporary test to ensure functionality doesn't change when the filtering feature is disabled.
// This is an exact copy of 'TestMultiTopicDecoupleSink'.
// TODO(#1804): remove this test when enabling the feature by default.
func TestMultiTopicDecoupleSinkWithoutIngressFiltering(t *testing.T) {
	// If the broker has no targets, it will drop events at ingress without sending them
	// to pub/sub. So we add a target with no filter to the broker to ensure events are not
	// dropped due to ingress filtering.
	brokerTargets := map[string]*config.Target{"target": {
		CellTenantType: config.CellTenantType_BROKER,
	}}

	type brokerTestCase struct {
		broker  *config.CellTenantKey
		topic   string
		wantErr bool
	}
	tests := []struct {
		name         string
		brokerConfig *config.TargetsConfig
		cases        []brokerTestCase
	}{
		{
			name: "happy path single broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_READY,
						},
						Targets: brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker: config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:  "test_topic_1",
				},
			},
		},
		{
			name: "happy path multiple brokers",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_READY,
						},
						Targets: brokerTargets,
					},
					"test_ns_2/test_broker_2": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_2",
							State: config.State_READY,
						},
						Targets: brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker: config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:  "test_topic_1",
				},
				{
					broker: config.TestOnlyBrokerKey("test_ns_2", "test_broker_2"),
					topic:  "test_topic_2",
				},
			},
		},
		{
			name: "broker doesn't exist in config",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "broker is not ready",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "test_topic_1",
							State: config.State_UNKNOWN,
						},
						Targets: brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "decouple queue is nil for broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type:          config.CellTenantType_BROKER,
						DecoupleQueue: nil,
						Targets:       brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
		{
			name: "empty topic for broker",
			brokerConfig: &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type: config.CellTenantType_BROKER,
						DecoupleQueue: &config.Queue{
							Topic: "",
						},
						Targets: brokerTargets,
					},
				},
			},
			cases: []brokerTestCase{
				{
					broker:  config.TestOnlyBrokerKey("test_ns_1", "test_broker_1"),
					topic:   "test_topic_1",
					wantErr: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := logtest.TestContextWithLogger(t)
			psSrv := pstest.NewServer()
			defer psSrv.Close()
			psClient := createPubsubClient(ctx, t, psSrv)

			brokerConfig := memory.NewTargets(tt.brokerConfig)
			for i, testCase := range tt.cases {
				topic := psClient.Topic(testCase.topic)
				if exists, err := topic.Exists(ctx); err != nil {
					t.Fatal(err)
				} else if !exists {
					if topic, err = psClient.CreateTopic(ctx, testCase.topic); err != nil {
						t.Fatal(err)
					}
				}
				subscription, err := psClient.CreateSubscription(
					ctx, fmt.Sprintf("test-sub-%d", i), pubsub.SubscriptionConfig{Topic: topic})
				if err != nil {
					t.Fatal(err)
				}

				sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, pubsub.DefaultPublishSettings)
				// Send events
				event := createTestEvent(uuid.New().String())
				err = sink.Send(context.Background(), testCase.broker, *event)

				// Verify results.
				if testCase.wantErr && err == nil {
					t.Fatal("Want error but got nil")
				}
				if !testCase.wantErr && err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !testCase.wantErr {
					rctx, cancel := context.WithCancel(ctx)
					msgCh := make(chan *pubsub.Message, 1)
					subscription.Receive(rctx,
						func(ctx context.Context, m *pubsub.Message) {
							select {
							case msgCh <- m:
								cancel()
							case <-ctx.Done():
							}
							m.Ack()
						},
					)
					msg := <-msgCh
					if got, err := binding.ToEvent(ctx, cepubsub.NewMessage(msg)); err != nil {
						t.Error(err)
					} else if diff := cmp.Diff(event, got); diff != "" {
						t.Errorf("Output event doesn't match input, diff: %v", diff)
					}
				}
			}
		})
	}
}

func TestHasTrigger(t *testing.T) {
	DecoupleQueue := &config.Queue{Topic: "test_topic", State: config.State_READY}

	tests := []struct {
		name          string
		brokerTargets map[string]*config.Target
		hasTrigger    bool
	}{
		{
			name:       "broker with no target",
			hasTrigger: false,
		},
		{
			name:          "broker with empty target",
			brokerTargets: map[string]*config.Target{},
			hasTrigger:    false,
		},
		{
			name: "broker with target with no filters",
			brokerTargets: map[string]*config.Target{
				"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
				},
			},
			hasTrigger: true,
		},
		{
			name: "broker with target with matching filter",
			brokerTargets: map[string]*config.Target{
				"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type":   eventType,
						"source": "test-source",
					},
				},
			},
			hasTrigger: true,
		},
		{
			name: "broker with target with non-matching filter",
			brokerTargets: map[string]*config.Target{
				"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type":   eventType,
						"source": "some-random-source",
					},
				},
			},
			hasTrigger: false,
		},
		{
			name: "broker with multiple targets with one matching target filter",
			brokerTargets: map[string]*config.Target{
				"non_matching_target_1": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type":   eventType,
						"source": "some-random-source",
					},
				},
				"non_matching_target_2": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"source": "some-random-other-source",
					},
				},
				"matching_target_3": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type": eventType,
					},
				},
			},
			hasTrigger: true,
		},
		{
			name: "broker with multiple targets with no matching target filters",
			brokerTargets: map[string]*config.Target{
				"non_matching_target_1": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type":   eventType,
						"source": "some-random-source",
					},
				},
				"non_matching_target_2": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"source": "some-random-other-source",
					},
				},
				"non_matching_target_3": {
					CellTenantType: config.CellTenantType_BROKER,
					FilterAttributes: map[string]string{
						"type": eventType + "empty",
					},
				},
			},
			hasTrigger: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := logtest.TestContextWithLogger(t)
			psSrv := pstest.NewServer()
			defer psSrv.Close()
			psClient := createPubsubClient(ctx, t, psSrv)

			testBrokerConfig := &config.TargetsConfig{
				CellTenants: map[string]*config.CellTenant{
					"test_ns_1/test_broker_1": {
						Type:          config.CellTenantType_BROKER,
						DecoupleQueue: DecoupleQueue,
						Targets:       test.brokerTargets,
					},
				},
			}

			brokerConfig := memory.NewTargets(testBrokerConfig)
			sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, pubsub.DefaultPublishSettings)

			event := createTestEvent(uuid.New().String())

			hasTrigger := sink.hasTrigger(ctx, event)
			if hasTrigger != test.hasTrigger {
				t.Errorf("Sink says event has trigger %t which should be %t", hasTrigger, test.hasTrigger)
			}
		})
	}
}

func TestMultiTopicDecoupleSinkSendChecksFilter(t *testing.T) {
	// TODO(#1804): remove this mock when enabling the feature by default.
	origEnableEventFilterFunc := enableEventFilterFunc
	defer func() { enableEventFilterFunc = origEnableEventFilterFunc }()
	enableEventFilterFunc = func() bool {
		return true
	}

	filterCalled := false
	origEventFilterFunc := eventFilterFunc
	defer func() { eventFilterFunc = origEventFilterFunc }()
	eventFilterFunc = func(ctx context.Context, attrs map[string]string, event *event.Event) bool {
		filterCalled = true
		return true
	}

	ctx := logtest.TestContextWithLogger(t)
	psSrv := pstest.NewServer()
	defer psSrv.Close()
	psClient := createPubsubClient(ctx, t, psSrv)

	testBrokerConfig := &config.TargetsConfig{
		CellTenants: map[string]*config.CellTenant{
			"test_ns_1/test_broker_1": {
				Type:          config.CellTenantType_BROKER,
				DecoupleQueue: &config.Queue{Topic: "test_topic_1", State: config.State_READY},
				Targets: map[string]*config.Target{"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
				}},
			},
		},
	}

	brokerConfig := memory.NewTargets(testBrokerConfig)

	testTopic := "test_topic_1"
	topic := psClient.Topic(testTopic)
	if exists, err := topic.Exists(ctx); err != nil {
		t.Fatal(err)
	} else if !exists {
		if topic, err = psClient.CreateTopic(ctx, testTopic); err != nil {
			t.Fatal(err)
		}
	}

	sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, pubsub.DefaultPublishSettings)
	// Send event.
	event := createTestEvent(uuid.New().String())

	namespace := config.TestOnlyBrokerKey("test_ns_1", "test_broker_1")
	err := sink.Send(context.Background(), namespace, *event)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results.
	if !filterCalled {
		t.Errorf("Send did not call EventFilterFunc")
	}
}

// Temoporary test to ensure the filtering feature is disabled by default.
// TODO(#1804): remove this test when enabling the feature by default.
func TestMultiTopicDecoupleSinkSendDoesNotChecksFilterWhenFeatureDisabled(t *testing.T) {
	filterCalled := false
	origEventFilterFunc := eventFilterFunc
	defer func() { eventFilterFunc = origEventFilterFunc }()
	eventFilterFunc = func(ctx context.Context, attrs map[string]string, event *event.Event) bool {
		filterCalled = true
		return true
	}

	ctx := logtest.TestContextWithLogger(t)
	psSrv := pstest.NewServer()
	defer psSrv.Close()
	psClient := createPubsubClient(ctx, t, psSrv)

	testBrokerConfig := &config.TargetsConfig{
		CellTenants: map[string]*config.CellTenant{
			"test_ns_1/test_broker_1": {
				Type:          config.CellTenantType_BROKER,
				DecoupleQueue: &config.Queue{Topic: "test_topic_1", State: config.State_READY},
				Targets: map[string]*config.Target{"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
				}},
			},
		},
	}

	brokerConfig := memory.NewTargets(testBrokerConfig)

	testTopic := "test_topic_1"
	topic := psClient.Topic(testTopic)
	if exists, err := topic.Exists(ctx); err != nil {
		t.Fatal(err)
	} else if !exists {
		if topic, err = psClient.CreateTopic(ctx, testTopic); err != nil {
			t.Fatal(err)
		}
	}

	sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, pubsub.DefaultPublishSettings)
	// Send event.
	event := createTestEvent(uuid.New().String())

	namespace := config.TestOnlyBrokerKey("test_ns_1", "test_broker_1")
	err := sink.Send(context.Background(), namespace, *event)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results.
	if filterCalled {
		t.Errorf("Send called EventFilterFunc when feature is disabled")
	}
}

func TestMultiTopicDecoupleSinkUsesPublishSettings(t *testing.T) {
	ctx := logtest.TestContextWithLogger(t)
	psSrv := pstest.NewServer()
	defer psSrv.Close()
	psClient := createPubsubClient(ctx, t, psSrv)

	testBrokerConfig := &config.TargetsConfig{
		CellTenants: map[string]*config.CellTenant{
			"test_ns_1/test_broker_1": {
				Type:          config.CellTenantType_BROKER,
				DecoupleQueue: &config.Queue{Topic: "test_topic_1", State: config.State_READY},
				Targets: map[string]*config.Target{"target_1": {
					CellTenantType: config.CellTenantType_BROKER,
				}},
			},
		},
	}

	brokerConfig := memory.NewTargets(testBrokerConfig)

	testTopic := "test_topic_1"
	topic := psClient.Topic(testTopic)
	if exists, err := topic.Exists(ctx); err != nil {
		t.Fatal(err)
	} else if !exists {
		if topic, err = psClient.CreateTopic(ctx, testTopic); err != nil {
			t.Fatal(err)
		}
	}

	ce := createTestEvent(uuid.New().String())
	ce.SetData(event.ApplicationJSON, `{"hello": "world"}`)
	publishSettings := pubsub.DefaultPublishSettings
	// This is a purposely smaller than the event's data to cause an error.
	publishSettings.BufferedByteLimit = len(ce.Data()) - 1
	sink := NewMultiTopicDecoupleSink(ctx, brokerConfig, psClient, publishSettings)
	// Send event.

	namespace := config.TestOnlyBrokerKey("test_ns_1", "test_broker_1")
	err := sink.Send(context.Background(), namespace, *ce)
	if err == nil {
		t.Fatal("Expected an error due to the BufferedByteLimit being smaller than the event, actually none.")
	}
	if want, got := "bundler reached buffered byte limit", err.Error(); want != got {
		t.Fatalf("Unexpected error, expected %q, actually %q", want, got)
	}
}
