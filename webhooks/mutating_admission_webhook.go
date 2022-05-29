package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NOTE: RBAC not needed here.
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1,sideEffects=NoneOnDryRun

// PodMutator annotates Pods
type PodMutator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// PodMutator adds an annotation to every incoming pods.
func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {

	pod := &corev1.Pod{}
	err := m.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// Check enablement
	val, ok := pod.GetAnnotations()["openfeature.dev"]
	if ok {
		if val != "enabled" {
			m.Log.V(2).Info("openfeature.dev Annotation is not enabled")
			return admission.Allowed("openfeature is disabled")
		}
	}
	var featureFlagCustomResource corev1alpha1.FeatureFlagConfiguration
	// Check CustomResource
	val, ok = pod.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
	if !ok {
		return admission.Allowed("FeatureFlagConfiguration not found")
	} else {
		// Current limitation is to use the same namespace, this is easy to fix though
		// e.g. namespace/name check
		err = m.Client.Get(context.TODO(), client.ObjectKey{Name: val, Namespace: req.Namespace},
			&featureFlagCustomResource)
		if err != nil {
			return admission.Denied("FeatureFlagConfiguration not found")
		}
	}
	name := pod.Name
	if len(pod.GetOwnerReferences()) != 0 {
		name = pod.GetOwnerReferences()[0].Name
	}

	// TODO: this should be a short sha to avoid collisions
	configName := name
	// Create the agent configmap
	m.Client.Delete(context.TODO(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: req.Namespace,
		},
	}) // Delete the configmap if it exists
	m.Log.V(1).Info(fmt.Sprintf("Creating configmap %s", configName))
	if err := m.Client.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: req.Namespace,
			Annotations: map[string]string{
				"openfeature.dev/featureflagconfiguration": featureFlagCustomResource.Name,
			},
		},
		//TODO
		Data: map[string]string{
			"config.yaml": featureFlagCustomResource.Spec.FeatureFlagSpec,
		},
	}); err != nil {

		m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", configName, err.Error()))
		return admission.Errored(http.StatusInternalServerError, err)
	}

	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))
	// Inject the agent
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: "flagd-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configName,
				},
			},
		},
	})
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:  "flagd",
		Image: "ghcr.io/open-feature/flagd:main",
		Args: []string{
			"start", "-f", "/etc/flagd/config.yaml",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "flagd-config",
				MountPath: "/etc/flagd",
			},
		},
	})

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// PodMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (m *PodMutator) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}
