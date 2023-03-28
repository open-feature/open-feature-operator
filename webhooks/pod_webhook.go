package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	goErr "errors"

	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	controllers "github.com/open-feature/open-feature-operator/controllers/core/flagsourceconfiguration"
	"github.com/open-feature/open-feature-operator/pkg/types"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	appsV1 "k8s.io/api/apps/v1"
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
	OpenFeatureAnnotationPath                            = "metadata.annotations.openfeature.dev"
	OpenFeatureAnnotationPrefix                          = "openfeature.dev"
	AllowKubernetesSyncAnnotation                        = "allowkubernetessync"
	FlagSourceConfigurationAnnotation                    = "flagsourceconfiguration"
	FeatureFlagConfigurationAnnotation                   = "featureflagconfiguration"
	EnabledAnnotation                                    = "enabled"
	ProbeReadiness                                       = "/readyz"
	ProbeLiveness                                        = "/healthz"
	ProbeInitialDelay                                    = 5
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
	KubeProxyConfig           *controllers.KubeProxyConfiguration
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
	annotations := pod.GetAnnotations()
	enabled := m.checkOFEnabled(annotations)

	if !enabled {
		m.Log.V(2).Info(`openfeature.dev/enabled annotation is not set to "true"`)
		return admission.Allowed("OpenFeature is disabled")
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
	flagSourceConfigurationSpec, code, err := m.createFSConfigSpec(ctx, req, annotations, pod)
	if err != nil {
		return admission.Errored(code, err)
	}

	marshaledPod, err := m.injectSidecar(ctx, pod, flagSourceConfigurationSpec)
	if err != nil {
		if goErr.Is(err, &kubeProxyDeferError{}) {
			return admission.Denied(err.Error())
		}
		m.Log.Error(err, "unable to inject flagd sidecar")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (m *PodMutator) createFSConfigSpec(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (*v1alpha1.FlagSourceConfigurationSpec, int32, error) {
	// Check configuration
	fscNames := []string{}
	val, ok := annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, FlagSourceConfigurationAnnotation)]
	if ok {
		fscNames = parseList(val)
	}

	flagSourceConfigurationSpec, err := v1alpha1.NewFlagSourceConfigurationSpec()
	if err != nil {
		m.Log.V(1).Error(err, "unable to parse env var configuration", "webhook", "handle")
		return nil, http.StatusBadRequest, err
	}

	for _, fscName := range fscNames {
		ns, name := parseAnnotation(fscName, req.Namespace)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("failed to parse annotation %s error: %s", fscName, err.Error()))
			return nil, http.StatusBadRequest, err
		}
		fc := m.getFlagSourceConfiguration(ctx, ns, name)
		if reflect.DeepEqual(fc, v1alpha1.FlagSourceConfiguration{}) {
			m.Log.V(1).Info(fmt.Sprintf("FlagSourceConfiguration could not be found for %s", fscName))
			return nil, http.StatusBadRequest, err
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
			return nil, http.StatusInternalServerError, err
		}
	}
	return flagSourceConfigurationSpec, 0, nil
}

func (m *PodMutator) checkOFEnabled(annotations map[string]string) bool {
	val, ok := annotations[OpenFeatureAnnotationPrefix]
	if ok {
		m.Log.V(1).Info("DEPRECATED: The openfeature.dev annotation has been superseded by the openfeature.dev/enabled annotation. " +
			"Docs: https://github.com/open-feature/open-feature-operator/blob/main/docs/annotations.md")
		if val == "enabled" {
			return true
		}
	}
	val, ok = annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, EnabledAnnotation)]
	if ok {
		if val == "true" {
			return true
		}
	}
	return false
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

	for _, source := range flagSourceConfig.Sources {
		if source.Provider == "" {
			source.Provider = flagSourceConfig.DefaultSyncProvider
		}
		switch {
		case source.Provider.IsFilepath():
			if err := m.handleFilepathProvider(ctx, pod, &sidecar, source); err != nil {
				return nil, err
			}
		case source.Provider.IsKubernetes():
			if err := m.handleKubernetesProvider(ctx, pod, &sidecar, source); err != nil {
				return nil, err
			}
		case source.Provider.IsHttp():
			m.handleHttpProvider(&sidecar, source)
		case source.Provider.IsKubeProxy():
			if err := m.handleKubeProxy(ctx, &sidecar, source); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unrecognized sync provider in config: %s", source.Provider)
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

func (m *PodMutator) isKubeProxyReady(ctx context.Context) (bool, bool, error) {
	m.Client.Scheme()
	d := appsV1.Deployment{}
	err := m.Client.Get(ctx, client.ObjectKey{Name: controllers.KubeProxyDeploymentName, Namespace: m.KubeProxyConfig.Namespace}, &d)
	if err != nil {
		if errors.IsNotFound(err) {
			// does not exist, is not ready, no error
			return false, false, nil
		}
		// does not exist, is not ready, is in error
		return false, false, err
	}
	if d.Status.ReadyReplicas == 0 {
		// exists, not ready, no error
		if d.CreationTimestamp.Time.Before(time.Now().Add(-3 * time.Minute)) {
			return true, false, fmt.Errorf(
				"flagd-kube-proxy not ready after 3 minutes, was created at %s",
				d.CreationTimestamp.Time.String(),
			)
		}
		return true, false, nil
	}
	// exists, at least one replica ready, no error
	return true, true, nil
}

func (m *PodMutator) handleKubeProxy(ctx context.Context, sidecar *corev1.Container, source v1alpha1.Source) error {
	// does the proxy exist
	exists, ready, err := m.isKubeProxyReady(ctx)
	if err != nil {
		return err
	}
	if !exists || (exists && !ready) {
		return &kubeProxyDeferError{}
	}

	config := []types.SourceConfig{
		{
			Provider: "grpc",
			Selector: fmt.Sprintf("core.openfeature.dev/%s", source.Source),
			URI:      fmt.Sprintf("grpc://%s.%s.svc.cluster.local:%d", controllers.KubeProxyServiceName, m.KubeProxyConfig.Namespace, m.KubeProxyConfig.Port),
		},
	}
	configB, err := json.Marshal(config)
	if err != nil {
		return err
	}

	sidecar.Args = append(
		sidecar.Args,
		"--sources",
		string(configB),
		"--debug",
	)
	return nil
}

func (m *PodMutator) handleHttpProvider(sidecar *corev1.Container, source v1alpha1.Source) {
	// append args
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		source.Source,
	)
	if source.HttpSyncBearerToken != "" {
		sidecar.Args = append(
			sidecar.Args,
			"--bearer-token",
			source.HttpSyncBearerToken,
		)
	}
}

func (m *PodMutator) handleKubernetesProvider(ctx context.Context, pod *corev1.Pod, sidecar *corev1.Container, source v1alpha1.Source) error {
	ns, n := parseAnnotation(source.Source, pod.Namespace)
	// ensure that the FeatureFlagConfiguration exists
	ff := m.getFeatureFlag(ctx, ns, n)
	if ff.Name == "" {
		return fmt.Errorf("feature flag configuration %s/%s not found", ns, n)
	}
	// add permissions to pod
	if err := m.enableClusterRoleBinding(ctx, pod); err != nil {
		return err
	}
	// mark pod with annotation (required to backfill permissions if they are dropped)
	pod.Annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation)] = "true"
	// append args
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

func (m *PodMutator) handleFilepathProvider(ctx context.Context, pod *corev1.Pod, sidecar *corev1.Container, source v1alpha1.Source) error {
	// create config map
	ns, n := parseAnnotation(source.Source, pod.Namespace)
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
	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, v1alpha1.FeatureFlagConfigurationId(ns, n))
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
			v1alpha1.FeatureFlagConfigurationConfigMapKey(ns, n),
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
		newS := strings.TrimSpace(ss[i])
		if newS != "" { //function should not add empty values
			out = append(out, newS)
		}
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
