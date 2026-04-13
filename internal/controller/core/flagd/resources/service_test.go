// nolint:dupl
package resources

import (
	"context"
	"testing"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestFlagdService_getService(t *testing.T) {
	r := FlagdService{
		FlagdConfig: testFlagdConfig,
	}

	svc, err := r.GetResource(context.TODO(), &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			ServiceType: "ClusterIP",
		},
	})

	require.Nil(t, err)
	require.NotNil(t, svc)
	require.IsType(t, &v1.Service{}, svc)

	expectedPorts := map[string]struct {
		appProtocol string
		port        int32
	}{
		"flagd":   {"grpc", 8013},
		"ofrep":   {"http", 8016},
		"sync":    {"grpc", 8015},
		"metrics": {"http", 8014},
	}

	ports := svc.(*v1.Service).Spec.Ports
	require.Len(t, ports, len(expectedPorts), "unexpected number of ports")

	for _, port := range ports {
		expected, ok := expectedPorts[port.Name]
		require.True(t, ok, "unexpected port: %s", port.Name)
		require.Equal(t, expected.appProtocol, *port.AppProtocol)
		require.Equal(t, expected.port, port.Port)
		require.Equal(t, intstr.FromInt(int(expected.port)), port.TargetPort)
	}
}

func Test_areServicesEqual(t *testing.T) {
	type args struct {
		old client.Object
		new client.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "has changed",
			args: args{
				old: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeNodePort,
					},
				},
				new: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
			},
			want: false,
		},
		{
			name: "has not changed",
			args: args{
				old: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
				new: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
			},
			want: true,
		},
		{
			name: "old is not a service",
			args: args{
				old: &v1.ConfigMap{},
				new: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
			},
			want: false,
		},
		{
			name: "new is not a service",
			args: args{
				old: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
				new: &v1.ConfigMap{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &FlagdService{}
			got := s.AreObjectsEqual(tt.args.old, tt.args.new)

			require.Equal(t, tt.want, got)
		})
	}
}
