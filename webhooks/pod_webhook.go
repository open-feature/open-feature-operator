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
	featureFlagConfiguration "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	flagSourceConfiguration "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
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

// todo => we want to test for each ffconfig pre return in inject

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
	ready                     bool
}

func (m *PodMutator) IsReady(_ *http.Request) error {
	if m.ready {
		return nil
	}
	return goErr.New("pod mutator is not ready")
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
	fscNames := []string{}
	val, ok = pod.GetAnnotations()["openfeature.dev/flagsourceconfiguration"]
	if ok {
		fscNames = parseList(val)
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
	flagSourceConfigurationSpec, err := flagSourceConfiguration.NewFlagSourceConfigurationSpec()
	if err != nil {
		m.Log.V(1).Error(err, "unable to parse env var configuration", "webhook", "handle")
		return admission.Errored(http.StatusBadRequest, err)
	}

	for _, fscName := range fscNames {
		ns, name := parseAnnotation(fscName, req.Namespace)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to parse annotation %s error: %s", fscName, err.Error()))
			return admission.Errored(http.StatusBadRequest, err)
		}
		fc := m.getFlagSourceConfiguration(ctx, ns, name)
		if reflect.DeepEqual(fc, flagSourceConfiguration.FlagSourceConfiguration{}) {
			m.Log.V(1).Info(fmt.Sprintf("FlagSourceConfiguration could not be found for %s", fscName))
			return admission.Errored(http.StatusBadRequest, err)
		}
		flagSourceConfigurationSpec.Merge(&fc.Spec)
	}

	// maintain backwards compatibility of the openfeature.dev/featureflagconfiguration annotation
	ffConfigAnnotation, ffConfigAnnotationOk := pod.GetAnnotations()["openfeature.dev/featureflagconfiguration"]
	if ffConfigAnnotationOk {
		m.Log.V(1).Info("DEPRECATED: The openfeature.dev/featureflagconfiguration annotation has been superseded by the openfeature.dev/flagsourceconfiguration annotation. " +
			"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/annotations.md")
		if err := m.handleBackwardsCompatibility(ctx, flagSourceConfigurationSpec, ffConfigAnnotation, req.Namespace); err != nil {
			m.Log.Error(err, "unable to handle openfeature.dev/featureflagconfiguration annotation")
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	marshaledPod, err := m.injectSidecar(ctx, pod, flagSourceConfigurationSpec)
	if err != nil {
		m.Log.Error(err, "unable to inject flagd sidecar")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (m *PodMutator) handleBackwardsCompatibility(ctx context.Context, fcConfig *flagSourceConfiguration.FlagSourceConfigurationSpec, ffconfigAnnotation string, defaultNamespace string) error {
	for _, ffName := range parseList(ffconfigAnnotation) {
		ns, name := parseAnnotation(ffName, defaultNamespace)
		fsConfig := m.getFeatureFlag(ctx, ns, name)
		if reflect.DeepEqual(fsConfig, featureFlagConfiguration.FeatureFlagConfiguration{}) {
			return fmt.Errorf("FeatureFlagConfiguration %s not found", ffName)
		}
		if fsConfig.Spec.FlagDSpec != nil {
			if len(fsConfig.Spec.FlagDSpec.Envs) > 0 {
				fcConfig.EnvVars = append(fsConfig.Spec.FlagDSpec.Envs, fcConfig.EnvVars...)
			}
			if fsConfig.Spec.FlagDSpec.MetricsPort != 0 && fcConfig.MetricsPort == flagSourceConfiguration.DefaultMetricPort {
				fcConfig.MetricsPort = fsConfig.Spec.FlagDSpec.MetricsPort
			}
		}
		switch {
		case fsConfig.Spec.SyncProvider == nil:
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, flagSourceConfiguration.SyncProvider{
				Provider: fcConfig.DefaultSyncProvider,
				Source:   ffName,
			})
		case flagSourceConfiguration.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsKubernetes():
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, flagSourceConfiguration.SyncProvider{
				Provider: flagSourceConfiguration.SyncProviderKubernetes,
				Source:   ffName,
			})
		case flagSourceConfiguration.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsFilepath():
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, flagSourceConfiguration.SyncProvider{
				Provider: flagSourceConfiguration.SyncProviderFilepath,
				Source:   ffName,
			})
		case flagSourceConfiguration.SyncProviderType(fsConfig.Spec.SyncProvider.Name).IsHttp():
			if fsConfig.Spec.SyncProvider.HttpSyncConfiguration == nil {
				return fmt.Errorf("FeatureFlagConfiguration %s is missing HttpSyncConfiguration", ffName)
			}
			fcConfig.SyncProviders = append(fcConfig.SyncProviders, flagSourceConfiguration.SyncProvider{
				Provider:            flagSourceConfiguration.SyncProviderHttp,
				Source:              fsConfig.Spec.SyncProvider.HttpSyncConfiguration.Target,
				HttpSyncBearerToken: fsConfig.Spec.SyncProvider.HttpSyncConfiguration.BearerToken,
			})
		default:
			return fmt.Errorf("FeatureFlagConfiguration %s configuration is unrecognized", ffName)
		}
	}
	return nil
}

func (m *PodMutator) injectSidecar(
	ctx context.Context,
	pod *corev1.Pod,
	flagSourceConfig *flagSourceConfiguration.FlagSourceConfigurationSpec,
) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("Creating sidecar for pod %s/%s", pod.Namespace, pod.Name))
	sidecar := corev1.Container{
		Name:  "flagd",
		Image: fmt.Sprintf("%s:%s", flagSourceConfig.Image, flagSourceConfig.Tag),
		Args: []string{
			"start",
		},
		ImagePullPolicy: FlagDImagePullPolicy,
		VolumeMounts:    []corev1.VolumeMount{},
		Env:             []corev1.EnvVar{},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: flagSourceConfig.MetricsPort,
			},
		},
		SecurityContext: setSecurityContext(),
		Resources:       m.FlagDResourceRequirements,
	}

	for _, syncProvider := range flagSourceConfig.SyncProviders {
		switch {
		case syncProvider.Provider.IsFilepath():
			if err := m.handleFilepathProvider(ctx, pod, &sidecar, syncProvider); err != nil {
				return nil, err
			}
		case syncProvider.Provider.IsKubernetes():
			if err := m.handleKubernetesProvider(ctx, pod, &sidecar, syncProvider); err != nil {
				return nil, err
			}
		case syncProvider.Provider.IsHttp():
			m.handleHttpProvider(&sidecar, syncProvider)
		default:
			return nil, fmt.Errorf("unrecognized sync provider in config: %s", syncProvider.Provider)
		}
	}

	sidecar.Env = append(sidecar.Env, flagSourceConfig.ToEnvVars()...)
	for i := 0; i < len(pod.Spec.Containers); i++ {
		cntr := pod.Spec.Containers[i]
		cntr.Env = append(cntr.Env, sidecar.Env...)
	}

	// append sync provider args
	if flagSourceConfig.SyncProviderArgs != nil {
		for _, v := range flagSourceConfig.SyncProviderArgs {
			sidecar.Args = append(
				sidecar.Args,
				"--sync-provider-args",
				v,
			)
		}
	}

	pod.Spec.Containers = append(pod.Spec.Containers, sidecar)

	return json.Marshal(pod)
}

func (m *PodMutator) handleHttpProvider(sidecar *corev1.Container, syncProvider flagSourceConfiguration.SyncProvider) error {
	// append args
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		syncProvider.Source,
	)
	if syncProvider.HttpSyncBearerToken != "" {
		sidecar.Args = append(
			sidecar.Args,
			"--bearer-token",
			syncProvider.HttpSyncBearerToken,
		)
	}
	return nil
}

func (m *PodMutator) handleKubernetesProvider(ctx context.Context, pod *corev1.Pod, sidecar *corev1.Container, syncProvider flagSourceConfiguration.SyncProvider) error {
	// add permissions to pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return err
	}
	// append args
	ns, n := parseAnnotation(syncProvider.Source, pod.Namespace)
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		fmt.Sprintf(
			"core.openfeature.dev/%s/%s",
			ns,
			n,
		),
	)
	return nil
}

func (m *PodMutator) handleFilepathProvider(ctx context.Context, pod *corev1.Pod, sidecar *corev1.Container, syncProvider flagSourceConfiguration.SyncProvider) error {
	// create config map
	ns, n := parseAnnotation(syncProvider.Source, pod.Namespace)
	cm := corev1.ConfigMap{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: n, Namespace: ns}, &cm); errors.IsNotFound(err) {
		err := m.createConfigMap(ctx, ns, n, pod)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", n, err.Error()))
			return err
		}
	}

	// Add owner reference of the pod's owner
	if !podOwnerIsOwner(pod, cm) {
		reference := pod.OwnerReferences[0]
		reference.Controller = utils.FalseVal()
		cm.OwnerReferences = append(cm.OwnerReferences, reference)
		err := m.Client.Update(ctx, &cm)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", n, err.Error()))
		}
	}
	// mount configmap
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: n,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: n,
				},
			},
		},
	})
	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, featureFlagConfiguration.FeatureFlagConfigurationId(ns, n))
	sidecar.VolumeMounts = append(sidecar.VolumeMounts, corev1.VolumeMount{
		Name: n,
		// create a directory mount per featureFlag spec
		// file mounts will not work
		MountPath: mountPath,
	})
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		fmt.Sprintf("file:%s/%s",
			mountPath,
			featureFlagConfiguration.FeatureFlagConfigurationConfigMapKey(ns, n),
		),
	)
	return nil
}

// BackfillPermissions recovers the state of the flagd-kubernetes-sync role binding in the event of upgrade
func (m *PodMutator) BackfillPermissions(ctx context.Context) error {
	defer func() {
		m.ready = true
	}()
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

func (m *PodMutator) createConfigMap(ctx context.Context, namespace string, name string, pod *corev1.Pod) error {
	m.Log.V(1).Info(fmt.Sprintf("Creating configmap %s", name))
	references := []metav1.OwnerReference{
		pod.OwnerReferences[0],
	}
	references[0].Controller = utils.FalseVal()
	ff := m.getFeatureFlag(ctx, namespace, name)
	references = append(references, featureFlagConfiguration.GetFfReference(&ff))

	cm := featureFlagConfiguration.GenerateFfConfigMap(name, namespace, references, ff.Spec)

	return m.Client.Create(ctx, &cm)
}

func (m *PodMutator) getFeatureFlag(ctx context.Context, namespace string, name string) featureFlagConfiguration.FeatureFlagConfiguration {
	ffConfig := featureFlagConfiguration.FeatureFlagConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return featureFlagConfiguration.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func (m *PodMutator) getFlagSourceConfiguration(ctx context.Context, namespace string, name string) flagSourceConfiguration.FlagSourceConfiguration {
	fcConfig := flagSourceConfiguration.FlagSourceConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &fcConfig); errors.IsNotFound(err) {
		return flagSourceConfiguration.FlagSourceConfiguration{}
	}
	return fcConfig
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
