package v1beta1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ClusterFeatureFlagSource_EmbeddedSpec(t *testing.T) {
	cffs := &ClusterFeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared",
		},
		Spec: ClusterFeatureFlagSourceSpec{
			FeatureFlagSourceSpec: FeatureFlagSourceSpec{
				ManagementPort: 8014,
				Port:           8013,
				Evaluator:      "json",
				Sources: []Source{
					{
						Source:   "tenant-a/foo",
						Provider: common.SyncProviderKubernetes,
					},
				},
				EnvVarPrefix: "FLAGD",
				LogFormat:    "json",
			},
		},
	}

	// embedded fields are accessible via the outer spec
	require.Equal(t, int32(8013), cffs.Spec.Port)
	require.Equal(t, "tenant-a/foo", cffs.Spec.Sources[0].Source)
}

func Test_ClusterFeatureFlagSource_MethodPromotion(t *testing.T) {
	// Merge is defined on FeatureFlagSourceSpec; must be reachable on
	// ClusterFeatureFlagSourceSpec via embedding
	cffs := ClusterFeatureFlagSourceSpec{
		FeatureFlagSourceSpec: FeatureFlagSourceSpec{
			EnvVarPrefix: "OLD",
			Port:         1000,
		},
	}
	newSpec := &FeatureFlagSourceSpec{
		EnvVarPrefix: "NEW",
		Port:         2000,
	}
	cffs.Merge(newSpec)
	require.Equal(t, "NEW", cffs.EnvVarPrefix)
	require.Equal(t, int32(2000), cffs.Port)
}

func Test_ClusterFeatureFlagSource_DeepCopy(t *testing.T) {
	orig := &ClusterFeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared",
		},
		Spec: ClusterFeatureFlagSourceSpec{
			FeatureFlagSourceSpec: FeatureFlagSourceSpec{
				Port: 8013,
				Sources: []Source{
					{Source: "ns-a/foo"},
				},
				EnvVars: []v1.EnvVar{
					{Name: "FOO", Value: "bar"},
				},
			},
		},
	}
	clone := orig.DeepCopy()
	require.Equal(t, orig, clone)

	// mutating clone must not affect original
	clone.Spec.Sources[0].Source = "mutated/x"
	clone.Spec.EnvVars[0].Value = "changed"
	require.Equal(t, "ns-a/foo", orig.Spec.Sources[0].Source)
	require.Equal(t, "bar", orig.Spec.EnvVars[0].Value)
}

func Test_ClusterFeatureFlagSourceList_DeepCopy(t *testing.T) {
	list := &ClusterFeatureFlagSourceList{
		Items: []ClusterFeatureFlagSource{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "one"},
				Spec: ClusterFeatureFlagSourceSpec{
					FeatureFlagSourceSpec: FeatureFlagSourceSpec{Port: 1},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "two"},
				Spec: ClusterFeatureFlagSourceSpec{
					FeatureFlagSourceSpec: FeatureFlagSourceSpec{Port: 2},
				},
			},
		},
	}
	clone := list.DeepCopy()
	require.Equal(t, list, clone)
	clone.Items[0].Spec.Port = 999
	require.Equal(t, int32(1), list.Items[0].Spec.Port)
}
