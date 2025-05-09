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
	"github.com/krateoplatformops/authn/apis/core"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfig) DeepCopyInto(out *OIDCConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfig.
func (in *OIDCConfig) DeepCopy() *OIDCConfig {
	if in == nil {
		return nil
	}
	out := new(OIDCConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigList) DeepCopyInto(out *OIDCConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OIDCConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigList.
func (in *OIDCConfigList) DeepCopy() *OIDCConfigList {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCConfigSpec) DeepCopyInto(out *OIDCConfigSpec) {
	*out = *in
	if in.ClientSecret != nil {
		in, out := &in.ClientSecret, &out.ClientSecret
		*out = (*in).DeepCopy()
	}
	if in.RESTActionRef != nil {
		in, out := &in.RESTActionRef, &out.RESTActionRef
		*out = new(core.ObjectRef)
		**out = **in
	}
	if in.Graphics != nil {
		in, out := &in.Graphics, &out.Graphics
		*out = new(core.Graphics)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCConfigSpec.
func (in *OIDCConfigSpec) DeepCopy() *OIDCConfigSpec {
	if in == nil {
		return nil
	}
	out := new(OIDCConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenResponse) DeepCopyInto(out *TokenResponse) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenResponse.
func (in *TokenResponse) DeepCopy() *TokenResponse {
	if in == nil {
		return nil
	}
	out := new(TokenResponse)
	in.DeepCopyInto(out)
	return out
}
