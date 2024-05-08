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
	resources2 "github.com/open-feature/open-feature-operator/controllers/core/flagd/common"
	"github.com/open-feature/open-feature-operator/controllers/core/flagd/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FlagdReconciler reconciles a Flagd object
type FlagdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	FlagdConfig resources2.FlagdConfiguration

	ResourceReconciler IFlagdResourceReconciler

	FlagdDeployment resources.IFlagdResource
	FlagdService    resources.IFlagdResource
	FlagdIngress    resources.IFlagdResource
}

type IFlagdResourceReconciler interface {
	Reconcile(ctx context.Context, flagd *api.Flagd, obj client.Object, resource resources.IFlagdResource) error
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
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

	if err := r.ResourceReconciler.Reconcile(
		ctx,
		flagd,
		&appsv1.Deployment{},
		r.FlagdDeployment,
	); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ResourceReconciler.Reconcile(
		ctx,
		flagd,
		&v1.Service{},
		r.FlagdService,
	); err != nil {
		return ctrl.Result{}, err
	}

	if flagd.Spec.Ingress.Enabled {
		if err := r.ResourceReconciler.Reconcile(
			ctx,
			flagd,
			&networkingv1.Ingress{},
			r.FlagdIngress,
		); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlagdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Flagd{}).
		Complete(r)
}
