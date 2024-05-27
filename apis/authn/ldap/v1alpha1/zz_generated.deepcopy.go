//go:build !ignore_autogenerated

/*
Copyright 2023 Krateo SRL.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LDAPConfig) DeepCopyInto(out *LDAPConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LDAPConfig.
func (in *LDAPConfig) DeepCopy() *LDAPConfig {
	if in == nil {
		return nil
	}
	out := new(LDAPConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LDAPConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LDAPConfigList) DeepCopyInto(out *LDAPConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LDAPConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LDAPConfigList.
func (in *LDAPConfigList) DeepCopy() *LDAPConfigList {
	if in == nil {
		return nil
	}
	out := new(LDAPConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LDAPConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LDAPConfigSpec) DeepCopyInto(out *LDAPConfigSpec) {
	*out = *in
	if in.BindDN != nil {
		in, out := &in.BindDN, &out.BindDN
		*out = new(string)
		**out = **in
	}
	if in.BindSecret != nil {
		in, out := &in.BindSecret, &out.BindSecret
		*out = (*in).DeepCopy()
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LDAPConfigSpec.
func (in *LDAPConfigSpec) DeepCopy() *LDAPConfigSpec {
	if in == nil {
		return nil
	}
	out := new(LDAPConfigSpec)
	in.DeepCopyInto(out)
	return out
}
