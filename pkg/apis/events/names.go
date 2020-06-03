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

package events

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateName generates a name based on the Source information.
func GenerateName(obj metav1.ObjectMetaAccessor) string {
	meta := obj.GetObjectMeta()
	return fmt.Sprintf("cre-src_%s_%s_%s", meta.GetNamespace(), meta.GetName(), string(meta.GetUID()))
}
