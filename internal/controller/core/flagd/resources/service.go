package resources

import (
	"context"
	"reflect"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/controller/core/flagd/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdService struct {
	FlagdConfig resources.FlagdConfiguration
}

func (r FlagdService) AreObjectsEqual(o1 client.Object, o2 client.Object) bool {
	oldService, ok := o1.(*v1.Service)
	if !ok {
		return false
	}

	newService, ok := o2.(*v1.Service)
	if !ok {
		return false
	}

	return reflect.DeepEqual(oldService.Spec, newService.Spec)
}

func (r FlagdService) GetResource(_ context.Context, flagd *api.Flagd) (client.Object, error) {
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
	}, nil
}
