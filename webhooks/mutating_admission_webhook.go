package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	configv1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
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

	// Check configuration
	val, ok = pod.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
	if !ok {
		return admission.Allowed("FeatureFlagConfiguration not found")
	}

	// Check if the pod is static or orphaned
	if len(pod.GetOwnerReferences()) == 0 {
		return admission.Denied("static or orphaned pods cannot be mutated")
	}

	// Check for ConfigMap and create it if it doesn't exist
	cm := corev1.ConfigMap{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: val, Namespace: req.Namespace}, &cm); errors.IsNotFound(err) {
		err := m.CreateConfigMap(ctx, val, req.Namespace, pod)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", val, err.Error()))
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	if !CheckOwnerReference(pod, cm) {
		reference := pod.OwnerReferences[0]
		reference.Controller = m.falseVal()
		cm.OwnerReferences = append(cm.OwnerReferences, reference)
		err := m.Client.Update(ctx, &cm)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", val, err.Error()))
		}
	}

	marshaledPod, err := m.InjectSidecar(pod, val)
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

func CheckOwnerReference(pod *corev1.Pod, cm corev1.ConfigMap) bool {
	for _, cmOwner := range cm.OwnerReferences {
		for _, podOwner := range pod.OwnerReferences {
			if cmOwner.UID == podOwner.UID {
				return true
			}
		}
	}
	return false
}

func (m *PodMutator) CreateConfigMap(ctx context.Context, name string, namespace string, pod *corev1.Pod) error {
	m.Log.V(1).Info(fmt.Sprintf("Creating configmap %s", name))
	references := []metav1.OwnerReference{
		pod.OwnerReferences[0],
	}
	references[0].Controller = m.falseVal()
	ff := m.GetFeatureFlag(ctx, name, namespace)
	if ff.Name != "" {
		references = append(references, m.GetFfReference(&ff))
	}

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/featureflagconfiguration": name,
			},
			OwnerReferences: references,
		},
		Data: map[string]string{
			"config.yaml": ff.Spec.FeatureFlagSpec,
		},
	}
	return m.Client.Create(ctx, &cm)
}

func (m *PodMutator) GetFeatureFlag(ctx context.Context, name string, namespace string) configv1alpha1.FeatureFlagConfiguration {
	ffConfig := configv1alpha1.FeatureFlagConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return configv1alpha1.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func (m *PodMutator) InjectSidecar(pod *corev1.Pod, configMap string) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))
	// Inject the agent
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: "flagd-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMap,
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
	return json.Marshal(pod)
}

func (m *PodMutator) GetFfReference(ff *configv1alpha1.FeatureFlagConfiguration) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: ff.APIVersion,
		Kind:       ff.Kind,
		Name:       ff.Name,
		UID:        ff.UID,
		Controller: m.trueVal(),
	}
}

func (m *PodMutator) trueVal() *bool {
	b := true
	return &b
}

func (m *PodMutator) falseVal() *bool {
	b := false
	return &b
}
