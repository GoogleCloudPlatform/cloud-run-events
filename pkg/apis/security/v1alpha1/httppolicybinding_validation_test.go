/*
Copyright 2020 Google LLC.

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

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/tracker"
)

func TestHTTPPolicyBindingValidation(t *testing.T) {
	cases := []struct {
		name    string
		pb      HTTPPolicyBinding
		wantErr *apis.FieldError
	}{{
		name: "valid",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Name:      "subject",
						Namespace: "foo",
					},
				},
				Policy: duckv1.KReference{
					Name:      "policy",
					Namespace: "foo",
				},
			},
		},
	}, {
		name: "subject namespace mismatch",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Name:      "subject",
						Namespace: "bar",
					},
				},
				Policy: duckv1.KReference{
					Name:      "policy",
					Namespace: "foo",
				},
			},
		},
		wantErr: apis.ErrInvalidValue("bar", "namespace").ViaField("subject").ViaField("spec"),
	}, {
		name: "subject name and selector not specified",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Namespace: "foo",
					},
				},
				Policy: duckv1.KReference{
					Name:      "policy",
					Namespace: "foo",
				},
			},
		},
		wantErr: apis.ErrMissingOneOf("name", "selector").ViaField("subject").ViaField("spec"),
	}, {
		name: "policy namespace mismatch",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Name:      "subject",
						Namespace: "foo",
					},
				},
				Policy: duckv1.KReference{
					Name:      "policy",
					Namespace: "bar",
				},
			},
		},
		wantErr: apis.ErrInvalidValue("bar", "namespace").ViaField("policy").ViaField("spec"),
	}, {
		name: "policy API specified",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Name:      "subject",
						Namespace: "foo",
					},
				},
				Policy: duckv1.KReference{
					APIVersion: "other.policy",
					Name:       "policy",
					Namespace:  "foo",
				},
			},
		},
		wantErr: apis.ErrDisallowedFields("apiVersion", "kind").ViaField("policy").ViaField("spec"),
	}, {
		name: "policy kind specified",
		pb: HTTPPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-binding",
				Namespace: "foo",
			},
			Spec: PolicyBindingSpec{
				BindingSpec: duckv1alpha1.BindingSpec{
					Subject: tracker.Reference{
						Name:      "subject",
						Namespace: "foo",
					},
				},
				Policy: duckv1.KReference{
					Kind:      "other.kind",
					Name:      "policy",
					Namespace: "foo",
				},
			},
		},
		wantErr: apis.ErrDisallowedFields("apiVersion", "kind").ViaField("policy").ViaField("spec"),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.pb.Validate(context.Background())
			if diff := cmp.Diff(tc.wantErr.Error(), gotErr.Error()); diff != "" {
				t.Errorf("HTTPPolicyBinding.Validate (-want, +got) = %v", diff)
			}
		})
	}
}
