package common

import (
	"context"
	"fmt"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	appsV1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconcileErrorInterval            = 10 * time.Second
	ReconcileSuccessInterval          = 120 * time.Second
	FinalizerName                     = "featureflagconfiguration.core.openfeature.dev/finalizer"
	OpenFeatureAnnotationPath         = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	FlagSourceConfigurationAnnotation = "flagsourceconfiguration"
	OpenFeatureAnnotationRoot         = "openfeature.dev"
)

func FlagSourceConfigurationIndex(o client.Object) []string {
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
	if _, ok := deployment.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", FlagSourceConfigurationAnnotation)]; ok {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

func FindFlagConfig(ctx context.Context, c client.Client, namespace string, name string) (*v1alpha1.FeatureFlagConfiguration, error) {
	ffConfig := &v1alpha1.FeatureFlagConfiguration{}
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
