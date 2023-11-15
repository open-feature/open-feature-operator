package v1alpha3

import (
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha3/common"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v2"
)

func TestFlagSourceConfiguration_ConvertFrom(t *testing.T) {
	tt := true
	tests := []struct {
		name    string
		srcObj  *v1alpha1.FlagSourceConfiguration
		wantErr bool
		wantObj *FlagSourceConfiguration
	}{
		{
			name: "Test that conversion from v1alpha1 to v1alpha3 works",
			srcObj: &v1alpha1.FlagSourceConfiguration{
				TypeMeta: v1.TypeMeta{
					Kind:       "FlagSourceConfiguration",
					APIVersion: "core.openfeature.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "flagsourceconfig1",
					Namespace: "default",
				},
				Spec: v1alpha1.FlagSourceConfigurationSpec{
					MetricsPort: 20,
					Port:        21,
					SocketPath:  "path",
					Evaluator:   "eval",
					Image:       "img",
					Tag:         "tag",
					Sources: []v1alpha1.Source{
						{
							Source:              "source",
							Provider:            "provider",
							HttpSyncBearerToken: "token",
							TLS:                 true,
							CertPath:            "etc/cert.ca",
							ProviderID:          "app",
							Selector:            "source=database",
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "env1",
							Value: "val1",
						},
						{
							Name:  "env2",
							Value: "val2",
						},
					},
					DefaultSyncProvider: v1alpha1.SyncProviderType("provider"),
					LogFormat:           "log",
					EnvVarPrefix:        "pre",
					RolloutOnChange:     &tt,
					DebugLogging:        common.FalseVal(),
					OtelCollectorUri:    "",
				},
			},
			wantErr: false,
			wantObj: &FlagSourceConfiguration{
				ObjectMeta: v1.ObjectMeta{
					Name:      "flagsourceconfig1",
					Namespace: "default",
				},
				Spec: FlagSourceConfigurationSpec{
					MetricsPort: 20,
					Port:        21,
					SocketPath:  "path",
					Evaluator:   "eval",
					Image:       "img",
					Tag:         "tag",
					Sources: []Source{
						{
							Source:              "source",
							Provider:            "provider",
							HttpSyncBearerToken: "token",
							TLS:                 true,
							CertPath:            "etc/cert.ca",
							ProviderID:          "app",
							Selector:            "source=database",
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "env1",
							Value: "val1",
						},
						{
							Name:  "env2",
							Value: "val2",
						},
					},
					DefaultSyncProvider: "provider",
					LogFormat:           "log",
					EnvVarPrefix:        "pre",
					RolloutOnChange:     &tt,
					DebugLogging:        common.FalseVal(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &FlagSourceConfiguration{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       FlagSourceConfigurationSpec{},
				Status:     FlagSourceConfigurationStatus{},
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

func TestFlagSourceConfiguration_ConvertTo(t *testing.T) {
	tt := true
	tests := []struct {
		name    string
		src     *FlagSourceConfiguration
		wantErr bool
		wantObj *v1alpha1.FlagSourceConfiguration
	}{
		{
			name: "Test that conversion from v1alpha3 to v1alpha1 works",
			src: &FlagSourceConfiguration{
				TypeMeta: v1.TypeMeta{
					Kind:       "FlagSourceConfiguration",
					APIVersion: "core.openfeature.dev/v1alpha3",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "flagsourceconfig1",
					Namespace: "default",
				},
				Spec: FlagSourceConfigurationSpec{
					MetricsPort: 20,
					Port:        21,
					SocketPath:  "path",
					Evaluator:   "eval",
					Image:       "img",
					Tag:         "tag",
					Sources: []Source{
						{
							Source:              "source",
							Provider:            "provider",
							HttpSyncBearerToken: "token",
							TLS:                 false,
							CertPath:            "etc/cert.ca",
							ProviderID:          "app",
							Selector:            "source=database",
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "env1",
							Value: "val1",
						},
						{
							Name:  "env2",
							Value: "val2",
						},
					},
					DefaultSyncProvider: "provider",
					LogFormat:           "log",
					EnvVarPrefix:        "pre",
					RolloutOnChange:     &tt,
					DebugLogging:        common.FalseVal(),
				},
			},
			wantErr: false,
			wantObj: &v1alpha1.FlagSourceConfiguration{
				ObjectMeta: v1.ObjectMeta{
					Name:      "flagsourceconfig1",
					Namespace: "default",
				},
				Spec: v1alpha1.FlagSourceConfigurationSpec{
					MetricsPort: 20,
					Port:        21,
					SocketPath:  "path",
					Evaluator:   "eval",
					Image:       "img",
					Tag:         "tag",
					Sources: []v1alpha1.Source{
						{
							Source:              "source",
							Provider:            "provider",
							HttpSyncBearerToken: "token",
							TLS:                 false,
							CertPath:            "etc/cert.ca",
							ProviderID:          "app",
							Selector:            "source=database",
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "env1",
							Value: "val1",
						},
						{
							Name:  "env2",
							Value: "val2",
						},
					},
					DefaultSyncProvider: v1alpha1.SyncProviderType("provider"),
					LogFormat:           "log",
					EnvVarPrefix:        "pre",
					RolloutOnChange:     &tt,
					DebugLogging:        common.FalseVal(),
					OtelCollectorUri:    "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := v1alpha1.FlagSourceConfiguration{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       v1alpha1.FlagSourceConfigurationSpec{},
				Status:     v1alpha1.FlagSourceConfigurationStatus{},
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

func TestFlagSourceConfiguration_ConvertFrom_Errorcase(t *testing.T) {
	// A random different object is used here to simulate a different API version
	testObj := v2.ExternalJob{}

	dst := &FlagSourceConfiguration{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{},
		Spec:       FlagSourceConfigurationSpec{},
		Status:     FlagSourceConfigurationStatus{},
	}

	if err := dst.ConvertFrom(&testObj); err == nil {
		t.Errorf("ConvertFrom() error = %v", err)
	} else {
		require.ErrorIs(t, err, common.ErrCannotCastFlagSourceConfiguration)
	}
}

func TestFlagSourceConfiguration_ConvertTo_Errorcase(t *testing.T) {
	testObj := FlagSourceConfiguration{}

	// A random different object is used here to simulate a different API version
	dst := v2.ExternalJob{}

	if err := testObj.ConvertTo(&dst); err == nil {
		t.Errorf("ConvertTo() error = %v", err)
	} else {
		require.ErrorIs(t, err, common.ErrCannotCastFlagSourceConfiguration)
	}
}
