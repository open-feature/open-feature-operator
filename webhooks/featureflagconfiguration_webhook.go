package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NOTE: RBAC not needed here.
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/validate-v1alpha1-featureflagconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.openfeature.dev,resources=featureflagconfigurations,verbs=create;update,versions=v1alpha1,name=validate.featureflagconfiguration.openfeature.dev,admissionReviewVersions=v1

// FeatureFlagConfigurationValidator annotates Pods
type FeatureFlagConfigurationValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// FeatureFlagConfigurationValidator adds an annotation to every incoming pods.
func (m *FeatureFlagConfigurationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {

	config := corev1alpha1.FeatureFlagConfiguration{}
	err := m.decoder.Decode(req, &config)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if config.Spec.FeatureFlagSpec != "" {
		if !m.isJSON(config.Spec.FeatureFlagSpec) {
			return admission.Denied(fmt.Sprintf("FeatureFlagSpec is not valid JSON: %s", config.Spec.FeatureFlagSpec))
		}
		err = m.validateJSONSchema(utils.GetSchema(), config.Spec.FeatureFlagSpec)
		if err != nil {
			return admission.Denied(fmt.Sprintf("FeatureFlagSpec is not valid JSON: %s", err.Error()))
		}
	}
	return admission.Allowed("")
}

// FeatureFlagConfigurationValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (m *FeatureFlagConfigurationValidator) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}

func (m *FeatureFlagConfigurationValidator) isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func (m *FeatureFlagConfigurationValidator) validateJSONSchema(schemaJSON string, inputJSON string) error {

	schemaLoader := gojsonschema.NewBytesLoader([]byte(schemaJSON))
	valuesLoader := gojsonschema.NewBytesLoader([]byte(inputJSON))
	result, err := gojsonschema.Validate(schemaLoader, valuesLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var sb strings.Builder
		for _, desc := range result.Errors() {
			sb.WriteString(fmt.Sprintf("- %s\n", desc))
		}
		return errors.New(sb.String())
	}
	return nil
}
