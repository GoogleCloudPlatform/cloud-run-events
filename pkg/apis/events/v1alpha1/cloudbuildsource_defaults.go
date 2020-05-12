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

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	metadataClient "github.com/google/knative-gcp/pkg/gclient/metadata"

	"knative.dev/pkg/ptr"
)

const (
	DefaultTopic = "cloud-builds"
)

func (bs *CloudBuildSource) SetDefaults(ctx context.Context) {
	bs.Spec.SetDefaults(ctx)
	duckv1alpha1.SetClusterNameAnnotation(&bs.ObjectMeta, metadataClient.NewDefaultMetadataClient())
	duckv1alpha1.SetAutoscalingAnnotationsDefaults(ctx, &bs.ObjectMeta)
}

func (bss *CloudBuildSourceSpec) SetDefaults(ctx context.Context) {
	bss.SetPubSubDefaults()
	if bss.Topic == nil {
		topic := DefaultTopic
		bss.Topic = ptr.String(topic)
	}

}
