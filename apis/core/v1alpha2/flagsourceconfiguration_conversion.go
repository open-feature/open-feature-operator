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
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (ffc *FlagSourceConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(ffc).
		Complete()
}

func (src *FlagSourceConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.FlagSourceConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec = v1alpha1.FlagSourceConfigurationSpec{
		MetricsPort:   src.Spec.MetricsPort,
		Port:          src.Spec.Port,
		SocketPath:    src.Spec.SocketPath,
		Evaluator:     src.Spec.Evaluator,
		Image:         src.Spec.Image,
		Tag:           src.Spec.Tag,
		SyncProviders: []v1alpha1.SyncProvider{},
	}
	return nil
}

func (dst *FlagSourceConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.FlagSourceConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec = FlagSourceConfigurationSpec{
		MetricsPort: src.Spec.MetricsPort,
		Port:        src.Spec.Port,
		SocketPath:  src.Spec.SocketPath,
		Evaluator:   src.Spec.Evaluator,
		Image:       src.Spec.Image,
		Tag:         src.Spec.Tag,
	}
	return nil
}
