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

package v1alpha2

import (
	"fmt"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha2/common"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *FlagSourceConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha1.FlagSourceConfiguration)

	if !ok {
		return fmt.Errorf("type %T %w", dstRaw, common.ErrCannotCastFlagSourceConfiguration)
	}

	// Copy equal stuff to new object
	// DO NOT COPY TypeMeta
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = v1alpha1.FlagSourceConfigurationSpec{
		MetricsPort:         src.Spec.MetricsPort,
		Port:                src.Spec.Port,
		SocketPath:          src.Spec.SocketPath,
		Evaluator:           src.Spec.Evaluator,
		Image:               src.Spec.Image,
		Tag:                 src.Spec.Tag,
		Sources:             []v1alpha1.Source{},
		SyncProviderArgs:    src.Spec.SyncProviderArgs,
		LogFormat:           src.Spec.LogFormat,
		DefaultSyncProvider: v1alpha1.SyncProviderType(src.Spec.DefaultSyncProvider),
		ProbesEnabled:       src.Spec.ProbesEnabled,
		DebugLogging:        common.FalseVal(),
		OtelCollectorUri:    src.Spec.OtelCollectorUri,
	}
	return nil
}

func (dst *FlagSourceConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha1.FlagSourceConfiguration)

	if !ok {
		return fmt.Errorf("type %T %w", srcRaw, common.ErrCannotCastFlagSourceConfiguration)
	}

	// Copy equal stuff to new object
	// DO NOT COPY TypeMeta
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec = FlagSourceConfigurationSpec{
		MetricsPort:         src.Spec.MetricsPort,
		Port:                src.Spec.Port,
		SocketPath:          src.Spec.SocketPath,
		Evaluator:           src.Spec.Evaluator,
		Image:               src.Spec.Image,
		Tag:                 src.Spec.Tag,
		SyncProviderArgs:    src.Spec.SyncProviderArgs,
		LogFormat:           src.Spec.LogFormat,
		DefaultSyncProvider: string(src.Spec.DefaultSyncProvider),
		ProbesEnabled:       src.Spec.ProbesEnabled,
	}
	return nil
}
