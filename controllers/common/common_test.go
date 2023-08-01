package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestFlagSourceConfigurationIndex(t *testing.T) {
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
								fmt.Sprintf("openfeature.dev/%s", FlagSourceConfigurationAnnotation): "true",
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
			out := FlagSourceConfigurationIndex(tt.obj)
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
				t.Errorf("podOwnerIsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}
