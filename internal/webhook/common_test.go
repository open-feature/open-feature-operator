package webhook

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/types"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestOpenFeatureEnabledAnnotationIndex(t *testing.T) {

	tests := []struct {
		name string
		o    client.Object
		want []string
	}{
		{
			name: "not a pod",
			o:    &corev1.ConfigMap{},
			want: []string{"false"},
		}, {
			name: "no annotations",
			o:    &corev1.Pod{},
			want: []string{"false"},
		}, {
			name: "annotated wrong",
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"test/ann": "nope", "openfeature.dev/allowkubernetessync": "false"}}},
			want: []string{"false"},
		}, {
			name: "annotated with enabled index",
			o:    &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"openfeature.dev/allowkubernetessync": "true"}}},
			want: []string{"true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OpenFeatureEnabledAnnotationIndex(tt.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenFeatureEnabledAnnotationIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodMutator_checkOFEnabled(t *testing.T) {

	tests := []struct {
		name        string
		annotations map[string]string
		want        bool
	}{
		{
			name:        "enabled",
			annotations: map[string]string{fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation): "true"},
			want:        true,
		}, {
			name:        "disabled",
			annotations: map[string]string{fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.EnabledAnnotation): "false"},
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkOFEnabled(tt.annotations); got != tt.want {
				t.Errorf("checkOFEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseList(t *testing.T) {

	tests := []struct {
		name string
		s    string
		want []string
	}{
		{
			name: "empty string",
			s:    "",
			want: []string{},
		}, {
			name: "nice list with spaces",
			s:    "annotation1, annotation2,    annotation4 , annotation3,",
			want: []string{"annotation1", "annotation2", "annotation4", "annotation3"},
		}, {
			name: "list with no spaces",
			s:    "annotation1, annotation2,annotation4, annotation3",
			want: []string{"annotation1", "annotation2", "annotation4", "annotation3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseList(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodMutator_containsK8sProvider(t *testing.T) {

	tests := []struct {
		name    string
		sources []api.Source
		want    bool
	}{
		{
			name:    "empty",
			sources: []api.Source{},
			want:    false,
		},
		{
			name: "non-kubernetes",
			sources: []api.Source{
				{Provider: apicommon.SyncProviderFilepath},
				{Provider: apicommon.SyncProviderGrpc},
			},
			want: false,
		},
		{
			name: "kubernetes",
			sources: []api.Source{
				{Provider: apicommon.SyncProviderFilepath},
				{Provider: apicommon.SyncProviderKubernetes},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsK8sProvider(tt.sources); got != tt.want {
				t.Errorf("containsK8sProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewFeatureFlagSourceSpec(t *testing.T) {
	env := types.EnvConfig{
		SidecarManagementPort: 80,
		SidecarPort:           88,
		SidecarSocketPath:     "socket-path",
		SidecarEvaluator:      "evaluator",
		SidecarProviderArgs:   "arg1,arg2,arg3",
		SidecarSyncProvider:   "kubernetes",
		SidecarEnvVarPrefix:   "pre",
		SidecarLogFormat:      "log",
		SidecarProbesEnabled:  true,
	}

	f := false
	tt := true

	expected := &api.FeatureFlagSourceSpec{
		ManagementPort:      int32(80),
		Port:                int32(88),
		SocketPath:          "socket-path",
		Evaluator:           "evaluator",
		Sources:             []api.Source{},
		EnvVars:             []corev1.EnvVar{},
		SyncProviderArgs:    []string{"arg1", "arg2", "arg3"},
		DefaultSyncProvider: apicommon.SyncProviderKubernetes,
		EnvVarPrefix:        "pre",
		LogFormat:           "log",
		RolloutOnChange:     nil,
		DebugLogging:        &f,
		OtelCollectorUri:    "",
		ProbesEnabled:       &tt,
	}

	require.Equal(t, expected, NewFeatureFlagSourceSpec(env))
}

func Test_shouldUseSidecar(t *testing.T) {
	require.True(t, shouldUseSidecar(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation): "value",
	}))

	require.False(t, shouldUseSidecar(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation): "value",
	}))
}

func Test_resolveNamespacedRef(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		podNS   string
		wantNS  string
		wantN   string
		wantErr error
	}{
		{
			name:   "bare name",
			token:  "foo",
			podNS:  "ns-a",
			wantNS: "ns-a",
			wantN:  "foo",
		},
		{
			name:   "same-namespace explicit form (back-compat)",
			token:  "ns-a/foo",
			podNS:  "ns-a",
			wantNS: "ns-a",
			wantN:  "foo",
		},
		{
			name:    "cross-namespace rejected",
			token:   "ns-b/foo",
			podNS:   "ns-a",
			wantErr: ErrCrossNamespaceRef,
		},
		{
			name:   "empty token yields empty name (matches ParseAnnotation)",
			token:  "",
			podNS:  "ns-a",
			wantNS: "ns-a",
			wantN:  "",
		},
		{
			name:    "leading slash treated as cross-namespace (empty ns != podNS)",
			token:   "/foo",
			podNS:   "ns-a",
			wantErr: ErrCrossNamespaceRef,
		},
		{
			name:    "three-part path: first segment mismatch is cross-namespace",
			token:   "ns-b/foo/bar",
			podNS:   "ns-a",
			wantErr: ErrCrossNamespaceRef,
		},
		{
			name:   "three-part path: first segment matches podNS keeps rest as name",
			token:  "ns-a/foo/bar",
			podNS:  "ns-a",
			wantNS: "ns-a",
			wantN:  "foo/bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, n, err := resolveNamespacedRef(tt.token, tt.podNS)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantNS, ns)
			require.Equal(t, tt.wantN, n)
		})
	}
}

func Test_shouldUseClusterFeatureFlagSource(t *testing.T) {
	require.True(t, shouldUseClusterFeatureFlagSource(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterFeatureFlagSourceAnnotation): "shared",
	}))

	require.False(t, shouldUseClusterFeatureFlagSource(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation): "foo",
	}))

	require.False(t, shouldUseClusterFeatureFlagSource(map[string]string{}))
}

func Test_shouldUseClusterInProcessConfiguration(t *testing.T) {
	require.True(t, shouldUseClusterInProcessConfiguration(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterInProcessConfigurationAnnotation): "shared",
	}))

	require.False(t, shouldUseClusterInProcessConfiguration(map[string]string{
		fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation): "foo",
	}))

	require.False(t, shouldUseClusterInProcessConfiguration(map[string]string{}))
}

func Test_getClusterFeatureFlagSource(t *testing.T) {
	seeded := &api.ClusterFeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{Name: "shared"},
		Spec: api.ClusterFeatureFlagSourceSpec{
			FeatureFlagSourceSpec: api.FeatureFlagSourceSpec{
				Port: 8013,
			},
		},
	}
	m := &PodMutator{Client: NewClient(false, seeded)}

	t.Run("found", func(t *testing.T) {
		got, err := m.getClusterFeatureFlagSource(context.Background(), "shared")
		require.NoError(t, err)
		require.Equal(t, "shared", got.Name)
		require.Equal(t, int32(8013), got.Spec.Port)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := m.getClusterFeatureFlagSource(context.Background(), "missing")
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))
	})
}

func Test_getClusterInProcessConfiguration(t *testing.T) {
	seeded := &api.ClusterInProcessConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "shared"},
		Spec: api.ClusterInProcessConfigurationSpec{
			InProcessConfigurationSpec: api.InProcessConfigurationSpec{
				Port: 8015,
				Host: "localhost",
			},
		},
	}
	m := &PodMutator{Client: NewClient(false, seeded)}

	t.Run("found", func(t *testing.T) {
		got, err := m.getClusterInProcessConfiguration(context.Background(), "shared")
		require.NoError(t, err)
		require.Equal(t, "shared", got.Name)
		require.Equal(t, "localhost", got.Spec.Host)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := m.getClusterInProcessConfiguration(context.Background(), "missing")
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))
	})
}

// sanity check on errors.Is semantics for wrapped ErrCrossNamespaceRef
func Test_ErrCrossNamespaceRef_ErrorsIs(t *testing.T) {
	_, _, err := resolveNamespacedRef("ns-b/foo", "ns-a")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCrossNamespaceRef))
}
