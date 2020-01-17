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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/webhook/resourcesemantics"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	kngcpduck "github.com/google/knative-gcp/pkg/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudAuditLogsSource is a specification for a Cloud Audit Log event source.
type CloudAuditLogsSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudAuditLogsSourceSpec   `json:"spec"`
	Status CloudAuditLogsSourceStatus `json:"status"`
}

var (
	_ kmeta.OwnerRefable           = (*CloudAuditLogsSource)(nil)
	_ resourcesemantics.GenericCRD = (*CloudAuditLogsSource)(nil)
	_ kngcpduck.PubSubable         = (*CloudAuditLogsSource)(nil)
	_                              = duck.VerifyType(&CloudAuditLogsSource{}, &duckv1.Conditions{})
)

const (
	SinkReady         apis.ConditionType = "SinkReady"
	AuditLogEventType                    = "com.google.cloud.auditlog.event"
)

var auditLogsSourceCondSet = apis.NewLivingConditionSet(
	duckv1alpha1.PullSubscriptionReady,
	duckv1alpha1.TopicReady,
	SinkReady,
)

type CloudAuditLogsSourceSpec struct {
	// This brings in the PubSub based Source Specs. Includes:
	duckv1alpha1.PubSubSpec `json:",inline"`

	// The CloudAuditLogsSource will pull events matching the following
	// parameters:

	// The GCP service providing audit logs. Required.
	ServiceName string `json:"serviceName"`
	// The name of the service method or operation. For API calls,
	// this should be the name of the API method. Required.
	MethodName string `json:"methodName"`
	// The resource or collection that is the target of the
	// operation. The name is a scheme-less URI, not including the
	// API service name.
	ResourceName string `json:"resourceName,omitempty"`
}

type CloudAuditLogsSourceStatus struct {
	duckv1alpha1.PubSubStatus `json:",inline"`

	// ID of the Stackdriver sink used to publish audit log messages.
	StackdriverSink string `json:"stackdriverSink,omitempty"`
}

func (*CloudAuditLogsSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("CloudAuditLogsSource")
}

///Methods for pubsubable interface

// PubSubSpec returns the PubSubSpec portion of the Spec.
func (s *CloudAuditLogsSource) PubSubSpec() *duckv1alpha1.PubSubSpec {
	return &s.Spec.PubSubSpec
}

// PubSubStatus returns the PubSubStatus portion of the Status.
func (s *CloudAuditLogsSource) PubSubStatus() *duckv1alpha1.PubSubStatus {
	return &s.Status.PubSubStatus
}

// ConditionSet returns the apis.ConditionSet of the embedding object
func (*CloudAuditLogsSource) ConditionSet() *apis.ConditionSet {
	return &auditLogsSourceCondSet
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CloudAuditLogsSourceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []CloudAuditLogsSource `json:"items"`
}
