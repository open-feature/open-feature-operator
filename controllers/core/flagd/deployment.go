package flagd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IFlagdDeployment interface {
	Reconcile(ctx context.Context, flagd *api.Flagd, owner metav1.OwnerReference) (*ctrl.Result, error)
}

type FlagdDeployment struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	FlagdInjector flagdinjector.IFlagdContainerInjector
	FlagdConfig   FlagdConfiguration
}

func (r *FlagdDeployment) Reconcile(ctx context.Context, flagd *api.Flagd, owner metav1.OwnerReference) (*ctrl.Result, error) {
	exists := false
	existingDeployment := &appsv1.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: flagd.Namespace,
		Name:      flagd.Name,
	}, existingDeployment)

	if err == nil {
		exists = true
	} else if err != nil && !errors.IsNotFound(err) {
		r.Log.Error(err, fmt.Sprintf("Failed to get Flagd deployment '%s/%s'", flagd.Namespace, flagd.Name))
		return &ctrl.Result{}, err
	}

	// check if the deployment is managed by the operator.
	// if not, do not continue to not mess with anything user generated
	if !common.IsManagedByOFO(existingDeployment) {
		r.Log.Info(fmt.Sprintf("Found existing deployment '%s/%s' that is not managed by OFO. Will not proceed with deployment", flagd.Namespace, flagd.Name))
		return &ctrl.Result{}, nil
	}

	newDeployment, err := r.getFlagdDeployment(ctx, flagd, owner)

	if exists && !reflect.DeepEqual(existingDeployment, newDeployment) {
		if err := r.Client.Update(ctx, newDeployment); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to update Flagd deployment '%s/%s'", flagd.Namespace, flagd.Name))
			return &ctrl.Result{}, err
		}
	} else {
		if err := r.Client.Create(ctx, newDeployment); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create Flagd deployment '%s/%s'", flagd.Namespace, flagd.Name))
			return &ctrl.Result{}, err
		}
	}
	return nil, nil
}

func (r *FlagdDeployment) getFlagdDeployment(ctx context.Context, flagd *api.Flagd, owner metav1.OwnerReference) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app":                          flagd.Name,
		"app.kubernetes.io/name":       flagd.Name,
		"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
		"app.kubernetes.io/version":    r.FlagdConfig.Tag,
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            flagd.Name,
			Namespace:       flagd.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{owner},
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

	return deployment, nil
}
