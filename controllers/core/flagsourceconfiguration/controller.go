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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/open-feature/open-feature-operator/controllers/common"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	envVarPodNamespace           = "POD_NAMESPACE"
	envVarProxyImage             = "KUBE_PROXY_IMAGE"
	envVarProxyTag               = "KUBE_PROXY_TAG"
	envVarProxyPort              = "KUBE_PROXY_PORT"
	envVarProxyMetricsPort       = "KUBE_PROXY_METRICS_PORT"
	envVarProxyDebugLogging      = "KUBE_PROXY_DEBUG_LOGGING"
	defaultKubeProxyImage        = "ghcr.io/open-feature/kube-flagd-proxy"
	defaultKubeProxyTag          = "v0.1.2" //KUBE_PROXY_TAG_RENOVATE
	defaultKubeProxyPort         = 8015
	defaultKubeProxyMetricsPort  = 8016
	defaultKubeProxyDebugLogging = false
	defaultKubeProxyNamespace    = "open-feature-operator-system"
)

// FlagSourceConfigurationReconciler reconciles a FlagSourceConfiguration object
type FlagSourceConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ReqLogger contains the Logger of this controller
	Log             logr.Logger
	KubeProxyConfig *KubeProxyConfiguration
}

type KubeProxyConfiguration struct {
	Port         int
	MetricsPort  int
	DebugLogging bool
	Image        string
	Tag          string
	Namespace    string
}

//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=flagsourceconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
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
	if err := r.Client.Create(ctx, r.newFlagdKubeProxyManifest()); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	r.Log.Info("deploying the kube-flagd-proxy service")
	if err := r.Client.Create(ctx, r.newFlagdKubeProxyServiceManifest()); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *FlagSourceConfigurationReconciler) newFlagdKubeProxyServiceManifest() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyServiceName,
			Namespace: r.KubeProxyConfig.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": KubeProxyDeploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "flagd-kube-proxy",
					Port:       int32(r.KubeProxyConfig.Port),
					TargetPort: intstr.FromInt(r.KubeProxyConfig.Port),
				},
			},
		},
	}
}

func (r *FlagSourceConfigurationReconciler) newFlagdKubeProxyManifest() *appsV1.Deployment {
	replicas := int32(1)
	args := []string{
		"start",
		"--metrics-port",
		fmt.Sprintf("%d", r.KubeProxyConfig.MetricsPort),
	}
	if r.KubeProxyConfig.DebugLogging {
		args = append(args, "--debug")
	}
	return &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyDeploymentName,
			Namespace: r.KubeProxyConfig.Namespace,
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
							Image: fmt.Sprintf("%s:%s", r.KubeProxyConfig.Image, r.KubeProxyConfig.Tag),
							Name:  KubeProxyDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "port",
									ContainerPort: int32(r.KubeProxyConfig.Port),
								},
								{
									Name:          "metrics-port",
									ContainerPort: int32(r.KubeProxyConfig.MetricsPort),
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

func (r *FlagSourceConfigurationReconciler) doesKubeProxyExist(ctx context.Context) (bool, error) {
	r.Client.Scheme()
	d := appsV1.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: KubeProxyDeploymentName, Namespace: r.KubeProxyConfig.Namespace}, &d)
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

func NewKubeProxyConfig() (*KubeProxyConfiguration, error) {
	kpc := &KubeProxyConfiguration{
		Port:         defaultKubeProxyPort,
		MetricsPort:  defaultKubeProxyMetricsPort,
		DebugLogging: defaultKubeProxyDebugLogging,
		Image:        defaultKubeProxyImage,
		Tag:          defaultKubeProxyTag,
		Namespace:    defaultKubeProxyNamespace,
	}
	ns, ok := os.LookupEnv(envVarPodNamespace)
	if ok {
		kpc.Namespace = ns
	}
	kpi, ok := os.LookupEnv(envVarProxyImage)
	if ok {
		kpc.Image = kpi
	}
	kpt, ok := os.LookupEnv(envVarProxyTag)
	if ok {
		kpc.Tag = kpt
	}
	portString, ok := os.LookupEnv(envVarProxyPort)
	if ok {
		port, err := strconv.Atoi(portString)
		if err != nil {
			return kpc, fmt.Errorf("could not parse %s env var: %w", envVarProxyPort, err)
		}
		kpc.Port = port
	}
	metricsPortString, ok := os.LookupEnv(envVarProxyMetricsPort)
	if ok {
		metricsPort, err := strconv.Atoi(metricsPortString)
		if err != nil {
			return kpc, fmt.Errorf("could not parse %s env var: %w", envVarProxyMetricsPort, err)
		}
		kpc.MetricsPort = metricsPort
	}
	kpDebugLogging, ok := os.LookupEnv(envVarProxyDebugLogging)
	if ok {
		debugLogging, err := strconv.ParseBool(kpDebugLogging)
		if err != nil {
			return kpc, fmt.Errorf("could not parse %s env var: %w", envVarProxyDebugLogging, err)
		}
		kpc.DebugLogging = debugLogging
	}
	return kpc, nil
}
