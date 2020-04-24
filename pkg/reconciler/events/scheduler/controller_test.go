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

package scheduler

import (
	"testing"

	iamtesting "github.com/google/knative-gcp/pkg/reconciler/testing"
	"knative.dev/pkg/configmap"
	logtesting "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"

	// Fake injection informers
	_ "github.com/google/knative-gcp/pkg/client/clientset/versioned/typed/pubsub/v1alpha1/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/client/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/events/v1alpha1/cloudschedulersource/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/pullsubscription/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/topic/fake"
	_ "github.com/google/knative-gcp/pkg/reconciler/testing"
	_ "knative.dev/pkg/client/injection/kube/informers/batch/v1/job/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount/fake"
)

func TestNew(t *testing.T) {
	defer logtesting.ClearAll()
	ctx, _ := SetupFakeContext(t)

	c := newController(ctx, configmap.NewStaticWatcher(), iamtesting.NoopIAMPolicyManager)

	if c == nil {
		t.Fatal("Expected newControllerWithIAMPolicyManager to return a non-nil value")
	}
}
