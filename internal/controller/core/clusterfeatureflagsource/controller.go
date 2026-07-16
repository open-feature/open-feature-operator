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

package clusterfeatureflagsource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/internal/common/utils"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ClusterFeatureFlagSourceReconciler reconciles a ClusterFeatureFlagSource object
type ClusterFeatureFlagSourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	FlagdProxy        *flagdproxy.FlagdProxyHandler
	FlagdProxyBackoff *utils.ExponentialBackoff
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clusterfeatureflagsources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clusterfeatureflagsources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clusterfeatureflagsources/finalizers,verbs=update

func (r *ClusterFeatureFlagSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Searching for ClusterFeatureFlagSource")

	cffsConfig := &api.ClusterFeatureFlagSource{}
	if err := r.Client.Get(ctx, req.NamespacedName, cffsConfig); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s resource not found. Ignoring since object must be deleted", req.NamespacedName))
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get the %s", req.NamespacedName))
		return r.finishReconcile(err, false)
	}

	needsFlagdProxy := false
	for _, source := range cffsConfig.Spec.Sources {
		if source.Provider.IsFlagdProxy() {
			r.Log.Info(fmt.Sprintf("clusterfeatureflagsource %s requires flagd-proxy", req.NamespacedName))
			needsFlagdProxy = true
		}
	}

	if needsFlagdProxy {
		if err := r.FlagdProxy.HandleFlagdProxy(ctx); err != nil {
			r.Log.Error(err, "error handling the flagd-proxy deployment")
			return reconcile.Result{RequeueAfter: r.FlagdProxyBackoff.Next()}, err
		}
		r.FlagdProxyBackoff.Reset()
	}

	if cffsConfig.Spec.RolloutOnChange == nil || !*cffsConfig.Spec.RolloutOnChange {
		return r.finishReconcile(nil, false)
	}

	err := r.handleDeploymentUpdate(ctx, cffsConfig.Name)
	return r.finishReconcile(err, false)
}

func (r *ClusterFeatureFlagSourceReconciler) handleDeploymentUpdate(ctx context.Context, cffsName string) error {
	// list all deployments cluster-wide that opt into a clusterfeatureflagsource
	deployList := &appsV1.DeploymentList{}
	if err := r.Client.List(ctx, deployList, client.MatchingFields{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.ClusterFeatureFlagSourceAnnotation): "true",
	}); err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get the deployments with annotation %s/%s", common.OpenFeatureAnnotationPath, common.ClusterFeatureFlagSourceAnnotation))
		return err
	}

	for _, deployment := range deployList.Items {
		annotation, ok := deployment.Spec.Template.Annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationRoot, common.ClusterFeatureFlagSourceAnnotation)]
		if !ok {
			continue
		}
		if isUsingConfiguration(cffsName, annotation) {
			r.Log.Info(fmt.Sprintf("restarting deployment %s/%s", deployment.Namespace, deployment.Name))
			deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			if err := r.Client.Update(ctx, &deployment); err != nil {
				r.Log.V(1).Error(err, fmt.Sprintf("Failed to update Deployment: %s/%s", deployment.Namespace, deployment.Name))
				continue
			}
		}
	}

	return nil
}

// isUsingConfiguration reports whether a comma-separated cluster annotation names cffsName.
// Cluster-scoped references are bare names (no namespace form).
func isUsingConfiguration(cffsName string, annotation string) bool {
	for _, target := range strings.Split(annotation, ",") {
		if strings.TrimSpace(target) == cffsName {
			return true
		}
	}
	return false
}

func (r *ClusterFeatureFlagSourceReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := common.ReconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling ClusterFeatureFlagSource with error: %w")
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling ClusterFeatureFlagSource")
	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterFeatureFlagSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.ClusterFeatureFlagSource{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
