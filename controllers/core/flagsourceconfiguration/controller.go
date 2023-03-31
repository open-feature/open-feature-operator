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

package flagsourceconfiguration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/open-feature/open-feature-operator/controllers/common"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
)

const (
	FlagdProxyDeploymentName     = "flagd-proxy"
	FlagdProxyServiceAccountName = "open-feature-operator-flagd-proxy"
	FlagdProxyServiceName        = "flagd-proxy-svc"

	envVarPodNamespace            = "POD_NAMESPACE"
	envVarProxyImage              = "KUBE_PROXY_IMAGE"
	envVarProxyTag                = "KUBE_PROXY_TAG"
	envVarProxyPort               = "KUBE_PROXY_PORT"
	envVarProxyMetricsPort        = "KUBE_PROXY_METRICS_PORT"
	envVarProxyDebugLogging       = "KUBE_PROXY_DEBUG_LOGGING"
	defaultFlagdProxyImage        = "ghcr.io/open-feature/flagd-proxy"
	defaultFlagdProxyTag          = "v0.1.2" //KUBE_PROXY_TAG_RENOVATE
	defaultFlagdProxyPort         = 8015
	defaultFlagdProxyMetricsPort  = 8016
	defaultFlagdProxyDebugLogging = false
	defaultFlagdProxyNamespace    = "open-feature-operator-system"
)

// FlagSourceConfigurationReconciler reconciles a FlagSourceConfiguration object
type FlagSourceConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ReqLogger contains the Logger of this controller
	Log        logr.Logger
	FlagdProxy *FlagdProxyHandler
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;create
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FlagSourceConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Searching for FlagSourceConfiguration")

	// Fetch the FlagSourceConfiguration from the cache
	fsConfig := &corev1alpha1.FlagSourceConfiguration{}
	if err := r.Client.Get(ctx, req.NamespacedName, fsConfig); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(fmt.Sprintf("%s resource not found. Ignoring since object must be deleted", req.NamespacedName))
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get the %s", req.NamespacedName))
		return r.finishReconcile(err, false)
	}

	for _, source := range fsConfig.Spec.Sources {
		if source.Provider.IsFlagdProxy() {
			r.Log.Info(fmt.Sprintf("flagsourceconfiguration %s uses flagd-proxy, checking deployment", req.NamespacedName))
			if err := r.FlagdProxy.handleFlagdProxy(ctx); err != nil {
				r.Log.Error(err, "error handling the flagd-proxy deployment")
			}
			break
		}
	}

	if fsConfig.Spec.RolloutOnChange == nil || !*fsConfig.Spec.RolloutOnChange {
		return r.finishReconcile(nil, false)
	}

	// Object has been updated, so, we can restart any deployments that are using this annotation
	// => 	we know there has been an update because we are using the GenerationChangedPredicate filter
	// 		and our resource exists within the cluster
	deployList := &appsV1.DeploymentList{}
	if err := r.Client.List(ctx, deployList, client.MatchingFields{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.FlagSourceConfigurationAnnotation): "true",
	}); err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get the pods with annotation %s/%s", common.OpenFeatureAnnotationPath, common.FlagSourceConfigurationAnnotation))
		return r.finishReconcile(err, false)
	}

	// Loop through all deployments containing the openfeature.dev/flagsourceconfiguration annotation
	// and trigger a restart for any which have our resource listed as a configuration
	for _, deployment := range deployList.Items {
		annotations := deployment.Spec.Template.Annotations
		annotation, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationRoot, common.FlagSourceConfigurationAnnotation)]
		if !ok {
			continue
		}
		if r.isUsingConfiguration(fsConfig.Namespace, fsConfig.Name, deployment.Namespace, annotation) {
			r.Log.Info(fmt.Sprintf("restarting deployment %s/%s", deployment.Namespace, deployment.Name))
			deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			if err := r.Client.Update(ctx, &deployment); err != nil {
				r.Log.V(1).Error(err, fmt.Sprintf("Failed to update Deployment: %s/%s", deployment.Namespace, deployment.Name))
				continue
			}
		}
	}

	return r.finishReconcile(nil, false)
}

func (r *FlagSourceConfigurationReconciler) isUsingConfiguration(namespace string, name string, deploymentNamespace string, annotation string) bool {
	s := strings.Split(annotation, ",") // parse annotation list
	for _, target := range s {
		ss := strings.Split(strings.TrimSpace(target), "/")
		if len(ss) != 2 {
			target = fmt.Sprintf("%s/%s", deploymentNamespace, target)
		}
		if target == fmt.Sprintf("%s/%s", namespace, name) {
			return true
		}
	}
	return false
}

func (r *FlagSourceConfigurationReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := common.ReconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling FlagSourceConfiguration with error: %w")
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling FlagSourceConfiguration")
	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlagSourceConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.FlagSourceConfiguration{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// we are only interested in update events for this reconciliation loop
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
