package webhooks

import (
	"context"
	"encoding/json"
	derror "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	"github.com/xeipuuv/gojsonschema"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NOTE: RBAC not needed here.
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:webhook:path=/validate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=validate.openfeature.dev,admissionReviewVersions=v1,sideEffects=NoneOnDryRun

// PodValidator annotates Pods
type PodValidator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// PodValidator adds an annotation to every incoming pods.
func (m *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := m.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Check configuration
	val, ok := pod.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
	if !ok {
		m.Log.V(2).Info("FeatureFlagConfiguration not found, allowing admission")
		return admission.Allowed("FeatureFlagConfiguration not found")
	}
	// Check for ConfigMap and create it if it doesn't exist
	cm := corev1.ConfigMap{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: val, Namespace: req.Namespace}, &cm); errors.IsNotFound(err) {
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("configmap %s is missing and should have been created by mutate.openfeature.dev: %s", val, err.Error()))
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}
	data := cm.Data["config.json"]
	if data != "" {
		if !m.isJSON(data) {
			m.Log.V(1).Info("config.json is not valid JSON")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("config.json is not valid JSON"))
		} else {
			m.Log.V(1).Info("config.json is valid JSON")
			if err := validateJSONSchema(utils.GetSchema(), data); err != nil {
				m.Log.V(1).Info(fmt.Sprintf("config.json does not conform to Open Feature schema %s", err.Error()))
				return admission.Errored(http.StatusBadRequest, err)
			} else {
				m.Log.V(1).Info("config.json conforms to Open Feature schema")
			}
		}
	}

	return admission.Allowed("")
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
		return derror.New(sb.String())
	}
	return nil
}

// PodMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (m *PodValidator) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}
func (m *PodValidator) isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
