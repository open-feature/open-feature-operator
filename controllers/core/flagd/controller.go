/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flagd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var operatorOwnerReference *metav1.OwnerReference

// FlagdReconciler reconciles a Flagd object
type FlagdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	FlagdInjector flagdinjector.IFlagdContainerInjector
	FlagdConfig   FlagdConfiguration

	operatorOwnerReference *metav1.OwnerReference
}

type FlagdConfiguration struct {
	Port           int
	ManagementPort int
	DebugLogging   bool
	Image          string
	Tag            string

	OperatorNamespace      string
	OperatorDeploymentName string
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/finalizers,verbs=update
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;services/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=featureflagsources/finalizers,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *FlagdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Searching for FeatureFlagSource")

	// Fetch the Flagd resource
	flagd := &api.Flagd{}
	if err := r.Client.Get(ctx, req.NamespacedName, flagd); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(fmt.Sprintf("Flagd resource '%s' not found. Ignoring since object must be deleted", req.NamespacedName))
			return ctrl.Result{Requeue: false}, nil
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get Flagd resource '%s'", req.NamespacedName))
		return ctrl.Result{}, err
	}

	result, err := r.reconcileFlagdDeployment(ctx, req, flagd)
	if err != nil || result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

func (r *FlagdReconciler) reconcileFlagdDeployment(ctx context.Context, req ctrl.Request, flagd *api.Flagd) (*ctrl.Result, error) {
	exists := false
	existingDeployment := &v1.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: flagd.Namespace,
		Name:      flagd.Name,
	}, existingDeployment)

	if err == nil {
		exists = true
	} else if err != nil && !errors.IsNotFound(err) {
		r.Log.Error(err, fmt.Sprintf("Failed to get Flagd deployment '%s'", req.NamespacedName))
		return &ctrl.Result{}, err
	}

	// check if the deployment is managed by the operator.
	// if not, do not continue to not mess with anything user generated
	if !common.IsManagedByOFO(existingDeployment) {
		r.Log.Info(fmt.Sprintf("Found existing deployment '%s' that is not managed by OFO. Will not proceed with deployment", req.NamespacedName))
		return &ctrl.Result{}, nil
	}

	newDeployment, err := r.getFlagdDeployment(ctx, flagd)

	if exists && !reflect.DeepEqual(existingDeployment, newDeployment) {
		if err := r.Client.Update(ctx, newDeployment); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to update Flagd deployment '%s'", req.NamespacedName))
			return &ctrl.Result{}, err
		}
	} else {
		if err := r.Client.Create(ctx, newDeployment); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create Flagd deployment '%s'", req.NamespacedName))
			return &ctrl.Result{}, err
		}
	}
	return nil, nil
}

func shouldUpdate(old *v1.Deployment, new *v1.Deployment) bool {
	return !reflect.DeepEqual(old, new)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlagdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Flagd{}).
		Complete(r)
}

func (r *FlagdReconciler) getFlagdDeployment(ctx context.Context, flagd *api.Flagd) (*v1.Deployment, error) {

	ownerRef, err := r.getOwnerReference(ctx)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{
		"app":                          flagd.Name,
		"app.kubernetes.io/name":       flagd.Name,
		"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
		"app.kubernetes.io/version":    r.FlagdConfig.Tag,
	}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            flagd.Name,
			Namespace:       flagd.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*ownerRef},
		},
		Spec: v1.DeploymentSpec{
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

	err = r.FlagdInjector.InjectFlagd(ctx, &deployment.ObjectMeta, &deployment.Spec.Template.Spec, &featureFlagSource.Spec)
	if err != nil {
		return nil, fmt.Errorf("could not inject flagd container into deployment: %v", err)
	}

	return deployment, nil
}

func (r *FlagdReconciler) getOwnerReference(ctx context.Context) (*metav1.OwnerReference, error) {
	if r.operatorOwnerReference != nil {
		return r.operatorOwnerReference, nil
	}

	operatorDeployment := &v1.Deployment{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: r.FlagdConfig.OperatorDeploymentName, Namespace: r.FlagdConfig.OperatorNamespace}, operatorDeployment); err != nil {
		return nil, fmt.Errorf("unable to fetch operator deployment: %w", err)
	}

	r.operatorOwnerReference = &metav1.OwnerReference{
		UID:        operatorDeployment.GetUID(),
		Name:       operatorDeployment.GetName(),
		APIVersion: operatorDeployment.APIVersion,
		Kind:       operatorDeployment.Kind,
	}

	return r.operatorOwnerReference, nil
}
