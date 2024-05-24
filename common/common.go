package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcileErrorInterval                             = 10 * time.Second
	ReconcileSuccessInterval                           = 120 * time.Second
	FinalizerName                                      = "featureflag.core.openfeature.dev/finalizer"
	OpenFeatureAnnotationPath                          = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	OpenFeatureAnnotationRoot                          = "openfeature.dev"
	FlagdImagePullPolicy             corev1.PullPolicy = "Always"
	ClusterRoleBindingName           string            = "open-feature-operator-flagd-kubernetes-sync"
	AllowKubernetesSyncAnnotation                      = "allowkubernetessync"
	OpenFeatureAnnotationPrefix                        = "openfeature.dev"
	PodOpenFeatureAnnotationPath                       = "metadata.annotations.openfeature.dev"
	SourceConfigParam                                  = "--sources"
	ProbeReadiness                                     = "/readyz"
	ProbeLiveness                                      = "/healthz"
	ProbeInitialDelay                                  = 5
	FeatureFlagSourceAnnotation                        = "featureflagsource"
	EnabledAnnotation                                  = "enabled"
	ManagedByAnnotationKey                             = "app.kubernetes.io/managed-by"
	ManagedByAnnotationValue                           = "open-feature-operator"
	OperatorDeploymentName                             = "open-feature-operator-controller-manager"
	InProcessConfigurationAnnotation                   = "inprocessconfiguration"
)

var ErrFlagdProxyNotReady = errors.New("flagd-proxy is not ready, deferring pod admission")
var ErrUnrecognizedSyncProvider = errors.New("unrecognized sync provider")

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
