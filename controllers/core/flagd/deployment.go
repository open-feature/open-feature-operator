package flagd

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdDeployment struct {
	client.Client
	Log logr.Logger

	FlagdInjector flagdinjector.IFlagdContainerInjector
	FlagdConfig   FlagdConfiguration

	ResourceReconciler *ResourceReconciler
}

func (r *FlagdDeployment) Reconcile(ctx context.Context, flagd *api.Flagd) error {
	return r.ResourceReconciler.Reconcile(
		ctx,
		flagd,
		&appsv1.Deployment{},
		func() (client.Object, error) {
			return r.getFlagdDeployment(ctx, flagd)
		},
		func(old client.Object, new client.Object) bool {
			return areDeploymentsEqual(old, new)
		},
	)
}

func (r *FlagdDeployment) getFlagdDeployment(ctx context.Context, flagd *api.Flagd) (*appsv1.Deployment, error) {
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
		Namespace: flagd.Spec.FeatureFlagSourceRef.Namespace,
		Name:      flagd.Spec.FeatureFlagSourceRef.Name,
	}, featureFlagSource); err != nil {
		return nil, fmt.Errorf("could not look up feature flag source for flagd: %v", err)
	}

	err := r.FlagdInjector.InjectFlagd(ctx, &deployment.ObjectMeta, &deployment.Spec.Template.Spec, &featureFlagSource.Spec)
	if err != nil {
		return nil, fmt.Errorf("could not inject flagd container into deployment: %v", err)
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

func areDeploymentsEqual(old client.Object, new client.Object) bool {
	oldDeployment, ok := old.(*appsv1.Deployment)
	if !ok {
		return false
	}

	newDeployment, ok := new.(*appsv1.Deployment)
	if !ok {
		return false
	}

	return reflect.DeepEqual(oldDeployment.Spec, newDeployment.Spec)
}
