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
	"os"
	"strconv"
	"strings"
	"time"

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

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
)

const (
	OpenFeatureAnnotationPath         = "spec.template.metadata.annotations.openfeature.dev/openfeature.dev"
	FlagSourceConfigurationAnnotation = "flagsourceconfiguration"
	OpenFeatureAnnotationRoot         = "openfeature.dev"
	KubeProxyDeploymentName           = "kube-proxy"
	KubeProxyServiceAccountName       = "open-feature-operator-kube-proxy"
	KubeProxyServiceName              = "kube-proxy-svc"
)

var (
	CurrentNamespace      = "open-feature-operator-system"
	kubeProxyImage        = "ghcr.io/open-feature/kube-flagd-proxy"
	kubeProxyTag          = "v0.1.2"
	KubeProxyPort         = 8015
	kubeProxyMetricsPort  = 8016
	kubeProxyDebugLogging = false
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
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/finalizers,verbs=update

func (m *FlagSourceConfigurationReconciler) Init(ctx context.Context) error {
	ns, ok := os.LookupEnv("POD_WEBHOOK")
	if ok {
		CurrentNamespace = ns
	}
	kpi, ok := os.LookupEnv("KUBE_PROXY_IMAGE")
	if ok {
		kubeProxyImage = kpi
	}
	kpt, ok := os.LookupEnv("KUBE_PROXY_TAG")
	if ok {
		kubeProxyTag = kpt
	}
	portString, ok := os.LookupEnv("KUBE_PROXY_PORT")
	if ok {
		port, err := strconv.Atoi(portString)
		if err != nil {
			return fmt.Errorf("could not parse KUBE_PROXY_TAG env var: %w", err)
		}
		KubeProxyPort = port
	}
	kpDebugLogging, ok := os.LookupEnv("KUBE_PROXY_DEBUG_LOGGING")
	if ok {
		debugLogging, err := strconv.ParseBool(kpDebugLogging)
		if err != nil {
			return fmt.Errorf("could not parse KUBE_PROXY_DEBUG_LOGGING env var: %w", err)
		}
		kubeProxyDebugLogging = debugLogging
	}
	return nil
}

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
			r.Log.Info(fmt.Sprintf("%s resource not found. Ignoring since object must be deleted", req.NamespacedName))
			return r.finishReconcile(nil, false)
		}
		r.Log.Error(err, fmt.Sprintf("Failed to get the %s", req.NamespacedName))
		return r.finishReconcile(err, false)
	}
	for _, source := range fsConfig.Spec.Sources {
		if source.Provider.IsKubeProxy() {
			r.Log.Info(fmt.Sprintf("flagsourceconfiguration %s uses kube-proxy, checking deployment", req.NamespacedName))
			if err := r.handleKubeProxy(ctx); err != nil {
				r.Log.Error(err, "error handling the kube-flagd-proxy deployment")
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
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, FlagSourceConfigurationAnnotation): "true",
	}); err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get the pods with annotation %s/%s", OpenFeatureAnnotationPath, FlagSourceConfigurationAnnotation))
		return r.finishReconcile(err, false)
	}

	// Loop through all deployments containing the openfeature.dev/flagsourceconfiguration annotation
	// and trigger a restart for any which have our resource listed as a configuration
	for _, deployment := range deployList.Items {
		annotations := deployment.Spec.Template.Annotations
		annotation, ok := annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationRoot, FlagSourceConfigurationAnnotation)]
		if !ok {
			continue
		}
		if isUsingConfiguration(fsConfig.Namespace, fsConfig.Name, deployment.Namespace, annotation) {
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

func (r *FlagSourceConfigurationReconciler) handleKubeProxy(ctx context.Context) error {
	exists, err := r.doesKubeProxyExist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return r.deployKubeProxy(ctx)
	}
	return nil
}

func (r *FlagSourceConfigurationReconciler) deployKubeProxy(ctx context.Context) error {
	r.Log.Info("deploying the kube-flagd-proxy")
	if err := r.Client.Create(ctx, newFlagdKubeProxyManifest()); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	r.Log.Info("deploying the kube-flagd-proxy service")
	if err := r.Client.Create(ctx, newFlagdKubeProxyServiceManifest()); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func newFlagdKubeProxyServiceManifest() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyServiceName,
			Namespace: CurrentNamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": KubeProxyDeploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "flagd-kube-proxy",
					Port:       int32(KubeProxyPort),
					TargetPort: intstr.FromInt(KubeProxyPort),
				},
			},
		},
	}
}

func newFlagdKubeProxyManifest() *appsV1.Deployment {
	replicas := int32(1)
	args := []string{
		"start",
		"--metrics-port",
		fmt.Sprintf("%d", kubeProxyMetricsPort),
	}
	if kubeProxyDebugLogging {
		args = append(args, "--debug")
	}
	return &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyDeploymentName,
			Namespace: CurrentNamespace,
			Labels: map[string]string{
				"app": KubeProxyDeploymentName,
			},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": KubeProxyDeploymentName,
				},
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                    KubeProxyDeploymentName,
						"app.kubernetes.io/name": KubeProxyDeploymentName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: KubeProxyServiceAccountName,
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf("%s:%s", kubeProxyImage, kubeProxyTag),
							Name:  KubeProxyDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "port",
									ContainerPort: int32(KubeProxyPort),
								},
								{
									Name:          "metrics-port",
									ContainerPort: int32(kubeProxyMetricsPort),
								},
							},
							Args: args,
						},
					},
				},
			},
		},
	}
}

func (f *FlagSourceConfigurationReconciler) doesKubeProxyExist(ctx context.Context) (bool, error) {
	f.Client.Scheme()
	d := appsV1.Deployment{}
	err := f.Client.Get(ctx, client.ObjectKey{Name: KubeProxyDeploymentName, Namespace: CurrentNamespace}, &d)
	if err != nil {
		if errors.IsNotFound(err) {
			// does not exist, is not ready, no error
			return false, nil
		}
		// does not exist, is not ready, is in error
		return false, err
	}
	// exists, at least one replica ready, no error
	return true, nil
}

func isUsingConfiguration(namespace string, name string, deploymentNamespace string, annotation string) bool {
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
		interval := reconcileErrorInterval
		if requeueImmediate {
			interval = 0
		}
		r.Log.Error(err, "Finished Reconciling FlagSourceConfiguration with error: %w")
		return ctrl.Result{Requeue: true, RequeueAfter: interval}, err
	}
	r.Log.Info("Finished Reconciling FlagSourceConfiguration")
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
