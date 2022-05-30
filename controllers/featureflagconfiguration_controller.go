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

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
)

// FeatureFlagConfigurationReconciler reconciles a FeatureFlagConfiguration object
type FeatureFlagConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=featureflagconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=featureflagconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=featureflagconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FeatureFlagConfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *FeatureFlagConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	podList := &v1.PodList{}

	// I'd like to be able to do some field-matching with the ListOptions so we don't have to iterate,
	// but they can't seem to be used for annotations.
	// we may want our webhook to also add labels for easy querying?
	if err := r.List(ctx, podList); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// find any pods that are associated with this CR
	for _, pod := range podList.Items {
if val, ok := pod.ObjectMeta.Annotations["openfeature.dev/featureflagconfiguration"]; ok {
if val == req.Name {
			configMapList := &v1.ConfigMapList{}
			
			// query configMaps
if err := r.List(ctx, configMapList) ; err != nil {
  return ctrl.Result{}, errors.New("error listing")
}
			for _, configMap := range configMapList.Items {
				// find the configMap matching the pod name (this is how our webhook names them for now, might want something else long-term)
				if (configMap.Name == pod.Name) {

					// get the new contents by querying our CR based on the request data.
					featureFlagConfiguration := &configv1alpha1.FeatureFlagConfiguration{}
if err := r.Get(ctx, req.NamespacedName, featureFlagConfiguration) ; err != nil {
  return ctrl.Result{}, errors.New("error getting custom resource")
}

					// update the config map with the new contents.
					configMap.Data["config.yaml"] = featureFlagConfiguration.Spec.FeatureFlagSpec
					r.Update(ctx, &configMap)
					logger.Info(fmt.Sprintf("Successfully updated configMap %s", configMap.Name));
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FeatureFlagConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.FeatureFlagConfiguration{}).
		Complete(r)
}
