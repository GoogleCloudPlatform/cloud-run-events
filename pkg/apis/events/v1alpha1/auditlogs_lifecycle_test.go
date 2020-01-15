/*
Copyright 2019 Google LLC.

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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

func TestAuditLogsSourceStatusIsReady(t *testing.T) {
	tests := []struct {
		name                string
		s                   *AuditLogsSourceStatus
		wantConditionStatus corev1.ConditionStatus
		want                bool
	}{{
		name: "initialized",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			return s
		}(),
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name: "the status of topic is false",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			s.MarkPullSubscriptionReady()
			s.MarkSinkReady()
			s.MarkTopicFailed("test", "the status of topic is false")
			return s
		}(),
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name: "the status of topic is unknown",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			s.MarkPullSubscriptionReady()
			s.MarkSinkReady()
			s.MarkTopicUnknown("test", "the status of topic is unknown")
			return s
		}(),
		wantConditionStatus: corev1.ConditionUnknown,
	},
		{
			name: "the status of pullsubscription is false",
			s: func() *AuditLogsSourceStatus {
				s := &AuditLogsSourceStatus{}
				s.InitializeConditions()
				s.MarkTopicReady()
				s.MarkSinkReady()
				s.MarkPullSubscriptionFailed("test", "the status of pullsubscription is false")
				return s
			}(),
			wantConditionStatus: corev1.ConditionFalse,
		}, {
			name: "the status of pullsubscription is unknown",
			s: func() *AuditLogsSourceStatus {
				s := &AuditLogsSourceStatus{}
				s.InitializeConditions()
				s.MarkTopicReady()
				s.MarkSinkReady()
				s.MarkPullSubscriptionUnknown("test", "the status of pullsubscription is unknown")
				return s
			}(),
			wantConditionStatus: corev1.ConditionUnknown,
		},
		{
			name: "sink is not ready",
			s: func() *AuditLogsSourceStatus {
				s := &AuditLogsSourceStatus{}
				s.InitializeConditions()
				s.MarkTopicReady()
				s.MarkPullSubscriptionReady()
				s.MarkSinkNotReady("test", "sink is not ready")
				return s
			}(),
			wantConditionStatus: corev1.ConditionFalse,
		}, {
			name: "ready",
			s: func() *AuditLogsSourceStatus {
				s := &AuditLogsSourceStatus{}
				s.InitializeConditions()
				s.MarkTopicReady()
				s.MarkPullSubscriptionReady()
				s.MarkSinkReady()
				return s
			}(),
			wantConditionStatus: corev1.ConditionTrue,
			want:true,
		}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotConditionStatus := test.s.GetTopLevelCondition().Status
			got := test.s.IsReady()
			if gotConditionStatus != test.wantConditionStatus {
				t.Errorf("unexpected condition status: want %v, got %v", test.wantConditionStatus, gotConditionStatus)
			}
			if got != test.want {
				t.Errorf("unexpected readiniess: want %v, got %v", test.want, got)
			}
		})
	}
}
func TestAuditLogsSourceGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		s         *AuditLogsSourceStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name:      "uninitialized",
		s:         &AuditLogsSourceStatus{},
		condQuery: SinkReady,
		want:      nil,
	}, {
		name: "initialized",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			return s
		}(),
		condQuery: SinkReady,
		want: &apis.Condition{
			Type:   SinkReady,
			Status: corev1.ConditionUnknown,
		},
	}, {
		name: "not ready",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			s.MarkSinkNotReady("NotReady", "test message")
			return s
		}(),
		condQuery: SinkReady,
		want: &apis.Condition{
			Type:    SinkReady,
			Status:  corev1.ConditionFalse,
			Reason:  "NotReady",
			Message: "test message",
		},
	}, {
		name: "ready",
		s: func() *AuditLogsSourceStatus {
			s := &AuditLogsSourceStatus{}
			s.InitializeConditions()
			s.MarkSinkReady()
			return s
		}(),
		condQuery: SinkReady,
		want: &apis.Condition{
			Type:   SinkReady,
			Status: corev1.ConditionTrue,
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.s.GetCondition(test.condQuery)
			ignoreTime := cmpopts.IgnoreFields(apis.Condition{},
				"LastTransitionTime", "Severity")
			if diff := cmp.Diff(test.want, got, ignoreTime); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}
