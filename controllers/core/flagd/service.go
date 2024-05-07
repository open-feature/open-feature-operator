package flagd

import (
	"context"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdService struct {
	client.Client

	Log         logr.Logger
	FlagdConfig FlagdConfiguration

	ResourceReconciler *ResourceReconciler
}

func (r FlagdService) Reconcile(ctx context.Context, flagd *api.Flagd) error {
	return r.ResourceReconciler.Reconcile(
		ctx,
		flagd,
		&v1.Service{},
		func() (client.Object, error) {
			return r.getService(flagd), nil
		},
		func(old client.Object, new client.Object) bool {
			oldService, ok := old.(*v1.Service)
			if !ok {
				return false
			}

			newService, ok := new.(*v1.Service)
			if !ok {
				return false
			}

			return reflect.DeepEqual(oldService.Spec, newService.Spec)
		},
	)
}

func (r FlagdService) getService(flagd *api.Flagd) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flagd.Name,
			Namespace: flagd.Namespace,
			Labels: map[string]string{
				"app":                          flagd.Name,
				"app.kubernetes.io/name":       flagd.Name,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    r.FlagdConfig.Tag,
			},
			Annotations: flagd.Annotations,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: flagd.APIVersion,
				Kind:       flagd.Kind,
				Name:       flagd.Name,
				UID:        flagd.UID,
			}},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": flagd.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name: "flagd",
					Port: int32(r.FlagdConfig.FlagdPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(r.FlagdConfig.FlagdPort),
					},
				},
				{
					Name: "ofrep",
					Port: int32(r.FlagdConfig.OFREPPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(r.FlagdConfig.OFREPPort),
					},
				},
				{
					Name: "sync",
					Port: int32(r.FlagdConfig.SyncPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(r.FlagdConfig.SyncPort),
					},
				},
				{
					Name: "metrics",
					Port: int32(r.FlagdConfig.ManagementPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(r.FlagdConfig.ManagementPort),
					},
				},
			},
			Type: flagd.Spec.ServiceType,
		},
	}
}
