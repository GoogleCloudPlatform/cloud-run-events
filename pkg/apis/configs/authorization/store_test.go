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

package authorization

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
	logtesting "knative.dev/pkg/logging/testing"

	. "knative.dev/pkg/configmap/testing"
)

var ignoreStuff = cmp.Options{
	cmpopts.IgnoreUnexported(resource.Quantity{}),
}

func TestStoreLoadWithContext(t *testing.T) {
	store := NewStore(logtesting.TestLogger(t))

	_, defaultsConfig := ConfigMapsFromTestFile(t, ConfigName, defaulterKey)

	store.OnConfigChanged(defaultsConfig)

	config := FromContextOrDefaults(store.ToContext(context.Background()))

	t.Run("defaults", func(t *testing.T) {
		expected, _ := NewDefaultsConfigFromConfigMap(defaultsConfig)
		if diff := cmp.Diff(expected, config.AuthorizationDefaults, ignoreStuff...); diff != "" {
			t.Errorf("Unexpected defaults config (-want, +got): %v", diff)
			t.Fatalf("Unexpected defaults config (-want, +got): %v", diff)
		}
	})
}
