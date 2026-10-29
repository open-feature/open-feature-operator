package common

import (
	"context"
	"fmt"
	"testing"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFeatureFlagSourceIndex(t *testing.T) {
	tests := []struct {
		name string
		obj  client.Object
		out  []string
	}{
		{
			name: "non-deployment object",
			obj:  &appsV1.DaemonSet{},
			out:  []string{"false"},
		},
		{
			name: "no annotations",
			obj:  &appsV1.Deployment{},
			out:  []string{"false"},
		},
		{
			name: "not existing right annotation",
			obj: &appsV1.Deployment{
				Spec: appsV1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"silly": "some",
							},
						},
					},
				},
			},
			out: []string{"false"},
		},
		{
			name: "existing annotation",
			obj: &appsV1.Deployment{
				Spec: appsV1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("openfeature.dev/%s", FeatureFlagSourceAnnotation): "true",
							},
						},
					},
				},
			},
			out: []string{"true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := FeatureFlagSourceIndex(tt.obj)
			require.Equal(t, tt.out, out)
		})

	}
}

func TestSharedOwnership(t *testing.T) {
	tests := []struct {
		name   string
		owner1 []metav1.OwnerReference
		owner2 []metav1.OwnerReference
		want   bool
	}{{
		name: "same owner uid",
		owner1: []metav1.OwnerReference{
			{
				UID: "12345",
			},
		},
		owner2: []metav1.OwnerReference{
			{
				UID: "12345",
			},
		},
		want: true,
	},
		{
			name:   "pod and cm have different owners",
			owner1: []metav1.OwnerReference{},
			owner2: []metav1.OwnerReference{
				{
					UID: "12345",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SharedOwnership(tt.owner1, tt.owner2); got != tt.want {
				t.Errorf("SharedOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindFlagConfig(t *testing.T) {
	ff := &api.FeatureFlag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	tests := []struct {
		name    string
		ns      string
		obj     *api.FeatureFlag
		want    *api.FeatureFlag
		wantErr bool
	}{
		{
			name:    "test",
			ns:      "default",
			obj:     ff,
			want:    ff,
			wantErr: false,
		},
		{
			name:    "non-existing",
			ns:      "default",
			obj:     ff,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.AddToScheme(scheme.Scheme)
			require.Nil(t, err)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tt.obj).Build()

			got, err := FindFlagConfig(context.TODO(), fakeClient, tt.ns, tt.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindFlagConfig() = expected error %t, got %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				require.Equal(t, tt.want.Name, got.Name)
				require.Equal(t, tt.want.Namespace, got.Namespace)
			}

		})
	}
}

func TestPinSourceProviders(t *testing.T) {
	spec := &api.FeatureFlagSourceSpec{
		DefaultSyncProvider: apicommon.SyncProviderHttp,
		Sources: []api.Source{
			{Source: "https://example.com/flags.json"},
			{Source: "flags", Provider: apicommon.SyncProviderKubernetes},
		},
	}
	PinSourceProviders(spec, apicommon.SyncProviderKubernetes)
	require.Equal(t, apicommon.SyncProviderHttp, spec.Sources[0].Provider)
	require.Equal(t, apicommon.SyncProviderKubernetes, spec.Sources[1].Provider)

	spec = &api.FeatureFlagSourceSpec{Sources: []api.Source{{Source: "flags"}}}
	PinSourceProviders(spec, apicommon.SyncProviderKubernetes)
	require.Equal(t, apicommon.SyncProviderKubernetes, spec.Sources[0].Provider, "falls back when the spec has no default")
}

func TestValidateNamespacedSources(t *testing.T) {
	tests := []struct {
		name    string
		source  api.Source
		wantErr bool
	}{
		{name: "bare kubernetes name", source: api.Source{Source: "flags", Provider: apicommon.SyncProviderKubernetes}},
		{name: "own-namespace kubernetes", source: api.Source{Source: "ns-a/flags", Provider: apicommon.SyncProviderKubernetes}},
		{name: "other-namespace kubernetes", source: api.Source{Source: "ns-b/flags", Provider: apicommon.SyncProviderKubernetes}, wantErr: true},
		{name: "other-namespace file", source: api.Source{Source: "ns-b/flags", Provider: apicommon.SyncProviderFilepath}, wantErr: true},
		{name: "other-namespace flagd-proxy", source: api.Source{Source: "ns-b/flags", Provider: apicommon.SyncProviderFlagdProxy}, wantErr: true},
		{name: "http URL is not a cluster ref", source: api.Source{Source: "https://example.com/ns-b/flags", Provider: apicommon.SyncProviderHttp}},
		{name: "grpc target is not a cluster ref", source: api.Source{Source: "ns-b/flags", Provider: apicommon.SyncProviderGrpc}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespacedSources([]api.Source{tt.source}, "ns-a")
			if tt.wantErr {
				require.ErrorIs(t, err, ErrCrossNamespaceSource)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateClusterSources(t *testing.T) {
	tests := []struct {
		name    string
		source  api.Source
		wantErr bool
	}{
		{name: "qualified kubernetes", source: api.Source{Source: "ns-b/flags", Provider: apicommon.SyncProviderKubernetes}},
		{name: "bare kubernetes", source: api.Source{Source: "flags", Provider: apicommon.SyncProviderKubernetes}, wantErr: true},
		{name: "bare file", source: api.Source{Source: "flags", Provider: apicommon.SyncProviderFilepath}, wantErr: true},
		{name: "empty namespace", source: api.Source{Source: "/flags", Provider: apicommon.SyncProviderKubernetes}, wantErr: true},
		{name: "empty name", source: api.Source{Source: "ns-b/", Provider: apicommon.SyncProviderFlagdProxy}, wantErr: true},
		{name: "too many segments", source: api.Source{Source: "a/b/c", Provider: apicommon.SyncProviderKubernetes}, wantErr: true},
		{name: "s3 URL", source: api.Source{Source: "s3://bucket/key.json", Provider: apicommon.SyncProviderS3}},
		{name: "http URL", source: api.Source{Source: "https://example.com/flags.json", Provider: apicommon.SyncProviderHttp}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClusterSources([]api.Source{tt.source})
			if tt.wantErr {
				require.ErrorIs(t, err, ErrUnqualifiedClusterSource)
				return
			}
			require.NoError(t, err)
		})
	}
}
