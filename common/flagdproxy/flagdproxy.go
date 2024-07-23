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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FlagdProxyDeploymentName     = "flagd-proxy"
	FlagdProxyServiceAccountName = "open-feature-operator-flagd-proxy"
	FlagdProxyServiceName        = "flagd-proxy-svc"
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

func (f *FlagdProxyHandler) createObject(ctx context.Context, obj client.Object) error {
	return f.Client.Create(ctx, obj)
}

func (f *FlagdProxyHandler) updateObject(ctx context.Context, obj client.Object) error {
	return f.Client.Update(ctx, obj)
}

func (f *FlagdProxyHandler) HandleFlagdProxy(ctx context.Context) error {
	exists, deployment, err := f.doesFlagdProxyExist(ctx)
	if err != nil {
		return err
	}

	ownerReference, err := f.getOwnerReference(ctx)
	if err != nil {
		return err
	}
	newDeployment := f.newFlagdProxyManifest(ownerReference)
	newService := f.newFlagdProxyServiceManifest(ownerReference)

	if !exists {
		f.Log.Info("flagd-proxy Deployment does not exist, creating")
		return f.deployFlagdProxy(ctx, f.createObject, newDeployment, newService)
	}
	// flagd-proxy exists, need to check if we should update it
	if f.shouldUpdateFlagdProxy(deployment, newDeployment) {
		f.Log.Info("flagd-proxy Deployment out of sync, updating")
		return f.deployFlagdProxy(ctx, f.updateObject, newDeployment, newService)
	}
	f.Log.Info("flagd-proxy Deployment up-to-date")
	return nil
}

func (f *FlagdProxyHandler) deployFlagdProxy(ctx context.Context, createUpdateFunc CreateUpdateFunc, deployment *appsV1.Deployment, service *corev1.Service) error {
	f.Log.Info("deploying the flagd-proxy")
	if err := createUpdateFunc(ctx, deployment); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	f.Log.Info("deploying the flagd-proxy service")
	if err := createUpdateFunc(ctx, service); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (f *FlagdProxyHandler) newFlagdProxyServiceManifest(ownerReference *metav1.OwnerReference) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyServiceName,
			Namespace:       f.config.Namespace,
			OwnerReferences: []metav1.OwnerReference{*ownerReference},
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

func (f *FlagdProxyHandler) newFlagdProxyManifest(ownerReference *metav1.OwnerReference) *appsV1.Deployment {
	replicas := int32(1)
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
		"app":                          FlagdProxyDeploymentName,
		"app.kubernetes.io/name":       FlagdProxyDeploymentName,
		"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
		"app.kubernetes.io/version":    f.config.Tag,
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
		ObjectMeta: metav1.ObjectMeta{
			Name:      FlagdProxyDeploymentName,
			Namespace: f.config.Namespace,
			Labels: map[string]string{
				"app":                          FlagdProxyDeploymentName,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    f.config.Tag,
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
				},
			},
		},
	}
}

func (f *FlagdProxyHandler) doesFlagdProxyExist(ctx context.Context) (bool, *appsV1.Deployment, error) {
	d := &appsV1.Deployment{}
	err := f.Client.Get(ctx, client.ObjectKey{Name: FlagdProxyDeploymentName, Namespace: f.config.Namespace}, d)
	if err != nil {
		if errors.IsNotFound(err) {
			// does not exist, is not ready, no error
			return false, nil, nil
		}
		// does not exist, is not ready, is in error
		return false, nil, err
	}
	return true, d, nil
}

func (f *FlagdProxyHandler) shouldUpdateFlagdProxy(old, new *appsV1.Deployment) bool {
	if !common.IsManagedByOFO(old) {
		f.Log.Info("flagd-proxy Deployment not managed by OFO")
		return false
	}
	return !reflect.DeepEqual(old.Spec, new.Spec)
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
