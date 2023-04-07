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
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha3 "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"
	"github.com/open-feature/open-feature-operator/controllers/common"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	flagdCRDName           string = "Flagd"
	flagdContainerName     string = "flagd"
	flagdSelectorLabel     string = "flagd"
	clusterRoleBindingName string = "open-feature-operator-flagd-kubernetes-sync"
)

// FlagdReconciler reconciles a Flagd object
type FlagdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Flagd object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FlagdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx)
	r.Log.Info("Reconciling " + flagdCRDName)

	flagd := &corev1alpha3.Flagd{}
	if err := r.Client.Get(ctx, req.NamespacedName, flagd); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(flagdCRDName + " resource not found. Ignoring since object must be deleted")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, "Failed to get the "+flagdCRDName)
		return r.finishReconcile(err, false)
	}
	ns := flagd.Namespace
	flagdOwnerReferences := []metav1.OwnerReference{
		{
			APIVersion: flagd.APIVersion,
			Kind:       flagd.Kind,
			Name:       flagd.Name,
			UID:        flagd.UID,
			Controller: utils.FalseVal(),
		},
	}

	fsConfigSpec, err := corev1alpha1.NewFlagSourceConfigurationSpec()
	if err != nil {
		r.Log.Error(err, "unable to parse env var configuration")
		return r.finishReconcile(nil, false)
	}

	// fetch the FlagSourceConfiguration
	fsConfig := &corev1alpha1.FlagSourceConfiguration{}
	fsConfigNs, fsConfigName := ParseAnnotation(flagd.Spec.FlagSourceConfiguration, flagd.Namespace)
	if err := r.Client.Get(ctx,
		client.ObjectKey{Namespace: fsConfigNs, Name: fsConfigName}, fsConfig,
	); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Error(fmt.Errorf("%s/%s not found", fsConfigNs, fsConfigName), "FlagSourceConfiguration not found")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get FlagSourceConfiguration %s/%s", fsConfigNs, fsConfigName))
		return r.finishReconcile(err, false)
	}

	fsConfigSpec.Merge(&fsConfig.Spec)

	var selectorLabels map[string]string
	// retrieve service and attach selector labels to deployment
	if flagd.Spec.Service != "" {
		selectorLabels, err = r.serviceSelectorLabels(ctx, ns, flagd.Spec.Service)
		if err != nil {
			return r.finishReconcile(nil, false)
		}
	}

	// check for existing deployment
	deployment := &appsV1.Deployment{}
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: ns, Name: flagd.Name}, deployment,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the deployment %s/%s", ns, flagd.Name))
			return r.finishReconcile(nil, false)
		} else {
			deployment.Name = flagd.Name
			deployment.Namespace = ns
			deployment.Spec = flagd.Spec.DeploymentSpec
		}
	} else {
		deployment.Name = flagd.Name
		deployment.Namespace = ns
		deployment.Spec = flagd.Spec.DeploymentSpec
	}

	container := flagProviderContainer(deployment)
	container.Name = flagdContainerName
	container.Image = fmt.Sprintf("%s:%s", fsConfigSpec.Image, fsConfigSpec.Tag)
	container.Args = []string{"start"}
	container.ImagePullPolicy = corev1.PullAlways // TODO: configurable
	container.VolumeMounts = []corev1.VolumeMount{}
	container.Env = fsConfigSpec.EnvVars
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "metrics",
			ContainerPort: fsConfigSpec.MetricsPort,
		},
	}
	container.SecurityContext = nil // TODO

	deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	if err := HandleSourcesProviders(ctx, r.Log, r.Client, fsConfigSpec, fsConfigNs, constant.Namespace, flagd.Spec.ServiceAccountName,
		flagdOwnerReferences, &deployment.Spec.Template.Spec, deployment.Spec.Template.ObjectMeta, &container,
	); err != nil {
		r.Log.Error(err, "handle source providers")
		return r.finishReconcile(nil, false)
	}

	mergeFlagProviderContainer(deployment, container)

	deployment.Spec.Template.Spec.ServiceAccountName = flagd.Spec.ServiceAccountName
	deployment.OwnerReferences = flagdOwnerReferences

	applyDeploymentLabelsAndSelector(deployment, flagd, selectorLabels)

	if err := r.Client.Create(ctx, deployment); err != nil {
		r.Log.Error(err, "Failed to create deployment")
		return r.finishReconcile(nil, false)
	}

	return r.finishReconcile(nil, false)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlagdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha3.Flagd{}).
		// we are only interested in update events for this reconciliation loop
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *FlagdReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := common.ReconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling "+flagdCRDName)
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling " + flagdCRDName)
	return ctrl.Result{Requeue: false}, nil
}

func (r *FlagdReconciler) serviceSelectorLabels(ctx context.Context, serviceNs, serviceName string) (map[string]string, error) {
	svc := &corev1.Service{}
	r.Log.V(1).Info(fmt.Sprintf("Fetching service: %s/%s", serviceNs, serviceName))
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: serviceNs, Name: serviceName}, svc,
	); err != nil {
		r.Log.Error(err,
			fmt.Sprintf("Failed to get the service %s/%s", serviceNs, serviceName))
		return nil, err
	}

	return svc.Spec.Selector, nil
}

func applyDeploymentLabelsAndSelector(deployment *appsV1.Deployment, flagd *corev1alpha1.Flagd, selectorLabels map[string]string) {
	if deployment.Labels == nil {
		deployment.Labels = make(map[string]string)
	}
	deployment.Labels[flagdSelectorLabel] = flagd.Name
	for key, value := range selectorLabels {
		deployment.Labels[key] = value
	}
	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = make(map[string]string)
	}
	deployment.Spec.Template.Labels[flagdSelectorLabel] = flagd.Name
	if deployment.Spec.Selector == nil || deployment.Spec.Selector.MatchLabels == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: make(map[string]string)}
	}
	deployment.Spec.Selector.MatchLabels[flagdSelectorLabel] = flagd.Name
}

func flagProviderContainer(deployment *appsV1.Deployment) corev1.Container {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == flagdContainerName {
			return container
		}
	}

	return corev1.Container{}
}

func mergeFlagProviderContainer(deployment *appsV1.Deployment, flagProviderContainer corev1.Container) {
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		deployment.Spec.Template.Spec.Containers = []corev1.Container{flagProviderContainer}
		return
	}

	for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
		existingFlagProviderContainer := deployment.Spec.Template.Spec.Containers[i]
		if existingFlagProviderContainer.Name == flagdContainerName {
			deployment.Spec.Template.Spec.Containers[i] = flagProviderContainer
			return
		}
	}

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, flagProviderContainer)
}
