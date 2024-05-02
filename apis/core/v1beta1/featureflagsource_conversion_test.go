package v1beta1

import (
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/apis/core/v1beta2"
	v1beta2common "github.com/open-feature/open-feature-operator/apis/core/v1beta2/common"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFeatureFlagSource_ConvertFrom(t *testing.T) {
	ttt := true
	ff := false
	tests := []struct {
		name    string
		srcObj  *v1beta2.FeatureFlagSource
		wantErr bool
		wantObj *FeatureFlagSource
	}{
		{
			name: "Test that conversion from v1beta2 to v1beta1 works",
			srcObj: &v1beta2.FeatureFlagSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       "FeatureFlagSource",
					APIVersion: "core.openfeature.dev/v1beta2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ffs",
					Namespace: "",
					Labels: map[string]string{
						"some-label": "some-label-value",
					},
					Annotations: map[string]string{
						"some-annotation": "some-annotation-value",
					},
				},
				Spec: v1beta2.FeatureFlagSourceSpec{
					EnvVarPrefix: "prefix",
					RPC: &v1beta2.RPCConf{
						ManagementPort: int32(9999),
						Port:           int32(8888),
						SocketPath:     "path",
						Evaluator:      "eval",
						Sources: []v1beta2.Source{
							{
								Source:              "source",
								Provider:            "prov",
								HttpSyncBearerToken: "token",
								TLS:                 true,
								CertPath:            "certpath",
								ProviderID:          "id",
								Selector:            "selector",
								Interval:            uint32(5),
							},
						},
						EnvVars: []corev1.EnvVar{
							{
								Name:  "name",
								Value: "val",
							},
							{
								Name:  "name2",
								Value: "val2",
							},
						},
						SyncProviderArgs:    []string{"some", "args"},
						DefaultSyncProvider: v1beta2common.SyncProviderKubernetes,
						LogFormat:           "log",
						RolloutOnChange:     false,
						DebugLogging:        false,
						OtelCollectorUri:    "otel",
						ProbesEnabled:       true,
					},
				},
			},
			wantObj: &FeatureFlagSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ffs",
					Namespace: "",
					Labels: map[string]string{
						"some-label": "some-label-value",
					},
					Annotations: map[string]string{
						"some-annotation": "some-annotation-value",
					},
				},
				Spec: FeatureFlagSourceSpec{
					EnvVarPrefix:   "prefix",
					ManagementPort: int32(9999),
					Port:           int32(8888),
					SocketPath:     "path",
					Evaluator:      "eval",
					Sources: []Source{
						{
							Source:              "source",
							Provider:            "prov",
							HttpSyncBearerToken: "token",
							TLS:                 true,
							CertPath:            "certpath",
							ProviderID:          "id",
							Selector:            "selector",
							Interval:            uint32(5),
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "name",
							Value: "val",
						},
						{
							Name:  "name2",
							Value: "val2",
						},
					},
					SyncProviderArgs:    []string{"some", "args"},
					DefaultSyncProvider: common.SyncProviderType("kubernetes"),
					LogFormat:           "log",
					RolloutOnChange:     &ff,
					DebugLogging:        &ff,
					OtelCollectorUri:    "otel",
					ProbesEnabled:       &ttt,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &FeatureFlagSource{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       FeatureFlagSourceSpec{},
				Status:     FeatureFlagSourceStatus{},
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

func TestFeatureFlagSource_ConvertTo(t *testing.T) {
	ttt := true
	ff := false
	tests := []struct {
		name    string
		srcObj  *FeatureFlagSource
		wantErr bool
		wantObj *v1beta2.FeatureFlagSource
	}{
		{
			name: "Test that conversion from v1beta2 to v1beta1 works",
			wantObj: &v1beta2.FeatureFlagSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ffs",
					Namespace: "",
					Labels: map[string]string{
						"some-label": "some-label-value",
					},
					Annotations: map[string]string{
						"some-annotation": "some-annotation-value",
					},
				},
				Spec: v1beta2.FeatureFlagSourceSpec{
					EnvVarPrefix: "prefix",
					RPC: &v1beta2.RPCConf{
						ManagementPort: int32(9999),
						Port:           int32(8888),
						SocketPath:     "path",
						Evaluator:      "eval",
						Sources: []v1beta2.Source{
							{
								Source:              "source",
								Provider:            "prov",
								HttpSyncBearerToken: "token",
								TLS:                 true,
								CertPath:            "certpath",
								ProviderID:          "id",
								Selector:            "selector",
								Interval:            uint32(5),
							},
						},
						EnvVars: []corev1.EnvVar{
							{
								Name:  "name",
								Value: "val",
							},
							{
								Name:  "name2",
								Value: "val2",
							},
						},
						SyncProviderArgs:    []string{"some", "args"},
						DefaultSyncProvider: v1beta2common.SyncProviderKubernetes,
						LogFormat:           "log",
						RolloutOnChange:     false,
						DebugLogging:        false,
						OtelCollectorUri:    "otel",
						ProbesEnabled:       true,
					},
				},
			},
			srcObj: &FeatureFlagSource{
				TypeMeta: metav1.TypeMeta{
					Kind:       "FeatureFlagSource",
					APIVersion: "core.openfeature.dev/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ffs",
					Namespace: "",
					Labels: map[string]string{
						"some-label": "some-label-value",
					},
					Annotations: map[string]string{
						"some-annotation": "some-annotation-value",
					},
				},
				Spec: FeatureFlagSourceSpec{
					EnvVarPrefix:   "prefix",
					ManagementPort: int32(9999),
					Port:           int32(8888),
					SocketPath:     "path",
					Evaluator:      "eval",
					Sources: []Source{
						{
							Source:              "source",
							Provider:            "prov",
							HttpSyncBearerToken: "token",
							TLS:                 true,
							CertPath:            "certpath",
							ProviderID:          "id",
							Selector:            "selector",
							Interval:            uint32(5),
						},
					},
					EnvVars: []corev1.EnvVar{
						{
							Name:  "name",
							Value: "val",
						},
						{
							Name:  "name2",
							Value: "val2",
						},
					},
					SyncProviderArgs:    []string{"some", "args"},
					DefaultSyncProvider: common.SyncProviderType("kubernetes"),
					LogFormat:           "log",
					RolloutOnChange:     &ff,
					DebugLogging:        &ff,
					OtelCollectorUri:    "otel",
					ProbesEnabled:       &ttt,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &v1beta2.FeatureFlagSource{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1beta2.FeatureFlagSourceSpec{},
				Status:     v1beta2.FeatureFlagSourceStatus{},
			}
			if err := tt.srcObj.ConvertTo(dst); (err != nil) != tt.wantErr {
				t.Errorf("ConvertFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantObj != nil {
				require.Equal(t, tt.wantObj, dst, "Object was not converted correctly")
			}
		})
	}
}
