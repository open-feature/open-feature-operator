package flagdproxy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/types"
	"golang.org/x/exp/maps"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FlagdProxyDeploymentName          = "flagd-proxy"
	FlagdProxyServiceAccountName      = "open-feature-operator-flagd-proxy"
	FlagdProxyServiceName             = "flagd-proxy-svc"
	FlagdProxyPodDisruptionBudgetName = "flagd-proxy-pdb"
)

type FlagdProxyHandler struct {
	client.Client
	config *FlagdProxyConfiguration
	Log    logr.Logger
}

type CreateUpdateFunc func(ctx context.Context, obj client.Object) error

type FlagdProxyConfiguration struct {
	Port                   int
	ManagementPort         int
	DebugLogging           bool
	Image                  string
	Tag                    string
	Replicas               int
	Namespace              string
	OperatorDeploymentName string
	ImagePullSecrets       []string
	Labels                 map[string]string
	Annotations            map[string]string
}

func NewFlagdProxyConfiguration(env types.EnvConfig, imagePullSecrets []string, labels map[string]string, annotations map[string]string) *FlagdProxyConfiguration {
	return &FlagdProxyConfiguration{
		Image:                  env.FlagdProxyImage,
		Tag:                    env.FlagdProxyTag,
		Namespace:              env.PodNamespace,
		OperatorDeploymentName: common.OperatorDeploymentName,
		Port:                   env.FlagdProxyPort,
		ManagementPort:         env.FlagdProxyManagementPort,
		DebugLogging:           env.FlagdProxyDebugLogging,
		Replicas:               env.FlagdProxyReplicaCount,
		ImagePullSecrets:       imagePullSecrets,
		Labels:                 labels,
		Annotations:            annotations,
	}
}

func NewFlagdProxyHandler(config *FlagdProxyConfiguration, client client.Client, logger logr.Logger) *FlagdProxyHandler {
	return &FlagdProxyHandler{
		config: config,
		Client: client,
		Log:    logger,
	}
}

func (f *FlagdProxyHandler) Config() *FlagdProxyConfiguration {
	return f.config
}

func specDiffers(a, b client.Object) (bool, error) {
	if a == nil || b == nil {
		return false, fmt.Errorf("object is nil")
	}

	// Compare only spec based on the object type
	switch a.(type) {
	case *corev1.Service:
		return !reflect.DeepEqual(a.(*corev1.Service).Spec, b.(*corev1.Service).Spec), nil
	case *appsV1.Deployment:
		return !reflect.DeepEqual(a.(*appsV1.Deployment).Spec, b.(*appsV1.Deployment).Spec), nil
	case *policyv1.PodDisruptionBudget:
		return !reflect.DeepEqual(a.(*policyv1.PodDisruptionBudget).Spec, b.(*policyv1.PodDisruptionBudget).Spec), nil
	default:
		return false, fmt.Errorf("unsupported object type")
	}
}

// ensureFlagdProxyResource ensures that the given object is reconciled in the cluster. If the object does not exist, it will be created.
func (f *FlagdProxyHandler) ensureFlagdProxyResource(ctx context.Context, obj client.Object) error {
	if obj == nil {
		return fmt.Errorf("object is nil")
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var old = obj.DeepCopyObject().(client.Object)

		// Try to get the existing object
		err := f.Client.Get(ctx, client.ObjectKey{Name: old.GetName(), Namespace: old.GetNamespace()}, old)
		notFound := errors.IsNotFound(err)
		if err != nil && !notFound {
			return err
		}

		// If the object exists but is not managed by OFO, return an error
		if !notFound && !common.IsManagedByOFO(old) {
			return fmt.Errorf("%s not managed by OFO", obj.GetName())
		}

		// If the object is not found, we will create it
		if notFound {
			return f.Client.Create(ctx, obj)
		}

		// If the object is found, update if necessary
		needsUpdate, err := specDiffers(obj, old)
		if err != nil {
			return err
		}

		if needsUpdate {
			obj.SetResourceVersion(old.GetResourceVersion())
			return f.Client.Update(ctx, obj)
		}

		return nil
	})
}

// HandleFlagdProxy ensures flagd-proxy kubernetes components are configured properly
func (f *FlagdProxyHandler) HandleFlagdProxy(ctx context.Context) error {
	var err error

	ownerRef, err := f.getOwnerReference(ctx)
	if err != nil {
		return err
	}

	if err = f.ensureFlagdProxyResource(ctx, f.newFlagdProxyDeployment(ownerRef)); err != nil {
		return err
	}

	if err = f.ensureFlagdProxyResource(ctx, f.newFlagdProxyService(ownerRef)); err != nil {
		return err
	}

	err = f.ensureFlagdProxyResource(ctx, f.newFlagdProxyPodDisruptionBudget(ownerRef))
	return err
}

func (f *FlagdProxyHandler) newFlagdProxyService(ownerReference *metav1.OwnerReference) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyServiceName,
			Namespace:       f.config.Namespace,
			OwnerReferences: []metav1.OwnerReference{*ownerReference},
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name":      FlagdProxyDeploymentName,
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "flagd-proxy",
					Port:       int32(f.config.Port),
					TargetPort: intstr.FromInt(f.config.Port),
				},
			},
		},
	}
}

func (f *FlagdProxyHandler) newFlagdProxyPodDisruptionBudget(ownerReference *metav1.OwnerReference) *policyv1.PodDisruptionBudget {

	// Only require pods to be available if there is >1 replica configured (HA setup)
	minReplicas := intstr.FromInt(0)
	if f.config.Replicas > 1 {
		minReplicas = intstr.FromInt(f.config.Replicas / 2)
	}

	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyPodDisruptionBudgetName,
			Namespace:       f.config.Namespace,
			OwnerReferences: []metav1.OwnerReference{*ownerReference},
			Labels: map[string]string{
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      FlagdProxyDeploymentName,
					common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
				},
			},
		},
	}
}

func (f *FlagdProxyHandler) newFlagdProxyDeployment(ownerReference *metav1.OwnerReference) *appsV1.Deployment {
	replicas := int32(f.config.Replicas)
	args := []string{
		"start",
		"--management-port",
		fmt.Sprintf("%d", f.config.ManagementPort),
	}
	if f.config.DebugLogging {
		args = append(args, "--debug")
	}
	imagePullSecrets := []corev1.LocalObjectReference{}
	for _, secret := range f.config.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{
			Name: secret,
		})
	}
	flagdLabels := map[string]string{
		"app":                         FlagdProxyDeploymentName,
		"app.kubernetes.io/name":      FlagdProxyDeploymentName,
		common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
		"app.kubernetes.io/version":   f.config.Tag,
	}
	if len(f.config.Labels) > 0 {
		maps.Copy(flagdLabels, f.config.Labels)
	}

	// No "built-in" annotations to merge at this time. If adding them follow the same pattern as labels.
	flagdAnnotations := map[string]string{}
	if len(f.config.Annotations) > 0 {
		maps.Copy(flagdAnnotations, f.config.Annotations)
	}

	return &appsV1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      FlagdProxyDeploymentName,
			Namespace: f.config.Namespace,
			Labels: map[string]string{
				"app":                         FlagdProxyDeploymentName,
				common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":   f.config.Tag,
			},
			OwnerReferences: []metav1.OwnerReference{*ownerReference},
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": FlagdProxyDeploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      flagdLabels,
					Annotations: flagdAnnotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: FlagdProxyServiceAccountName,
					ImagePullSecrets:   imagePullSecrets,
					Containers: []corev1.Container{
						{
							Image: fmt.Sprintf("%s:%s", f.config.Image, f.config.Tag),
							Name:  FlagdProxyDeploymentName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "port",
									ContainerPort: int32(f.config.Port),
								},
								{
									Name:          "management-port",
									ContainerPort: int32(f.config.ManagementPort),
								},
							},
							Args: args,
						},
					},
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app.kubernetes.io/name":      FlagdProxyDeploymentName,
									common.ManagedByAnnotationKey: common.ManagedByAnnotationValue,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (f *FlagdProxyHandler) getOperatorDeployment(ctx context.Context) (*appsV1.Deployment, error) {
	d := &appsV1.Deployment{}
	if err := f.Client.Get(ctx, client.ObjectKey{Name: f.config.OperatorDeploymentName, Namespace: f.config.Namespace}, d); err != nil {
		return nil, fmt.Errorf("unable to fetch operator deployment: %w", err)
	}
	return d, nil
}

func (f *FlagdProxyHandler) getOwnerReference(ctx context.Context) (*metav1.OwnerReference, error) {
	operatorDeployment, err := f.getOperatorDeployment(ctx)
	if err != nil {
		f.Log.Error(err, "unable to create owner reference for open-feature-operator")
		return nil, err
	}

	return &metav1.OwnerReference{
		UID:        operatorDeployment.GetUID(),
		Name:       operatorDeployment.GetName(),
		APIVersion: operatorDeployment.APIVersion,
		Kind:       operatorDeployment.Kind,
	}, nil
}
