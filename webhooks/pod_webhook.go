package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-feature/open-feature-operator/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"reflect"
	"strings"
	"time"

	goErr "errors"

	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
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
	FlagDImagePullPolicy               corev1.PullPolicy = "Always"
	clusterRoleBindingName             string            = "open-feature-operator-flagd-kubernetes-sync"
	rootFileSyncMountPath              string            = "/etc/flagd"
	OpenFeatureAnnotationPath                            = "metadata.annotations.openfeature.dev/openfeature.dev"
	OpenFeatureAnnotationPrefix                          = "openfeature.dev"
	AllowKubernetesSyncAnnotation                        = "allowkubernetessync"
	FlagSourceConfigurationAnnotation                    = "flagsourceconfiguration"
	FeatureFlagConfigurationAnnotation                   = "featureflagconfiguration"
	EnabledAnnotation                                    = "enabled"
	ProbeReadiness                                       = "/readyz"
	ProbeLiveness                                        = "/healthz"
	ProbeInitialDelay                                    = 5
	SourceConfigParam                                    = "--sources"
)

// NOTE: RBAC not needed here.

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.openfeature.dev,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;update;

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
	val, ok := pod.GetAnnotations()[OpenFeatureAnnotationPrefix]
	if ok {
		m.Log.V(1).Info("DEPRECATED: The openfeature.dev annotation has been superseded by the openfeature.dev/enabled annotation. " +
			"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/annotations.md")
		if val == "enabled" {
			enabled = true
		}
	}
	val, ok = pod.GetAnnotations()[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation)]
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
	val, ok = pod.GetAnnotations()[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FlagSourceConfigurationAnnotation)]
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
	flagSourceConfigurationSpec, err := v1alpha1.NewFlagSourceConfigurationSpec()
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
		if reflect.DeepEqual(fc, v1alpha1.FlagSourceConfiguration{}) {
			m.Log.V(1).Info(fmt.Sprintf("FlagSourceConfiguration could not be found for %s", fscName))
			return admission.Errored(http.StatusBadRequest, err)
		}
		flagSourceConfigurationSpec.Merge(&fc.Spec)
	}

	// maintain backwards compatibility of the openfeature.dev/featureflagconfiguration annotation
	ffConfigAnnotation, ffConfigAnnotationOk := pod.GetAnnotations()[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FeatureFlagConfigurationAnnotation)]
	if ffConfigAnnotationOk {
		m.Log.V(1).Info("DEPRECATED: The openfeature.dev/featureflagconfiguration annotation has been superseded by the openfeature.dev/flagsourceconfiguration annotation. " +
			"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/annotations.md")
		if err := m.handleFeatureFlagConfigurationAnnotation(ctx, flagSourceConfigurationSpec, ffConfigAnnotation, req.Namespace); err != nil {
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

func (m *PodMutator) injectSidecar(
	ctx context.Context,
	pod *corev1.Pod,
	flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec,
) ([]byte, error) {
	m.Log.V(1).Info(fmt.Sprintf("creating sidecar for pod %s/%s", pod.Namespace, pod.Name))
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

	// Enable probes
	if *flagSourceConfig.ProbesEnabled {
		sidecar.LivenessProbe = buildProbe(ProbeLiveness, int(flagSourceConfig.MetricsPort))
		sidecar.ReadinessProbe = buildProbe(ProbeReadiness, int(flagSourceConfig.MetricsPort))
	}

	sources, err := m.buildSources(ctx, flagSourceConfig, pod, &sidecar)
	if err != nil {
		return nil, err
	}

	err = m.appendSources(sources, &sidecar)
	if err != nil {
		return nil, err
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

// buildSources builds types.SourceConfig collection to be used by the sidecar.
func (m *PodMutator) buildSources(ctx context.Context, flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec,
	pod *corev1.Pod, sidecar *corev1.Container) ([]types.SourceConfig, error) {

	var sourceCfgCollection []types.SourceConfig

	for _, source := range flagSourceConfig.Sources {
		if source.Provider == "" {
			source.Provider = flagSourceConfig.DefaultSyncProvider
		}

		var sourceCfg types.SourceConfig
		var err error

		switch {
		case source.Provider.IsKubernetes():
			sourceCfg, err = m.toKubernetesConfig(ctx, pod, source)
			if err != nil {
				return []types.SourceConfig{}, err
			}
		case source.Provider.IsFilepath():
			sourceCfg, err = m.handleFilepathProvider(ctx, pod, sidecar, source)
			if err != nil {
				return []types.SourceConfig{}, err
			}
		case source.Provider.IsHttp():
			sourceCfg = m.toHttpProviderConfig(source)
		case source.Provider.IsGrpc():
			sourceCfg = m.toGrpcProviderConfig(source)
		default:
			return []types.SourceConfig{}, fmt.Errorf("unrecognized sync provider in config: %s", source.Provider)
		}

		sourceCfgCollection = append(sourceCfgCollection, sourceCfg)

	}

	return sourceCfgCollection, nil
}

// toHttpProviderConfig generate types.SourceConfig for http provider
func (m *PodMutator) toHttpProviderConfig(source v1alpha1.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:         source.Source,
		Provider:    string(v1alpha1.SyncProviderHttp),
		BearerToken: source.HttpSyncBearerToken,
	}
}

// toGrpcProviderConfig generate types.SourceConfig for grpc provider
func (m *PodMutator) toGrpcProviderConfig(source v1alpha1.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:        source.Source,
		Provider:   string(v1alpha1.SyncProviderGrpc),
		CertPath:   source.CertPath,
		ProviderID: source.ProviderID,
		Selector:   source.Selector,
	}
}

// toKubernetesConfig generate types.SourceConfig for K8s provider. Further, this updates underlying pod permissions &
// annotations
func (m *PodMutator) toKubernetesConfig(ctx context.Context,
	pod *corev1.Pod, source v1alpha1.Source) (types.SourceConfig, error) {

	ns, n := parseAnnotation(source.Source, pod.Namespace)

	// ensure that the FeatureFlagConfiguration exists
	ff := m.getFeatureFlag(ctx, ns, n)
	if ff.Name == "" {
		return types.SourceConfig{}, fmt.Errorf("feature flag configuration %s/%s not found", ns, n)
	}

	// add permissions to pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return types.SourceConfig{}, err
	}

	// mark pod with annotation (required to backfill permissions if they are dropped)
	pod.Annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation)] = "true"

	// build K8s config
	return types.SourceConfig{
		URI:      fmt.Sprintf("core.openfeature.dev/%s/%s", ns, n),
		Provider: string(v1alpha1.SyncProviderKubernetes),
	}, nil
}

func (m *PodMutator) appendSources(sources []types.SourceConfig, sidecar *corev1.Container) error {
	bytes, err := json.Marshal(sources)
	if err != nil {
		return err
	}

	sidecar.Args = append(sidecar.Args, SourceConfigParam, string(bytes))
	return nil
}

// handleFilepathProvider generate types.SourceConfig for file provider. Further, this creates config map & mount it
func (m *PodMutator) handleFilepathProvider(ctx context.Context,
	pod *corev1.Pod, sidecar *corev1.Container, source v1alpha1.Source) (types.SourceConfig, error) {

	// create config map
	ns, n := parseAnnotation(source.Source, pod.Namespace)
	cm := corev1.ConfigMap{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: n, Namespace: ns}, &cm); errors.IsNotFound(err) {
		err := m.createConfigMap(ctx, ns, n, pod)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", n, err.Error()))
			return types.SourceConfig{}, err
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

	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, v1alpha1.FeatureFlagConfigurationId(ns, n))
	sidecar.VolumeMounts = append(sidecar.VolumeMounts, corev1.VolumeMount{
		Name: n,
		// create a directory mount per featureFlag spec
		// file mounts will not work
		MountPath: mountPath,
	})

	return types.SourceConfig{
		URI: fmt.Sprintf("file:%s/%s",
			mountPath,
			v1alpha1.FeatureFlagConfigurationConfigMapKey(ns, n),
		),
		Provider: string(v1alpha1.SyncProviderFilepath),
	}, nil
}

// BackfillPermissions recovers the state of the flagd-kubernetes-sync role binding in the event of upgrade
func (m *PodMutator) BackfillPermissions(ctx context.Context) error {
	defer func() {
		m.ready = true
	}()
	for i := 0; i < 5; i++ {
		// fetch all pods with the fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation) annotation set to "true"
		podList := &corev1.PodList{}
		err := m.Client.List(ctx, podList, client.MatchingFields{
			fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, AllowKubernetesSyncAnnotation): "true",
		})
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
					fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, AllowKubernetesSyncAnnotation),
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
	if ff.Name == "" {
		return fmt.Errorf("feature flag configuration %s/%s not found", namespace, name)
	}
	references = append(references, v1alpha1.GetFfReference(&ff))

	cm := v1alpha1.GenerateFfConfigMap(name, namespace, references, ff.Spec)

	return m.Client.Create(ctx, &cm)
}

func (m *PodMutator) getFeatureFlag(ctx context.Context, namespace string, name string) v1alpha1.FeatureFlagConfiguration {
	ffConfig := v1alpha1.FeatureFlagConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return v1alpha1.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func (m *PodMutator) getFlagSourceConfiguration(ctx context.Context, namespace string, name string) v1alpha1.FlagSourceConfiguration {
	fcConfig := v1alpha1.FlagSourceConfiguration{}
	if err := m.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &fcConfig); errors.IsNotFound(err) {
		return v1alpha1.FlagSourceConfiguration{}
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
	val, ok := pod.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", AllowKubernetesSyncAnnotation)]
	if ok && val == "true" {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

// buildProbe generates a http corev1.Probe with provided endpoint, port and with ProbeInitialDelay
func buildProbe(path string, port int) *corev1.Probe {
	httpGetAction := &corev1.HTTPGetAction{
		Path:   path,
		Port:   intstr.FromInt(port),
		Scheme: corev1.URISchemeHTTP,
	}

	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: httpGetAction,
		},
		InitialDelaySeconds: ProbeInitialDelay,
	}
}
