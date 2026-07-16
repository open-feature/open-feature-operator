package clusterfeatureflagsource

import (
	"context"
	"fmt"
	"testing"

	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/flagdproxy"
	commontypes "github.com/open-feature/open-feature-operator/internal/common/types"
	"github.com/open-feature/open-feature-operator/internal/common/utils"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	proxyTag       = "v0.9.4"
	cffsName       = "shared-config"
	deploymentName = "test-deploy"
	deployNS       = "consumer-ns"
	proxyNS        = "proxy-ns"
)

func TestClusterFeatureFlagSourceReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		name                 string
		cffs                 *api.ClusterFeatureFlagSource
		deployment           *appsv1.Deployment
		expectRestart        bool
		flagdProxyDeployment bool
	}{
		{
			// matching consumer restarts when the source changes and rollout is enabled
			name:          "deployment gets restarted with rollout",
			cffs:          createTestCFFS(cffsName, true, apicommon.SyncProviderHttp),
			deployment:    createConsumerDeployment(deploymentName, deployNS, cffsName),
			expectRestart: true,
		},
		{
			// rollout disabled -> no restart
			name:          "deployment without rollout",
			cffs:          createTestCFFS(cffsName, false, apicommon.SyncProviderHttp),
			deployment:    createConsumerDeployment(deploymentName, deployNS, cffsName),
			expectRestart: false,
		},
		{
			// annotation names a different source -> no restart
			name:          "deployment references a different source",
			cffs:          createTestCFFS(cffsName, true, apicommon.SyncProviderHttp),
			deployment:    createConsumerDeployment(deploymentName, deployNS, "some-other-source"),
			expectRestart: false,
		},
		{
			// flagd-proxy source triggers the proxy deployment
			name:                 "flagd-proxy source provisions proxy",
			cffs:                 createTestCFFS(cffsName, true, apicommon.SyncProviderFlagdProxy),
			deployment:           nil,
			flagdProxyDeployment: true,
		},
	}

	require.Nil(t, api.AddToScheme(scheme.Scheme))

	req := ctrl.Request{NamespacedName: types.NamespacedName{Name: cffsName}}
	ctx := context.TODO()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithIndex(&appsv1.Deployment{}, fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.ClusterFeatureFlagSourceAnnotation), common.ClusterFeatureFlagSourceIndex)

			objs := []client.Object{tt.cffs}
			if tt.deployment != nil {
				objs = append(objs, tt.deployment)
			}
			if tt.flagdProxyDeployment {
				objs = append(objs, createOperatorDeployment(proxyNS))
			}
			fakeClient := builder.WithObjects(objs...).Build()

			kpConfig := flagdproxy.NewFlagdProxyConfiguration(commontypes.EnvConfig{
				FlagdProxyImage: "ghcr.io/open-feature/flagd-proxy",
				FlagdProxyTag:   proxyTag,
			}, nil, nil, nil)
			kpConfig.Namespace = proxyNS
			kph := flagdproxy.NewFlagdProxyHandler(kpConfig, fakeClient, ctrl.Log.WithName("cffs-proxy"))

			r := &ClusterFeatureFlagSourceReconciler{
				Client:            fakeClient,
				Log:               ctrl.Log.WithName("cffs-controller"),
				Scheme:            fakeClient.Scheme(),
				FlagdProxy:        kph,
				FlagdProxyBackoff: &utils.ExponentialBackoff{StartDelay: 0, MaxDelay: 0},
			}

			_, err := r.Reconcile(ctx, req)
			require.Nil(t, err)

			if tt.deployment != nil {
				got := &appsv1.Deployment{}
				require.Nil(t, fakeClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: deployNS}, got))
				restartedAt := got.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]
				if tt.expectRestart {
					require.NotEmpty(t, restartedAt)
				} else {
					require.Empty(t, restartedAt)
				}
			}

			if tt.flagdProxyDeployment {
				proxy := &appsv1.Deployment{}
				require.Nil(t, fakeClient.Get(ctx, types.NamespacedName{Name: flagdproxy.FlagdProxyDeploymentName, Namespace: proxyNS}, proxy))
				require.Len(t, proxy.Spec.Template.Spec.Containers, 1)
				require.Equal(t, "ghcr.io/open-feature/flagd-proxy:"+proxyTag, proxy.Spec.Template.Spec.Containers[0].Image)
			}
		})
	}
}

// a matching consumer in a different namespace is also restarted (proves cluster-wide reach)
func TestClusterFeatureFlagSourceReconciler_RestartsAcrossNamespaces(t *testing.T) {
	require.Nil(t, api.AddToScheme(scheme.Scheme))
	ctx := context.TODO()

	depA := createConsumerDeployment("dep-a", "ns-a", cffsName)
	depB := createConsumerDeployment("dep-b", "ns-b", cffsName)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithIndex(&appsv1.Deployment{}, fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.ClusterFeatureFlagSourceAnnotation), common.ClusterFeatureFlagSourceIndex).
		WithObjects(createTestCFFS(cffsName, true, apicommon.SyncProviderHttp), depA, depB).
		Build()

	r := &ClusterFeatureFlagSourceReconciler{
		Client: fakeClient,
		Log:    ctrl.Log.WithName("cffs-controller"),
		Scheme: fakeClient.Scheme(),
	}

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: cffsName}})
	require.Nil(t, err)

	for _, d := range []struct{ name, ns string }{{"dep-a", "ns-a"}, {"dep-b", "ns-b"}} {
		got := &appsv1.Deployment{}
		require.Nil(t, fakeClient.Get(ctx, types.NamespacedName{Name: d.name, Namespace: d.ns}, got))
		require.NotEmpty(t, got.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"], "%s/%s should have restarted", d.ns, d.name)
	}
}

func Test_isUsingConfiguration(t *testing.T) {
	require.True(t, isUsingConfiguration("shared", "shared"))
	require.True(t, isUsingConfiguration("shared", "a, shared, b"))
	require.False(t, isUsingConfiguration("shared", "other"))
	require.False(t, isUsingConfiguration("shared", ""))
}

func createConsumerDeployment(name, ns, sourceName string) *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.ClusterFeatureFlagSourceAnnotation): "true",
						fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationRoot, common.ClusterFeatureFlagSourceAnnotation): sourceName,
					},
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "busybox"}},
				},
			},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
		},
	}
}

func createTestCFFS(name string, rollout bool, provider apicommon.SyncProviderType) *api.ClusterFeatureFlagSource {
	return &api.ClusterFeatureFlagSource{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: api.ClusterFeatureFlagSourceSpec{
			FeatureFlagSourceSpec: api.FeatureFlagSourceSpec{
				Sources: []api.Source{
					{Source: "my-source", Provider: provider},
				},
				RolloutOnChange: &rollout,
			},
		},
	}
}

func createOperatorDeployment(ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: common.OperatorDeploymentName},
	}
}
