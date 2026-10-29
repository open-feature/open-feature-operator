package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ClusterInProcessConfiguration_EmbeddedSpec(t *testing.T) {
	cipc := &ClusterInProcessConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared",
		},
		Spec: ClusterInProcessConfigurationSpec{
			InProcessConfigurationSpec: InProcessConfigurationSpec{
				Port: 8015,
				Host: "localhost",
				TLS:  false,
			},
		},
	}

	require.Equal(t, int32(8015), cipc.Spec.Port)
	require.Equal(t, "localhost", cipc.Spec.Host)
}

func Test_ClusterInProcessConfiguration_MethodPromotion(t *testing.T) {
	cipc := ClusterInProcessConfigurationSpec{
		InProcessConfigurationSpec: InProcessConfigurationSpec{
			Host: "old-host",
			Port: 1000,
		},
	}
	newSpec := &InProcessConfigurationSpec{
		Host: "new-host",
		Port: 2000,
	}
	cipc.Merge(newSpec)
	require.Equal(t, "new-host", cipc.Host)
	require.Equal(t, int32(2000), cipc.Port)
}

func Test_ClusterInProcessConfiguration_DeepCopy(t *testing.T) {
	orig := &ClusterInProcessConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "shared"},
		Spec: ClusterInProcessConfigurationSpec{
			InProcessConfigurationSpec: InProcessConfigurationSpec{
				Port: 8015,
				Host: "localhost",
				EnvVars: []v1.EnvVar{
					{Name: "FOO", Value: "bar"},
				},
			},
		},
	}
	clone := orig.DeepCopy()
	require.Equal(t, orig, clone)

	clone.Spec.EnvVars[0].Value = "changed"
	require.Equal(t, "bar", orig.Spec.EnvVars[0].Value)
}

func Test_ClusterInProcessConfigurationList_DeepCopy(t *testing.T) {
	list := &ClusterInProcessConfigurationList{
		Items: []ClusterInProcessConfiguration{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "one"},
				Spec: ClusterInProcessConfigurationSpec{
					InProcessConfigurationSpec: InProcessConfigurationSpec{Port: 1},
				},
			},
		},
	}
	clone := list.DeepCopy()
	require.Equal(t, list, clone)
	clone.Items[0].Spec.Port = 999
	require.Equal(t, int32(1), list.Items[0].Spec.Port)
}
