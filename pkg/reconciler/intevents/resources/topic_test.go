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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/knative-gcp/pkg/apis/duck"
	duckv1beta1 "github.com/google/knative-gcp/pkg/apis/duck/v1beta1"
	"github.com/google/knative-gcp/pkg/apis/events/v1beta1"
	inteventsv1beta1 "github.com/google/knative-gcp/pkg/apis/intevents/v1beta1"
	metadatatesting "github.com/google/knative-gcp/pkg/gclient/metadata/testing"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestMakeTopicWithCloudStorageSource(t *testing.T) {
	source := &v1beta1.CloudStorageSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "storage-name",
			Namespace: "storage-namespace",
			UID:       "storage-uid",
			Annotations: map[string]string{
				duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
			},
		},
		Spec: v1beta1.CloudStorageSourceSpec{
			PubSubSpec: duckv1beta1.PubSubSpec{
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
				},
			},
			Bucket: "this-bucket",
		},
	}
	args := &TopicArgs{
		Namespace: source.Namespace,
		Name:      source.Name,
		Spec:      &source.Spec.PubSubSpec,
		Owner:     source,
		Topic:     "topic-abc",
		Labels: map[string]string{
			"receive-adapter": "storage.events.cloud.google.com",
			"source":          source.Name,
		},
		Annotations: map[string]string{
			duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
		},
	}
	got := MakeTopic(args)

	yes := true
	want := &inteventsv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "storage-namespace",
			Name:      "storage-name",
			Labels: map[string]string{
				"receive-adapter": "storage.events.cloud.google.com",
				"source":          "storage-name",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "events.cloud.google.com/v1beta1",
				Kind:               "CloudStorageSource",
				Name:               "storage-name",
				UID:                "storage-uid",
				Controller:         &yes,
				BlockOwnerDeletion: &yes,
			}},
			Annotations: map[string]string{
				duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
			},
		},
		Spec: inteventsv1beta1.TopicSpec{
			Secret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "eventing-secret-name",
				},
				Key: "eventing-secret-key",
			},
			Project:           "project-123",
			Topic:             "topic-abc",
			PropagationPolicy: inteventsv1beta1.TopicPolicyCreateDelete,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestMakeTopicWithCloudSchedulerSource(t *testing.T) {
	source := &v1beta1.CloudSchedulerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scheduler-name",
			Namespace: "scheduler-namespace",
			UID:       "scheduler-uid",
			Annotations: map[string]string{
				duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
			},
		},
		Spec: v1beta1.CloudSchedulerSourceSpec{
			PubSubSpec: duckv1beta1.PubSubSpec{
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
				},
			},
		},
	}
	args := &TopicArgs{
		Namespace: source.Namespace,
		Name:      source.Name,
		Spec:      &source.Spec.PubSubSpec,
		Owner:     source,
		Topic:     "topic-abc",
		Labels: map[string]string{
			"receive-adapter": "scheduler.events.cloud.google.com",
			"source":          source.Name,
		},
		Annotations: map[string]string{
			duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
		},
	}
	got := MakeTopic(args)

	yes := true
	want := &inteventsv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "scheduler-namespace",
			Name:      "scheduler-name",
			Labels: map[string]string{
				"receive-adapter": "scheduler.events.cloud.google.com",
				"source":          "scheduler-name",
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "events.cloud.google.com/v1beta1",
				Kind:               "CloudSchedulerSource",
				Name:               "scheduler-name",
				UID:                "scheduler-uid",
				Controller:         &yes,
				BlockOwnerDeletion: &yes,
			}},
			Annotations: map[string]string{
				duck.ClusterNameAnnotation: metadatatesting.FakeClusterName,
			},
		},
		Spec: inteventsv1beta1.TopicSpec{
			Secret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "eventing-secret-name",
				},
				Key: "eventing-secret-key",
			},
			Project:           "project-123",
			Topic:             "topic-abc",
			PropagationPolicy: inteventsv1beta1.TopicPolicyCreateDelete,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}
