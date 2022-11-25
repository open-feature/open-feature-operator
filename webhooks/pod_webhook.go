package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	ffNames := strings.Split(val, ", ")

	// Check if the pod is static or orphaned
	if len(pod.GetOwnerReferences()) == 0 {
		return admission.Denied("static or orphaned pods cannot be mutated")
	}

	// Check for the correct clusterrolebinding for the pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return admission.Denied(err.Error())
	}

	ffConfigs := []*corev1alpha1.FeatureFlagConfiguration{}
	for _, ffName := range ffNames {
		ns, name := parseAnnotation(ffName, req.Namespace)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to parse annotation %s error: %s", ffName, err.Error()))
			return admission.Errored(http.StatusBadRequest, err)
		}
		// Check to see whether the FeatureFlagConfiguration has service or sync overrides
		ff := m.getFeatureFlag(ctx, name, ns)
		if ff.Spec.SyncProvider != nil && !ff.Spec.SyncProvider.IsKubernetes() {
			// Check for ConfigMap and create it if it doesn't exist (only required if sync provider isn't kubernetes)
			cm := corev1.ConfigMap{}
			if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: req.Namespace}, &cm); errors.IsNotFound(err) {
				err := m.createConfigMap(ctx, name, req.Namespace, pod)
				if err != nil {
					m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", ffName, err.Error()))
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
					m.Log.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", ffName, err.Error()))
				}
			}
		}
		ffConfigs = append(ffConfigs, &ff)
	}

	marshaledPod, err := m.injectSidecar(pod, ffConfigs)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func parseAnnotation(s string, defaultNs string) (string, string) {
	ss := strings.Split(s, "/")
	if len(ss) == 2 {
		return ss[0], ss[1]
	}
	return defaultNs, s
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

func (m *PodMutator) injectSidecar(pod *corev1.Pod, featureFlags []*corev1alpha1.FeatureFlagConfiguration) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))
	commandSequence := []string{
		"start",
	}
	var envs []corev1.EnvVar
	var volumeMounts []corev1.VolumeMount

	for _, featureFlag := range featureFlags {
		if featureFlag.Spec.FlagDSpec != nil {
			if featureFlag.Spec.FlagDSpec.MetricsPort != 0 {
				flagdMetricsPort = featureFlag.Spec.FlagDSpec.MetricsPort
			}
			envs = append(envs, featureFlag.Spec.FlagDSpec.Envs...)
		}
		switch {
		// kubernetes sync is the default state
		case featureFlag.Spec.SyncProvider != nil || featureFlag.Spec.SyncProvider.IsKubernetes():
			fmt.Printf("FeatureFlagConfiguration %s using kubernetes sync implementation\n", featureFlag.Name)
			commandSequence = append(
				commandSequence,
				"--uri",
				fmt.Sprintf(
					"core.openfeature.dev/%s/%s",
					featureFlag.ObjectMeta.Namespace,
					featureFlag.ObjectMeta.Name,
				),
			)
			// if http is explicitly set
		case featureFlag.Spec.SyncProvider.IsHttp():
			fmt.Printf("FeatureFlagConfiguration %s using http sync implementation\n", featureFlag.Name)
			if featureFlag.Spec.SyncProvider.HttpSyncConfiguration != nil {
				commandSequence = append(
					commandSequence,
					"--uri",
					featureFlag.Spec.SyncProvider.HttpSyncConfiguration.Target,
					"--bearer-token",
					featureFlag.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
				)
			} else {
				fmt.Printf("FeatureFlagConfiguration %s is missing a httpSyncConfiguration\n", featureFlag.Name)
			}
			// if filepath is explicitly set
		case featureFlag.Spec.SyncProvider.IsFilepath():
			fmt.Printf("FeatureFlagConfiguration %s using filepath sync implementation\n", featureFlag.Name)
			commandSequence = append(
				commandSequence,
				"--uri",
				fmt.Sprintf("file://etc/flagd/%s.json", featureFlag.Name),
			)
			pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
				Name: featureFlag.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: featureFlag.Name,
						},
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      featureFlag.Name,
				MountPath: "/etc/flagd/",
			})
		default:
			return nil, fmt.Errorf(
				"sync provider for ffconfig %s not recognized: %s",
				featureFlag.Name,
				featureFlag.Spec.SyncProvider.Name,
			)
		}
	}

	if os.Getenv("FLAGD_VERSION") != "" {
		FlagDTag = os.Getenv("FLAGD_VERSION")
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
		SecurityContext: setSecurityContext(),
	})
	return json.Marshal(pod)
}

func setSecurityContext() *corev1.SecurityContext {
	// user and group have been set to 65532 to mirror the configuration in the Dockerfile
	user := int64(65532)
	group := int64(65532)
	return &corev1.SecurityContext{
		// flagd does not require any additional capabilities, no bits set
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"all",
			},
		},
		RunAsUser:  &user,
		RunAsGroup: &group,
		Privileged: utils.FalseVal(),
		// Prevents misconfiguration from allowing access to resources on host
		RunAsNonRoot: utils.TrueVal(),
		// Prevent container gaining more privileges than its parent process
		AllowPrivilegeEscalation: utils.FalseVal(),
		ReadOnlyRootFilesystem:   utils.TrueVal(),
		// SeccompProfile defines the systems calls that can be made by the container
		SeccompProfile: &corev1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
}
