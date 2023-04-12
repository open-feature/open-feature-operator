package common

import (
	"fmt"
	"time"

	appsV1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CrdName                           = "FeatureFlagConfiguration"
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
