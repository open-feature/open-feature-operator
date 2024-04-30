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

package v1beta2

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var featureflagsourcelog = logf.Log.WithName("featureflagsource-resource")

func (r *FeatureFlagSource) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-core-openfeature-dev-v1beta2-featureflagsource,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.openfeature.dev,resources=featureflagsources,verbs=create;update,versions=v1beta2,name=vfeatureflagsource.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &FeatureFlagSource{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureFlagSource) ValidateCreate() error {
	featureflagsourcelog.Info("validate create", "name", r.Name)
	return validateFFS(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureFlagSource) ValidateUpdate(old runtime.Object) error {
	featureflagsourcelog.Info("validate update", "name", r.Name)
	return validateFFS(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FeatureFlagSource) ValidateDelete() error {
	featureflagsourcelog.Info("validate delete", "name", r.Name)
	return nil
}

func validateFFS(ffs *FeatureFlagSource) error {
	if ffs.Spec.InProces != nil && ffs.Spec.RPC != nil {
		return fmt.Errorf("rpc and in-process evaluation cannot be set together")
	}
	return nil
}
