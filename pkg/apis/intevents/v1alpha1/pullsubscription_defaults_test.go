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

package v1alpha1

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gcpauthtesthelper "github.com/google/knative-gcp/pkg/apis/configs/gcpauth/testhelper"
	"github.com/google/knative-gcp/pkg/apis/duck"
	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	testingMetadataClient "github.com/google/knative-gcp/pkg/gclient/metadata/testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

func TestPullSubscriptionDefaults(t *testing.T) {

	defaultRetentionDuration := defaultRetentionDuration
	defaultAckDeadline := defaultAckDeadline

	tests := []struct {
		name  string
		start *PullSubscription
		want  *PullSubscription
	}{{
		name: "non-nil structured",
		start: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModeCloudEventsStructured,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "my-cloud-key",
						},
						Key: "test.json",
					},
				},
			},
		},
		want: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModeCloudEventsStructured,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "my-cloud-key",
						},
						Key: "test.json",
					},
				},
			},
		},
	}, {
		name: "non-nil push",
		start: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode: ModePushCompatible,
			},
		},
		want: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModePushCompatible,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &gcpauthtesthelper.Secret,
				},
			},
		},
	}, {
		name: "non-nil invalid",
		start: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode: "invalid",
			},
		},
		want: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModeCloudEventsBinary,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &gcpauthtesthelper.Secret,
				},
			},
		},
	}, {
		name: "nil",
		start: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{},
		},
		want: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModeCloudEventsBinary,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &gcpauthtesthelper.Secret,
				},
			},
		},
	}, {
		name: "nil secret",
		start: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{},
		},
		want: &PullSubscription{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
				},
			},
			Spec: PullSubscriptionSpec{
				Mode:              ModeCloudEventsBinary,
				RetentionDuration: ptr.String(defaultRetentionDuration.String()),
				AckDeadline:       ptr.String(defaultAckDeadline.String()),
				PubSubSpec: duckv1alpha1.PubSubSpec{
					Secret: &gcpauthtesthelper.Secret,
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.start
			got.SetDefaults(gcpauthtesthelper.ContextWithDefaults())

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("failed to get expected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestPullSubscriptionDefaults_NoChange(t *testing.T) {
	days2 := 2 * 24 * time.Hour
	secs60 := 60 * time.Second
	want := &PullSubscription{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				duck.ClusterNameAnnotation: testingMetadataClient.FakeClusterName,
			},
		},
		Spec: PullSubscriptionSpec{
			Mode:              ModeCloudEventsBinary,
			AckDeadline:       ptr.String(secs60.String()),
			RetentionDuration: ptr.String(days2.String()),
			PubSubSpec: duckv1alpha1.PubSubSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-cloud-key",
					},
					Key: "test.json",
				},
			},
		},
	}

	got := want.DeepCopy()
	got.SetDefaults(gcpauthtesthelper.ContextWithDefaults())
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("failed to get expected (-want, +got) = %v", diff)
	}
}
