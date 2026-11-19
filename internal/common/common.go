package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/internal/common/utils"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcileErrorInterval                                    = 10 * time.Second
	ReconcileSuccessInterval                                  = 120 * time.Second
	FinalizerName                                             = "featureflag.core.openfeature.dev/finalizer"
	OpenFeatureAnnotationPath                                 = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	OpenFeatureAnnotationRoot                                 = "openfeature.dev"
	FlagdImagePullPolicy                    corev1.PullPolicy = "Always"
	ClusterRoleBindingName                  string            = "open-feature-operator-flagd-kubernetes-sync"
	KubernetesSyncRoleName                  string            = "open-feature-operator-flagd-kubernetes-sync"
	KubernetesSyncRoleBindingName           string            = "open-feature-operator-flagd-kubernetes-sync"
	AllowKubernetesSyncAnnotation                             = "allowkubernetessync"
	KubernetesSyncScopeAnnotation                             = "kubernetessyncscope"
	KubernetesSyncScopeNamespace                              = "namespace"
	KubernetesSyncScopeCluster                                = "cluster"
	OpenFeatureAnnotationPrefix                               = "openfeature.dev"
	PodOpenFeatureAnnotationPath                              = "metadata.annotations.openfeature.dev"
	SourceConfigParam                                         = "--sources"
	ProbeReadiness                                            = "/readyz"
	ProbeLiveness                                             = "/healthz"
	ProbeInitialDelay                                         = 5
	FeatureFlagSourceAnnotation                               = "featureflagsource"
	ClusterFeatureFlagSourceAnnotation                        = "clusterfeatureflagsource"
	EnabledAnnotation                                         = "enabled"
	ManagedByAnnotationKey                                    = "app.kubernetes.io/managed-by"
	ManagedByAnnotationValue                                  = "open-feature-operator"
	OperatorDeploymentName                                    = "open-feature-operator-controller-manager"
	InProcessConfigurationAnnotation                          = "inprocessconfiguration"
	ClusterInProcessConfigurationAnnotation                   = "clusterinprocessconfiguration"
	FlagdGrpcService                                          = "flagd.evaluation.v1.Service"
	FlagdGrpcServicePath                                      = "/" + FlagdGrpcService
	SyncGrpcService                                           = "flagd.sync.v1.Service"
	SyncGrpcServicePath                                       = "/" + SyncGrpcService
	OFREPHttpServicePath                                      = "/ofrep"
)

var ErrFlagdProxyNotReady = errors.New("flagd-proxy is not ready, deferring pod admission")
var ErrUnrecognizedSyncProvider = errors.New("unrecognized sync provider")
var ErrCrossNamespaceSource = errors.New("kubernetes, file and flagd-proxy sources of a FeatureFlagSource must be in its own namespace; use a ClusterFeatureFlagSource to share flags across namespaces")
var ErrUnqualifiedClusterSource = errors.New("kubernetes, file and flagd-proxy sources of a ClusterFeatureFlagSource must use the form <namespace>/<name>")

// PinSourceProviders gives each source without a provider its own spec's default, so merging
// several specs cannot change which provider reads it.
func PinSourceProviders(spec *api.FeatureFlagSourceSpec, fallback apicommon.SyncProviderType) {
	provider := spec.DefaultSyncProvider
	if provider == "" {
		provider = fallback
	}
	for i := range spec.Sources {
		if spec.Sources[i].Provider == "" {
			spec.Sources[i].Provider = provider
		}
	}
}

// ValidateNamespacedSources rejects in-cluster sources outside namespace. Run it after PinSourceProviders.
func ValidateNamespacedSources(sources []api.Source, namespace string) error {
	for _, s := range sources {
		if !readsClusterObject(s.Provider) {
			continue
		}
		if ns, _ := utils.ParseAnnotation(s.Source, namespace); ns != namespace {
			return fmt.Errorf("%w: %q", ErrCrossNamespaceSource, s.Source)
		}
	}
	return nil
}

// ValidateClusterSources rejects in-cluster sources without an explicit namespace. Run it after PinSourceProviders.
func ValidateClusterSources(sources []api.Source) error {
	for _, s := range sources {
		if !readsClusterObject(s.Provider) {
			continue
		}
		ns, name, ok := strings.Cut(s.Source, "/")
		if !ok || ns == "" || name == "" || strings.Contains(name, "/") {
			return fmt.Errorf("%w: %q", ErrUnqualifiedClusterSource, s.Source)
		}
	}
	return nil
}

func readsClusterObject(p apicommon.SyncProviderType) bool {
	return p.IsKubernetes() || p.IsFilepath() || p.IsFlagdProxy()
}

func FeatureFlagSourceIndex(o client.Object) []string {
	deployment, ok := o.(*appsV1.Deployment)
	if !ok {
		return []string{
			"false",
		}
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		return []string{
			"false",
		}
	}
	if _, ok := deployment.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", FeatureFlagSourceAnnotation)]; ok {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

func ClusterFeatureFlagSourceIndex(o client.Object) []string {
	deployment, ok := o.(*appsV1.Deployment)
	if !ok {
		return []string{"false"}
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		return []string{"false"}
	}
	if _, ok := deployment.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", ClusterFeatureFlagSourceAnnotation)]; ok {
		return []string{"true"}
	}
	return []string{"false"}
}

func FindFlagConfig(ctx context.Context, c client.Client, namespace string, name string) (*api.FeatureFlag, error) {
	ffConfig := &api.FeatureFlag{}
	if err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, ffConfig); err != nil {
		return nil, err
	}
	return ffConfig, nil
}

// SharedOwnership returns true if any of the owner references match in the given slices
func SharedOwnership(ownerReferences1, ownerReferences2 []metav1.OwnerReference) bool {
	for _, owner1 := range ownerReferences1 {
		for _, owner2 := range ownerReferences2 {
			if owner1.UID == owner2.UID {
				return true
			}
		}
	}
	return false
}

func IsManagedByOFO(obj client.Object) bool {
	val, ok := obj.GetLabels()[ManagedByAnnotationKey]
	return ok && val == ManagedByAnnotationValue
}
