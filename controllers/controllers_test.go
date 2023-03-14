package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testNamespace  = "test-namespace"
	fsConfigName   = "test-config"
	deploymentName = "test-deploy"
)

func createTestDeployment() {
	deploy := &appsv1.Deployment{}
	deploy.Name = "test-deploy"
	deploy.Namespace = testNamespace
	deploy.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	deploy.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationRoot, "enabled")] = "true"
	deploy.Spec.Template.ObjectMeta.Annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationRoot, FlagSourceConfigurationAnnotation)] = fmt.Sprintf("%s/%s", testNamespace, fsConfigName)
	deploy.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}
	deploy.Spec.Template.ObjectMeta.Labels = map[string]string{
		"app": "test",
	}
	deploy.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "test",
			Image: "busybox",
			Args: []string{
				"sleep",
				"1000",
			},
		},
	}
	err := k8sClient.Create(testCtx, deploy)
	Expect(err).ShouldNot(HaveOccurred())
}

func createTestFSConfig() *v1alpha1.FlagSourceConfiguration {
	fsConfig := &v1alpha1.FlagSourceConfiguration{}
	rolloutOnChange := true
	fsConfig.Namespace = testNamespace
	fsConfig.Name = fsConfigName
	fsConfig.Spec.Image = deploymentName
	fsConfig.Spec.Sources = []v1alpha1.Source{
		{
			Source:   "not-real.com",
			Provider: "http",
		},
	}
	fsConfig.Spec.RolloutOnChange = &rolloutOnChange
	err := k8sClient.Create(testCtx, fsConfig)
	Expect(err).ShouldNot(HaveOccurred())
	return fsConfig
}

func setup() {
	ns := &corev1.Namespace{}
	ns.Name = testNamespace
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())
}

var _ = Describe("flagsourceconfiguration controller tests", func() {

	It("should restart annotated deployments", func() {
		config := createTestFSConfig()
		createTestDeployment()

		// get deployment + set var equal to restartedAt annotation value
		deployment := &appsv1.Deployment{}
		err := k8sClient.Get(testCtx, client.ObjectKey{Name: deploymentName, Namespace: testNamespace}, deployment)
		Expect(err).ShouldNot(HaveOccurred())
		restartAt := deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"]

		// update the fsconfig
		config.Spec.Image = "image-2"
		err = k8sClient.Update(testCtx, config)
		Expect(err).ShouldNot(HaveOccurred())

		// fetch deployment and test if it has been updated
		maxRetries := 5
		notRestartedError := fmt.Errorf("deployment has not been restarted after %d seconds", maxRetries)
		for i := 0; i < maxRetries; i++ {
			err = k8sClient.Get(testCtx, client.ObjectKey{Name: deploymentName, Namespace: testNamespace}, deployment)
			Expect(err).ShouldNot(HaveOccurred())
			if deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] != restartAt {
				notRestartedError = nil
			}
		}
		Expect(notRestartedError).ShouldNot(HaveOccurred())
	})
})
