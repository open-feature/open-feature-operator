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
	"strings"
	"time"

	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
)

const (
	OpenFeatureAnnotationPath         = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	FlagSourceConfigurationAnnotation = "flagsourceconfiguration"
	OpenFeatureAnnotationRoot         = "openfeature.dev"
)

// FlagSourceConfigurationReconciler reconciles a FlagSourceConfiguration object
type FlagSourceConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ReqLogger contains the Logger of this controller
	Log logr.Logger
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;update;patch;
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FlagSourceConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx)

	// Fetch the FlagSourceConfiguration from the cache
	fsConfig := &corev1alpha1.FlagSourceConfiguration{}
	if err := r.Client.Get(ctx, req.NamespacedName, fsConfig); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(crdName + " resource not found. Ignoring since object must be deleted")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, "Failed to get the "+crdName)
		return r.finishReconcile(err, false)
	}

	if fsConfig.Spec.RolloutOnChange == nil || !*fsConfig.Spec.RolloutOnChange {
		return r.finishReconcile(nil, false)
	}

	// Object has been updated, so, we can restart any deployments that are using this annotation
	// => 	we know there has been an update because we are using the GenerationChangedPredicate filter
	// 		and our resource exists within the cluster
	deployList := &appsV1.DeploymentList{}
	if err := r.Client.List(ctx, deployList, client.MatchingFields{
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, FlagSourceConfigurationAnnotation): "true",
	}); err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get the pods with annotation %s/%s", OpenFeatureAnnotationPath, FlagSourceConfigurationAnnotation))
		return r.finishReconcile(err, false)
	}

	// Loop through all deployments containing the openfeature.dev/featureflagconfiguration annotation
	// and trigger a restart for any which have our resource listed as a configuration
	for _, deployment := range deployList.Items {
		fmt.Println("handling deployment ", deployment.Name)
		annotations := deployment.Spec.Template.Annotations
		fmt.Println(annotations)
		annotation, ok := annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationRoot, FlagSourceConfigurationAnnotation)]
		if !ok {
			continue
		}
		if isUsingConfiguration(fsConfig.Namespace, fsConfig.Name, deployment.Namespace, annotation) {
			fmt.Println("found to be using annotation")
			deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			if err := r.Client.Update(ctx, &deployment); err != nil {
				r.Log.V(1).Error(err, fmt.Sprintf("Failed to update Deployment: %s", err.Error()))
				continue
			}
		}
	}

	return r.finishReconcile(nil, false)
}

func isUsingConfiguration(namespace string, name string, deploymentNamespace string, annotation string) bool {
	s := strings.Split(annotation, ",") // parse annotation list
	for _, target := range s {
		fmt.Println(target)
		ss := strings.Split(strings.TrimSpace(target), "/")
		fmt.Println(ss)
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
		interval := reconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling "+crdName+" with error: %w")
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling " + crdName)
	return ctrl.Result{Requeue: false}, nil
}

func FlagSourceConfigurationIndex(o client.Object) []string {
	deployment := o.(*appsV1.Deployment)
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

// SetupWithManager sets up the controller with the Manager.
func (r *FlagSourceConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.FlagSourceConfiguration{}).
		// we are only interested in update events for this reconciliation loop
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
