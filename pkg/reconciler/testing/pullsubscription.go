/*
Copyright 2019 The Knative Authors

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

package testing

import (
	"context"
	"time"

	authorizationtesthelper "github.com/google/knative-gcp/pkg/apis/configs/authorization/testhelper"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/google/knative-gcp/pkg/apis/intevents/v1alpha1"
)

// PullSubscriptionOption enables further configuration of a PullSubscription.
type PullSubscriptionOption func(*v1alpha1.PullSubscription)

// NewPullSubscription creates a PullSubscription with PullSubscriptionOptions
func NewPullSubscription(name, namespace string, so ...PullSubscriptionOption) *v1alpha1.PullSubscription {
	s := &v1alpha1.PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

// NewPullSubscriptionWithNoDefaults creates a PullSubscription with
// PullSubscriptionOptions but does not set defaults.
func NewPullSubscriptionWithNoDefaults(name, namespace string, so ...PullSubscriptionOption) *v1alpha1.PullSubscription {
	s := &v1alpha1.PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	return s
}

// NewPullSubscriptionWithoutNamespace creates a PullSubscription with PullSubscriptionOptions but without a specific namespace
func NewPullSubscriptionWithoutNamespace(name string, so ...PullSubscriptionOption) *v1alpha1.PullSubscription {
	s := &v1alpha1.PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

func WithPullSubscriptionUID(uid types.UID) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.UID = uid
	}
}

func WithPullSubscriptionGenerateName(generateName string) PullSubscriptionOption {
	return func(c *v1alpha1.PullSubscription) {
		c.ObjectMeta.GenerateName = generateName
	}
}

// WithInitPullSubscriptionConditions initializes the PullSubscriptions's conditions.
func WithInitPullSubscriptionConditions(s *v1alpha1.PullSubscription) {
	s.Status.InitializeConditions()
}

func WithPullSubscriptionSink(gvk metav1.GroupVersionKind, name string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec.Sink = duckv1.Destination{
			Ref: &duckv1.KReference{
				APIVersion: apiVersion(gvk),
				Kind:       gvk.Kind,
				Name:       name,
			},
		}
	}
}

func WithPullSubscriptionTransformer(gvk metav1.GroupVersionKind, name string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec.Transformer = &duckv1.Destination{
			Ref: &duckv1.KReference{
				APIVersion: apiVersion(gvk),
				Kind:       gvk.Kind,
				Name:       name,
			},
		}
	}
}

func WithPullSubscriptionMarkSink(uri *apis.URL) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkSink(uri)
	}
}

func WithPullSubscriptionMarkTransformer(uri *apis.URL) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkTransformer(uri)
	}
}

func WithPullSubscriptionMarkNoTransformer(reason, message string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkNoTransformer(reason, message)
	}
}

func WithPullSubscriptionMarkSubscribed(subscriptionID string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkSubscribed(subscriptionID)
	}
}

func WithPullSubscriptionSubscriptionID(subscriptionID string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.SubscriptionID = subscriptionID
	}
}

func WithPullSubscriptionProjectID(projectID string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.ProjectID = projectID
	}
}

func WithPullSubscriptionTransformerURI(uri *apis.URL) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.TransformerURI = uri
	}
}

func WithPullSubscriptionMarkNoSubscription(reason, message string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkNoSubscription(reason, message)
	}
}

func WithPullSubscriptionMarkDeployed(ps *v1alpha1.PullSubscription) {
	ps.Status.MarkDeployed()
}

func WithPullSubscriptionSpec(spec v1alpha1.PullSubscriptionSpec) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec = spec
		s.Spec.SetDefaults(apis.WithinParent(context.Background(), s.ObjectMeta))
	}
}

// Same as withPullSubscriptionSpec but does not set defaults
func WithPullSubscriptionSpecWithNoDefaults(spec v1alpha1.PullSubscriptionSpec) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec = spec
	}
}

func WithPullSubscriptionReady(sink *apis.URL) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.InitializeConditions()
		s.Status.MarkSink(sink)
		s.Status.MarkDeployed()
		s.Status.MarkSubscribed("subID")
	}
}

func WithPullSubscriptionFailed() PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.InitializeConditions()
		s.Status.MarkNoSink("InvalidSink",
			`failed to get ref &ObjectReference{Kind:Sink,Namespace:testnamespace,Name:sink,UID:,APIVersion:testing.cloud.google.com/v1alpha1,ResourceVersion:,FieldPath:,}: sinks.testing.cloud.google.com "sink" not found`)

	}
}

func WithPullSubscriptionUnknown() PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.InitializeConditions()
	}
}

func WithPullSubscriptionJobFailure(subscriptionID, reason, message string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.SubscriptionID = subscriptionID
		s.Status.MarkNoSubscription(reason, message)
	}
}

func WithPullSubscriptionSinkNotFound() PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.MarkNoSink("InvalidSink",
			`failed to get ref &ObjectReference{Kind:Sink,Namespace:testnamespace,Name:sink,UID:,APIVersion:testing.cloud.google.com/v1alpha1,ResourceVersion:,FieldPath:,}: sinks.testing.cloud.google.com "sink" not found`)
	}
}

func WithPullSubscriptionDeleted(s *v1alpha1.PullSubscription) {
	t := metav1.NewTime(time.Unix(1e9, 0))
	s.ObjectMeta.SetDeletionTimestamp(&t)
}

func WithPullSubscriptionOwnerReferences(ownerReferences []metav1.OwnerReference) PullSubscriptionOption {
	return func(c *v1alpha1.PullSubscription) {
		c.ObjectMeta.OwnerReferences = ownerReferences
	}
}

func WithPullSubscriptionLabels(labels map[string]string) PullSubscriptionOption {
	return func(c *v1alpha1.PullSubscription) {
		c.ObjectMeta.Labels = labels
	}
}

func WithPullSubscriptionAnnotations(annotations map[string]string) PullSubscriptionOption {
	return func(c *v1alpha1.PullSubscription) {
		c.ObjectMeta.Annotations = annotations
	}
}

func WithPullSubscriptionFinalizers(finalizers ...string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Finalizers = finalizers
	}
}

func WithPullSubscriptionStatusObservedGeneration(generation int64) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.Status.ObservedGeneration = generation
	}
}

func WithPullSubscriptionObjectMetaGeneration(generation int64) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.ObjectMeta.Generation = generation
	}
}

func WithPullSubscriptionReadyStatus(status corev1.ConditionStatus, reason, message string) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Status.Conditions = []apis.Condition{{
			Type:    apis.ConditionReady,
			Status:  status,
			Reason:  reason,
			Message: message,
		}}
	}
}

func WithPullSubscriptionMode(mode v1alpha1.ModeType) PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec.Mode = mode
	}
}

func WithPullSubscriptionDefaultAuthorization() PullSubscriptionOption {
	return func(s *v1alpha1.PullSubscription) {
		s.Spec.PubSubSpec.SetPubSubDefaults(authorizationtesthelper.ContextWithDefaults())
	}
}
