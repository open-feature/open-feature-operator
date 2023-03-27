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
