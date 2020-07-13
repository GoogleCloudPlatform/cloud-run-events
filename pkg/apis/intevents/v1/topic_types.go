/*
 * Copyright 2020 The Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/webhook/resourcesemantics"

	"github.com/google/knative-gcp/pkg/apis/duck/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Topic is a resource representing a Topic backed by Google Cloud Pub/Sub.
type Topic struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Topic.
	Spec TopicSpec `json:"spec,omitempty"`

	// Status represents the current state of the Topic. This data may be out of
	// date.
	// +optional
	Status TopicStatus `json:"status,omitempty"`
}

// Check that PullSubscription can be converted to other versions.
var _ apis.Convertible = (*Topic)(nil)

// Check that Topic can be validated, can be defaulted, and has immutable fields.
var _ runtime.Object = (*Topic)(nil)
var _ resourcesemantics.GenericCRD = (*Topic)(nil)

// Check that Topic implements the Conditions duck type.
var _ = duck.VerifyType(&Topic{}, &duckv1.Conditions{})

// Check that Topic implements the KRShaped duck type.
var _ duckv1.KRShaped = (*Topic)(nil)

// TopicSpec defines parameters for creating or publishing to a Cloud Pub/Sub
// Topic depending on the PropagationPolicy.
type TopicSpec struct {
	v1.IdentitySpec `json:",inline"`

	// Secret is the credential to be used to create and publish into the
	// Cloud Pub/Sub Topic. The value of the secret entry must be a service
	// account key in the JSON format
	// (see https://cloud.google.com/iam/docs/creating-managing-service-account-keys).
	Secret *corev1.SecretKeySelector `json:"secret,omitempty"`

	// Project is the ID of the Google Cloud Project that the Pub/Sub
	// Topic will be created in or used from.
	Project string `json:"project,omitempty"`

	// Topic is the ID of the Topic to create/use in Google Cloud Pub/Sub.
	Topic string `json:"topic,omitempty"`

	//PropagationPolicy defines how Topic controls the Cloud Pub/Sub topic for
	// lifecycle changes. Defaults to TopicPolicyCreateNoDelete if empty.
	PropagationPolicy PropagationPolicyType `json:"propagationPolicy,omitempty"`

	// EnablePublisher controls the creation of an HTTP publisher endpoint. If set to true, then
	// a publisher will be created and this Topic will be Addressable (have status.address). If set
	// to false, then no publisher will be created and this custom object represents the creation
	// and deletion of a GCP Pub/Sub Topic only.
	// Defaults to true.
	// +optional
	EnablePublisher *bool `json:"publisher,omitempty"`
}

// PropagationPolicyType defines enum type for TopicPolicy
type PropagationPolicyType string

const (
	// TopicPolicyCreateDelete defines the Cloud Pub/Sub topic management
	// policy for creating topic (if not present), and deleting topic when the
	// Topic resource is deleted.
	TopicPolicyCreateDelete PropagationPolicyType = "CreateDelete"

	// TopicPolicyCreateNoDelete defines the Cloud Pub/Sub topic management
	// policy for creating topic (if not present), and not deleting topic when
	// the Topic resource is deleted.
	TopicPolicyCreateNoDelete PropagationPolicyType = "CreateNoDelete"

	// TopicPolicyNoCreateNoDelete defines the Cloud Pub/Sub topic
	// management policy for only using existing topics, and not deleting
	// topic when the Topic resource is deleted.
	TopicPolicyNoCreateNoDelete PropagationPolicyType = "NoCreateNoDelete"
)

var topicCondSet = apis.NewLivingConditionSet(
	TopicConditionTopicExists,
)

const (
	// TopicConditionReady has status True when all subconditions below have
	// been set to True.
	TopicConditionReady = apis.ConditionReady

	// TopicConditionAddressable has status true when this Topic meets the
	// Addressable contract and has a non-empty hostname.
	TopicConditionAddressable apis.ConditionType = "Addressable"

	// TopicConditionTopicExists has status True when the Topic has had a
	// Pub/Sub topic created for it.
	TopicConditionTopicExists apis.ConditionType = "TopicExists"

	// TopicConditionPublisherReady has status True when the Topic has had
	// its publisher deployment created and ready.
	TopicConditionPublisherReady apis.ConditionType = "PublisherReady"
)

// TopicStatus represents the current state of a Topic.
type TopicStatus struct {
	v1.IdentityStatus `json:",inline"`

	// Topic is Addressable. It currently exposes the endpoint as a
	// fully-qualified DNS name which will distribute traffic over the
	// provided targets from inside the cluster.
	//
	// It generally has the form {Topic}.{namespace}.svc.{cluster domain name}
	duckv1.AddressStatus `json:",inline"`

	// ProjectID is the resolved project ID in use by the Topic.
	// +optional
	ProjectID string `json:"projectId,omitempty"`

	// TopicID is the created topic ID used by the Topic.
	// +optional
	TopicID string `json:"topicId,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopicList is a collection of Pub/Sub backed Topics.
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for Pub/Sub backed Topic.
func (t *Topic) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Topic")
}

// Methods for identifiable interface.
// IdentitySpec returns the IdentitySpec portion of the Spec.
func (t *Topic) IdentitySpec() *v1.IdentitySpec {
	return &t.Spec.IdentitySpec
}

// IdentityStatus returns the IdentityStatus portion of the Status.
func (t *Topic) IdentityStatus() *v1.IdentityStatus {
	return &t.Status.IdentityStatus
}

// ConditionSet returns the apis.ConditionSet of the embedding object
func (t *Topic) ConditionSet() *apis.ConditionSet {
	return &topicCondSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*Topic) GetConditionSet() apis.ConditionSet {
	return topicCondSet
}

// GetStatus retrieves the status of the Topic. Implements the KRShaped interface.
func (t *Topic) GetStatus() *duckv1.Status {
	return &t.Status.Status
}
