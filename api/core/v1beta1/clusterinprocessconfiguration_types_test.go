package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Test_ClusterInProcessConfiguration_EmbeddedSpec(t *testing.T) {
	cipc := &ClusterInProcessConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared",
		},
		Spec: ClusterInProcessConfigurationSpec{
			InProcessConfigurationSpec: InProcessConfigurationSpec{
				Port: ptr.To(int32(8015)),
				Host: ptr.To("localhost"),
				TLS:  ptr.To(false),
			},
		},
	}

	require.Equal(t, int32(8015), *cipc.Spec.Port)
	require.Equal(t, "localhost", *cipc.Spec.Host)
}

func Test_ClusterInProcessConfiguration_MethodPromotion(t *testing.T) {
	cipc := ClusterInProcessConfigurationSpec{
		InProcessConfigurationSpec: InProcessConfigurationSpec{
			Host: ptr.To("old-host"),
			Port: ptr.To(int32(1000)),
		},
	}
	newSpec := &InProcessConfigurationSpec{
		Host: ptr.To("new-host"),
		Port: ptr.To(int32(2000)),
	}
	cipc.Merge(newSpec)
	require.Equal(t, "new-host", *cipc.Host)
	require.Equal(t, int32(2000), *cipc.Port)
}

func Test_ClusterInProcessConfiguration_DeepCopy(t *testing.T) {
	orig := &ClusterInProcessConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "shared"},
		Spec: ClusterInProcessConfigurationSpec{
			InProcessConfigurationSpec: InProcessConfigurationSpec{
				Port: ptr.To(int32(8015)),
				Host: ptr.To("localhost"),
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
					InProcessConfigurationSpec: InProcessConfigurationSpec{Port: ptr.To(int32(1))},
				},
			},
		},
	}
	clone := list.DeepCopy()
	require.Equal(t, list, clone)
	clone.Items[0].Spec.Port = ptr.To(int32(999))
	require.Equal(t, int32(1), *list.Items[0].Spec.Port)
}
