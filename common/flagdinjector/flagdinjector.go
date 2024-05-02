package flagdinjector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta2"
	apicommon "github.com/open-feature/open-feature-operator/apis/core/v1beta2/common"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/open-feature/open-feature-operator/common/utils"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rootFileSyncMountPath = "/etc/flagd"
)

type IFlagdContainerInjector interface {
	InjectFlagd(
		ctx context.Context,
		objectMeta *metav1.ObjectMeta,
		podSpec *corev1.PodSpec,
		flagSourceConfig *api.FeatureFlagSourceSpec,
	) error

	EnableClusterRoleBinding(
		ctx context.Context,
		namespace,
		serviceAccountName string,
	) error
}

type FlagdContainerInjector struct {
	Client                    client.Client
	Logger                    logr.Logger
	FlagdProxyConfig          *flagdproxy.FlagdProxyConfiguration
	FlagdResourceRequirements corev1.ResourceRequirements
	Image                     string
	Tag                       string
}

//nolint:gocyclo
func (fi *FlagdContainerInjector) InjectFlagd(
	ctx context.Context,
	objectMeta *metav1.ObjectMeta,
	podSpec *corev1.PodSpec,
	flagSourceConfig *api.FeatureFlagSourceSpec,
) error {
	fi.Logger.V(1).Info(fmt.Sprintf("creating flagdContainer for pod %s/%s", objectMeta.Namespace, objectMeta.Name))
	flagdContainer := fi.generateBasicFlagdContainer(flagSourceConfig)

	// Enable probes
	if flagSourceConfig.RPC.ProbesEnabled {
		flagdContainer.LivenessProbe = buildProbe(common.ProbeLiveness, int(flagSourceConfig.RPC.ManagementPort))
		flagdContainer.ReadinessProbe = buildProbe(common.ProbeReadiness, int(flagSourceConfig.RPC.ManagementPort))
	}

	if err := fi.handleSidecarSources(ctx, objectMeta, podSpec, flagSourceConfig, &flagdContainer); err != nil {
		return err
	}

	flagdContainer.Env = append(flagdContainer.Env, flagSourceConfig.ToEnvVarsRPC()...)
	for i := 0; i < len(podSpec.Containers); i++ {
		podSpec.Containers[i].Env = append(podSpec.Containers[i].Env, flagdContainer.Env...)
	}

	// append sync provider args
	if len(flagSourceConfig.RPC.SyncProviderArgs) > 0 {
		for _, v := range flagSourceConfig.RPC.SyncProviderArgs {
			flagdContainer.Args = append(
				flagdContainer.Args,
				"--sync-provider-args",
				v,
			)
		}
	}

	// set --debug flag if enabled
	if flagSourceConfig.RPC.DebugLogging {
		flagdContainer.Args = append(
			flagdContainer.Args,
			"--debug",
		)
	}

	// set --otel-collector-uri flag if enabled
	if flagSourceConfig.RPC.OtelCollectorUri != "" {
		flagdContainer.Args = append(
			flagdContainer.Args,
			"--metrics-exporter",
			"otel",
		)

		flagdContainer.Args = append(
			flagdContainer.Args,
			"--otel-collector-uri",
			flagSourceConfig.RPC.OtelCollectorUri,
		)
	}

	if len(flagSourceConfig.RPC.Resources.Requests) != 0 {
		flagdContainer.Resources.Requests = flagSourceConfig.RPC.Resources.Requests
	}

	if len(flagSourceConfig.RPC.Resources.Limits) != 0 {
		flagdContainer.Resources.Limits = flagSourceConfig.RPC.Resources.Limits
	}

	addFlagdContainer(podSpec, flagdContainer)

	return nil
}

// EnableClusterRoleBinding enables the open-feature-operator-flagd-kubernetes-sync cluster role binding for the given
// service account under the given namespace (required for kubernetes sync provider)
func (fi *FlagdContainerInjector) EnableClusterRoleBinding(ctx context.Context, namespace, serviceAccountName string) error {
	serviceAccount := client.ObjectKey{
		Name:      determineServiceAccountName(serviceAccountName),
		Namespace: namespace,
	}

	// Check if the service account exists
	fi.Logger.V(1).Info(fmt.Sprintf("Fetching serviceAccount: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
	sa := corev1.ServiceAccount{}
	if err := fi.Client.Get(ctx, serviceAccount, &sa); err != nil {
		fi.Logger.V(1).Info(fmt.Sprintf("ServiceAccount not found: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
		return err
	}

	fi.Logger.V(1).Info(fmt.Sprintf("Fetching clusterrolebinding: %s", common.ClusterRoleBindingName))
	// Fetch service account if it exists
	crb := rbacv1.ClusterRoleBinding{}
	if err := fi.Client.Get(ctx, client.ObjectKey{Name: common.ClusterRoleBindingName}, &crb); errors.IsNotFound(err) {
		fi.Logger.V(1).Info(fmt.Sprintf("ClusterRoleBinding not found: %s", common.ClusterRoleBindingName))
		return err
	}

	if !fi.isServiceAccountSet(&crb, serviceAccount) {
		return fi.updateServiceAccount(ctx, &crb, serviceAccount)
	}

	return nil
}

func determineServiceAccountName(name string) string {
	if name == "" {
		return "default"
	}
	return name
}

func (fi *FlagdContainerInjector) isServiceAccountSet(crb *rbacv1.ClusterRoleBinding, serviceAccount client.ObjectKey) bool {
	for _, subject := range crb.Subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == serviceAccount.Name && subject.Namespace == serviceAccount.Namespace {
			fi.Logger.V(1).Info(fmt.Sprintf("ClusterRoleBinding already exists for service account: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
			return true
		}
	}
	return false
}

func (fi *FlagdContainerInjector) updateServiceAccount(ctx context.Context, crb *rbacv1.ClusterRoleBinding, serviceAccount client.ObjectKey) error {
	fi.Logger.V(1).Info(fmt.Sprintf("Updating ClusterRoleBinding %s for service account: %s/%s", crb.Name,
		serviceAccount.Namespace, serviceAccount.Name))
	crb.Subjects = append(crb.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceAccount.Name,
		Namespace: serviceAccount.Namespace,
	})
	if err := fi.Client.Update(ctx, crb); err != nil {
		fi.Logger.V(1).Info(fmt.Sprintf("Failed to update ClusterRoleBinding: %s", err.Error()))
		return err
	}
	fi.Logger.V(1).Info(fmt.Sprintf("Updated ClusterRoleBinding: %s", crb.Name))
	return nil
}

func (fi *FlagdContainerInjector) handleSidecarSources(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, flagSourceConfig *api.FeatureFlagSourceSpec, sidecar *corev1.Container) error {
	sources, err := fi.buildSources(ctx, objectMeta, flagSourceConfig, podSpec, sidecar)
	if err != nil {
		return err
	}

	err = appendSources(sources, sidecar)
	if err != nil {
		return err
	}
	return nil
}

func (fi *FlagdContainerInjector) buildSources(ctx context.Context, objectMeta *metav1.ObjectMeta, flagSourceConfig *api.FeatureFlagSourceSpec, podSpec *corev1.PodSpec, sidecar *corev1.Container) ([]types.SourceConfig, error) {
	var sourceCfgCollection []types.SourceConfig

	for _, source := range flagSourceConfig.RPC.Sources {
		if source.Provider == "" {
			source.Provider = flagSourceConfig.RPC.DefaultSyncProvider
		}

		sourceCfg, err := fi.newSourceConfig(ctx, source, objectMeta, podSpec, sidecar)
		if err != nil {
			return []types.SourceConfig{}, err
		}

		sourceCfgCollection = append(sourceCfgCollection, *sourceCfg)

	}

	return sourceCfgCollection, nil
}

func (fi *FlagdContainerInjector) newSourceConfig(ctx context.Context, source api.Source, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, sidecar *corev1.Container) (*types.SourceConfig, error) {
	sourceCfg := types.SourceConfig{}
	var err error = nil

	switch {
	case source.Provider.IsKubernetes():
		sourceCfg, err = fi.toKubernetesProviderConfig(ctx, objectMeta, podSpec, source)
	case source.Provider.IsFilepath():
		sourceCfg, err = fi.toFilepathProviderConfig(ctx, objectMeta, podSpec, sidecar, source)
	case source.Provider.IsHttp():
		sourceCfg = fi.toHttpProviderConfig(source)
	case source.Provider.IsGrpc():
		sourceCfg = fi.toGrpcProviderConfig(source)
	case source.Provider.IsFlagdProxy():
		sourceCfg, err = fi.toFlagdProxyConfig(ctx, objectMeta, source)
	default:
		err = fmt.Errorf("could not add provider %s: %w", source.Provider, common.ErrUnrecognizedSyncProvider)
	}

	return &sourceCfg, err
}

func (fi *FlagdContainerInjector) toFilepathProviderConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, sidecar *corev1.Container, source api.Source) (types.SourceConfig, error) {
	// create config map
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)
	cm := corev1.ConfigMap{}
	if err := fi.Client.Get(ctx, client.ObjectKey{Name: n, Namespace: ns}, &cm); errors.IsNotFound(err) {
		err := fi.createConfigMap(ctx, ns, n, objectMeta.OwnerReferences)
		if err != nil {
			fi.Logger.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", n, err.Error()))
			return types.SourceConfig{}, err
		}
	}

	// Add owner reference of the pod's owner
	if !common.SharedOwnership(objectMeta.OwnerReferences, cm.OwnerReferences) {
		fi.updateCMOwnerReference(ctx, objectMeta, cm)
	}

	// mount configmap
	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: n,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: n,
				},
			},
		},
	})

	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, utils.FeatureFlagId(ns, n))
	sidecar.VolumeMounts = append(sidecar.VolumeMounts, corev1.VolumeMount{
		Name: n,
		// create a directory mount per featureFlag spec
		// file mounts will not work
		MountPath: mountPath,
	})

	return types.SourceConfig{
		URI:      fmt.Sprintf("%s/%s", mountPath, utils.FeatureFlagConfigMapKey(ns, n)),
		Provider: "file",
	}, nil
}

func (fi *FlagdContainerInjector) updateCMOwnerReference(ctx context.Context, objectMeta *metav1.ObjectMeta, cm corev1.ConfigMap) {
	if len(objectMeta.OwnerReferences) == 0 {
		return
	}
	reference := objectMeta.OwnerReferences[0]
	reference.Controller = utils.FalseVal()
	cm.OwnerReferences = append(cm.OwnerReferences, reference)
	err := fi.Client.Update(ctx, &cm)
	if err != nil {
		fi.Logger.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", cm.Name, err.Error()))
	}
}

func (fi *FlagdContainerInjector) toHttpProviderConfig(source api.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:         source.Source,
		Provider:    string(apicommon.SyncProviderHttp),
		BearerToken: source.HttpSyncBearerToken,
		Interval:    source.Interval,
	}
}

func (fi *FlagdContainerInjector) toGrpcProviderConfig(source api.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:        source.Source,
		Provider:   string(apicommon.SyncProviderGrpc),
		TLS:        source.TLS,
		CertPath:   source.CertPath,
		ProviderID: source.ProviderID,
		Selector:   source.Selector,
	}
}

func (fi *FlagdContainerInjector) toFlagdProxyConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, source api.Source) (types.SourceConfig, error) {
	// does the proxy exist
	exists, ready, err := fi.isFlagdProxyReady(ctx)
	if err != nil {
		return types.SourceConfig{}, err
	}
	if !exists || (exists && !ready) {
		return types.SourceConfig{}, common.ErrFlagdProxyNotReady
	}
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)
	return types.SourceConfig{
		Provider: "grpc",
		Selector: fmt.Sprintf("core.openfeature.dev/%s/%s", ns, n),
		URI:      fmt.Sprintf("%s.%s.svc.cluster.local:%d", flagdproxy.FlagdProxyServiceName, fi.FlagdProxyConfig.Namespace, fi.FlagdProxyConfig.Port),
	}, nil
}

func (fi *FlagdContainerInjector) isFlagdProxyReady(ctx context.Context) (bool, bool, error) {
	d := appsV1.Deployment{}
	err := fi.Client.Get(ctx, client.ObjectKey{Name: flagdproxy.FlagdProxyDeploymentName, Namespace: fi.FlagdProxyConfig.Namespace}, &d)
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
				"flagd-proxy not ready after 3 minutes, was created at %s: %w",
				d.CreationTimestamp.Time.String(),
				common.ErrFlagdProxyNotReady,
			)
		}
		return true, false, nil
	}
	// exists, at least one replica ready, no error
	return true, true, nil
}

func (fi *FlagdContainerInjector) toKubernetesProviderConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, source api.Source) (types.SourceConfig, error) {
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)

	// ensure that the FeatureFlag exists
	if _, err := common.FindFlagConfig(ctx, fi.Client, ns, n); err != nil {
		return types.SourceConfig{}, fmt.Errorf("could not retrieve featureflag %s/%s: %w", ns, n, err)
	}

	// add permissions to pod
	if err := fi.EnableClusterRoleBinding(ctx, objectMeta.Namespace, podSpec.ServiceAccountName); err != nil {
		return types.SourceConfig{}, err
	}

	// mark pod with annotation (required to backfill permissions if they are dropped)
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = map[string]string{}
	}
	objectMeta.Annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.AllowKubernetesSyncAnnotation)] = "true"

	// build K8s config
	return types.SourceConfig{
		URI:      fmt.Sprintf("%s/%s", ns, n),
		Provider: string(apicommon.SyncProviderKubernetes),
	}, nil
}

func (fi *FlagdContainerInjector) generateBasicFlagdContainer(flagSourceConfig *api.FeatureFlagSourceSpec) corev1.Container {
	return corev1.Container{
		Name:  "flagd",
		Image: fmt.Sprintf("%s:%s", fi.Image, fi.Tag),
		Args: []string{
			"start",
			"--management-port",
			fmt.Sprintf("%d", flagSourceConfig.RPC.ManagementPort),
		},
		ImagePullPolicy: common.FlagdImagePullPolicy,
		VolumeMounts:    []corev1.VolumeMount{},
		Env:             []corev1.EnvVar{},
		Ports: []corev1.ContainerPort{
			{
				Name:          "management",
				ContainerPort: flagSourceConfig.RPC.ManagementPort,
			},
		},
		SecurityContext: getSecurityContext(),
		Resources:       fi.FlagdResourceRequirements,
	}
}

func (fi *FlagdContainerInjector) createConfigMap(ctx context.Context, namespace, name string, ownerReferences []metav1.OwnerReference) error {
	fi.Logger.V(1).Info(fmt.Sprintf("Creating configmap %s", name))
	references := []metav1.OwnerReference{}
	if len(ownerReferences) > 0 {
		references = append(references, ownerReferences[0])
		references[0].Controller = utils.FalseVal()
	}
	ff, err := common.FindFlagConfig(ctx, fi.Client, namespace, name)
	if err != nil {
		return fmt.Errorf("could not retrieve featureflag %s/%s: %w", namespace, name, err)
	}

	references = append(references, ff.GetReference())

	cm, err := ff.GenerateConfigMap(name, namespace, references)
	if err != nil {
		return fmt.Errorf("could generate configmap for featureflag %s/%s: %w", namespace, name, err)
	}

	return fi.Client.Create(ctx, cm)
}

func addFlagdContainer(spec *corev1.PodSpec, flagdContainer corev1.Container) {
	for idx, container := range spec.Containers {
		if container.Name == flagdContainer.Name {
			spec.Containers[idx] = flagdContainer
			return
		}
	}
	spec.Containers = append(spec.Containers, flagdContainer)
}

func appendSources(sources []types.SourceConfig, sidecar *corev1.Container) error {
	if len(sources) == 0 {
		return nil
	}

	bytes, err := json.Marshal(sources)
	if err != nil {
		return err
	}

	sidecar.Args = append(sidecar.Args, common.SourceConfigParam, string(bytes))
	return nil
}

func getSecurityContext() *corev1.SecurityContext {
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
		InitialDelaySeconds: common.ProbeInitialDelay,
	}
}
