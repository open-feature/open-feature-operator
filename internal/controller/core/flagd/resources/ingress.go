package resources

import (
	"context"
	"reflect"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/controller/core/flagd/common"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdIngress struct {
	FlagdConfig resources.FlagdConfiguration
}

func (r FlagdIngress) AreObjectsEqual(o1 client.Object, o2 client.Object) bool {
	oldIngress, ok := o1.(*networkingv1.Ingress)
	if !ok {
		return false
	}

	newIngress, ok := o2.(*networkingv1.Ingress)
	if !ok {
		return false
	}

	return reflect.DeepEqual(oldIngress.Spec, newIngress.Spec)
}

func (r FlagdIngress) GetResource(_ context.Context, flagd *api.Flagd) (client.Object, error) {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flagd.Name,
			Namespace: flagd.Namespace,
			Labels: map[string]string{
				"app":                          flagd.Name,
				"app.kubernetes.io/name":       flagd.Name,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    r.FlagdConfig.Tag,
			},
			Annotations: flagd.Spec.Ingress.Annotations,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: flagd.APIVersion,
				Kind:       flagd.Kind,
				Name:       flagd.Name,
				UID:        flagd.UID,
			}},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: flagd.Spec.Ingress.IngressClassName,
			TLS:              flagd.Spec.Ingress.TLS,
			Rules:            r.getRules(flagd),
		},
	}, nil
}

func (r FlagdIngress) getRules(flagd *api.Flagd) []networkingv1.IngressRule {
	rules := make([]networkingv1.IngressRule, len(flagd.Spec.Ingress.Hosts))
	for i, host := range flagd.Spec.Ingress.Hosts {
		rules[i] = r.getRule(flagd, host)
	}
	return rules
}

func (r FlagdIngress) getRule(flagd *api.Flagd, host string) networkingv1.IngressRule {
	pathType := networkingv1.PathTypePrefix
	if flagd.Spec.Ingress.PathType != "" {
		pathType = flagd.Spec.Ingress.PathType
	}
	return networkingv1.IngressRule{
		Host: host,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     getFlagdPath(flagd.Spec.Ingress),
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: flagd.Name,
								Port: networkingv1.ServiceBackendPort{
									Number: int32(r.FlagdConfig.FlagdPort),
								},
							},
							Resource: nil,
						},
					},
					{
						Path:     getOFREPPath(flagd.Spec.Ingress),
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: flagd.Name,
								Port: networkingv1.ServiceBackendPort{
									Number: int32(r.FlagdConfig.OFREPPort),
								},
							},
							Resource: nil,
						},
					},
					{
						Path:     getSyncPath(flagd.Spec.Ingress),
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: flagd.Name,
								Port: networkingv1.ServiceBackendPort{
									Number: int32(r.FlagdConfig.SyncPort),
								},
							},
							Resource: nil,
						},
					},
				},
			},
		},
	}
}

func getFlagdPath(i api.IngressSpec) string {
	path := common.FlagdGrpcServicePath
	if i.FlagdPath != "" {
		path = i.FlagdPath
	}
	return path
}

func getOFREPPath(i api.IngressSpec) string {
	path := common.OFREPHttpServicePath
	if i.OFREPPath != "" {
		path = i.OFREPPath
	}
	return path
}

func getSyncPath(i api.IngressSpec) string {
	path := common.SyncGrpcServicePath
	if i.SyncPath != "" {
		path = i.SyncPath
	}
	return path
}
