/*
Copyright 2019 Google LLC

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

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/kmeta"
	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	pubsubv1alpha1 "github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
)

// MakePullSubscription creates the spec for, but does not create, a GCP PullSubscription
// for a given GCS.
func MakePullSubscription(source *v1alpha1.Storage, project string) *pubsubv1alpha1.PullSubscription {
	labels := map[string]string{
		"receive-adapter": "gcssource",
	}

	pubsubSecret := source.Spec.GCSSecret
	if source.Spec.PullSubscriptionSecret != nil {
		pubsubSecret = *source.Spec.PullSubscriptionSecret
	}

	return &pubsubv1alpha1.PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      source.Name,
			Namespace: source.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(source),
			},
		},
		Spec: pubsubv1alpha1.PullSubscriptionSpec{
			Secret:  &pubsubSecret,
			Project: project,
			Topic:   source.Status.Topic,
			Sink: apisv1alpha1.Destination{
				ObjectReference: &source.Spec.Sink,
			},
		},
	}
}
