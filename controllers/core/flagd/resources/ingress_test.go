// nolint:dupl
package resources

import (
	"context"
	"testing"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestFlagdIngress_getIngress(t *testing.T) {
	r := FlagdIngress{
		FlagdConfig: testFlagdConfig,
	}

	ingressResult, err := r.GetResource(context.TODO(), &api.Flagd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flagd",
			Namespace: "my-namespace",
		},
		Spec: api.FlagdSpec{
			Ingress: api.IngressSpec{
				Enabled: true,
				Annotations: map[string]string{
					"foo": "bar",
				},
				Hosts: []string{
					"flagd.test",
					"flagd.service",
				},
				TLS: []networkingv1.IngressTLS{
					{
						Hosts: []string{
							"flagd.test",
							"flagd.service",
						},
						SecretName: "my-secret",
					},
				},
				IngressClassName: strPtr("nginx"),
			},
		},
	})

	require.Nil(t, err)

	pathType := networkingv1.PathTypePrefix

	require.NotNil(t, ingressResult)
	require.Equal(t, networkingv1.IngressSpec{
		IngressClassName: strPtr("nginx"),
		TLS: []networkingv1.IngressTLS{
			{
				Hosts: []string{
					"flagd.test",
					"flagd.service",
				},
				SecretName: "my-secret",
			},
		},
		Rules: []networkingv1.IngressRule{
			{
				Host: "flagd.test",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/flagd",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
										Port: networkingv1.ServiceBackendPort{
											Number: int32(testFlagdConfig.FlagdPort),
										},
									},
									Resource: nil,
								},
							},
							{
								Path:     "/ofrep",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
										Port: networkingv1.ServiceBackendPort{
											Number: int32(r.FlagdConfig.OFREPPort),
										},
									},
									Resource: nil,
								},
							},
							{
								Path:     "/sync",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
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
			},
			{
				Host: "flagd.service",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/flagd",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
										Port: networkingv1.ServiceBackendPort{
											Number: int32(testFlagdConfig.FlagdPort),
										},
									},
									Resource: nil,
								},
							},
							{
								Path:     "/ofrep",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
										Port: networkingv1.ServiceBackendPort{
											Number: int32(r.FlagdConfig.OFREPPort),
										},
									},
									Resource: nil,
								},
							},
							{
								Path:     "/sync",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "my-flagd",
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
			},
		},
	}, ingressResult.(*networkingv1.Ingress).Spec)

}

func strPtr(s string) *string {
	return &s
}

func Test_areIngressesEqual(t *testing.T) {
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
				old: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("nginx"),
					},
				},
				new: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("kong"),
					},
				},
			},
			want: false,
		},
		{
			name: "has not changed",
			args: args{
				old: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("nginx"),
					},
				},
				new: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("nginx"),
					},
				},
			},
			want: true,
		},
		{
			name: "old is not a service",
			args: args{
				old: &v1.ConfigMap{},
				new: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("nginx"),
					},
				},
			},
			want: false,
		},
		{
			name: "new is not a service",
			args: args{
				old: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						IngressClassName: strPtr("nginx"),
					},
				},
				new: &v1.ConfigMap{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			i := &FlagdIngress{}
			got := i.AreObjectsEqual(tt.args.old, tt.args.new)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_getFlagdPath(t *testing.T) {
	type args struct {
		i api.IngressSpec
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default path",
			args: args{
				i: api.IngressSpec{},
			},
			want: defaultFlagdPath,
		},
		{
			name: "custom path",
			args: args{
				i: api.IngressSpec{
					FlagdPath: "my-path",
				},
			},
			want: "my-path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFlagdPath(tt.args.i); got != tt.want {
				t.Errorf("getFlagdPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOFREPPath(t *testing.T) {
	type args struct {
		i api.IngressSpec
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default path",
			args: args{
				i: api.IngressSpec{},
			},
			want: defaultOFREPPath,
		},
		{
			name: "custom path",
			args: args{
				i: api.IngressSpec{
					OFREPPath: "my-path",
				},
			},
			want: "my-path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOFREPPath(tt.args.i); got != tt.want {
				t.Errorf("getOFREPPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSyncPath(t *testing.T) {
	type args struct {
		i api.IngressSpec
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default path",
			args: args{
				i: api.IngressSpec{},
			},
			want: defaultSyncPath,
		},
		{
			name: "custom path",
			args: args{
				i: api.IngressSpec{
					SyncPath: "my-path",
				},
			},
			want: "my-path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSyncPath(tt.args.i); got != tt.want {
				t.Errorf("getSyncPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
