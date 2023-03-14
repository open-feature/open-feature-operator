/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	corev1alpha1 "github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	corev1alpha2 "github.com/open-feature/open-feature-operator/apis/core/v1alpha2"
	corev1alpha3 "github.com/open-feature/open-feature-operator/apis/core/v1alpha3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient           client.Client
	testEnv             *envtest.Environment
	testCtx, testCancel = context.WithCancel(context.Background())
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")

	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha2.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = corev1alpha3.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		Scheme:                scheme,
		CRDInstallOptions:     envtest.CRDInstallOptions{Scheme: scheme},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
		Host:           testEnv.WebhookInstallOptions.LocalServingHost,
		Port:           testEnv.WebhookInstallOptions.LocalServingPort,
		CertDir:        testEnv.WebhookInstallOptions.LocalServingCertDir,
	})
	Expect(err).ToNot(HaveOccurred())

	err = mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&appsV1.Deployment{},
		fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, FlagSourceConfigurationAnnotation),
		FlagSourceConfigurationIndex,
	)
	Expect(err).ToNot(HaveOccurred())

	err = (&FlagSourceConfigurationReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err := mgr.Start(testCtx)
		Expect(err).ToNot(HaveOccurred())
	}()

	setup()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	testCancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
