//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

package v1alpha3

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagSourceConfiguration) DeepCopyInto(out *FlagSourceConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagSourceConfiguration.
func (in *FlagSourceConfiguration) DeepCopy() *FlagSourceConfiguration {
	if in == nil {
		return nil
	}
	out := new(FlagSourceConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlagSourceConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagSourceConfigurationList) DeepCopyInto(out *FlagSourceConfigurationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FlagSourceConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagSourceConfigurationList.
func (in *FlagSourceConfigurationList) DeepCopy() *FlagSourceConfigurationList {
	if in == nil {
		return nil
	}
	out := new(FlagSourceConfigurationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlagSourceConfigurationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagSourceConfigurationSpec) DeepCopyInto(out *FlagSourceConfigurationSpec) {
	*out = *in
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make([]Source, len(*in))
		copy(*out, *in)
	}
	if in.EnvVars != nil {
		in, out := &in.EnvVars, &out.EnvVars
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SyncProviderArgs != nil {
		in, out := &in.SyncProviderArgs, &out.SyncProviderArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RolloutOnChange != nil {
		in, out := &in.RolloutOnChange, &out.RolloutOnChange
		*out = new(bool)
		**out = **in
	}
	if in.ProbesEnabled != nil {
		in, out := &in.ProbesEnabled, &out.ProbesEnabled
		*out = new(bool)
		**out = **in
	}
	if in.DebugLogging != nil {
		in, out := &in.DebugLogging, &out.DebugLogging
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagSourceConfigurationSpec.
func (in *FlagSourceConfigurationSpec) DeepCopy() *FlagSourceConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(FlagSourceConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagSourceConfigurationStatus) DeepCopyInto(out *FlagSourceConfigurationStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagSourceConfigurationStatus.
func (in *FlagSourceConfigurationStatus) DeepCopy() *FlagSourceConfigurationStatus {
	if in == nil {
		return nil
	}
	out := new(FlagSourceConfigurationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Flagd) DeepCopyInto(out *Flagd) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Flagd.
func (in *Flagd) DeepCopy() *Flagd {
	if in == nil {
		return nil
	}
	out := new(Flagd)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Flagd) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagdList) DeepCopyInto(out *FlagdList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Flagd, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagdList.
func (in *FlagdList) DeepCopy() *FlagdList {
	if in == nil {
		return nil
	}
	out := new(FlagdList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlagdList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagdSpec) DeepCopyInto(out *FlagdSpec) {
	*out = *in
	in.DeploymentSpec.DeepCopyInto(&out.DeploymentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagdSpec.
func (in *FlagdSpec) DeepCopy() *FlagdSpec {
	if in == nil {
		return nil
	}
	out := new(FlagdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlagdStatus) DeepCopyInto(out *FlagdStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlagdStatus.
func (in *FlagdStatus) DeepCopy() *FlagdStatus {
	if in == nil {
		return nil
	}
	out := new(FlagdStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Source) DeepCopyInto(out *Source) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Source.
func (in *Source) DeepCopy() *Source {
	if in == nil {
		return nil
	}
	out := new(Source)
	in.DeepCopyInto(out)
	return out
}
