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

package v1beta1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	brokerConditionReady = apis.Condition{
		Type:   eventingv1beta1.BrokerConditionReady,
		Status: corev1.ConditionTrue,
	}

	brokerConditionAddressable = apis.Condition{
		Type:   eventingv1beta1.BrokerConditionAddressable,
		Status: corev1.ConditionTrue,
	}

	brokerConditionBrokerCell = apis.Condition{
		Type:   BrokerConditionBrokerCell,
		Status: corev1.ConditionTrue,
	}

	brokerConditionBrokerCellFalse = apis.Condition{
		Type:   BrokerConditionBrokerCell,
		Status: corev1.ConditionFalse,
	}

	brokerConditionTopic = apis.Condition{
		Type:   BrokerConditionTopic,
		Status: corev1.ConditionTrue,
	}

	brokerConditionSubscription = apis.Condition{
		Type:   BrokerConditionSubscription,
		Status: corev1.ConditionTrue,
	}

	brokerConditionConfig = apis.Condition{
		Type:   BrokerConditionConfig,
		Status: corev1.ConditionTrue,
	}
)

func TestBrokerGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		ts        *BrokerStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name: "single condition",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						brokerConditionReady,
					},
				},
			},
		},
		condQuery: apis.ConditionReady,
		want:      &brokerConditionReady,
	}, {
		name: "multiple conditions",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						brokerConditionAddressable,
						brokerConditionBrokerCell,
						brokerConditionTopic,
					},
				},
			},
		},
		condQuery: BrokerConditionBrokerCell,
		want:      &brokerConditionBrokerCell,
	}, {
		name: "multiple conditions, condition false",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						brokerConditionAddressable,
						brokerConditionBrokerCellFalse,
						brokerConditionTopic,
					},
				},
			},
		},
		condQuery: BrokerConditionBrokerCell,
		want:      &brokerConditionBrokerCellFalse,
	}, {
		name: "unknown condition",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						brokerConditionAddressable,
					},
				},
			},
		},
		condQuery: apis.ConditionType("foo"),
		want:      nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ts.GetCondition(test.condQuery)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}

func TestBrokerInitializeConditions(t *testing.T) {
	tests := []struct {
		name string
		ts   *BrokerStatus
		want *BrokerStatus
	}{{
		name: "empty",
		ts:   &BrokerStatus{},
		want: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   eventingv1beta1.BrokerConditionAddressable,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionBrokerCell,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionConfig,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   eventingv1beta1.BrokerConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionSubscription,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionTopic,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}, {
		name: "one false",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   eventingv1beta1.BrokerConditionAddressable,
						Status: corev1.ConditionFalse,
					}},
				},
			},
		},
		want: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   eventingv1beta1.BrokerConditionAddressable,
						Status: corev1.ConditionFalse,
					}, {
						Type:   BrokerConditionBrokerCell,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionConfig,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   eventingv1beta1.BrokerConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionSubscription,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionTopic,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}, {
		name: "one true",
		ts: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   eventingv1beta1.BrokerConditionAddressable,
						Status: corev1.ConditionTrue,
					}},
				},
			},
		},
		want: &BrokerStatus{
			BrokerStatus: eventingv1beta1.BrokerStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   eventingv1beta1.BrokerConditionAddressable,
						Status: corev1.ConditionTrue,
					}, {
						Type:   BrokerConditionBrokerCell,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionConfig,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   eventingv1beta1.BrokerConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionSubscription,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   BrokerConditionTopic,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}}

	ignoreAllButTypeAndStatus := cmpopts.IgnoreFields(
		apis.Condition{},
		"LastTransitionTime", "Message", "Reason", "Severity")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.ts.InitializeConditions()
			if diff := cmp.Diff(test.want, test.ts, ignoreAllButTypeAndStatus); diff != "" {
				t.Errorf("unexpected conditions (-want, +got) = %v", diff)
			}
		})
	}
}

func TestBrokerConditionStatus(t *testing.T) {
	tests := []struct {
		name                string
		addressStatus       string
		brokerCellStatus    string
		subscriptionStatus  string
		topicStatus         string
		configStatus        string
		wantConditionStatus corev1.ConditionStatus
	}{{
		name:                "all happy",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionTrue,
	}, {
		name:                "subscription sad",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "false",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "subscription unknown",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "unknown",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "topic sad",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "false",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "topic unknown",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "unknown",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "address missing",
		addressStatus:       "false",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "ingress false",
		addressStatus:       "true",
		brokerCellStatus:    "false",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "config false",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "false",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "config unknown",
		addressStatus:       "true",
		brokerCellStatus:    "true",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "unknown",
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "brokerCell false",
		addressStatus:       "true",
		brokerCellStatus:    "false",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "brokerCell unknown",
		addressStatus:       "true",
		brokerCellStatus:    "unknown",
		subscriptionStatus:  "true",
		topicStatus:         "true",
		configStatus:        "true",
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "all sad",
		addressStatus:       "false",
		brokerCellStatus:    "false",
		subscriptionStatus:  "false",
		topicStatus:         "false",
		configStatus:        "false",
		wantConditionStatus: corev1.ConditionFalse,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &BrokerStatus{}
			if test.addressStatus == "true" {
				bs.SetAddress(apis.HTTP("example.com"))
			} else {
				bs.SetAddress(nil)
			}
			if test.brokerCellStatus == "true" {
				bs.MarkBrokerCellReady()
			} else if test.brokerCellStatus == "false" {
				bs.MarkBrokerCelllFailed("Unable to create brokercell", "induced failure")
			} else {
				bs.MarkBrokerCelllUnknown("Unable to create brokercell", "induced unknown")
			}
			if test.subscriptionStatus == "true" {
				bs.MarkSubscriptionReady()
			} else if test.subscriptionStatus == "false" {
				bs.MarkSubscriptionFailed("Unable to create PubSub subscription", "induced failure")
			} else {
				bs.MarkSubscriptionUnknown("Unable to create PubSub subscription", "induced unknown")
			}
			if test.topicStatus == "true" {
				bs.MarkTopicReady()
			} else if test.topicStatus == "false" {
				bs.MarkTopicFailed("Unable to create PubSub topic", "induced failure")
			} else {
				bs.MarkTopicUnknown("Unable to create PubSub topic", "induced unknown")
			}
			if test.configStatus == "true" {
				bs.MarkConfigReady()
			} else if test.configStatus == "false" {
				bs.MarkConfigFailed("Unable to reconstruct/update config", "induced failure")
			} else {
				bs.MarkConfigUnknown("Unable to reconstruct/update config", "induced unknown")
			}
			got := bs.GetTopLevelCondition().Status
			if test.wantConditionStatus != got {
				t.Errorf("unexpected readiness: want %v, got %v", test.wantConditionStatus, got)
			}
			happy := bs.IsReady()
			switch test.wantConditionStatus {
			case corev1.ConditionTrue:
				if !happy {
					t.Error("expected happy true, got false")
				}
			case corev1.ConditionFalse, corev1.ConditionUnknown:
				if happy {
					t.Error("expected happy false, got true")
				}
			}
		})
	}
}
