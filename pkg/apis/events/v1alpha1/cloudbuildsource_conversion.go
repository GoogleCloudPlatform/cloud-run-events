/*
Copyright 2020 Google LLC

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
	"fmt"

	"github.com/google/knative-gcp/pkg/apis/convert"
	"github.com/google/knative-gcp/pkg/apis/events/v1beta1"
	"knative.dev/pkg/apis"
)

// ConvertTo implements apis.Convertible.
// Converts source (from v1beta1.CloudBuildSource) into v1alpha1.CloudBuildSource.
func (source *CloudBuildSource) ConvertTo(_ context.Context, to apis.Convertible) error {
	switch sink := to.(type) {
	case *v1beta1.CloudBuildSource:
		sink.ObjectMeta = source.ObjectMeta
		sink.Spec.PubSubSpec = convert.ToV1beta1PubSubSpec(source.Spec.PubSubSpec)
		sink.Status.PubSubStatus = convert.ToV1beta1PubSubStatus(source.Status.PubSubStatus)
		return nil
	default:
		return fmt.Errorf("unknown conversion, got: %T", sink)

	}
}

// ConvertFrom implements apis.Convertible.
// Converts obj from v1alpha1.CloudBuildSource into v1beta1.CloudBuildSource.
func (sink *CloudBuildSource) ConvertFrom(_ context.Context, from apis.Convertible) error {
	switch source := from.(type) {
	case *v1beta1.CloudBuildSource:
		sink.ObjectMeta = source.ObjectMeta
		sink.Spec.PubSubSpec = convert.FromV1beta1PubSubSpec(source.Spec.PubSubSpec)
		sink.Status.PubSubStatus = convert.FromV1beta1PubSubStatus(source.Status.PubSubStatus)
		return nil
	default:
		return fmt.Errorf("unknown conversion, got: %T", source)
	}
}
