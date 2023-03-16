package v1alpha2

import (
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha2/common"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v2"
)

func TestFeatureFlagConfiguration_ConvertFrom(t *testing.T) {
	tests := []struct {
		name    string
		srcObj  *v1alpha1.FeatureFlagConfiguration
		wantErr bool
		wantObj *FeatureFlagConfiguration
	}{
		{
			name: "Test that conversion from v1alpha1 to v1alpha2 works",
			srcObj: &v1alpha1.FeatureFlagConfiguration{
				TypeMeta: v1.TypeMeta{
					Kind:       "FeatureFlagConfiguration",
					APIVersion: "core.openfeature.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "FeatureFlagconfig1",
					Namespace: "default",
				},
				Spec: v1alpha1.FeatureFlagConfigurationSpec{
					ServiceProvider: &v1alpha1.FeatureFlagServiceProvider{
						Name: "name1",
						Credentials: &corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod1",
						},
					},
					SyncProvider: &v1alpha1.FeatureFlagSyncProvider{
						Name: "syncprovider1",
						HttpSyncConfiguration: &v1alpha1.HttpSyncConfiguration{
							Target:      "target1",
							BearerToken: "token",
						},
					},
					FlagDSpec: &v1alpha1.FlagDSpec{
						MetricsPort: 22,
						Envs: []corev1.EnvVar{
							{
								Name:  "var1",
								Value: "val1",
							},
							{
								Name:  "var2",
								Value: "val2",
							},
						},
					},
					FeatureFlagSpec: `{"flags":{"flag1":{"state":"ok","variants":"variant1","defaultVariant":"default"}}}`,
				},
			},
			wantErr: false,
			wantObj: &FeatureFlagConfiguration{
				ObjectMeta: v1.ObjectMeta{
					Name:      "FeatureFlagconfig1",
					Namespace: "default",
				},
				Spec: FeatureFlagConfigurationSpec{
					ServiceProvider: &FeatureFlagServiceProvider{
						Name: "name1",
						Credentials: &corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod1",
						},
					},
					SyncProvider: &FeatureFlagSyncProvider{
						Name: "syncprovider1",
						HttpSyncConfiguration: &HttpSyncConfiguration{
							Target:      "target1",
							BearerToken: "token",
						},
					},
					FlagDSpec: &FlagDSpec{
						Envs: []corev1.EnvVar{
							{
								Name:  "var1",
								Value: "val1",
							},
							{
								Name:  "var2",
								Value: "val2",
							},
						},
					},
					FeatureFlagSpec: FeatureFlagSpec{
						Flags: map[string]FlagSpec{
							"flag1": {
								State:          "ok",
								DefaultVariant: "default",
								Variants:       []byte(`"variant1"`),
							},
						},
					},
				},
			},
		},
		{
			name: "unable to unmarshal featureflagspec json",
			srcObj: &v1alpha1.FeatureFlagConfiguration{
				TypeMeta: v1.TypeMeta{
					Kind:       "FeatureFlagConfiguration",
					APIVersion: "core.openfeature.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "FeatureFlagconfig1",
					Namespace: "default",
				},
				Spec: v1alpha1.FeatureFlagConfigurationSpec{
					ServiceProvider: &v1alpha1.FeatureFlagServiceProvider{
						Name: "name1",
						Credentials: &corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod1",
						},
					},
					SyncProvider: &v1alpha1.FeatureFlagSyncProvider{
						Name: "syncprovider1",
						HttpSyncConfiguration: &v1alpha1.HttpSyncConfiguration{
							Target:      "target1",
							BearerToken: "token",
						},
					},
					FlagDSpec: &v1alpha1.FlagDSpec{
						MetricsPort: 22,
						Envs: []corev1.EnvVar{
							{
								Name:  "var1",
								Value: "val1",
							},
							{
								Name:  "var2",
								Value: "val2",
							},
						},
					},
					FeatureFlagSpec: `invalid`,
				},
			},
			wantErr: true,
			wantObj: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &FeatureFlagConfiguration{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       FeatureFlagConfigurationSpec{},
				Status:     FeatureFlagConfigurationStatus{},
			}
			if err := dst.ConvertFrom(tt.srcObj); (err != nil) != tt.wantErr {
				t.Errorf("ConvertFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantObj != nil {
				require.Equal(t, tt.wantObj, dst, "Object was not converted correctly")
			}
		})
	}
}

func TestFeatureFlagConfiguration_ConvertTo(t *testing.T) {
	tests := []struct {
		name    string
		src     *FeatureFlagConfiguration
		wantErr bool
		wantObj *v1alpha1.FeatureFlagConfiguration
	}{
		{
			name: "Test that conversion from v1alpha2 to v1alpha1 works",
			src: &FeatureFlagConfiguration{
				TypeMeta: v1.TypeMeta{
					Kind:       "FeatureFlagConfiguration",
					APIVersion: "core.openfeature.dev/v1alpha2",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "FeatureFlagconfig1",
					Namespace: "default",
				},
				Spec: FeatureFlagConfigurationSpec{
					ServiceProvider: &FeatureFlagServiceProvider{
						Name: "name1",
						Credentials: &corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod1",
						},
					},
					SyncProvider: &FeatureFlagSyncProvider{
						Name: "syncprovider1",
						HttpSyncConfiguration: &HttpSyncConfiguration{
							Target:      "target1",
							BearerToken: "token",
						},
					},
					FlagDSpec: &FlagDSpec{
						Envs: []corev1.EnvVar{
							{
								Name:  "var1",
								Value: "val1",
							},
							{
								Name:  "var2",
								Value: "val2",
							},
						},
					},
					FeatureFlagSpec: FeatureFlagSpec{
						Flags: map[string]FlagSpec{
							"flag1": {
								State:          "ok",
								DefaultVariant: "default",
								Variants:       []byte(`"variant1"`),
							},
						},
					},
				},
			},
			wantErr: false,
			wantObj: &v1alpha1.FeatureFlagConfiguration{
				ObjectMeta: v1.ObjectMeta{
					Name:      "FeatureFlagconfig1",
					Namespace: "default",
				},
				Spec: v1alpha1.FeatureFlagConfigurationSpec{
					ServiceProvider: &v1alpha1.FeatureFlagServiceProvider{
						Name: "name1",
						Credentials: &corev1.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod1",
						},
					},
					SyncProvider: &v1alpha1.FeatureFlagSyncProvider{
						Name: "syncprovider1",
						HttpSyncConfiguration: &v1alpha1.HttpSyncConfiguration{
							Target:      "target1",
							BearerToken: "token",
						},
					},
					FlagDSpec: &v1alpha1.FlagDSpec{
						Envs: []corev1.EnvVar{
							{
								Name:  "var1",
								Value: "val1",
							},
							{
								Name:  "var2",
								Value: "val2",
							},
						},
					},
					FeatureFlagSpec: `{"flags":{"flag1":{"state":"ok","variants":"variant1","defaultVariant":"default"}}}`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := v1alpha1.FeatureFlagConfiguration{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       v1alpha1.FeatureFlagConfigurationSpec{},
				Status:     v1alpha1.FeatureFlagConfigurationStatus{},
			}
			if err := tt.src.ConvertTo(&dst); (err != nil) != tt.wantErr {
				t.Errorf("ConvertTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantObj != nil {
				require.Equal(t, tt.wantObj, &dst, "Object was not converted correctly")
			}
		})
	}
}

func TestFeatureFlagConfiguration_ConvertFrom_Errorcase(t *testing.T) {
	// A random different object is used here to simulate a different API version
	testObj := v2.ExternalJob{}

	dst := &FeatureFlagConfiguration{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{},
		Spec:       FeatureFlagConfigurationSpec{},
		Status:     FeatureFlagConfigurationStatus{},
	}

	if err := dst.ConvertFrom(&testObj); err == nil {
		t.Errorf("ConvertFrom() error = %v", err)
	} else {
		require.ErrorIs(t, err, common.ErrCannotCastFeatureFlagConfiguration)
	}
}

func TestFeatureFlagConfiguration_ConvertTo_Errorcase(t *testing.T) {
	testObj := FeatureFlagConfiguration{}

	// A random different object is used here to simulate a different API version
	dst := v2.ExternalJob{}

	if err := testObj.ConvertTo(&dst); err == nil {
		t.Errorf("ConvertTo() error = %v", err)
	} else {
		require.ErrorIs(t, err, common.ErrCannotCastFeatureFlagConfiguration)
	}
}
