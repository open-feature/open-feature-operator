package v1alpha1

import (
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1/common"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_FeatureFlagConfiguration(t *testing.T) {
	ffConfig := FeatureFlagConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      "ffconf1",
			Namespace: "test",
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: "ver",
					Kind:       "kind",
					Name:       "ffconf1",
					UID:        types.UID("5"),
					Controller: common.TrueVal(),
				},
			},
		},
		Spec: FeatureFlagConfigurationSpec{
			FeatureFlagSpec: "spec",
		},
	}

	name := "cmname"
	namespace := "cmnamespace"
	references := []v1.OwnerReference{
		{
			APIVersion: "ver",
			Kind:       "kind",
			Name:       "ffconf1",
			UID:        types.UID("5"),
			Controller: common.TrueVal(),
		},
	}

	require.Equal(t, v1.OwnerReference{
		APIVersion: ffConfig.APIVersion,
		Kind:       ffConfig.Kind,
		Name:       ffConfig.Name,
		UID:        ffConfig.UID,
		Controller: common.TrueVal(),
	}, ffConfig.GetReference())

	require.Equal(t, corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/featureflagconfiguration": name,
			},
			OwnerReferences: references,
		},
		Data: map[string]string{
			"cmnamespace_cmname.flagd.json": "spec",
		},
	}, ffConfig.GenerateConfigMap(name, namespace, references))

	require.False(t, ffConfig.Spec.ServiceProvider.IsSet())

	ffConfig.Spec.ServiceProvider = &FeatureFlagServiceProvider{
		Name: "",
	}

	require.False(t, ffConfig.Spec.ServiceProvider.IsSet())

	ffConfig.Spec.ServiceProvider = &FeatureFlagServiceProvider{
		Name: "prov",
	}

	require.True(t, ffConfig.Spec.ServiceProvider.IsSet())
}
