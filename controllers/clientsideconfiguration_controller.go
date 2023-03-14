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
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"strings"
)

const (
	clientSideConfigurationCRDName        = "ClientSideConfiguration"
	clientSideDeploymentName              = "clientsidedeployment"
	clientSideServiceName                 = "clientsideservice"
	clientSideAppName                     = "clientsideapp"
	clientSideGatewayListenerName         = "clientsidegatewaylistener"
	clientSideServicePort          int32  = 80
	clusterRoleBindingName         string = "open-feature-operator-flagd-kubernetes-sync"
)

// ClientSideConfigurationReconciler reconciles a ClientSideConfiguration object
type ClientSideConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clientsideconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clientsideconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clientsideconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClientSideConfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClientSideConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx)
	r.Log.Info("Reconciling " + clientSideConfigurationCRDName)

	csconf := &corev1alpha1.ClientSideConfiguration{}
	if err := r.Client.Get(ctx, req.NamespacedName, csconf); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Info(clientSideConfigurationCRDName + " resource not found. Ignoring since object must be deleted")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, "Failed to get the "+clientSideConfigurationCRDName)
		return r.finishReconcile(err, false)
	}
	ns := csconf.Namespace

	// check for existing client side deployment
	deployment := &appsV1.Deployment{}
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: ns, Name: clientSideDeploymentName}, deployment,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the deployment %s/%s", ns, clientSideDeploymentName))
			return r.finishReconcile(err, false)
		} else {
			deployment.Name = clientSideDeploymentName
			deployment.Namespace = ns
		}
	} else {
		// TODO: delete deployment
		deployment.Name = clientSideDeploymentName
		deployment.Namespace = ns
	}

	fsConfigSpec, err := corev1alpha1.NewFlagSourceConfigurationSpec()
	if err != nil {
		r.Log.Error(err, "unable to parse env var configuration")
		return r.finishReconcile(nil, false)
	}

	// fetch the FlagSourceConfiguration
	fsConfig := &corev1alpha1.FlagSourceConfiguration{}
	if err := r.Client.Get(ctx,
		client.ObjectKey{Namespace: ns, Name: csconf.Spec.FlagSourceConfiguration}, fsConfig, // TODO: use namespace from spec
	); err != nil {
		if errors.IsNotFound(err) {
			// taking down all associated K8s resources is handled by K8s
			r.Log.Error(fmt.Errorf("%s/%s not found", ns, csconf.Spec.FlagSourceConfiguration), "FlagSourceConfiguration not found")
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get FlagSourceConfiguration %s/%s", ns, csconf.Spec.FlagSourceConfiguration))
		return r.finishReconcile(err, false)
	}

	fsConfigSpec.Merge(&fsConfig.Spec)

	// create service if it doesn't exist, update if it does
	svc := &corev1.Service{}
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: ns, Name: clientSideServiceName}, svc,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the service %s/%s", ns, clientSideServiceName))
			return r.finishReconcile(err, false)
		} else {
			svc.Name = clientSideServiceName
			svc.Namespace = ns
			svc.Spec.Selector = map[string]string{
				"app": clientSideAppName,
			}
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			svc.Spec.Ports = []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       clientSideServicePort,
					TargetPort: intstr.FromInt(int(fsConfigSpec.Port)),
				},
			}

			if err := r.Client.Create(ctx, svc); err != nil {
				r.Log.Error(err, "Failed to create service")
				return r.finishReconcile(nil, false)
			}
		}
	} else {
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Protocol:   corev1.ProtocolTCP,
				Port:       clientSideServicePort,
				TargetPort: intstr.FromInt(int(fsConfigSpec.Port)),
			},
		}

		if err := r.Client.Update(ctx, svc); err != nil {
			r.Log.Error(err, "Failed to update service")
			return r.finishReconcile(nil, false)
		}
	}

	// create gateway if it doesn't exist, update if it does
	namespacesFromSame := gatewayv1beta1.NamespacesFromSame
	hostname := gatewayv1beta1.Hostname(csconf.Spec.HTTPRouteHostname)
	gateway := &gatewayv1beta1.Gateway{}
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: ns, Name: csconf.Spec.GatewayName}, gateway,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the gateway %s/%s", ns, csconf.Spec.GatewayName))
			return r.finishReconcile(err, false)
		}
		gateway.Name = csconf.Spec.GatewayName
		gateway.Namespace = ns
		gateway.Spec.GatewayClassName = gatewayv1beta1.ObjectName(csconf.Spec.GatewayClassName)
		gateway.Spec.Listeners = []gatewayv1beta1.Listener{
			{
				Name:     clientSideGatewayListenerName,
				Hostname: &hostname,
				Protocol: gatewayv1beta1.HTTPProtocolType,
				Port:     gatewayv1beta1.PortNumber(csconf.Spec.HTTPRoutePort),
				AllowedRoutes: &gatewayv1beta1.AllowedRoutes{
					Namespaces: &gatewayv1beta1.RouteNamespaces{
						From: &namespacesFromSame,
					},
				},
			},
		}

		if err := r.Client.Create(ctx, gateway); err != nil {
			r.Log.Error(err, "Failed to create gateway")
			return r.finishReconcile(nil, false)
		}
	} else {
		gateway.Spec.GatewayClassName = gatewayv1beta1.ObjectName(csconf.Spec.GatewayClassName)
		listener := gatewayv1beta1.Listener{
			Name:     clientSideGatewayListenerName,
			Hostname: &hostname,
			Protocol: gatewayv1beta1.HTTPProtocolType,
			Port:     gatewayv1beta1.PortNumber(csconf.Spec.HTTPRoutePort),
			AllowedRoutes: &gatewayv1beta1.AllowedRoutes{
				Namespaces: &gatewayv1beta1.RouteNamespaces{
					From: &namespacesFromSame,
				},
			},
		}

		listenerExists := false
		for i := 0; i < len(gateway.Spec.Listeners); i++ {
			if gateway.Spec.Listeners[i].Name == clientSideGatewayListenerName {
				gateway.Spec.Listeners[i] = listener
				listenerExists = true
				break
			}
		}

		if !listenerExists {
			gateway.Spec.Listeners = append(gateway.Spec.Listeners, listener)
		}

		if err := r.Client.Update(ctx, gateway); err != nil {
			r.Log.Error(err, "Failed to update gateway")
			return r.finishReconcile(nil, false)
		}
	}

	// create gateway http route if it doesn't exist
	httpRoute := &gatewayv1beta1.HTTPRoute{}
	httpRoutePort := gatewayv1beta1.PortNumber(clientSideServicePort)
	httpRouteSectionName := gatewayv1beta1.SectionName(clientSideGatewayListenerName)
	httpRouteHostname := gatewayv1beta1.Hostname(csconf.Spec.HTTPRouteHostname)
	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: ns, Name: csconf.Spec.HTTPRouteName}, httpRoute,
	); err != nil {
		if !errors.IsNotFound(err) {
			r.Log.Error(err,
				fmt.Sprintf("Failed to get the gateway http route %s/%s", ns, csconf.Spec.HTTPRouteName))
			return r.finishReconcile(nil, false)
		} else {
			httpRoute.Name = csconf.Spec.HTTPRouteName
			httpRoute.Namespace = ns
			httpRoute.Spec.ParentRefs = []gatewayv1beta1.ParentReference{
				{
					Name:        gatewayv1beta1.ObjectName(csconf.Spec.GatewayName),
					SectionName: &httpRouteSectionName,
				},
			}
			httpRoute.Spec.Hostnames = []gatewayv1beta1.Hostname{httpRouteHostname}
			httpRoute.Spec.Rules = []gatewayv1beta1.HTTPRouteRule{
				{
					BackendRefs: []gatewayv1beta1.HTTPBackendRef{
						{
							BackendRef: gatewayv1beta1.BackendRef{
								BackendObjectReference: gatewayv1beta1.BackendObjectReference{
									Name: clientSideServiceName,
									Port: &httpRoutePort,
								},
							},
						},
					},
				},
			}

			if err := r.Client.Create(ctx, httpRoute); err != nil {
				r.Log.Error(err, "Failed to create gateway http route")
				return r.finishReconcile(nil, false)
			}
		}
	}

	flagdContainer := corev1.Container{
		Name:  "flagd",
		Image: fmt.Sprintf("%s:%s", fsConfigSpec.Image, fsConfigSpec.Tag),
		Args: []string{
			"start",
			"--cors-origin=" + csconf.Spec.CorsAllowOrigin,
		},
		ImagePullPolicy: corev1.PullAlways, // TODO: configurable
		VolumeMounts:    []corev1.VolumeMount{},
		Env:             []corev1.EnvVar{},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: fsConfigSpec.MetricsPort,
			},
		},
		SecurityContext: nil, // TODO
		// TODO resource limits
	}

	for _, source := range fsConfigSpec.Sources {
		if source.Provider == "" {
			source.Provider = fsConfigSpec.DefaultSyncProvider
		}
		switch {
		case source.Provider.IsKubernetes():
			if err := r.handleKubernetesProvider(ctx, ns, csconf.Spec.ServiceAccountName, &flagdContainer, source); err != nil {
				r.Log.Error(err, "Failed to handle kubernetes provider")
				return r.finishReconcile(nil, false)
			}
		default:
			r.Log.Error(fmt.Errorf("%s", source.Provider), "Unsupported source")
			return r.finishReconcile(nil, false)
		}
	}

	deployment.Spec.Template.Spec.ServiceAccountName = csconf.Spec.ServiceAccountName
	labels := map[string]string{
		"app": clientSideAppName,
	}
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

func (r *ClientSideConfigurationReconciler) enableClusterRoleBinding(ctx context.Context, namespace, serviceAccountName string) error {
	serviceAccount := client.ObjectKey{
		Name:      serviceAccountName,
		Namespace: namespace,
	}
	if serviceAccountName == "" {
		serviceAccount.Name = "default"
	}
	// Check if the service account exists
	r.Log.V(1).Info(fmt.Sprintf("Fetching serviceAccount: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
	sa := corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, serviceAccount, &sa); err != nil {
		r.Log.V(1).Info(fmt.Sprintf("ServiceAccount not found: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
		return err
	}
	r.Log.V(1).Info(fmt.Sprintf("Fetching clusterrolebinding: %s", clusterRoleBindingName))
	// Fetch service account if it exists
	crb := rbacv1.ClusterRoleBinding{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: clusterRoleBindingName}, &crb); errors.IsNotFound(err) {
		r.Log.V(1).Info(fmt.Sprintf("ClusterRoleBinding not found: %s", clusterRoleBindingName))
		return err
	}
	found := false
	for _, subject := range crb.Subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == serviceAccount.Name && subject.Namespace == serviceAccount.Namespace {
			r.Log.V(1).Info(fmt.Sprintf("ClusterRoleBinding already exists for service account: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
			found = true
		}
	}
	if !found {
		r.Log.V(1).Info(fmt.Sprintf("Updating ClusterRoleBinding %s for service account: %s/%s", crb.Name,
			serviceAccount.Namespace, serviceAccount.Name))
		crb.Subjects = append(crb.Subjects, rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      serviceAccount.Name,
			Namespace: serviceAccount.Namespace,
		})
		if err := r.Client.Update(ctx, &crb); err != nil {
			r.Log.V(1).Info(fmt.Sprintf("Failed to update ClusterRoleBinding: %s", err.Error()))
			return err
		}
	}
	r.Log.V(1).Info(fmt.Sprintf("Updated ClusterRoleBinding: %s", crb.Name))

	return nil
}

func (r *ClientSideConfigurationReconciler) handleKubernetesProvider(ctx context.Context, namespace, serviceAccountName string, container *corev1.Container, source corev1alpha1.Source) error {
	ns, n := parseAnnotation(source.Source, namespace)
	// ensure that the FeatureFlagConfiguration exists
	ff := r.getFeatureFlag(ctx, ns, n)
	if ff.Name == "" {
		return fmt.Errorf("feature flag configuration %s/%s not found", ns, n)
	}
	if err := r.enableClusterRoleBinding(ctx, namespace, serviceAccountName); err != nil {
		return fmt.Errorf("enableClusterRoleBinding: %w", err)
	}
	// append args
	container.Args = append(
		container.Args,
		"--uri",
		fmt.Sprintf(
			"core.openfeature.dev/%s/%s",
			ns,
			n,
		),
	)
	return nil
}

func (r *ClientSideConfigurationReconciler) getFeatureFlag(ctx context.Context, namespace string, name string) corev1alpha1.FeatureFlagConfiguration {
	ffConfig := corev1alpha1.FeatureFlagConfiguration{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ffConfig); errors.IsNotFound(err) {
		return corev1alpha1.FeatureFlagConfiguration{}
	}
	return ffConfig
}

func parseAnnotation(s string, defaultNs string) (string, string) {
	ss := strings.Split(s, "/")
	if len(ss) == 2 {
		return ss[0], ss[1]
	}
	return defaultNs, s
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClientSideConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.ClientSideConfiguration{}).
		// we are only interested in update events for this reconciliation loop
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *ClientSideConfigurationReconciler) finishReconcile(err error, requeueImmediate bool) (ctrl.Result, error) {
	if err != nil {
		interval := reconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling "+clientSideConfigurationCRDName)
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling " + clientSideConfigurationCRDName)
	return ctrl.Result{Requeue: false}, nil
}
