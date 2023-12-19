package flagdproxy

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/common/types"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ManagedByAnnotationValue     = "open-feature-operator"
	FlagdProxyDeploymentName     = "flagd-proxy"
	FlagdProxyServiceAccountName = "open-feature-operator-flagd-proxy"
	FlagdProxyServiceName        = "flagd-proxy-svc"
	operatorDeploymentName       = "open-feature-operator-controller-manager"
)

type FlagdProxyHandler struct {
	client.Client
	config *FlagdProxyConfiguration
	Log    logr.Logger
}

type FlagdProxyConfiguration struct {
	Port                   int
	ManagementPort         int
	DebugLogging           bool
	Image                  string
	Tag                    string
	Namespace              string
	OperatorDeploymentName string
}

func NewFlagdProxyConfiguration(env types.EnvConfig) *FlagdProxyConfiguration {
	return &FlagdProxyConfiguration{
		Image:                  env.FlagdProxyImage,
		Tag:                    env.FlagdProxyTag,
		Namespace:              env.PodNamespace,
		OperatorDeploymentName: operatorDeploymentName,
		Port:                   env.FlagdProxyPort,
		ManagementPort:         env.FlagdProxyManagementPort,
		DebugLogging:           env.FlagdProxyDebugLogging,
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

func (f *FlagdProxyHandler) HandleFlagdProxy(ctx context.Context) error {
	exists, err := f.doesFlagdProxyExist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return f.deployFlagdProxy(ctx)
	}
	return nil
}

func (f *FlagdProxyHandler) deployFlagdProxy(ctx context.Context) error {
	ownerReferences := []metav1.OwnerReference{}
	ownerReference, err := f.getOwnerReference(ctx)
	if err != nil {
		f.Log.Error(err, "unable to create owner reference for open-feature-operator, not appending")
	} else {
		ownerReferences = append(ownerReferences, ownerReference)
	}

	f.Log.Info("deploying the flagd-proxy")
	if err := f.Client.Create(ctx, f.newFlagdProxyManifest(ownerReferences)); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	f.Log.Info("deploying the flagd-proxy service")
	if err := f.Client.Create(ctx, f.newFlagdProxyServiceManifest(ownerReferences)); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (f *FlagdProxyHandler) newFlagdProxyServiceManifest(ownerReferences []metav1.OwnerReference) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            FlagdProxyServiceName,
			Namespace:       f.config.Namespace,
			OwnerReferences: ownerReferences,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name":       FlagdProxyDeploymentName,
				"app.kubernetes.io/managed-by": ManagedByAnnotationValue,
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

func (f *FlagdProxyHandler) newFlagdProxyManifest(ownerReferences []metav1.OwnerReference) *appsV1.Deployment {
	replicas := int32(1)
	args := []string{
		"start",
		"--management-port",
		fmt.Sprintf("%d", f.config.ManagementPort),
	}
	if f.config.DebugLogging {
		args = append(args, "--debug")
	}
	return &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FlagdProxyDeploymentName,
			Namespace: f.config.Namespace,
			Labels: map[string]string{
				"app":                          FlagdProxyDeploymentName,
				"app.kubernetes.io/managed-by": ManagedByAnnotationValue,
				"app.kubernetes.io/version":    f.config.Tag,
			},
			OwnerReferences: ownerReferences,
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
					Labels: map[string]string{
						"app":                          FlagdProxyDeploymentName,
						"app.kubernetes.io/name":       FlagdProxyDeploymentName,
						"app.kubernetes.io/managed-by": ManagedByAnnotationValue,
						"app.kubernetes.io/version":    f.config.Tag,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: FlagdProxyServiceAccountName,
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

func (f *FlagdProxyHandler) doesFlagdProxyExist(ctx context.Context) (bool, error) {
	d := &appsV1.Deployment{}
	err := f.Client.Get(ctx, client.ObjectKey{Name: FlagdProxyDeploymentName, Namespace: f.config.Namespace}, d)
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

func (f *FlagdProxyHandler) getOwnerReference(ctx context.Context) (metav1.OwnerReference, error) {
	d := &appsV1.Deployment{}
	if err := f.Client.Get(ctx, client.ObjectKey{Name: f.config.OperatorDeploymentName, Namespace: f.config.Namespace}, d); err != nil {
		return metav1.OwnerReference{}, fmt.Errorf("unable to fetch operator deployment to create owner reference: %w", err)
	}
	return metav1.OwnerReference{
		UID:        d.GetUID(),
		Name:       d.GetName(),
		APIVersion: d.APIVersion,
		Kind:       d.Kind,
	}, nil

}
