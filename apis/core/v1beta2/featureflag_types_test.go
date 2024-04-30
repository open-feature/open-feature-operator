package v1beta2

import (
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta2/common"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_FeatureFlag(t *testing.T) {
	ff := FeatureFlag{
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
		Spec: FeatureFlagSpec{
			FlagSpec: FlagSpec{
				Flags: map[string]Flag{},
			},
		},
	}

	require.Equal(t, v1.OwnerReference{
		APIVersion: ff.APIVersion,
		Kind:       ff.Kind,
		Name:       ff.Name,
		UID:        ff.UID,
		Controller: common.TrueVal(),
	}, ff.GetReference())

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

	cm, _ := ff.GenerateConfigMap(name, namespace, references)

	require.Equal(t, corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/featureflag": name,
			},
			OwnerReferences: references,
		},
		Data: map[string]string{
			"cmnamespace_cmname.flagd.json": "{\"flags\":{}}",
		},
	}, *cm)
}
