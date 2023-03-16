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
	"encoding/json"
	"fmt"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha2/common"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *FeatureFlagConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha1.FeatureFlagConfiguration)

	if !ok {
		return fmt.Errorf("type %T %w", dstRaw, common.ErrCannotCastFeatureFlagConfiguration)
	}

	// Copy equal stuff to new object
	// DO NOT COPY TypeMeta
	dst.ObjectMeta = src.ObjectMeta

	if src.Spec.ServiceProvider != nil {
		dst.Spec.ServiceProvider = &v1alpha1.FeatureFlagServiceProvider{
			Name:        src.Spec.ServiceProvider.Name,
			Credentials: src.Spec.ServiceProvider.Credentials,
		}
	}

	if src.Spec.SyncProvider != nil {
		dst.Spec.SyncProvider = &v1alpha1.FeatureFlagSyncProvider{Name: src.Spec.SyncProvider.Name}
		if src.Spec.SyncProvider.HttpSyncConfiguration != nil {
			dst.Spec.SyncProvider.HttpSyncConfiguration = &v1alpha1.HttpSyncConfiguration{
				Target:      src.Spec.SyncProvider.HttpSyncConfiguration.Target,
				BearerToken: src.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
			}
		}
	}

	if src.Spec.FlagDSpec != nil {
		dst.Spec.FlagDSpec = &v1alpha1.FlagDSpec{Envs: src.Spec.FlagDSpec.Envs}
	}

	featureFlagSpecB, err := json.Marshal(src.Spec.FeatureFlagSpec)
	if err != nil {
		return fmt.Errorf("featureflagspec: %w", err)
	}

	dst.Spec.FeatureFlagSpec = string(featureFlagSpecB)

	return nil
}

func (dst *FeatureFlagConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha1.FeatureFlagConfiguration)

	if !ok {
		return fmt.Errorf("type %T %w", srcRaw, common.ErrCannotCastFeatureFlagConfiguration)
	}

	// Copy equal stuff to new object
	// DO NOT COPY TypeMeta
	dst.ObjectMeta = src.ObjectMeta

	if src.Spec.ServiceProvider != nil {
		dst.Spec.ServiceProvider = &FeatureFlagServiceProvider{
			Name:        src.Spec.ServiceProvider.Name,
			Credentials: src.Spec.ServiceProvider.Credentials,
		}
	}

	if src.Spec.SyncProvider != nil {
		dst.Spec.SyncProvider = &FeatureFlagSyncProvider{
			Name: string(src.Spec.SyncProvider.Name),
		}
		if src.Spec.SyncProvider.HttpSyncConfiguration != nil {
			dst.Spec.SyncProvider.HttpSyncConfiguration = &HttpSyncConfiguration{
				Target:      src.Spec.SyncProvider.HttpSyncConfiguration.Target,
				BearerToken: src.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
			}
		}
	}

	if src.Spec.FlagDSpec != nil {
		dst.Spec.FlagDSpec = &FlagDSpec{Envs: src.Spec.FlagDSpec.Envs}
	}

	var featureFlagSpec FeatureFlagSpec
	if err := json.Unmarshal([]byte(src.Spec.FeatureFlagSpec), &featureFlagSpec); err != nil {
		return fmt.Errorf("featureflagspec: %w", err)
	}

	dst.Spec.FeatureFlagSpec = featureFlagSpec

	return nil
}
