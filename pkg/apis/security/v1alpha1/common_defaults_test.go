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

	corev1 "k8s.io/api/core/v1"
)

func TestJWTDefaults(t *testing.T) {
	j := &JWTSpec{}
	j.SetDefaults(context.Background())
	if j.JwtHeader != "Authorization" {
		t.Errorf("default JwtHeader got=%s want=Authorization", j.JwtHeader)
	}
}

func TestPolicyBindingSpecDefaults(t *testing.T) {
	spec := &PolicyBindingSpec{Policy: &corev1.ObjectReference{}}
	spec.SetDefaults(context.Background(), "test-namespace")
	if spec.Subject.Namespace != "test-namespace" {
		t.Errorf("spec.Subject.Namespace got=%s want=test-namespace", spec.Subject.Namespace)
	}
	if spec.Policy.Namespace != "test-namespace" {
		t.Errorf("spec.Policy.Namespace got=%s want=test-namespace", spec.Policy.Namespace)
	}
}
