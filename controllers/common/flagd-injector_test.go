package common

import (
	"context"
	"github.com/go-logr/logr/testr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFlagdContainerInjector_InjectFlagdKubernetesSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: v1alpha1.SyncProviderKubernetes,
		},
	}

	flagSourceConfig.SyncProviderArgs = []string{"arg-1", "arg-2"}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"my-namespace/server-side\",\"provider\":\"kubernetes\"}]", "--sync-provider-args", "arg-1", "--sync-provider-args", "arg-2", "--debug"}

	require.Equal(t, expectedDeployment, deployment)

	// verify the update of the ClusterRoleBinding
	cbr := &rbacv1.ClusterRoleBinding{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: constant.ClusterRoleBindingName}, cbr)

	require.Nil(t, err)

	require.Len(t, cbr.Subjects, 1)
	require.Equal(t, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      "default",
		Namespace: namespace,
	}, cbr.Subjects[0])
}

func TestFlagdContainerInjector_InjectFlagdFilePathSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: v1alpha1.SyncProviderFilepath,
		},
	}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil
	expectedDeployment.Spec.Template.Spec.Volumes = []v1.Volume{
		{
			Name: "server-side",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "server-side",
					},
				},
			},
		},
	}

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"/etc/flagd/my-namespace_server-side/my-namespace_server-side.flagd.json\",\"provider\":\"file\"}]", "--debug"}
	expectedDeployment.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "server-side",
			ReadOnly:  false,
			MountPath: "/etc/flagd/my-namespace_server-side",
		},
	}

	require.Equal(t, expectedDeployment, deployment)

	// verify the creation of the referenced ConfigMap
	cm := &v1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: deployment.Spec.Template.Spec.Volumes[0].ConfigMap.Name, Namespace: namespace}, cm)
	require.Nil(t, err)
}

func TestFlagdContainerInjector_InjectFlagdFilePathSource_UpdateReferencedConfigMap(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	// create the ConfigMap we refer to in the flag source
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-side",
			Namespace: namespace,
		},
	}

	err := fakeClient.Create(context.Background(), cm)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	ownerRef := metav1.OwnerReference{
		APIVersion: "v1alpha2",
		Kind:       "Flagd",
		Name:       "my-flagd",
		UID:        "1234",
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-deployment",
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: v1alpha1.SyncProviderFilepath,
		},
	}

	err = fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil
	expectedDeployment.OwnerReferences = []metav1.OwnerReference{ownerRef}
	expectedDeployment.Spec.Template.Spec.Volumes = []v1.Volume{
		{
			Name: "server-side",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "server-side",
					},
				},
			},
		},
	}

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"/etc/flagd/my-namespace_server-side/my-namespace_server-side.flagd.json\",\"provider\":\"file\"}]", "--debug"}
	expectedDeployment.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "server-side",
			ReadOnly:  false,
			MountPath: "/etc/flagd/my-namespace_server-side",
		},
	}

	require.Equal(t, expectedDeployment, deployment)

	// verify the creation of the referenced ConfigMap
	cm = &v1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: deployment.Spec.Template.Spec.Volumes[0].ConfigMap.Name, Namespace: namespace}, cm)
	require.Nil(t, err)

	require.Equal(t, deployment.OwnerReferences[0].Name, cm.OwnerReferences[0].Name)
	require.Equal(t, deployment.OwnerReferences[0].APIVersion, cm.OwnerReferences[0].APIVersion)
	require.Equal(t, deployment.OwnerReferences[0].Kind, cm.OwnerReferences[0].Kind)
	require.Equal(t, deployment.OwnerReferences[0].UID, cm.OwnerReferences[0].UID)
}

func TestFlagdContainerInjector_InjectHttpSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Source:              "http://localhost:8013",
			HttpSyncBearerToken: "my-token",
			Provider:            v1alpha1.SyncProviderHttp,
		},
	}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"http://localhost:8013\",\"provider\":\"http\",\"bearerToken\":\"my-token\"}]", "--debug"}

	require.Equal(t, expectedDeployment, deployment)
}

func TestFlagdContainerInjector_InjectGrpcSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Source:     "grpc://localhost:8013",
			Provider:   v1alpha1.SyncProviderGrpc,
			TLS:        true,
			CertPath:   "cert-path",
			ProviderID: "provider-id",
			Selector:   "selector",
		},
	}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"grpc://localhost:8013\",\"provider\":\"grpc\",\"certPath\":\"cert-path\",\"tls\":true,\"providerID\":\"provider-id\",\"selector\":\"selector\"}]", "--debug"}

	require.Equal(t, expectedDeployment, deployment)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyNotAvailable(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Provider: v1alpha1.SyncProviderFlagdProxy,
		},
	}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	// expect an error here because we do not have a flagd proxy in our cluster
	require.NotNil(t, err)
	require.ErrorIs(t, err, constant.ErrFlagdProxyNotReady)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyNotReady(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	flagdProxyDeployment := &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: FlagdProxyDeploymentName, Namespace: namespace},
	}

	err := fakeClient.Create(context.Background(), flagdProxyDeployment)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Provider: v1alpha1.SyncProviderFlagdProxy,
		},
	}

	err = fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)
	require.NotNil(t, err)
	require.ErrorIs(t, err, constant.ErrFlagdProxyNotReady)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyIsReady(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	flagdProxyDeployment := &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: FlagdProxyDeploymentName, Namespace: namespace},
	}

	err := fakeClient.Create(context.Background(), flagdProxyDeployment)
	require.Nil(t, err)

	flagdProxyDeployment.Status.ReadyReplicas = 1

	err = fakeClient.Status().Update(context.Background(), flagdProxyDeployment)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Provider: v1alpha1.SyncProviderFlagdProxy,
		},
	}

	err = fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"grpc://flagd-proxy-svc.my-namespace.svc.cluster.local:8013\",\"provider\":\"grpc\",\"selector\":\"core.openfeature.dev/my-namespace/\"}]", "--debug"}

	require.Equal(t, expectedDeployment, deployment)
}

func TestFlagdContainerInjector_InjectDefaultSyncProvider(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.DefaultSyncProvider = v1alpha1.SyncProviderGrpc

	flagSourceConfig.Sources = []v1alpha1.Source{{}}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil

	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]", "--debug"}

	require.Equal(t, expectedDeployment, deployment)
}

func TestFlagdContainerInjector_Inject_FlagdContainerAlreadyPresent(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "flagd",
						},
					},
				},
			},
		},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedDeployment := getExpectedDeployment(namespace)

	expectedDeployment.Annotations = nil
	expectedDeployment.Spec.Template.Spec.Containers[0].Args = []string{"start", "--debug"}

	require.Equal(t, expectedDeployment, deployment)
}

func TestFlagdContainerInjector_InjectUnknownSyncProvider(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagDResourceRequirements: getResourceRequirements(),
	}

	deployment := appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{},
	}

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []v1alpha1.Source{
		{
			Provider: "unknown",
		},
	}

	err := fi.InjectFlagd(context.Background(), &deployment.ObjectMeta, &deployment.Spec.Template.Spec, flagSourceConfig)

	require.NotNil(t, err)
	require.ErrorIs(t, err, constant.ErrUnrecognizedSyncProvider)
}

func initContainerInjectionTestEnv() (string, client.WithWatch) {
	namespace := "my-namespace"

	_ = v1alpha1.AddToScheme(scheme.Scheme)

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}

	cbr := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.ClusterRoleBindingName,
		},
	}

	ffConfig := &v1alpha1.FeatureFlagConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-side",
			Namespace: namespace,
		},
		Spec: v1alpha1.FeatureFlagConfigurationSpec{},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).WithObjects(ffConfig, serviceAccount, cbr)

	client := fakeClient.Build()
	return namespace, client
}

func getFlagSourceConfigSpec() *v1alpha1.FlagSourceConfigurationSpec {
	debugLogging := true
	probesEnabled := true

	return &v1alpha1.FlagSourceConfigurationSpec{
		MetricsPort: 8014,
		Port:        8013,
		Image:       "flagd",
		Tag:         "0.5.0",
		EnvVars: []v1.EnvVar{
			{
				Name:  "my-env-var",
				Value: "my-value",
			},
		},
		EnvVarPrefix:  "flagd",
		ProbesEnabled: &probesEnabled,
		DebugLogging:  &debugLogging,
	}
}

func getExpectedDeployment(namespace string) appsV1.Deployment {
	return appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/allowkubernetessync": "true",
			},
		},
		Spec: appsV1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:       "flagd",
							Image:      "flagd:0.5.0",
							WorkingDir: "",
							Ports: []v1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: int32(8014),
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "flagd_my-env-var",
									Value: "my-value",
								},
								{
									Name:  "flagd_EVALUATOR",
									Value: "",
								},
								{
									Name:  "flagd_LOG_FORMAT",
									Value: "",
								},
							},
							Resources: getResourceRequirements(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.IntOrString{Type: 0, IntVal: 8014, StrVal: ""},
										Host:   "",
										Scheme: "HTTP",
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      0,
								PeriodSeconds:       0,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.IntOrString{Type: 0, IntVal: 8014, StrVal: ""},
										Host:   "",
										Scheme: "HTTP",
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      0,
								PeriodSeconds:       0,
								SuccessThreshold:    0,
								FailureThreshold:    0,
							},
							VolumeMounts:             []v1.VolumeMount{},
							TerminationMessagePath:   "",
							TerminationMessagePolicy: "",
							ImagePullPolicy:          "Always",
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{
										"all",
									},
								},
								Privileged:               boolPtr(false),
								RunAsUser:                intPtr(65532),
								RunAsGroup:               intPtr(65532),
								RunAsNonRoot:             boolPtr(true),
								ReadOnlyRootFilesystem:   boolPtr(true),
								AllowPrivilegeEscalation: boolPtr(false),
								SeccompProfile: &v1.SeccompProfile{
									Type: "RuntimeDefault",
								},
							},
						},
					},
				},
			},
		},
	}
}

func intPtr(i int64) *int64 {
	return &i
}

func getProxyConfig() *FlagdProxyConfiguration {
	return &FlagdProxyConfiguration{
		Port:         8013,
		MetricsPort:  8014,
		DebugLogging: false,
		Image:        "flagd",
		Tag:          "0.5.0",
		Namespace:    "my-namespace",
	}
}

func getResourceRequirements() v1.ResourceRequirements {
	cpuReq, _ := resource.ParseQuantity("0.2")
	cpuLimit, _ := resource.ParseQuantity("0.5")
	memReq, _ := resource.ParseQuantity("10M")
	memLimit, _ := resource.ParseQuantity("20M")
	return v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    cpuLimit,
			v1.ResourceMemory: memLimit,
		},
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    cpuReq,
			v1.ResourceMemory: memReq,
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
