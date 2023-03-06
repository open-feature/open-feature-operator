package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	schemas "github.com/open-feature/schemas/json"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	"github.com/xeipuuv/gojsonschema"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NOTE: RBAC not needed here.
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/validate-v1alpha3-featureflagconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.openfeature.dev,resources=featureflagconfigurations,verbs=create;update,versions=v1alpha3,name=validate.featureflagconfiguration.openfeature.dev,admissionReviewVersions=v1

// FeatureFlagConfigurationValidator annotates Pods
type FeatureFlagConfigurationValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// Handle validates a FeatureFlagConfiguration
func (m *FeatureFlagConfigurationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	config := api.FeatureFlagConfiguration{}
	err := m.decoder.Decode(req, &config)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if config.Spec.FeatureFlagSpec != "" {
		if !m.isJSON(config.Spec.FeatureFlagSpec) {
			return admission.Denied(fmt.Sprintf("FeatureFlagSpec is not valid JSON: %s", config.Spec.FeatureFlagSpec))
		}
		err = validateJSONSchema(schemas.FlagdDefinitions, config.Spec.FeatureFlagSpec)
		if err != nil {
			return admission.Denied(fmt.Sprintf("FeatureFlagSpec is not valid JSON: %s", err.Error()))
		}
	}

	if config.Spec.ServiceProvider != nil && config.Spec.ServiceProvider.Credentials != nil {
		// Check the provider and whether it has an existing secret
		providerKeySecret := corev1.Secret{}
		if err := m.Client.Get(ctx, client.ObjectKey{
			Name:      config.Spec.ServiceProvider.Credentials.Name,
			Namespace: config.Spec.ServiceProvider.Credentials.Namespace,
		}, &providerKeySecret); errors.IsNotFound(err) {
			return admission.Denied("credentials secret not found")
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

func validateJSONSchema(schemaJSON string, inputJSON string) error {
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
		return errors.NewBadRequest(sb.String())
	}
	return nil
}
