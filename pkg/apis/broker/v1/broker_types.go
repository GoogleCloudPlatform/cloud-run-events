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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

const (
	// BrokerClass is the annotation value to use when creating a
	// Google Cloud Broker object.
	BrokerClass = "googlecloud"
)

// +genclient
// +genreconciler:class=eventing.knative.dev/broker.class
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Broker collects a pool of events that are consumable using Triggers. Brokers
// provide a well-known endpoint for event delivery that senders can use with
// minimal knowledge of the event routing strategy. Receivers use Triggers to
// request delivery of events from a Broker's pool to a specific URL or
// Addressable endpoint.
type Broker struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Broker.
	Spec eventingv1.BrokerSpec `json:"spec,omitempty"`

	// Status represents the current state of the Broker. This data may be out of
	// date.
	// +optional
	Status BrokerStatus `json:"status,omitempty"`
}

var (
	// Check that Broker can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*Broker)(nil)
	_ apis.Defaultable = (*Broker)(nil)

	// Check that Broker can return its spec untyped.
	_ apis.HasSpec = (*Broker)(nil)

	_ runtime.Object = (*Broker)(nil)

	// Check that we can create OwnerReferences to a Broker.
	_ kmeta.OwnerRefable = (*Broker)(nil)

	// Check that we implement KRShaped.
	_ duckv1.KRShaped = (*Broker)(nil)
)

// BrokerStatus represents the current state of a Broker.
type BrokerStatus struct {
	// Inherits core eventing BrokerStatus.
	// even with this change, the webhook seems to drop unknown fields. May need to alter the mutating webhook.
	eventingv1.BrokerStatus `json:",inline"`

	//TODO these fields don't work yet.
	//TODO this requires updating the eventing webhook to allow unknown fields. Since the only unknown
	// fields required are in status, maybe we can use a separate webhook just for broker and trigger
	// status with allow unknown fields set.

	// ProjectID is the resolved project ID in use by the Broker's pubsub resources.
	// +optional
	//ProjectID string `json:"projectId,omitempty"`

	// TopicID is the created topic ID used by the Broker.
	// +optional
	//TopicID string `json:"topicId,omitempty"`

	// SubscriptionID is the created subscription ID used by the Broker.
	// +optional
	//SubscriptionID string `json:"subscriptionId,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BrokerList is a collection of Brokers.
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Broker `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (b *Broker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Broker")
}

// GetUntypedSpec returns the spec of the Broker.
func (b *Broker) GetUntypedSpec() interface{} {
	return b.Spec
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*Broker) GetConditionSet() apis.ConditionSet {
	return brokerCondSet
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (b *Broker) GetStatus() *duckv1.Status {
	return &b.Status.Status
}
