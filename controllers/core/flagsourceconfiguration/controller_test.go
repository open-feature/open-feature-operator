package flagsourceconfiguration

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/controllers/common"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFlagSourceConfigurationReconciler_Reconcile(t *testing.T) {
	const (
		testNamespace  = "test-namespace"
		fsConfigName   = "test-config"
		deploymentName = "test-deploy"
	)

	deployment := createTestDeployment(fsConfigName, testNamespace, deploymentName)
	fsConfig := createTestFSConfig(fsConfigName, testNamespace, deploymentName)

	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(deployment, fsConfig).WithIndex(&appsv1.Deployment{}, fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.FlagSourceConfigurationAnnotation), common.FlagSourceConfigurationIndex).Build()

	r := &FlagSourceConfigurationReconciler{
		Client: fakeClient,
		Log:    ctrl.Log.WithName("flagsourceconfiguration-controller"),
		Scheme: fakeClient.Scheme(),
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: testNamespace,
			Name:      fsConfigName,
		},
	}

	ctx := context.TODO()

	deployment2 := &appsv1.Deployment{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: testNamespace}, deployment2)
	require.Nil(t, err)
	restartAt := deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]
	require.Equal(t, "", restartAt)

	r.Reconcile(ctx, req)

	err = fakeClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: testNamespace}, deployment)
	require.Nil(t, err)

	require.NotEqual(t, restartAt, deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"])

}

func createTestDeployment(fsConfigName string, testNamespace string, deploymentName string) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPath, common.FlagSourceConfigurationAnnotation): "true",
						fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationRoot, common.FlagSourceConfigurationAnnotation): fmt.Sprintf("%s/%s", testNamespace, fsConfigName),
					},
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "busybox",
							Args: []string{
								"sleep",
								"1000",
							},
						},
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}

	return deployment
}

func createTestFSConfig(fsConfigName string, testNamespace string, deploymentName string) *v1alpha1.FlagSourceConfiguration {
	rolloutOnChange := true
	fsConfig := &v1alpha1.FlagSourceConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fsConfigName,
			Namespace: testNamespace,
		},
		Spec: v1alpha1.FlagSourceConfigurationSpec{
			Image: deploymentName,
			Sources: []v1alpha1.Source{
				{
					Source:   "not-real.com",
					Provider: "http",
				},
			},
			RolloutOnChange: &rolloutOnChange,
		},
	}

	return fsConfig
}
