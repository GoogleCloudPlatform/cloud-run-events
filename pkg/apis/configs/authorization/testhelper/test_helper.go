/*
Copyright 2020 Google LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authorizationtesthelper

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/knative-gcp/pkg/apis/configs/authorization"
)

var Secret = corev1.SecretKeySelector{
	LocalObjectReference: corev1.LocalObjectReference{
		Name: "google-cloud-key",
	},
	Key: "key.json",
}

func ContextWithDefaults() context.Context {
	return WithDefaults(context.Background())
}

func WithDefaults(ctx context.Context) context.Context {
	d, _ := authorization.NewDefaultsConfigFromMap(map[string]string{
		"default-auth-config": ` 
  clusterDefaults:
    secret:
      name: google-cloud-key
      key: key.json
`,
	})
	cfg := &authorization.Config{
		AuthorizationDefaults: d,
	}
	return authorization.ToContext(ctx, cfg)
}
