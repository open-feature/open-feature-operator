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
	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/controllers/common"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	flagServiceCRDName            = "FlagService"
	flagServicePortName           = "flag-service-port"
	flagServicePort        int32  = 80
	clusterRoleBindingName string = "open-feature-operator-flagd-kubernetes-sync"
)

// FlagServiceReconciler reconciles a FlagService object
type FlagServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// clusterrole
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagservices/finalizers,verbs=update

// role
//+kubebuilder:rbac:groups="",namespace=open-feature-operator-system,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FlagService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FlagServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx)
	r.Log.Info("Reconciling " + flagServiceCRDName)

	flagSvc := &corev1alpha1.FlagService{}
	if err := r.Client.Get(ctx, req.NamespacedName, flagSvc); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(flagServiceCRDName + " resource not found. Ignoring since object must be deleted")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, "Failed to get the "+flagServiceCRDName)
		return r.finishReconcile(err, false)
	}
	ns := flagSvc.Namespace
	flagSvcOwnerReferences := []metav1.OwnerReference{
		{
			APIVersion: flagSvc.APIVersion,
			Kind:       flagSvc.Kind,
			Name:       flagSvc.Name,
			UID:        flagSvc.UID,
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
	fsConfigNs, fsConfigName := ParseAnnotation(flagSvc.Spec.FlagSourceConfiguration, flagSvc.Namespace)
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

	flagSvcPort := corev1.ServicePort{
		Name:       flagServicePortName,
		Protocol:   corev1.ProtocolTCP,
		Port:       flagServicePort,
		TargetPort: intstr.FromInt(int(fsConfigSpec.Port)),
	}

	// create service if it doesn't exist, update if it does
	// service mutation is intentionally restricted to OFO's namespace
	svc := &corev1.Service{}
	r.Log.V(1).Info(fmt.Sprintf("Fetching service: %s/%s", constant.Namespace, flagSvc.Name))
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: constant.Namespace, Name: flagSvc.Name}, svc,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the service %s/%s", constant.Namespace, flagSvc.Name))
			return r.finishReconcile(err, false)
		} else {
			r.Log.V(1).Info(fmt.Sprintf("Service %s/%s not found", constant.Namespace, flagSvc.Name))
			svc.Spec = flagSvc.Spec.ServiceSpec
			svc.Name = flagSvc.Name
			svc.OwnerReferences = flagSvcOwnerReferences
			svc.Namespace = constant.Namespace
			svc.Spec.Selector = map[string]string{
				"app": flagSvc.Name,
			}
			svc.Labels = flagSvc.Labels
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			svc.Spec.Ports = mergePorts(svc.Spec.Ports, flagSvcPort)

			r.Log.V(1).Info(fmt.Sprintf("Creating Service %s/%s", svc.Namespace, svc.Name))
			if err := r.Client.Create(ctx, svc); err != nil {
				r.Log.Error(err, "Failed to create service")
				return r.finishReconcile(nil, false)
			}
		}
	} else {
		r.Log.V(1).Info(fmt.Sprintf("Service %s/%s found", svc.Namespace, svc.Name))
		svc.Spec = flagSvc.Spec.ServiceSpec
		svc.Spec.Ports = mergePorts(svc.Spec.Ports, flagSvcPort)

		r.Log.V(1).Info(fmt.Sprintf("Updating Service %s/%s", svc.Namespace, svc.Name))
		if err := r.Client.Update(ctx, svc); err != nil {
			r.Log.Error(err, "Failed to update service")
			return r.finishReconcile(nil, false)
		}
	}

	r.Log.V(1).Info(fmt.Sprintf("Service created/updated: %+v", svc))

	// check for existing deployment
	deployment := &appsV1.Deployment{}
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: constant.Namespace, Name: flagSvc.Name}, deployment,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the deployment %s/%s", ns, flagSvc.Name))
			return r.finishReconcile(err, false)
		} else {
			deployment.Name = flagSvc.Name
			deployment.Namespace = constant.Namespace
		}
	} else {
		// TODO: delete deployment
		deployment.Name = flagSvc.Name
		deployment.Namespace = constant.Namespace
	}

	flagdContainer := corev1.Container{
		Name:  "flagd",
		Image: fmt.Sprintf("%s:%s", fsConfigSpec.Image, fsConfigSpec.Tag),
		Args: []string{
			"start",
		},
		ImagePullPolicy: corev1.PullAlways, // TODO: configurable
		VolumeMounts:    []corev1.VolumeMount{},
		Env:             fsConfigSpec.EnvVars,
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: fsConfigSpec.MetricsPort,
			},
		},
		SecurityContext: nil, // TODO
		// TODO resource limits
	}

	deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	if err := HandleSourcesProviders(ctx, r.Log, r.Client, fsConfigSpec, fsConfigNs, constant.Namespace, flagSvc.Spec.ServiceAccountName,
		flagSvcOwnerReferences, &deployment.Spec.Template.Spec, deployment.Spec.Template.ObjectMeta, &flagdContainer,
	); err != nil {
		r.Log.Error(err, "handle source providers")
		return r.finishReconcile(nil, false)
	}

	deployment.Spec.Template.Spec.ServiceAccountName = flagSvc.Spec.ServiceAccountName
	labels := map[string]string{
		"app": flagSvc.Name,
	}
	deployment.OwnerReferences = flagSvcOwnerReferences
	deployment.Labels = labels
	deployment.Spec.Template.Labels = labels
	deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	deployment.Spec.Template.Spec.Containers = []corev1.Container{flagdContainer}

	if err := r.Client.Create(ctx, deployment); err != nil {
		r.Log.Error(err, "Failed to create deployment")
		return r.finishReconcile(nil, false)
	}

	return r.finishReconcile(nil, false)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlagServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.FlagService{}).
		// we are only interested in update events for this reconciliation loop
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *FlagServiceReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := common.ReconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling "+flagServiceCRDName)
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling " + flagServiceCRDName)
	return ctrl.Result{Requeue: false}, nil
}

func mergePorts(ports []corev1.ServicePort, port corev1.ServicePort) []corev1.ServicePort {
	for i := 0; i < len(ports); i++ {
		if ports[i].Name == port.Name {
			ports[i] = port
			return ports
		}
	}

	return append(ports, port)
}
