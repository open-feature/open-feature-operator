package common

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	"github.com/open-feature/open-feature-operator/pkg/types"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	rootFileSyncMountPath = "/etc/flagd"
)

type FlagdContainerInjector struct {
	Client                    client.Client
	Logger                    logr.Logger
	FlagdProxyConfig          *FlagdProxyConfiguration
	FlagDResourceRequirements corev1.ResourceRequirements
}

func (pi *FlagdContainerInjector) InjectFlagd(
	ctx context.Context,
	objectMeta *metav1.ObjectMeta,
	podSpec *corev1.PodSpec,
	flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec,
) error {
	pi.Logger.V(1).Info(fmt.Sprintf("creating flagdContainer for pod %s/%s", objectMeta.Namespace, objectMeta.Name))
	flagdContainer := pi.generateBasicFlagdContainer(flagSourceConfig)

	// Enable probes
	if flagSourceConfig.ProbesEnabled != nil && *flagSourceConfig.ProbesEnabled {
		flagdContainer.LivenessProbe = buildProbe(constant.ProbeLiveness, int(flagSourceConfig.MetricsPort))
		flagdContainer.ReadinessProbe = buildProbe(constant.ProbeReadiness, int(flagSourceConfig.MetricsPort))
	}

	if err := pi.handleSidecarSources(ctx, objectMeta, podSpec, flagSourceConfig, &flagdContainer); err != nil {
		return err
	}

	flagdContainer.Env = append(flagdContainer.Env, flagSourceConfig.ToEnvVars()...)
	for i := 0; i < len(podSpec.Containers); i++ {
		cntr := podSpec.Containers[i]
		cntr.Env = append(cntr.Env, flagdContainer.Env...)
	}

	// append sync provider args
	if flagSourceConfig.SyncProviderArgs != nil {
		for _, v := range flagSourceConfig.SyncProviderArgs {
			flagdContainer.Args = append(
				flagdContainer.Args,
				"--sync-provider-args",
				v,
			)
		}
	}

	// set --debug flag if enabled
	if flagSourceConfig.DebugLogging != nil && *flagSourceConfig.DebugLogging {
		flagdContainer.Args = append(
			flagdContainer.Args,
			"--debug",
		)
	}

	pi.addFlagdContainer(podSpec, flagdContainer)

	return nil
}

func (pi *FlagdContainerInjector) handleSidecarSources(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec, sidecar *corev1.Container) error {
	sources, err := pi.buildSources(ctx, objectMeta, flagSourceConfig, podSpec, sidecar)
	if err != nil {
		return err
	}

	err = pi.appendSources(sources, sidecar)
	if err != nil {
		return err
	}
	return nil
}

func (pi *FlagdContainerInjector) buildSources(ctx context.Context, objectMeta *metav1.ObjectMeta, flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec, podSpec *corev1.PodSpec, sidecar *corev1.Container) ([]types.SourceConfig, error) {
	var sourceCfgCollection []types.SourceConfig

	for _, source := range flagSourceConfig.Sources {
		if source.Provider == "" {
			source.Provider = flagSourceConfig.DefaultSyncProvider
		}

		var sourceCfg types.SourceConfig
		var err error

		switch {
		case source.Provider.IsKubernetes():
			sourceCfg, err = pi.toKubernetesProviderConfig(ctx, objectMeta, podSpec, source)
			if err != nil {
				return []types.SourceConfig{}, err
			}
		case source.Provider.IsFilepath():
			sourceCfg, err = pi.toFilepathProviderConfig(ctx, objectMeta, podSpec, sidecar, source)
			if err != nil {
				return []types.SourceConfig{}, err
			}
		case source.Provider.IsHttp():
			sourceCfg = pi.toHttpProviderConfig(source)
		case source.Provider.IsGrpc():
			sourceCfg = pi.toGrpcProviderConfig(source)
		case source.Provider.IsFlagdProxy():
			sourceCfg, err = pi.toFlagdProxyConfig(ctx, objectMeta, source)
			if err != nil {
				return []types.SourceConfig{}, err
			}
		default:
			return []types.SourceConfig{}, fmt.Errorf("unrecognized sync provider in config: %s", source.Provider)
		}

		sourceCfgCollection = append(sourceCfgCollection, sourceCfg)

	}

	return sourceCfgCollection, nil
}

//func HandleSourcesProviders(
//	ctx context.Context, log logr.Logger, c Client.Client, flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec, ns, serviceAccountNameSpace, serviceAccountName string,
//	ownerReferences []metav1.OwnerReference, podSpec *corev1.PodSpec, meta metav1.ObjectMeta, sidecar *corev1.Container,
//) error {
//	for _, source := range flagSourceConfig.Sources {
//		if source.Provider == "" {
//			source.Provider = flagSourceConfig.DefaultSyncProvider
//		}
//		switch {
//		case source.Provider.IsFilepath():
//			if err := toFilepathProviderConfig(ctx, log, c, ns, ownerReferences, podSpec, sidecar, source); err != nil {
//				return fmt.Errorf("toFilepathProviderConfig: %w", err)
//			}
//		case source.Provider.IsKubernetes():
//			if err := handleKubernetesProvider(ctx, log, c, ns, serviceAccountNameSpace, serviceAccountName, meta, sidecar, source); err != nil {
//				return fmt.Errorf("handleKubernetesProvider: %w", err)
//			}
//		case source.Provider.IsHttp():
//			handleHttpProvider(sidecar, source)
//		default:
//			return fmt.Errorf("unrecognized sync provider in config: %s", source.Provider)
//		}
//	}
//
//	return nil
//}

func (pi *FlagdContainerInjector) toFilepathProviderConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, sidecar *corev1.Container, source v1alpha1.Source) (types.SourceConfig, error) {
	// create config map
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)
	cm := corev1.ConfigMap{}
	if err := pi.Client.Get(ctx, client.ObjectKey{Name: n, Namespace: ns}, &cm); errors.IsNotFound(err) {
		err := CreateConfigMap(ctx, pi.Logger, pi.Client, ns, n, objectMeta.OwnerReferences)
		if err != nil {
			pi.Logger.V(1).Info(fmt.Sprintf("failed to create config map %s error: %s", n, err.Error()))
			return types.SourceConfig{}, err
		}
	}

	// Add owner reference of the pod's owner
	if !SharedOwnership(objectMeta.OwnerReferences, cm.OwnerReferences) {
		reference := objectMeta.OwnerReferences[0]
		reference.Controller = utils.FalseVal()
		cm.OwnerReferences = append(cm.OwnerReferences, reference)
		err := pi.Client.Update(ctx, &cm)
		if err != nil {
			pi.Logger.V(1).Info(fmt.Sprintf("failed to update owner reference for %s error: %s", n, err.Error()))
		}
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

	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, utils.FeatureFlagConfigurationId(ns, n))
	sidecar.VolumeMounts = append(sidecar.VolumeMounts, corev1.VolumeMount{
		Name: n,
		// create a directory mount per featureFlag spec
		// file mounts will not work
		MountPath: mountPath,
	})

	return types.SourceConfig{
		URI: fmt.Sprintf("%s/%s", mountPath, utils.FeatureFlagConfigurationConfigMapKey(ns, n)),
		// todo - this constant needs to be aligned with flagd. We have a mixed usage of file vs filepath
		Provider: "file",
	}, nil
}

func (pi *FlagdContainerInjector) toHttpProviderConfig(source v1alpha1.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:         source.Source,
		Provider:    string(v1alpha1.SyncProviderHttp),
		BearerToken: source.HttpSyncBearerToken,
	}
}

func (pi *FlagdContainerInjector) toGrpcProviderConfig(source v1alpha1.Source) types.SourceConfig {
	return types.SourceConfig{
		URI:        source.Source,
		Provider:   string(v1alpha1.SyncProviderGrpc),
		TLS:        source.TLS,
		CertPath:   source.CertPath,
		ProviderID: source.ProviderID,
		Selector:   source.Selector,
	}
}

func (pi *FlagdContainerInjector) toFlagdProxyConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, source v1alpha1.Source) (types.SourceConfig, error) {
	// does the proxy exist
	exists, ready, err := pi.isFlagdProxyReady(ctx)
	if err != nil {
		return types.SourceConfig{}, err
	}
	if !exists || (exists && !ready) {
		return types.SourceConfig{}, &types.FlagdProxyDeferError{}
	}
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)
	return types.SourceConfig{
		Provider: "grpc",
		Selector: fmt.Sprintf("core.openfeature.dev/%s/%s", ns, n),
		URI:      fmt.Sprintf("grpc://%s.%s.svc.cluster.local:%d", FlagdProxyServiceName, pi.FlagdProxyConfig.Namespace, pi.FlagdProxyConfig.Port),
	}, nil
}

func (pi *FlagdContainerInjector) isFlagdProxyReady(ctx context.Context) (bool, bool, error) {
	d := appsV1.Deployment{}
	err := pi.Client.Get(ctx, client.ObjectKey{Name: FlagdProxyDeploymentName, Namespace: pi.FlagdProxyConfig.Namespace}, &d)
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
				"flagd-proxy not ready after 3 minutes, was created at %s",
				d.CreationTimestamp.Time.String(),
			)
		}
		return true, false, nil
	}
	// exists, at least one replica ready, no error
	return true, true, nil
}

func (pi *FlagdContainerInjector) toKubernetesProviderConfig(ctx context.Context, objectMeta *metav1.ObjectMeta, podSpec *corev1.PodSpec, source v1alpha1.Source) (types.SourceConfig, error) {
	ns, n := utils.ParseAnnotation(source.Source, objectMeta.Namespace)

	// ensure that the FeatureFlagConfiguration exists
	ff := pi.getFeatureFlag(ctx, ns, n)
	if ff.Name == "" {
		return types.SourceConfig{}, fmt.Errorf("feature flag configuration %s/%s not found", ns, n)
	}

	// add permissions to pod
	if err := EnableClusterRoleBinding(ctx, pi.Logger, pi.Client, objectMeta.Namespace, podSpec.ServiceAccountName); err != nil {
		return types.SourceConfig{}, err
	}

	// mark pod with annotation (required to backfill permissions if they are dropped)
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = map[string]string{}
	}
	objectMeta.Annotations[fmt.Sprintf("%s/%s", constant.OpenFeatureAnnotationPrefix, constant.AllowKubernetesSyncAnnotation)] = "true"

	// build K8s config
	return types.SourceConfig{
		URI:      fmt.Sprintf("%s/%s", ns, n),
		Provider: string(v1alpha1.SyncProviderKubernetes),
	}, nil
}

func (pi *FlagdContainerInjector) getFeatureFlag(ctx context.Context, namespace string, name string) v1alpha1.FeatureFlagConfiguration {
	ffConfig := v1alpha1.FeatureFlagConfiguration{}
	// try to retrieve the FeatureFlagConfiguration
	if err := pi.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return v1alpha1.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func (pi *FlagdContainerInjector) appendSources(sources []types.SourceConfig, sidecar *corev1.Container) error {
	if len(sources) == 0 {
		return nil
	}

	bytes, err := json.Marshal(sources)
	if err != nil {
		return err
	}

	sidecar.Args = append(sidecar.Args, constant.SourceConfigParam, string(bytes))
	return nil
}

func (pi *FlagdContainerInjector) generateBasicFlagdContainer(flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec) corev1.Container {
	return corev1.Container{
		Name:  "flagd",
		Image: fmt.Sprintf("%s:%s", flagSourceConfig.Image, flagSourceConfig.Tag),
		Args: []string{
			"start",
		},
		ImagePullPolicy: constant.FlagDImagePullPolicy,
		VolumeMounts:    []corev1.VolumeMount{},
		Env:             []corev1.EnvVar{},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: flagSourceConfig.MetricsPort,
			},
		},
		SecurityContext: getSecurityContext(),
		Resources:       pi.FlagDResourceRequirements,
	}
}

func (pi *FlagdContainerInjector) addFlagdContainer(spec *corev1.PodSpec, flagdContainer corev1.Container) {
	for idx, container := range spec.Containers {
		if container.Name == flagdContainer.Name {
			spec.Containers[idx] = flagdContainer
			return
		}
	}
	spec.Containers = append(spec.Containers, flagdContainer)
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
		InitialDelaySeconds: constant.ProbeInitialDelay,
	}
}
