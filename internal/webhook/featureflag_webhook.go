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

package webhook

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	schema "github.com/open-feature/flagd-schemas/json"
	"github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type FeatureFlagCustomValidator struct{}

// log is for logging in this package.
var (
	featureFlagLog = logf.Log.WithName("featureflag-resource validator")
	compiledSchema *gojsonschema.Schema
	schemaInitOnce sync.Once
)

func (v *FeatureFlagCustomValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.FeatureFlag{}).
		WithValidator(&FeatureFlagCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-core-openfeature-dev-v1beta1-featureflag,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.openfeature.dev,resources=featureflags,verbs=create;update,versions=v1beta1,name=vfeatureflag.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &FeatureFlagCustomValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *FeatureFlagCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	featureFlag, ok := obj.(*v1beta1.FeatureFlag)

	if !ok {
		return nil, fmt.Errorf("expected a FeatureFlag object but got %T", obj)
	}

	featureFlagLog.Info("validate create", "name", featureFlag.Name)

	if err := validateFeatureFlagFlags(featureFlag.Spec.FlagSpec.Flags); err != nil {
		return []string{}, err
	}

	return []string{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *FeatureFlagCustomValidator) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	featureFlag, ok := newObj.(*v1beta1.FeatureFlag)

	if !ok {
		return nil, fmt.Errorf("expected a FeatureFlag object but got %T", newObj)
	}

	featureFlagLog.Info("validate update", "name", featureFlag.Name)

	if err := validateFeatureFlagFlags(featureFlag.Spec.FlagSpec.Flags); err != nil {
		return []string{}, err
	}

	return []string{}, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *FeatureFlagCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	featureFlag, ok := obj.(*v1beta1.FeatureFlag)

	if !ok {
		return nil, fmt.Errorf("expected a FeatureFlag object but got %T", obj)
	}

	featureFlagLog.Info("validate delete", "name", featureFlag.Name)

	return []string{}, nil
}

func validateFeatureFlagFlags(flags v1beta1.Flags) error {
	b, err := json.Marshal(flags)
	if err != nil {
		return err
	}

	documentLoader := gojsonschema.NewStringLoader(string(b))

	compiledSchema, err := initSchemas()
	if err != nil {
		return fmt.Errorf("unable to initialize Schema: %s", err.Error())
	}

	result, err := compiledSchema.Validate(documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		err = fmt.Errorf("")
		for _, desc := range result.Errors() {
			err = fmt.Errorf("%s", err.Error()+desc.Description()+"\n")
		}
	}

	return err
}

func initSchemas() (*gojsonschema.Schema, error) {
	var err error
	schemaInitOnce.Do(func() {
		schemaLoader := gojsonschema.NewSchemaLoader()
		err = schemaLoader.AddSchemas(gojsonschema.NewStringLoader(schema.TargetingSchema))
		if err == nil {
			compiledSchema, err = schemaLoader.Compile(gojsonschema.NewStringLoader(schema.FlagSchema))
		}
	})

	return compiledSchema, err
}
