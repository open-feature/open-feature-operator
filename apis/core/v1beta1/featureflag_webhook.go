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

package v1beta1

import (
	"encoding/json"
	"fmt"
	"os"

	_ "embed"

	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var featureflaglog = logf.Log.WithName("featureflag-resource")

const SchemaPath = "./../../flagd-schemas/json/"

func (ff *FeatureFlag) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(ff).
		Complete()
}

//+kubebuilder:webhook:path=/validate-core-openfeature-dev-v1beta1-featureflag,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.openfeature.dev,resources=featureflags,verbs=create;update,versions=v1beta1,name=vfeatureflag.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &FeatureFlag{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ff *FeatureFlag) ValidateCreate() error {
	featureflaglog.Info("validate create", "name", ff.Name)

	if err := validateFeatureFlagFlags(ff.Spec.FlagSpec.Flags); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ff *FeatureFlag) ValidateUpdate(old runtime.Object) error {
	featureflaglog.Info("validate update", "name", ff.Name)

	if err := validateFeatureFlagFlags(ff.Spec.FlagSpec.Flags); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ff *FeatureFlag) ValidateDelete() error {
	featureflaglog.Info("validate delete", "name", ff.Name)

	return nil
}

func validateFeatureFlagFlags(flags Flags) error {
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
			err = fmt.Errorf(err.Error() + desc.Description() + "\n")
		}
	}
	return err
}

func initSchemas() (*gojsonschema.Schema, error) {
	targetingSchema, err := os.ReadFile(SchemaPath + "targeting.json")
	if err != nil {
		return nil, err
	}
	flagsScheme, err := os.ReadFile(SchemaPath + "flags.json")
	if err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.AddSchemas(gojsonschema.NewStringLoader(string(targetingSchema)))
	compiledSchema, err := schemaLoader.Compile(gojsonschema.NewStringLoader(string(flagsScheme)))
	if err != nil {
		return nil, err
	}

	return compiledSchema, nil
}
