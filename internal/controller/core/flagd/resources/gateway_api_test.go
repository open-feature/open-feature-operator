package resources

import (
	"context"
	"testing"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayApiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var GatewayApiGroup = gatewayApiv1.Group("gateway.networking.k8s.io")
var GatewayKind = gatewayApiv1.Kind("Gateway")
var GatewayNamespace = gatewayApiv1.Namespace("my-gateway-namespace")
var GatewayName = gatewayApiv1.ObjectName("my-gateway")

func int32Ptr(i int32) *int32 {
	return &i
}

func TestFlagdGatewayApiHttpRoute_getHttpRoute(t *testing.T) {
	r := FlagdGatewayApiHttpRoute{
		FlagdConfig: testFlagdConfig,
	}

	routeResult, err := r.GetResource(context.TODO(), &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			GatewayApiRoutes: api.GatewayApiSpec{
				Enabled: true,
				Annotations: map[string]string{
					"foo": "bar",
				},
				Hosts: []string{
					"flagd.test",
					"flagd.service",
				},
				ParentRefs: []gatewayApiv1.ParentReference{
					{
						Group:     &GatewayApiGroup,
						Kind:      &GatewayKind,
						Namespace: &GatewayNamespace,
						Name:      GatewayName,
					},
				},
			},
		},
	})

	require.Nil(t, err)

	require.NotNil(t, routeResult)
	require.Equal(t, gatewayApiv1.HTTPRouteSpec{
		CommonRouteSpec: gatewayApiv1.CommonRouteSpec{
			ParentRefs: []gatewayApiv1.ParentReference{
				{
					Group:     (*gatewayApiv1.Group)(strPtr("gateway.networking.k8s.io")),
					Kind:      (*gatewayApiv1.Kind)(strPtr("Gateway")),
					Namespace: (*gatewayApiv1.Namespace)(strPtr("my-gateway-namespace")),
					Name:      "my-gateway",
				},
			},
		},
		Hostnames: []gatewayApiv1.Hostname{"flagd.test", "flagd.service"},
		Rules: []gatewayApiv1.HTTPRouteRule{
			{
				Matches: []gatewayApiv1.HTTPRouteMatch{
					{
						Path: &gatewayApiv1.HTTPPathMatch{
							Type:  (*gatewayApiv1.PathMatchType)(strPtr("PathPrefix")),
							Value: strPtr(common.OFREPHttpServicePath),
						},
					},
				},
				BackendRefs: []gatewayApiv1.HTTPBackendRef{
					{
						BackendRef: gatewayApiv1.BackendRef{
							BackendObjectReference: gatewayApiv1.BackendObjectReference{
								Kind:      (*gatewayApiv1.Kind)(strPtr("Service")),
								Name:      "my-flagd",
								Namespace: (*gatewayApiv1.Namespace)(strPtr("my-namespace")),
								Port:      (*gatewayApiv1.PortNumber)(int32Ptr(8016)),
							},
						},
					},
				},
			},
			{
				Matches: []gatewayApiv1.HTTPRouteMatch{
					{
						Path: &gatewayApiv1.HTTPPathMatch{
							Type:  (*gatewayApiv1.PathMatchType)(strPtr("PathPrefix")),
							Value: strPtr(common.FlagdGrpcServicePath),
						},
					},
				},
				BackendRefs: []gatewayApiv1.HTTPBackendRef{
					{
						BackendRef: gatewayApiv1.BackendRef{
							BackendObjectReference: gatewayApiv1.BackendObjectReference{
								Kind:      (*gatewayApiv1.Kind)(strPtr("Service")),
								Name:      "my-flagd",
								Namespace: (*gatewayApiv1.Namespace)(strPtr("my-namespace")),
								Port:      (*gatewayApiv1.PortNumber)(int32Ptr(8013)),
							},
						},
					},
				},
			},
			{
				Matches: []gatewayApiv1.HTTPRouteMatch{
					{
						Path: &gatewayApiv1.HTTPPathMatch{
							Type:  (*gatewayApiv1.PathMatchType)(strPtr("PathPrefix")),
							Value: strPtr(common.SyncGrpcServicePath),
						},
					},
				},
				BackendRefs: []gatewayApiv1.HTTPBackendRef{
					{
						BackendRef: gatewayApiv1.BackendRef{
							BackendObjectReference: gatewayApiv1.BackendObjectReference{
								Kind:      (*gatewayApiv1.Kind)(strPtr("Service")),
								Name:      "my-flagd",
								Namespace: (*gatewayApiv1.Namespace)(strPtr("my-namespace")),
								Port:      (*gatewayApiv1.PortNumber)(int32Ptr(8015)),
							},
						},
					},
				},
			},
		},
	},
		routeResult.(*gatewayApiv1.HTTPRoute).Spec)
}
