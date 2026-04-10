package v1beta1

import (
	"encoding/json"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
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
				Flags: Flags{},
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
			"cmnamespace_cmname.flagd.json": "{\"flags\":null}",
		},
	}, *cm)
}

func Test_FeatureFlag_MetadataPreserved(t *testing.T) {
	ff := FeatureFlag{
		ObjectMeta: v1.ObjectMeta{
			Name:      "ff-meta",
			Namespace: "test",
		},
		Spec: FeatureFlagSpec{
			FlagSpec: FlagSpec{
				Flags: Flags{
					FlagsMap: map[string]Flag{
						"color": {
							State:          "ENABLED",
							Variants:       json.RawMessage(`{"red":"red","blue":"blue"}`),
							DefaultVariant: "red",
							Metadata:       json.RawMessage(`{"flagSetId":"set-abc","custom":true}`),
						},
					},
				},
				Metadata: json.RawMessage(`{"flagSetId":"set-123","version":"1.0"}`),
			},
		},
	}

	cm, err := ff.GenerateConfigMap("ff-meta", "test", nil)
	require.NoError(t, err)

	flagData := cm.Data["test_ff-meta.flagd.json"]
	require.NotEmpty(t, flagData)

	// unmarshal and verify metadata survived round-trip
	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(flagData), &parsed))

	// flag-set level metadata
	require.JSONEq(t, `{"flagSetId":"set-123","version":"1.0"}`, string(parsed["metadata"]))

	// per-flag metadata
	var flags map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["flags"], &flags))
	require.JSONEq(t, `{"flagSetId":"set-abc","custom":true}`, string(flags["color"]["metadata"]))
}
