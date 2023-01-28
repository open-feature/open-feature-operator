package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	goErr "errors"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// we likely want these to be configurable, eventually
const (
	FlagDImagePullPolicy             corev1.PullPolicy = "Always"
	clusterRoleBindingName           string            = "open-feature-operator-flagd-kubernetes-sync"
	flagdMetricPortEnvVar            string            = "FLAGD_METRICS_PORT"
	rootFileSyncMountPath            string            = "/etc/flagd"
	OpenFeatureEnabledAnnotationPath                   = "metadata.annotations.openfeature.dev/enabled"
)

// NOTE: RBAC not needed here.

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.openfeature.dev,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*;

// PodMutator annotates Pods
type PodMutator struct {
	Client                    client.Client
	FlagDResourceRequirements corev1.ResourceRequirements
	decoder                   *admission.Decoder
	Log                       logr.Logger
}

// BackfillPermissions recovers the state of the flagd-kubernetes-sync role binding in the event of upgrade
func (m *PodMutator) BackfillPermissions(ctx context.Context) error {
	for i := 0; i < 5; i++ {
		// fetch all pods with the "openfeature.dev/enabled" annotation set to "true"
		podList := &corev1.PodList{}
		err := m.Client.List(ctx, podList, client.MatchingFields{OpenFeatureEnabledAnnotationPath: "true"})
		if err != nil {
			if !goErr.Is(err, &cache.ErrCacheNotStarted{}) {
				return err
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// add each new service account to the flagd-kubernetes-sync role binding
		for _, pod := range podList.Items {
			m.Log.V(1).Info(fmt.Sprintf("backfilling permissions for pod %s/%s", pod.Namespace, pod.Name))
			if err := m.enableClusterRoleBinding(ctx, &pod); err != nil {
				m.Log.Error(
					err,
					fmt.Sprintf("unable backfill permissions for pod %s/%s", pod.Namespace, pod.Name),
					"webhook",
					OpenFeatureEnabledAnnotationPath,
				)
			}
		}
		return nil
	}
	return goErr.New("unable to backfill permissions for the flagd-kubernetes-sync role binding: timeout")
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
	enabled := false
	val, ok := pod.GetAnnotations()["openfeature.dev"]
	if ok {
		m.Log.V(1).Info("DEPRECATED: The openfeature.dev annotation has been superseded by the openfeature.dev/enabled annotation. " +
			"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/annotations.md")
		if val == "enabled" {
			enabled = true
		}
	}
	val, ok = pod.GetAnnotations()["openfeature.dev/enabled"]
	if ok {
		if val == "true" {
			enabled = true
		}
	}

	if !enabled {
		m.Log.V(2).Info(`openfeature.dev/enabled annotation is not set to "true"`)
		return admission.Allowed("OpenFeature is disabled")
	}

	// Check configuration
	ffNames := []string{}
	val, ok = pod.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
	if ok {
		ffNames = parseList(val)
	}

	fcNames := []string{}
	val, ok = pod.GetAnnotations()["openfeature.dev/flagsourceconfiguration"]
	if ok {
		fcNames = parseList(val)
	}
	// Check if the pod is static or orphaned
	if len(pod.GetOwnerReferences()) == 0 {
		return admission.Denied("static or orphaned pods cannot be mutated")
	}

	// Check for the correct clusterrolebinding for the pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return admission.Denied(err.Error())
	}

	// merge any provided flagd specs
	flagSourceConfigurationSpec, err := corev1alpha1.NewFlagSourceConfigurationSpec()
	if err != nil {
		m.Log.V(1).Error(err, "unable to parse env var configuration", "webhook", "handle")
		return admission.Errored(http.StatusBadRequest, err)
	}

	for _, fcName := range fcNames {
		ns, name := parseAnnotation(fcName, req.Namespace)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to parse annotation %s error: %s", fcName, err.Error()))
			return admission.Errored(http.StatusBadRequest, err)
		}
		fc := m.getFlagSourceConfiguration(ctx, name, ns)
		if reflect.DeepEqual(fc, corev1alpha1.FlagSourceConfiguration{}) {
			m.Log.V(1).Info(fmt.Sprintf("FlagSourceConfiguration could not be found for %s", fcName))
			return admission.Errored(http.StatusBadRequest, err)
		}
		flagSourceConfigurationSpec.Merge(&fc.Spec)
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
		if reflect.DeepEqual(ff, corev1alpha1.FeatureFlagConfiguration{}) {
			m.Log.V(1).Info(fmt.Sprintf("FeatureFlagConfiguration could not be found for %s", ffName))
			return admission.Errored(http.StatusBadRequest, err)
		}
		if ff.Spec.SyncProvider == nil || ff.Spec.SyncProvider.Name == "" {
			ff.Spec.SyncProvider = &corev1alpha1.FeatureFlagSyncProvider{
				Name: flagSourceConfigurationSpec.DefaultSyncProvider,
			}
		}
		if !ff.Spec.SyncProvider.Name.IsKubernetes() {
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

	marshaledPod, err := m.injectSidecar(pod, flagSourceConfigurationSpec, ffConfigs)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func parseList(s string) []string {
	out := []string{}
	ss := strings.Split(s, ",")
	for i := 0; i < len(ss); i++ {
		out = append(out, strings.TrimSpace(ss[i]))
	}
	return out
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
	serviceAccount := client.ObjectKey{
		Name:      pod.Spec.ServiceAccountName,
		Namespace: pod.Namespace,
	}
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
	found := false
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

func (m *PodMutator) getFlagSourceConfiguration(ctx context.Context, name string, namespace string) corev1alpha1.FlagSourceConfiguration {
	fcConfig := corev1alpha1.FlagSourceConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &fcConfig); errors.IsNotFound(err) {
		return corev1alpha1.FlagSourceConfiguration{}
	}
	return fcConfig
}

func (m *PodMutator) injectSidecar(
	pod *corev1.Pod,
	flagdConfig *corev1alpha1.FlagSourceConfigurationSpec,
	featureFlags []*corev1alpha1.FeatureFlagConfiguration,
) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))

	commandSequence := []string{
		"start",
	}
	var envs []corev1.EnvVar
	var volumeMounts []corev1.VolumeMount

	for _, featureFlag := range featureFlags {
		if featureFlag.Spec.FlagDSpec != nil {
			m.Log.V(1).Info("DEPRECATED: The FlagDSpec property of the FeatureFlagConfiguration CRD has been superseded by " +
				"the FlagSourceConfiguration CRD." +
				"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/flagd_configuration.md")
			if featureFlag.Spec.FlagDSpec.MetricsPort != 0 && flagdConfig.MetricsPort == 8013 {
				flagdConfig.MetricsPort = featureFlag.Spec.FlagDSpec.MetricsPort
			}
			envs = append(envs, featureFlag.Spec.FlagDSpec.Envs...)
		}
		switch {
		// kubernetes sync is the default state
		case featureFlag.Spec.SyncProvider == nil || featureFlag.Spec.SyncProvider.Name.IsKubernetes():
			m.Log.V(1).Info(fmt.Sprintf("FeatureFlagConfiguration %s using kubernetes sync implementation", featureFlag.Name))
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
		case featureFlag.Spec.SyncProvider.Name.IsHttp():
			m.Log.V(1).Info(fmt.Sprintf("FeatureFlagConfiguration %s using http sync implementation", featureFlag.Name))
			if featureFlag.Spec.SyncProvider.HttpSyncConfiguration != nil {
				commandSequence = append(
					commandSequence,
					"--uri",
					featureFlag.Spec.SyncProvider.HttpSyncConfiguration.Target,
				)
				if featureFlag.Spec.SyncProvider.HttpSyncConfiguration.BearerToken != "" {
					commandSequence = append(
						commandSequence,
						"--bearer-token",
						featureFlag.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
					)
				}
			} else {
				err := fmt.Errorf("FeatureFlagConfiguration %s is missing a httpSyncConfiguration", featureFlag.Name)
				m.Log.V(1).Error(err, "unable to add http sync provider")
			}
			// if filepath is explicitly set
		case featureFlag.Spec.SyncProvider.Name.IsFilepath():
			m.Log.V(1).Info(fmt.Sprintf("FeatureFlagConfiguration %s using filepath sync implementation", featureFlag.Name))
			commandSequence = append(
				commandSequence,
				"--uri",
				fmt.Sprintf("file:%s/%s",
					fileSyncMountPath(featureFlag),
					corev1alpha1.FeatureFlagConfigurationConfigMapKey(featureFlag.Namespace, featureFlag.Name)),
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
				Name: featureFlag.Name,
				// create a directory mount per featureFlag spec
				// file mounts will not work
				MountPath: fileSyncMountPath(featureFlag),
			})
		default:
			err := fmt.Errorf(
				"sync provider for ffconfig %s not recognized: %s",
				featureFlag.Name,
				featureFlag.Spec.SyncProvider.Name,
			)
			m.Log.Error(err, err.Error())
		}
	}

	// append sync provider args
	if flagdConfig.SyncProviderArgs != nil {
		for _, v := range flagdConfig.SyncProviderArgs {
			commandSequence = append(
				commandSequence,
				"--sync-provider-args",
				v,
			)
		}
	}

	envs = append(envs, flagdConfig.ToEnvVars()...)
	for i := 0; i < len(pod.Spec.Containers); i++ {
		cntr := pod.Spec.Containers[i]
		cntr.Env = append(cntr.Env, envs...)
	}

	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:            "flagd",
		Image:           fmt.Sprintf("%s:%s", flagdConfig.Image, flagdConfig.Tag),
		Args:            commandSequence,
		ImagePullPolicy: FlagDImagePullPolicy,
		VolumeMounts:    volumeMounts,
		Env:             envs,
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: flagdConfig.MetricsPort,
			},
		},
		SecurityContext: setSecurityContext(),
		Resources:       m.FlagDResourceRequirements,
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

func fileSyncMountPath(featureFlag *corev1alpha1.FeatureFlagConfiguration) string {
	return fmt.Sprintf("%s/%s", rootFileSyncMountPath, corev1alpha1.FeatureFlagConfigurationId(featureFlag.Namespace, featureFlag.Name))
}

func OpenFeatureEnabledAnnotationIndex(o client.Object) []string {
	pod := o.(*corev1.Pod)
	if pod.ObjectMeta.Annotations == nil {
		return []string{
			"false",
		}
	}
	val, ok := pod.ObjectMeta.Annotations["openfeature.dev/enabled"]
	if ok && val == "true" {
		return []string{
			"true",
		}
	}
	val, ok = pod.ObjectMeta.Annotations["openfeature.dev"]
	if ok && val == "enabled" {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}
