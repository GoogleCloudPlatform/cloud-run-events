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

package v1

import (
	"testing"

	"github.com/google/knative-gcp/pkg/reconciler/intevents/resources"

	"github.com/google/go-cmp/cmp"
	gcpduckv1 "github.com/google/knative-gcp/pkg/apis/duck/v1"
	v1 "github.com/google/knative-gcp/pkg/apis/events/v1"
	inteventsv1 "github.com/google/knative-gcp/pkg/apis/intevents/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestMakePullSubscription(t *testing.T) {
	source := &v1.CloudStorageSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bucket-name",
			Namespace: "bucket-namespace",
			UID:       "bucket-uid",
		},
		Spec: v1.CloudStorageSourceSpec{
			Bucket: "this-bucket",
			PubSubSpec: gcpduckv1.PubSubSpec{
				Project: "project-123",
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "eventing-secret-name",
					},
					Key: "eventing-secret-key",
				},
				SourceSpec: duckv1.SourceSpec{
					Sink: duckv1.Destination{
						Ref: &duckv1.KReference{
							APIVersion: "v1",
							Kind:       "Kitchen",
							Name:       "sink",
						},
					},
					CloudEventOverrides: &duckv1.CloudEventOverrides{
						Extensions: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		},
	}
	args := &PullSubscriptionArgs{
		Namespace:   source.Namespace,
		Name:        source.Name,
		Spec:        &source.Spec.PubSubSpec,
		Owner:       source,
		Topic:       "topic-abc",
		AdapterType: "google.storage",
		Annotations: resources.GetAnnotations(nil, "storages.events.cloud.google.com"),
		Labels: map[string]string{
			"receive-adapter":                     "storage.events.cloud.google.com",
			"events.cloud.google.com/source-name": source.Name,
		},
	}
	got := MakePullSubscription(args)

	yes := true
	want := &inteventsv1.PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "bucket-namespace",
			Name:      "bucket-name",
			Labels: map[string]string{
				"receive-adapter":                     "storage.events.cloud.google.com",
				"events.cloud.google.com/source-name": "bucket-name",
			},
			Annotations: map[string]string{
				"metrics-resource-group": "storages.events.cloud.google.com",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "events.cloud.google.com/v1",
				Kind:               "CloudStorageSource",
				Name:               "bucket-name",
				UID:                "bucket-uid",
				Controller:         &yes,
				BlockOwnerDeletion: &yes,
			}},
		},
		Spec: inteventsv1.PullSubscriptionSpec{
			PubSubSpec: gcpduckv1.PubSubSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "eventing-secret-name",
					},
					Key: "eventing-secret-key",
				},
				Project: "project-123",
				SourceSpec: duckv1.SourceSpec{
					Sink: duckv1.Destination{
						Ref: &duckv1.KReference{
							APIVersion: "v1",
							Kind:       "Kitchen",
							Name:       "sink",
						},
					},
					CloudEventOverrides: &duckv1.CloudEventOverrides{
						Extensions: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			Topic:       "topic-abc",
			AdapterType: "google.storage",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}