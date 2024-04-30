package common

import (
	"context"
	"fmt"
	"testing"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta2"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFeatureFlagSourceIndex(t *testing.T) {
	tests := []struct {
		name string
		obj  client.Object
		out  []string
	}{
		{
			name: "non-deployment object",
			obj:  &appsV1.DaemonSet{},
			out:  []string{"false"},
		},
		{
			name: "no annotations",
			obj:  &appsV1.Deployment{},
			out:  []string{"false"},
		},
		{
			name: "not existing right annotation",
			obj: &appsV1.Deployment{
				Spec: appsV1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"silly": "some",
							},
						},
					},
				},
			},
			out: []string{"false"},
		},
		{
			name: "existing annotation",
			obj: &appsV1.Deployment{
				Spec: appsV1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("openfeature.dev/%s", FeatureFlagSourceAnnotation): "true",
							},
						},
					},
				},
			},
			out: []string{"true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := FeatureFlagSourceIndex(tt.obj)
			require.Equal(t, tt.out, out)
		})

	}
}

func TestSharedOwnership(t *testing.T) {
	tests := []struct {
		name   string
		owner1 []metav1.OwnerReference
		owner2 []metav1.OwnerReference
		want   bool
	}{{
		name: "same owner uid",
		owner1: []metav1.OwnerReference{
			{
				UID: "12345",
			},
		},
		owner2: []metav1.OwnerReference{
			{
				UID: "12345",
			},
		},
		want: true,
	},
		{
			name:   "pod and cm have different owners",
			owner1: []metav1.OwnerReference{},
			owner2: []metav1.OwnerReference{
				{
					UID: "12345",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SharedOwnership(tt.owner1, tt.owner2); got != tt.want {
				t.Errorf("SharedOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindFlagConfig(t *testing.T) {
	ff := &api.FeatureFlag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	tests := []struct {
		name    string
		ns      string
		obj     *api.FeatureFlag
		want    *api.FeatureFlag
		wantErr bool
	}{
		{
			name:    "test",
			ns:      "default",
			obj:     ff,
			want:    ff,
			wantErr: false,
		},
		{
			name:    "non-existing",
			ns:      "default",
			obj:     ff,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.AddToScheme(scheme.Scheme)
			require.Nil(t, err)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tt.obj).Build()

			got, err := FindFlagConfig(context.TODO(), fakeClient, tt.ns, tt.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindFlagConfig() = expected error %t, got %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				require.Equal(t, tt.want.Name, got.Name)
				require.Equal(t, tt.want.Namespace, got.Namespace)
			}

		})
	}
}
