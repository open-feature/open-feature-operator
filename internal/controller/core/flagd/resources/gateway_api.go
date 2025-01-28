package resources

import (
	"context"
	"reflect"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/controller/core/flagd/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayApiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type FlagdGatewayApiHttpRoute struct {
	FlagdConfig resources.FlagdConfiguration
}

func (r FlagdGatewayApiHttpRoute) AreObjectsEqual(o1 client.Object, o2 client.Object) bool {
	oldGateway, ok := o1.(*gatewayApiv1.HTTPRoute)
	if !ok {
		return false
	}

	newGateway, ok := o2.(*gatewayApiv1.HTTPRoute)
	if !ok {
		return false
	}

	return reflect.DeepEqual(oldGateway.Spec, newGateway.Spec)
}

func (r FlagdGatewayApiHttpRoute) GetResource(_ context.Context, flagd *api.Flagd) (client.Object, error) {
	return &gatewayApiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flagd.Name,
			Namespace: flagd.Namespace,
			Labels: map[string]string{
				"app":                          flagd.Name,
				"app.kubernetes.io/name":       flagd.Name,
				"app.kubernetes.io/managed-by": common.ManagedByAnnotationValue,
				"app.kubernetes.io/version":    r.FlagdConfig.Tag,
			},
			Annotations: flagd.Spec.GatewayApiRoutes.Annotations,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: flagd.APIVersion,
				Kind:       flagd.Kind,
				Name:       flagd.Name,
				UID:        flagd.UID,
			}},
		},
		Spec: gatewayApiv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayApiv1.CommonRouteSpec{
				ParentRefs: flagd.Spec.GatewayApiRoutes.ParentRefs,
			},
			Hostnames: getGatewayHostnames(flagd.Spec.GatewayApiRoutes.Hosts),
			Rules:     r.getRules(flagd),
		},
	}, nil
}

func (r FlagdGatewayApiHttpRoute) getRules(flagd *api.Flagd) []gatewayApiv1.HTTPRouteRule {
	pathTypePrefix := gatewayApiv1.PathMatchPathPrefix

	ofrepPathPrefix := common.OFREPHttpServicePath
	flagdServicePathPrefix := common.FlagdGrpcServicePath
	syncServicePathPrefix := common.SyncGrpcServicePath

	serviceKind := gatewayApiv1.Kind("Service")
	serviceNamespace := gatewayApiv1.Namespace(flagd.Namespace)
	serviceName := gatewayApiv1.ObjectName(flagd.Name)

	ofrepPort := gatewayApiv1.PortNumber(r.FlagdConfig.OFREPPort)
	flagdPort := gatewayApiv1.PortNumber(r.FlagdConfig.FlagdPort)
	syncPort := gatewayApiv1.PortNumber(r.FlagdConfig.SyncPort)

	return []gatewayApiv1.HTTPRouteRule{
		{
			Matches: []gatewayApiv1.HTTPRouteMatch{
				{
					Path: &gatewayApiv1.HTTPPathMatch{
						Type:  &pathTypePrefix,
						Value: &ofrepPathPrefix,
					},
				},
			},
			BackendRefs: []gatewayApiv1.HTTPBackendRef{
				{
					BackendRef: gatewayApiv1.BackendRef{
						BackendObjectReference: gatewayApiv1.BackendObjectReference{
							Kind:      &serviceKind,
							Namespace: &serviceNamespace,
							Name:      serviceName,
							Port:      &ofrepPort,
						},
					},
				},
			},
		},
		// The flagd and sync service could be served in a GRPC route but as we use the GRPC gateway for these functionalities too,
		// it is preferred to use a simple HTTP route:
		// https://gateway-api.sigs.k8s.io/api-types/grpcroute/#cross-serving
		{
			Matches: []gatewayApiv1.HTTPRouteMatch{
				{
					Path: &gatewayApiv1.HTTPPathMatch{
						Type:  &pathTypePrefix,
						Value: &flagdServicePathPrefix,
					},
				},
			},
			BackendRefs: []gatewayApiv1.HTTPBackendRef{
				{
					BackendRef: gatewayApiv1.BackendRef{
						BackendObjectReference: gatewayApiv1.BackendObjectReference{
							Kind:      &serviceKind,
							Namespace: &serviceNamespace,
							Name:      serviceName,
							Port:      &flagdPort,
						},
					},
				},
			},
		},
		{
			Matches: []gatewayApiv1.HTTPRouteMatch{
				{
					Path: &gatewayApiv1.HTTPPathMatch{
						Type:  &pathTypePrefix,
						Value: &syncServicePathPrefix,
					},
				},
			},
			BackendRefs: []gatewayApiv1.HTTPBackendRef{
				{
					BackendRef: gatewayApiv1.BackendRef{
						BackendObjectReference: gatewayApiv1.BackendObjectReference{
							Kind:      &serviceKind,
							Namespace: &serviceNamespace,
							Name:      serviceName,
							Port:      &syncPort,
						},
					},
				},
			},
		},
	}
}

func getGatewayHostnames(hosts []string) []gatewayApiv1.Hostname {
	if hosts == nil {
		return nil
	}

	hostnames := make([]gatewayApiv1.Hostname, len(hosts))
	for i, host := range hosts {
		hostnames[i] = gatewayApiv1.Hostname(host)
	}
	return hostnames
}
