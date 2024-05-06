package flagd

import (
	"context"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FlagdIngress struct {
	client.Client

	Log         logr.Logger
	FlagdConfig FlagdConfiguration

	ResourceReconciler *ResourceReconciler
}

func (r FlagdIngress) Reconcile(ctx context.Context, flagd *api.Flagd) (*ctrl.Result, error) {
	return r.ResourceReconciler.Reconcile(
		ctx,
		flagd,
		&networkingv1.Ingress{},
		func() (client.Object, error) {
			return r.getIngress(flagd), nil
		},
		func(old client.Object, new client.Object) bool {
			oldIngress, ok := old.(*networkingv1.Ingress)
			if !ok {
				return false
			}

			newIngress, ok := new.(*networkingv1.Ingress)
			if !ok {
				return false
			}

			return reflect.DeepEqual(oldIngress.Spec, newIngress.Spec)
		},
	)
}

func (r FlagdIngress) getIngress(flagd *api.Flagd) *networkingv1.Ingress {
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
				Name:       flagd.Kind,
				UID:        flagd.UID,
			}},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: flagd.Spec.Ingress.IngressClassName,
			DefaultBackend: &networkingv1.IngressBackend{
				Service:  nil,
				Resource: nil,
			},
			TLS:   flagd.Spec.Ingress.TLS,
			Rules: r.getRules(flagd),
		},
	}
}

func (r FlagdIngress) getRules(flagd *api.Flagd) []networkingv1.IngressRule {
	rules := make([]networkingv1.IngressRule, 2*len(flagd.Spec.Ingress.Hosts))
	for i, host := range flagd.Spec.Ingress.Hosts {
		rules[2*i] = r.getRule(flagd, host, "flagd", int32(r.FlagdConfig.FlagdPort))
		rules[2*i+1] = r.getRule(flagd, host, "ofrep", int32(r.FlagdConfig.OFREPPort))
	}
	return rules
}

func (r FlagdIngress) getRule(flagd *api.Flagd, host, path string, port int32) networkingv1.IngressRule {
	pathType := networkingv1.PathTypePrefix
	return networkingv1.IngressRule{
		Host: host,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: flagd.GetServiceName(),
								Port: networkingv1.ServiceBackendPort{
									Number: port,
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
