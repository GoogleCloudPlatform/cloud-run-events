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
	"golang.org/x/net/context"
	"google.golang.org/api/iam/v1"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	gServiceAccount = "test"
	kServiceAccount = &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	projectID   = "project"
	role        = "roles/iam.workloadIdentityUser"
	addbindings = []*iam.Binding{{
		Members: []string{"member1"},
		Role:    role,
	}}
	removebindings = []*iam.Binding{{
		Members: []string{"member1", "member2"},
		Role:    role,
	}}
)

func TestAddIamPolicyBinding(t *testing.T) {
	want := "failed to add iam policy binding: failed to get iam policy: googleapi: Error 400: Invalid service account email (test)., badRequest"
	gotb := AddIamPolicyBinding(context.Background(), projectID, &gServiceAccount, kServiceAccount)
	got := gotb.Error()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestRemoveIamPolicyBinding(t *testing.T) {
	want := "failed to remove iam policy binding: failed to get iam policy: googleapi: Error 400: Invalid service account email (test)., badRequest"
	gotb := RemoveIamPolicyBinding(context.Background(), projectID, &gServiceAccount, kServiceAccount)
	got := gotb.Error()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func TestMakeSetIamPolicyRequest(t *testing.T) {
	testCases := []struct {
		name string
		want *iam.SetIamPolicyRequest
		got  *iam.SetIamPolicyRequest
	}{{
		name: "Add iam policy binding",
		want: &iam.SetIamPolicyRequest{
			Policy: &iam.Policy{
				Bindings: []*iam.Binding{{
					Members: []string{"member1"},
					Role:    role,
				}, {
					Members: []string{"member2"},
					Role:    role,
				}},
			},
		},
		got: MakeSetIamPolicyRequest(addbindings, "add", "member2"),
	}, {
		name: "Remove iam policy binding",
		want: &iam.SetIamPolicyRequest{
			Policy: &iam.Policy{
				Bindings: []*iam.Binding{{
					Members: []string{"member1"},
					Role:    role,
				}},
			},
		},
		got: MakeSetIamPolicyRequest(removebindings, "remove", "member2"),
	}, {
		name: "invalid iam policy binding action",
		want: nil,
		got:  MakeSetIamPolicyRequest(removebindings, "plus", "member2"),
	}}

	for _, tc := range testCases {
		if diff := cmp.Diff(tc.want, tc.got); diff != "" {
			t.Errorf("unexpected (-want, +got) = %v", diff)
		}
	}
}
