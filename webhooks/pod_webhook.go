package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// we likely want these to be configurable, eventually
const (
	FlagDImagePullPolicy   corev1.PullPolicy = "Always"
	clusterRoleBindingName string            = "open-feature-operator-flagd-kubernetes-sync"
	flagdMetricPortEnvVar  string            = "FLAGD_METRICS_PORT"
)

var FlagDTag = "main"
var flagdMetricsPort int32 = 8014

// NOTE: RBAC not needed here.

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.openfeature.dev,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*;

// PodMutator annotates Pods
type PodMutator struct {
	Client  client.Client
	decoder *admission.Decoder
	Log     logr.Logger
}

// Handle injects the flagd sidecar (if the prerequisites are all met)
func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	defer func() {
		if err := recover(); err != nil {
			admission.Errored(http.StatusInternalServerError, fmt.Errorf("%v", err))
		}
	}()

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

	// Check for the correct clusterrolebinding for the pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return admission.Denied(err.Error())
	}

	// Check to see whether the FeatureFlagConfiguration has service or sync overrides
	ff := m.getFeatureFlag(ctx, val, req.Namespace)

	if ff.Spec.SyncProvider != nil && !ff.Spec.SyncProvider.IsKubernetes() {
		// Check for ConfigMap and create it if it doesn't exist (only required if sync provider isn't kubernetes)
		cm := corev1.ConfigMap{}
		if err := m.Client.Get(ctx, client.ObjectKey{Name: val, Namespace: req.Namespace}, &cm); errors.IsNotFound(err) {
			err := m.createConfigMap(ctx, val, req.Namespace, pod)
			if err != nil {
				m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", val, err.Error()))
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}

		// Add owner reference of the pod's owner
		if !podOwnerIsOwner(pod, cm) {
			reference := pod.OwnerReferences[0]
			reference.Controller = utils.FalseVal()
			cm.OwnerReferences = append(cm.OwnerReferences, reference)
			err := m.Client.Update(ctx, &cm)
			if err != nil {
				m.Log.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", val, err.Error()))
			}
		}
	}

	marshaledPod, err := m.injectSidecar(pod, val, &ff)
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

func podOwnerIsOwner(pod *corev1.Pod, cm corev1.ConfigMap) bool {
	for _, cmOwner := range cm.OwnerReferences {
		for _, podOwner := range pod.OwnerReferences {
			if cmOwner.UID == podOwner.UID {
				return true
			}
		}
	}
	return false
}

func (m *PodMutator) enableClusterRoleBinding(ctx context.Context, pod *corev1.Pod) error {

	var serviceAccount = client.ObjectKey{Name: pod.Spec.ServiceAccountName,
		Namespace: pod.Namespace}
	if pod.Spec.ServiceAccountName == "" {
		serviceAccount.Name = "default"
	}
	// Check if the service account exists
	m.Log.V(1).Info(fmt.Sprintf("Fetching serviceAccount: %s/%s", pod.Namespace, pod.Spec.ServiceAccountName))
	sa := corev1.ServiceAccount{}
	if err := m.Client.Get(ctx, serviceAccount, &sa); err != nil {
		m.Log.V(1).Info(fmt.Sprintf("ServiceAccount not found: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
		return err
	}
	m.Log.V(1).Info(fmt.Sprintf("Fetching clusterrolebinding: %s", clusterRoleBindingName))
	// Fetch service account if it exists
	crb := v1.ClusterRoleBinding{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: clusterRoleBindingName}, &crb); errors.IsNotFound(err) {
		m.Log.V(1).Info(fmt.Sprintf("ClusterRoleBinding not found: %s", clusterRoleBindingName))
		return err
	}
	var found = false
	for _, subject := range crb.Subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == serviceAccount.Name && subject.Namespace == serviceAccount.Namespace {
			m.Log.V(1).Info(fmt.Sprintf("ClusterRoleBinding already exists for service account: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
			found = true
		}
	}
	if !found {
		m.Log.V(1).Info(fmt.Sprintf("Updating ClusterRoleBinding %s for service account: %s/%s", crb.Name,
			serviceAccount.Namespace, serviceAccount.Name))
		crb.Subjects = append(crb.Subjects, v1.Subject{
			Kind:      "ServiceAccount",
			Name:      serviceAccount.Name,
			Namespace: serviceAccount.Namespace,
		})
		if err := m.Client.Update(ctx, &crb); err != nil {
			m.Log.V(1).Info(fmt.Sprintf("Failed to update ClusterRoleBinding: %s", err.Error()))
			return err
		}
	}
	m.Log.V(1).Info(fmt.Sprintf("Updated ClusterRoleBinding: %s", crb.Name))

	return nil
}

func (m *PodMutator) createConfigMap(ctx context.Context, name string, namespace string, pod *corev1.Pod) error {
	m.Log.V(1).Info(fmt.Sprintf("Creating configmap %s", name))
	references := []metav1.OwnerReference{
		pod.OwnerReferences[0],
	}
	references[0].Controller = utils.FalseVal()
	ff := m.getFeatureFlag(ctx, name, namespace)
	if ff.Name != "" {
		references = append(references, corev1alpha1.GetFfReference(&ff))
	}

	cm := corev1alpha1.GenerateFfConfigMap(name, namespace, references, ff.Spec)

	return m.Client.Create(ctx, &cm)
}

func (m *PodMutator) getFeatureFlag(ctx context.Context, name string, namespace string) corev1alpha1.FeatureFlagConfiguration {
	ffConfig := corev1alpha1.FeatureFlagConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return corev1alpha1.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func (m *PodMutator) injectSidecar(pod *corev1.Pod, configMap string, featureFlag *corev1alpha1.FeatureFlagConfiguration) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))

	commandSequence := []string{
		"start", "--uri", "/etc/flagd/config.json", "--disable-request-logging",
	}

	// FlagD is the default provider name externally
	if featureFlag.Spec.ServiceProvider != nil && featureFlag.Spec.ServiceProvider.Name != "flagd" {
		commandSequence = append(commandSequence, "--service-provider")
		commandSequence = append(commandSequence, "http")
	}

	var volumeMounts []corev1.VolumeMount
	// Adds the sync provider if it is set
	if featureFlag.Spec.SyncProvider != nil && !featureFlag.Spec.SyncProvider.IsKubernetes() {
		commandSequence = append(commandSequence, "--sync-provider")
		commandSequence = append(commandSequence, featureFlag.Spec.SyncProvider.Name)
		// inject config map as volume if sync provider not kubernetes
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
		volumeMounts = []corev1.VolumeMount{
			{
				Name:      "flagd-config",
				MountPath: "/etc/flagd",
			},
		}
	} else {
		featureFlag.Spec.SyncProvider = &corev1alpha1.FeatureFlagSyncProvider{
			Name: "kubernetes",
		}
		commandSequence = append(commandSequence, "--sync-provider")
		commandSequence = append(commandSequence, "kubernetes")
		commandSequence = append(commandSequence, "--sync-provider-args=namespace="+featureFlag.ObjectMeta.Namespace)
		commandSequence = append(commandSequence, "--sync-provider-args=featureflagconfiguration="+featureFlag.ObjectMeta.Name)
	}

	if os.Getenv("FLAGD_VERSION") != "" {
		FlagDTag = os.Getenv("FLAGD_VERSION")
	}

	var envs []corev1.EnvVar
	if featureFlag.Spec.FlagDSpec != nil {
		if featureFlag.Spec.FlagDSpec.MetricsPort != 0 {
			flagdMetricsPort = featureFlag.Spec.FlagDSpec.MetricsPort
		}
		envs = featureFlag.Spec.FlagDSpec.Envs
	}

	envs = append(envs, corev1.EnvVar{
		Name:  flagdMetricPortEnvVar,
		Value: fmt.Sprintf("%d", flagdMetricsPort),
	})

	for i := 0; i < len(pod.Spec.Containers); i++ {
		cntr := pod.Spec.Containers[i]
		cntr.Env = append(cntr.Env, envs...)
		pod.Spec.Containers[i] = cntr
	}

	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:            "flagd",
		Image:           "ghcr.io/open-feature/flagd:" + FlagDTag,
		Args:            commandSequence,
		ImagePullPolicy: FlagDImagePullPolicy,
		VolumeMounts:    volumeMounts,
		Env:             envs,
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: flagdMetricsPort,
			},
		},
	})
	return json.Marshal(pod)
}
