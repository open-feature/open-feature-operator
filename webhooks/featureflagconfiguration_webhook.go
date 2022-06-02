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

const (
	OFJsonSchema = `
	{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id": "https://openfeature.dev/flag.schema.json",
		"title": "OpenFeature Feature Flags",
		"type": "object",
		"patternProperties": {
		  "^[A-Za-z]+$": {
			"description": "The flag key that uniquely represents the flag.",
			"type": "object",
			"properties": {
			  "name": {
				"type": "string"
			  },
			  "description": {
				"type": "string"
			  },
			  "returnType": {
				"type": "string",
				"enum": ["boolean", "string", "number", "object"],
				"default": "boolean"
			  },
			  "variants": {
				"type": "object",
				"patternProperties": {
				  "^[A-Za-z]+$": {
					"properties": {
					  "value": {
						"type": ["string", "number", "boolean", "object"]
					  }
					}
				  },
				  "additionalProperties": false
				},
				"minProperties": 2,
				"default": { "enabled": true, "disabled": false }
			  },
			  "defaultVariant": {
				"type": "string",
				"default": "enabled"
			  },
			  "state": {
				"type": "string",
				"enum": ["enabled", "disabled"],
				"default": "enabled"
			  },
			  "rules": {
				"type": "array",
				"items": {
				  "$ref": "#/$defs/rule"
				},
				"default": []
			  }
			},
			"required": ["state"],
			"additionalProperties": false
		  }
		},
		"additionalProperties": false,
	  
		"$defs": {
		  "rule": {
			"type": "object",
			"description": "A rule that ",
			"properties": {
			  "action": {
				"description": "The action that should be taken if at least one condition evaluates to true.",
				"type": "object",
				"properties": {
				  "variant": {
					"type": "string",
					"description": "The variant that should be return if one of the conditions evaluates to true."
				  }
				},
				"required": ["variant"],
				"additionalProperties": false
			  },
			  "conditions": {
				"type": "array",
				"description": "The conditions that should that be evaluated.",
				"items": {
				  "type": "object",
				  "properties": {
					"context": {
					  "type": "string",
					  "description": "The context key that should be evaluated in this condition"
					},
					"op": {
					  "type": "string",
					  "description": "The operation that should be performed",
					  "enum": ["equals", "starts_with", "ends_with"]
					},
					"value": {
					  "type": "string",
					  "description": "The value that should be evaluated"
					}
				  },
				  "required": ["context", "op", "value"],
				  "additionalProperties": false
				}
			  }
			},
			"required": ["action", "conditions"],
			"additionalProperties": false
		  }
		}
	  }`
)

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
		err = m.validateJSONSchema(OFJsonSchema, config.Spec.FeatureFlagSpec)
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
