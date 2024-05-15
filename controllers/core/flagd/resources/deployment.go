package resources

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	"github.com/open-feature/open-feature-operator/controllers/core/flagd/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdDeployment struct {
	client.Client
	Log logr.Logger

	FlagdInjector flagdinjector.IFlagdContainerInjector
	FlagdConfig   resources.FlagdConfiguration
}

func (r *FlagdDeployment) AreObjectsEqual(o1 client.Object, o2 client.Object) bool {
	oldDeployment, ok := o1.(*appsv1.Deployment)
	if !ok {
		return false
	}

	newDeployment, ok := o2.(*appsv1.Deployment)
	if !ok {
		return false
	}

	return reflect.DeepEqual(oldDeployment.Spec, newDeployment.Spec)
}

func (r *FlagdDeployment) GetResource(ctx context.Context, flagd *api.Flagd) (client.Object, error) {
	labels := map[string]string{
		"app":                          flagd.Name,
		"app.kubernetes.io/name":       flagd.Name,
		"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
		"app.kubernetes.io/version":    r.FlagdConfig.Tag,
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flagd.Name,
			Namespace: flagd.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: flagd.APIVersion,
				Kind:       flagd.Kind,
				Name:       flagd.Name,
				UID:        flagd.UID,
			}},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: flagd.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": flagd.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: flagd.Spec.ServiceAccountName,
				},
			},
		},
	}

	featureFlagSource := &api.FeatureFlagSource{}

	if err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: flagd.Namespace,
		Name:      flagd.Spec.FeatureFlagSource,
	}, featureFlagSource); err != nil {
		return nil, fmt.Errorf("could not look up feature flag source for flagd: %w", err)
	}

	err := r.FlagdInjector.InjectFlagd(ctx, &deployment.ObjectMeta, &deployment.Spec.Template.Spec, &featureFlagSource.Spec)
	if err != nil {
		return nil, fmt.Errorf("could not inject flagd container into deployment: %w", err)
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil, errors.New("no flagd container has been injected into deployment")
	}

	deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
		{
			Name:          "management",
			ContainerPort: int32(r.FlagdConfig.ManagementPort),
		},
		{
			Name:          "flagd",
			ContainerPort: int32(r.FlagdConfig.FlagdPort),
		},
		{
			Name:          "ofrep",
			ContainerPort: int32(r.FlagdConfig.OFREPPort),
		},
		{
			Name:          "sync",
			ContainerPort: int32(r.FlagdConfig.SyncPort),
		},
	}

	return deployment, nil
}
