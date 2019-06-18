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

package v1alpha1

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

var (
	pullSubscriptionSpec = PullSubscriptionSpec{
		Secret: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "secret-name",
			},
			Key: "secret-key",
		},
		Project: "my-eventing-project",
		Topic:   "pubsub-topic",
		Sink: &corev1.ObjectReference{
			APIVersion: "foo",
			Kind:       "bar",
			Namespace:  "baz",
			Name:       "qux",
		},
		ServiceAccountName: "service-account-name",
	}
)

func TestPubSubCheckImmutableFields(t *testing.T) {
	testCases := map[string]struct {
		orig    interface{}
		updated PullSubscriptionSpec
		allowed bool
	}{
		"nil orig": {
			updated: pullSubscriptionSpec,
			allowed: true,
		},
		"Secret.Name changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "some-other-name",
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project:            pullSubscriptionSpec.Project,
				Topic:              pullSubscriptionSpec.Topic,
				Sink:               pullSubscriptionSpec.Sink,
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Secret.Key changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: "some-other-key",
				},
				Project:            pullSubscriptionSpec.Project,
				Topic:              pullSubscriptionSpec.Topic,
				Sink:               pullSubscriptionSpec.Sink,
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Project changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project:            "some-other-project",
				Topic:              pullSubscriptionSpec.Topic,
				Sink:               pullSubscriptionSpec.Sink,
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Topic changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project:            pullSubscriptionSpec.Project,
				Topic:              "some-other-topic",
				Sink:               pullSubscriptionSpec.Sink,
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Sink.APIVersion changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project: pullSubscriptionSpec.Project,
				Topic:   pullSubscriptionSpec.Topic,
				Sink: &corev1.ObjectReference{
					APIVersion: "some-other-api-version",
					Kind:       pullSubscriptionSpec.Sink.Kind,
					Namespace:  pullSubscriptionSpec.Sink.Namespace,
					Name:       pullSubscriptionSpec.Sink.Name,
				},
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Sink.Kind changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project: pullSubscriptionSpec.Project,
				Topic:   pullSubscriptionSpec.Topic,
				Sink: &corev1.ObjectReference{
					APIVersion: pullSubscriptionSpec.Sink.APIVersion,
					Kind:       "some-other-kind",
					Namespace:  pullSubscriptionSpec.Sink.Namespace,
					Name:       pullSubscriptionSpec.Sink.Name,
				},
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Sink.Namespace changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project: pullSubscriptionSpec.Project,
				Topic:   pullSubscriptionSpec.Topic,
				Sink: &corev1.ObjectReference{
					APIVersion: pullSubscriptionSpec.Sink.APIVersion,
					Kind:       pullSubscriptionSpec.Sink.Kind,
					Namespace:  "some-other-namespace",
					Name:       pullSubscriptionSpec.Sink.Name,
				},
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"Sink.Name changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project: pullSubscriptionSpec.Project,
				Topic:   pullSubscriptionSpec.Topic,
				Sink: &corev1.ObjectReference{
					APIVersion: pullSubscriptionSpec.Sink.APIVersion,
					Kind:       pullSubscriptionSpec.Sink.Kind,
					Namespace:  pullSubscriptionSpec.Sink.Namespace,
					Name:       "some-other-name",
				},
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"ServiceAccountName changed": {
			orig: &pullSubscriptionSpec,
			updated: PullSubscriptionSpec{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: pullSubscriptionSpec.Secret.Name,
					},
					Key: pullSubscriptionSpec.Secret.Key,
				},
				Project: pullSubscriptionSpec.Project,
				Topic:   pullSubscriptionSpec.Topic,
				Sink: &corev1.ObjectReference{
					APIVersion: pullSubscriptionSpec.Sink.APIVersion,
					Kind:       pullSubscriptionSpec.Sink.Kind,
					Namespace:  pullSubscriptionSpec.Sink.Namespace,
					Name:       "some-other-name",
				},
				ServiceAccountName: pullSubscriptionSpec.ServiceAccountName,
			},
			allowed: false,
		},
		"no change": {
			orig:    &pullSubscriptionSpec,
			updated: pullSubscriptionSpec,
			allowed: true,
		},
		"not spec": {
			orig:    []string{"wrong"},
			updated: pullSubscriptionSpec,
			allowed: true,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var orig *PullSubscription

			if tc.orig != nil {
				if spec, ok := tc.orig.(*PullSubscriptionSpec); ok {
					orig = &PullSubscription{
						Spec: *spec,
					}
				}
			}
			updated := &PullSubscription{
				Spec: tc.updated,
			}
			err := updated.CheckImmutableFields(context.TODO(), orig)
			if tc.allowed != (err == nil) {
				t.Fatalf("Unexpected immutable field check. Expected %v. Actual %v", tc.allowed, err)
			}
		})
	}
}
