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

package resources

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"

	"github.com/google/knative-gcp/pkg/apis/messaging/v1alpha1"
)

func TestGenerateTopicID(t *testing.T) {
	want := "cre-chan_default_foo_a-uid"
	got := GenerateTopicID(&v1alpha1.Channel{
		ObjectMeta: v1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			UID:       "a-uid",
		},
	})

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestGeneratePublisherName(t *testing.T) {
	want := "cre-foo-chan"
	got := GeneratePublisherName(&v1alpha1.Channel{
		ObjectMeta: v1.ObjectMeta{
			Name: "foo",
		},
	})

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestGeneratePublisherNameFromChannel(t *testing.T) {
	want := "cre-foo-chan"
	got := GeneratePublisherName(&v1alpha1.Channel{
		ObjectMeta: v1.ObjectMeta{
			Name: "cre-foo",
		},
	})

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestGenerateSubscriptionName(t *testing.T) {
	want := "cre-sub-a-uid"
	got := GeneratePullSubscriptionName("a-uid")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestExtractUIDFromSubscriptionName(t *testing.T) {
	want := "a-uid"
	got := ExtractUIDFromPullSubscriptionName("cre-sub-a-uid")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}
