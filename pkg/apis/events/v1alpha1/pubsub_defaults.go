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
	"time"

	"knative.dev/pkg/ptr"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

const (
	defaultRetentionDuration = 7 * 24 * time.Hour
	defaultAckDeadline       = 30 * time.Second
)

func (ps *PubSub) SetDefaults(ctx context.Context) {
	ps.Spec.SetDefaults(ctx)
}

func (pss *PubSubSpec) SetDefaults(ctx context.Context) {
	if pss.AckDeadline == nil {
		ackDeadline := defaultAckDeadline
		pss.AckDeadline = ptr.String(ackDeadline.String())
	}

	if pss.RetentionDuration == nil {
		retentionDuration := defaultRetentionDuration
		pss.RetentionDuration = ptr.String(retentionDuration.String())
	}

	if pss.Secret == nil || equality.Semantic.DeepEqual(pss.Secret, &corev1.SecretKeySelector{}) {
		pss.Secret = duckv1alpha1.DefaultGoogleCloudSecretSelector()
	}
}
