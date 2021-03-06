// +build !ignore_autogenerated

/*
Copyright 2021 Google LLC

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package gcpauth

import (
	v1 "k8s.io/api/core/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Defaults) DeepCopyInto(out *Defaults) {
	*out = *in
	if in.NamespaceDefaults != nil {
		in, out := &in.NamespaceDefaults, &out.NamespaceDefaults
		*out = make(map[string]ScopedDefaults, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	in.ClusterDefaults.DeepCopyInto(&out.ClusterDefaults)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Defaults.
func (in *Defaults) DeepCopy() *Defaults {
	if in == nil {
		return nil
	}
	out := new(Defaults)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopedDefaults) DeepCopyInto(out *ScopedDefaults) {
	*out = *in
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkloadIdentityMapping != nil {
		in, out := &in.WorkloadIdentityMapping, &out.WorkloadIdentityMapping
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopedDefaults.
func (in *ScopedDefaults) DeepCopy() *ScopedDefaults {
	if in == nil {
		return nil
	}
	out := new(ScopedDefaults)
	in.DeepCopyInto(out)
	return out
}
