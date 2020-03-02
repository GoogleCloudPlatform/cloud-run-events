// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicy) DeepCopyInto(out *EventPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicy.
func (in *EventPolicy) DeepCopy() *EventPolicy {
	if in == nil {
		return nil
	}
	out := new(EventPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicyBinding) DeepCopyInto(out *EventPolicyBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicyBinding.
func (in *EventPolicyBinding) DeepCopy() *EventPolicyBinding {
	if in == nil {
		return nil
	}
	out := new(EventPolicyBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventPolicyBinding) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicyBindingList) DeepCopyInto(out *EventPolicyBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EventPolicyBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicyBindingList.
func (in *EventPolicyBindingList) DeepCopy() *EventPolicyBindingList {
	if in == nil {
		return nil
	}
	out := new(EventPolicyBindingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventPolicyBindingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicyList) DeepCopyInto(out *EventPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EventPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicyList.
func (in *EventPolicyList) DeepCopy() *EventPolicyList {
	if in == nil {
		return nil
	}
	out := new(EventPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicyRuleSpec) DeepCopyInto(out *EventPolicyRuleSpec) {
	*out = *in
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.DataSchema != nil {
		in, out := &in.DataSchema, &out.DataSchema
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Subject != nil {
		in, out := &in.Subject, &out.Subject
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.ContentType != nil {
		in, out := &in.ContentType, &out.ContentType
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.MediaType != nil {
		in, out := &in.MediaType, &out.MediaType
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make([]KeyValuesMatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicyRuleSpec.
func (in *EventPolicyRuleSpec) DeepCopy() *EventPolicyRuleSpec {
	if in == nil {
		return nil
	}
	out := new(EventPolicyRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventPolicySpec) DeepCopyInto(out *EventPolicySpec) {
	*out = *in
	if in.JWT != nil {
		in, out := &in.JWT, &out.JWT
		*out = new(JWTSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]EventPolicyRuleSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventPolicySpec.
func (in *EventPolicySpec) DeepCopy() *EventPolicySpec {
	if in == nil {
		return nil
	}
	out := new(EventPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicy) DeepCopyInto(out *HTTPPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicy.
func (in *HTTPPolicy) DeepCopy() *HTTPPolicy {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicyBinding) DeepCopyInto(out *HTTPPolicyBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicyBinding.
func (in *HTTPPolicyBinding) DeepCopy() *HTTPPolicyBinding {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicyBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPPolicyBinding) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicyBindingList) DeepCopyInto(out *HTTPPolicyBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HTTPPolicyBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicyBindingList.
func (in *HTTPPolicyBindingList) DeepCopy() *HTTPPolicyBindingList {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicyBindingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPPolicyBindingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicyList) DeepCopyInto(out *HTTPPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HTTPPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicyList.
func (in *HTTPPolicyList) DeepCopy() *HTTPPolicyList {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicyRuleSpec) DeepCopyInto(out *HTTPPolicyRuleSpec) {
	*out = *in
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = new(RequestAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.Operations != nil {
		in, out := &in.Operations, &out.Operations
		*out = make([]RequestOperation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]KeyValuesMatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicyRuleSpec.
func (in *HTTPPolicyRuleSpec) DeepCopy() *HTTPPolicyRuleSpec {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicyRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPolicySpec) DeepCopyInto(out *HTTPPolicySpec) {
	*out = *in
	if in.JWT != nil {
		in, out := &in.JWT, &out.JWT
		*out = new(JWTSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]HTTPPolicyRuleSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPolicySpec.
func (in *HTTPPolicySpec) DeepCopy() *HTTPPolicySpec {
	if in == nil {
		return nil
	}
	out := new(HTTPPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JWTSpec) DeepCopyInto(out *JWTSpec) {
	*out = *in
	if in.ExcludePaths != nil {
		in, out := &in.ExcludePaths, &out.ExcludePaths
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.IncludePaths != nil {
		in, out := &in.IncludePaths, &out.IncludePaths
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JWTSpec.
func (in *JWTSpec) DeepCopy() *JWTSpec {
	if in == nil {
		return nil
	}
	out := new(JWTSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KeyValuesMatch) DeepCopyInto(out *KeyValuesMatch) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KeyValuesMatch.
func (in *KeyValuesMatch) DeepCopy() *KeyValuesMatch {
	if in == nil {
		return nil
	}
	out := new(KeyValuesMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyBindingSpec) DeepCopyInto(out *PolicyBindingSpec) {
	*out = *in
	in.BindingSpec.DeepCopyInto(&out.BindingSpec)
	if in.Policy != nil {
		in, out := &in.Policy, &out.Policy
		*out = new(v1.ObjectReference)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyBindingSpec.
func (in *PolicyBindingSpec) DeepCopy() *PolicyBindingSpec {
	if in == nil {
		return nil
	}
	out := new(PolicyBindingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyBindingStatus) DeepCopyInto(out *PolicyBindingStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyBindingStatus.
func (in *PolicyBindingStatus) DeepCopy() *PolicyBindingStatus {
	if in == nil {
		return nil
	}
	out := new(PolicyBindingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RequestAuth) DeepCopyInto(out *RequestAuth) {
	*out = *in
	if in.Principals != nil {
		in, out := &in.Principals, &out.Principals
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Claims != nil {
		in, out := &in.Claims, &out.Claims
		*out = make([]KeyValuesMatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RequestAuth.
func (in *RequestAuth) DeepCopy() *RequestAuth {
	if in == nil {
		return nil
	}
	out := new(RequestAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RequestOperation) DeepCopyInto(out *RequestOperation) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Paths != nil {
		in, out := &in.Paths, &out.Paths
		*out = make([]StringMatch, len(*in))
		copy(*out, *in)
	}
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RequestOperation.
func (in *RequestOperation) DeepCopy() *RequestOperation {
	if in == nil {
		return nil
	}
	out := new(RequestOperation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StringMatch.
func (in *StringMatch) DeepCopy() *StringMatch {
	if in == nil {
		return nil
	}
	out := new(StringMatch)
	in.DeepCopyInto(out)
	return out
}
