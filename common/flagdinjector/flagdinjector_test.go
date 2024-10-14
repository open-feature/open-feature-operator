package flagdinjector

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-logr/logr/testr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/apis/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/common/utils"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testTag   = "0.5.0"
	testImage = "flagd"
)

func TestFlagdContainerInjector_InjectDefaultSyncProvider(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.DefaultSyncProvider = apicommon.SyncProviderGrpc

	flagSourceConfig.Sources = []api.Source{{}}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectDefaultSyncProvider_WithDebugLogging(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.DefaultSyncProvider = apicommon.SyncProviderGrpc

	flagSourceConfig.DebugLogging = utils.TrueVal()

	flagSourceConfig.Sources = []api.Source{{}}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]", "--debug"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectDefaultSyncProvider_WithOtelCollectorUri(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.DefaultSyncProvider = apicommon.SyncProviderGrpc

	flagSourceConfig.OtelCollectorUri = "localhost:4317"

	flagSourceConfig.Sources = []api.Source{{}}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]", "--metrics-exporter", "otel", "--otel-collector-uri", "localhost:4317"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectDefaultSyncProvider_WithResources(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.DefaultSyncProvider = apicommon.SyncProviderGrpc

	flagSourceConfig.Resources = v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(256*1<<20, resource.BinarySI),
		},
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(256*1<<20, resource.BinarySI),
		},
	}

	flagSourceConfig.Sources = []api.Source{{}}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]"}
	expectedPod.Spec.Containers[1].Resources = flagSourceConfig.Resources

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectDefaultSyncProvider_WithSyncProviderArgs(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.SyncProviderArgs = []string{"arg-1", "arg-2"}

	flagSourceConfig.DefaultSyncProvider = apicommon.SyncProviderGrpc

	flagSourceConfig.Sources = []api.Source{{}}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"\",\"provider\":\"grpc\"}]", "--sync-provider-args", "arg-1", "--sync-provider-args", "arg-2"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectFlagdKubernetesSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: apicommon.SyncProviderKubernetes,
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"my-namespace/server-side\",\"provider\":\"kubernetes\"}]"}

	require.Equal(t, expectedPod, pod)

	// verify the update of the ClusterRoleBinding
	cbr := &rbacv1.ClusterRoleBinding{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: common.ClusterRoleBindingName}, cbr)

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
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: apicommon.SyncProviderFilepath,
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil
	expectedPod.Spec.Volumes = []v1.Volume{
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

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"/etc/flagd/my-namespace_server-side/my-namespace_server-side.flagd.json\",\"provider\":\"file\"}]"}
	expectedPod.Spec.Containers[1].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "server-side",
			ReadOnly:  false,
			MountPath: "/etc/flagd/my-namespace_server-side",
		},
	}

	require.Equal(t, expectedPod, pod)

	// verify the creation of the referenced ConfigMap
	cm := &v1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: pod.Spec.Volumes[0].ConfigMap.Name, Namespace: namespace}, cm)
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
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	ownerRef := metav1.OwnerReference{
		APIVersion: "v1alpha2",
		Kind:       "Flagd",
		Name:       "my-flagd",
		UID:        "1234",
	}

	pod := generatePod([]v1.Container{generateContainer()}, []metav1.OwnerReference{ownerRef}, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Source:   "my-namespace/server-side",
			Provider: apicommon.SyncProviderFilepath,
		},
	}

	err = fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil
	expectedPod.OwnerReferences = []metav1.OwnerReference{ownerRef}
	expectedPod.Spec.Volumes = []v1.Volume{
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

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"/etc/flagd/my-namespace_server-side/my-namespace_server-side.flagd.json\",\"provider\":\"file\"}]"}
	expectedPod.Spec.Containers[1].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "server-side",
			ReadOnly:  false,
			MountPath: "/etc/flagd/my-namespace_server-side",
		},
	}

	require.Equal(t, expectedPod, pod)

	// verify the creation of the referenced ConfigMap
	cm = &v1.ConfigMap{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: pod.Spec.Volumes[0].ConfigMap.Name, Namespace: namespace}, cm)
	require.Nil(t, err)

	require.Equal(t, pod.OwnerReferences[0].Name, cm.OwnerReferences[0].Name)
	require.Equal(t, pod.OwnerReferences[0].APIVersion, cm.OwnerReferences[0].APIVersion)
	require.Equal(t, pod.OwnerReferences[0].Kind, cm.OwnerReferences[0].Kind)
	require.Equal(t, pod.OwnerReferences[0].UID, cm.OwnerReferences[0].UID)
}

func TestFlagdContainerInjector_InjectHttpSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Source:              "http://localhost:8013",
			HttpSyncBearerToken: "my-token",
			Provider:            apicommon.SyncProviderHttp,
			Interval:            8,
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"http://localhost:8013\",\"provider\":\"http\",\"bearerToken\":\"my-token\",\"interval\":8}]"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectGrpcSource(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Source:     "grpc://localhost:8013",
			Provider:   apicommon.SyncProviderGrpc,
			TLS:        true,
			CertPath:   "cert-path",
			ProviderID: "provider-id",
			Selector:   "selector",
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"grpc://localhost:8013\",\"provider\":\"grpc\",\"certPath\":\"cert-path\",\"tls\":true,\"providerID\":\"provider-id\",\"selector\":\"selector\"}]"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyNotAvailable(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Provider: apicommon.SyncProviderFlagdProxy,
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	// expect an error here because we do not have a flagd proxy in our cluster
	require.NotNil(t, err)
	require.ErrorIs(t, err, common.ErrFlagdProxyNotReady)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyNotReady(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	flagdProxyDeployment := &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: flagdproxy.FlagdProxyDeploymentName, Namespace: namespace},
	}

	err := fakeClient.Create(context.Background(), flagdProxyDeployment)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Provider: apicommon.SyncProviderFlagdProxy,
		},
	}

	err = fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.NotNil(t, err)
	require.ErrorIs(t, err, common.ErrFlagdProxyNotReady)
}

func TestFlagdContainerInjector_InjectProxySource_ProxyIsReady(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	flagdProxyDeployment := &appsV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: flagdproxy.FlagdProxyDeploymentName, Namespace: namespace},
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
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Provider: apicommon.SyncProviderFlagdProxy,
		},
	}

	err = fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil

	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013", "--sources", "[{\"uri\":\"flagd-proxy-svc.my-namespace.svc.cluster.local:8013\",\"provider\":\"grpc\",\"selector\":\"core.openfeature.dev/my-namespace/\"}]"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_Inject_FlagdContainerAlreadyPresent(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer(), {
		Name: "flagd",
	}}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)
	require.Nil(t, err)

	expectedPod := getExpectedPod(namespace)

	expectedPod.Annotations = nil
	expectedPod.Spec.Containers[1].Args = []string{"start", "--management-port", "8014", "--port", "8013"}

	require.Equal(t, expectedPod, pod)
}

func TestFlagdContainerInjector_InjectUnknownSyncProvider(t *testing.T) {

	namespace, fakeClient := initContainerInjectionTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	pod := generatePod([]v1.Container{generateContainer()}, nil, namespace)

	flagSourceConfig := getFlagSourceConfigSpec()

	flagSourceConfig.Sources = []api.Source{
		{
			Provider: "unknown",
		},
	}

	err := fi.InjectFlagd(context.Background(), &pod.ObjectMeta, &pod.Spec, flagSourceConfig)

	require.NotNil(t, err)
	require.ErrorIs(t, err, common.ErrUnrecognizedSyncProvider)
}

func TestFlagdContainerInjector_createConfigMap(t *testing.T) {

	_ = api.AddToScheme(scheme.Scheme)

	fakeClientBuilder := fake.NewClientBuilder().
		WithScheme(scheme.Scheme)

	ownerUID := types.UID("123")
	tests := []struct {
		name          string
		flagdInjector *FlagdContainerInjector
		namespace     string
		confname      string
		ownerRefs     []metav1.OwnerReference
		wantErr       error
	}{
		{
			name: "featureflag not found",
			flagdInjector: &FlagdContainerInjector{
				Client: fakeClientBuilder.Build(),
				Logger: testr.New(t),
			},
			namespace: "myns",
			confname:  "mypod",
			ownerRefs: []metav1.OwnerReference{{}},
			wantErr:   errors.New("could not retrieve featureflag myns/mypod: featureflags.core.openfeature.dev \"mypod\" not found"),
		},
		{
			name: "featureflag found, config map created",
			flagdInjector: &FlagdContainerInjector{
				Client: fakeClientBuilder.WithObjects(&api.FeatureFlag{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myconf",
						Namespace: "myns",
						UID:       ownerUID,
					},
				}).Build(),
				Logger: testr.New(t),
			},
			namespace: "myns",
			confname:  "myconf",
			ownerRefs: []metav1.OwnerReference{{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.flagdInjector
			err := m.createConfigMap(context.TODO(), tt.namespace, tt.confname, tt.ownerRefs)

			if tt.wantErr == nil {
				require.Nil(t, err)
				ffConfig := v1.ConfigMap{}
				err := m.Client.Get(context.TODO(), client.ObjectKey{Name: tt.confname, Namespace: tt.namespace}, &ffConfig)
				require.Nil(t, err)
				require.Equal(t,
					map[string]string{
						"openfeature.dev/featureflag": tt.confname,
					},
					ffConfig.Annotations)
				require.EqualValues(t, utils.FalseVal(), ffConfig.OwnerReferences[0].Controller)
				require.Equal(t, ownerUID, ffConfig.OwnerReferences[1].UID)

			} else {
				t.Log("checking error", err)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
			}

		})
	}
}

func initContainerInjectionTestEnv() (string, client.WithWatch) {
	namespace := "my-namespace"

	_ = api.AddToScheme(scheme.Scheme)

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}

	cbr := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.ClusterRoleBindingName,
		},
	}

	ffConfig := &api.FeatureFlag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-side",
			Namespace: namespace,
		},
		Spec: api.FeatureFlagSpec{},
	}

	fakeClientBuilder := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).WithObjects(ffConfig, serviceAccount, cbr)

	fakeClient := fakeClientBuilder.Build()
	return namespace, fakeClient
}

func getFlagSourceConfigSpec() *api.FeatureFlagSourceSpec {
	probesEnabled := true

	return &api.FeatureFlagSourceSpec{
		ManagementPort: 8014,
		Port:           8013,
		EnvVars: []v1.EnvVar{
			{
				Name:  "my-env-var",
				Value: "my-value",
			},
		},
		EnvVarPrefix:  "flagd",
		ProbesEnabled: &probesEnabled,
	}
}

func getExpectedPod(namespace string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: namespace,
			Annotations: map[string]string{
				"openfeature.dev/allowkubernetessync": "true",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "image",
					Name:  "container",
					Env: []v1.EnvVar{
						{
							Name:  "flagd_my-env-var",
							Value: "my-value",
						},
						{
							Name:  "flagd_MANAGEMENT_PORT",
							Value: "8014",
						},
						{
							Name:  "flagd_PORT",
							Value: "8013",
						},
						{
							Name:  "flagd_EVALUATOR",
							Value: "",
						},
						{
							Name:  "flagd_LOG_FORMAT",
							Value: "",
						},
						{
							Name:  "flagd_RESOLVER",
							Value: "rpc",
						},
					},
				},
				{
					Name:       "flagd",
					Image:      "flagd:0.5.0",
					WorkingDir: "",
					Ports: []v1.ContainerPort{
						{
							Name:          "management",
							ContainerPort: int32(8014),
						},
						{
							Name:          "flagd",
							ContainerPort: int32(8013),
						},
					},
					Env: []v1.EnvVar{
						{
							Name:  "flagd_my-env-var",
							Value: "my-value",
						},
						{
							Name:  "flagd_MANAGEMENT_PORT",
							Value: "8014",
						},
						{
							Name:  "flagd_PORT",
							Value: "8013",
						},
						{
							Name:  "flagd_EVALUATOR",
							Value: "",
						},
						{
							Name:  "flagd_LOG_FORMAT",
							Value: "",
						},
						{
							Name:  "flagd_RESOLVER",
							Value: "rpc",
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
						Privileged:               utils.FalseVal(),
						RunAsUser:                intPtr(65532),
						RunAsGroup:               intPtr(65532),
						RunAsNonRoot:             utils.TrueVal(),
						ReadOnlyRootFilesystem:   utils.TrueVal(),
						AllowPrivilegeEscalation: utils.FalseVal(),
						SeccompProfile: &v1.SeccompProfile{
							Type: "RuntimeDefault",
						},
					},
				},
			},
		},
	}
}

func generatePod(containers []v1.Container, ownerRef []metav1.OwnerReference, ns string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-pod",
			Namespace:       ns,
			OwnerReferences: ownerRef,
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
	}
}

func generateContainer() v1.Container {
	return v1.Container{
		Image: "image",
		Name:  "container",
	}
}

func intPtr(i int64) *int64 {
	return &i
}

func getProxyConfig() *flagdproxy.FlagdProxyConfiguration {
	return &flagdproxy.FlagdProxyConfiguration{
		Port:           8013,
		ManagementPort: 8014,
		DebugLogging:   false,
		Image:          testImage,
		Tag:            testTag,
		Namespace:      "my-namespace",
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

func Test_getSecurityContext(t *testing.T) {
	user := int64(65532)
	group := int64(65532)
	want := &v1.SecurityContext{
		// flagd does not require any additional capabilities, no bits set
		Capabilities: &v1.Capabilities{
			Drop: []v1.Capability{
				"all",
			},
		},
		RunAsUser:  &user,
		RunAsGroup: &group,
		Privileged: utils.FalseVal(),
		// Prevents misconfiguration from allowing access to resources on host
		RunAsNonRoot: utils.TrueVal(),
		// Prevent container gaining more privileges than its parent process
		AllowPrivilegeEscalation: utils.FalseVal(),
		ReadOnlyRootFilesystem:   utils.TrueVal(),
		// SeccompProfile defines the systems calls that can be made by the container
		SeccompProfile: &v1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
	if got := getSecurityContext(); !reflect.DeepEqual(got, want) {
		t.Errorf("setSecurityContext() = %v, want %v", got, want)
	}
}

func TestFlagdContainerInjector_EnableClusterRoleBinding_AddDefaultServiceAccountName(t *testing.T) {
	enableClusterRoleBindingTest(t, "default", "")
}

func TestFlagdContainerInjector_EnableClusterRoleBinding_ServiceAccountName(t *testing.T) {
	enableClusterRoleBindingTest(t, "my-serviceaccount", "my-serviceaccount")
}

func enableClusterRoleBindingTest(t *testing.T, name string, input string) {
	namespace, fakeClient := initEnableClusterroleBindingTestEnv()

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.ClusterRoleBindingName,
		},
	}

	err := fakeClient.Create(context.Background(), serviceAccount)
	require.Nil(t, err)

	err = fakeClient.Create(context.Background(), crb)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	err = fi.EnableClusterRoleBinding(context.Background(), namespace, input)
	require.Nil(t, err)

	updatedCrb := &rbacv1.ClusterRoleBinding{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: crb.Name}, updatedCrb)

	require.Nil(t, err)

	require.Len(t, updatedCrb.Subjects, 1)
	require.Equal(t, name, updatedCrb.Subjects[0].Name)
	require.Equal(t, namespace, updatedCrb.Subjects[0].Namespace)
}

func TestFlagdContainerInjector_EnableClusterRoleBinding_ServiceAccountAlreadyIncluded(t *testing.T) {

	namespace, fakeClient := initEnableClusterroleBindingTestEnv()

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-serviceaccount",
			Namespace: namespace,
		},
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.ClusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
	}

	err := fakeClient.Create(context.Background(), serviceAccount)
	require.Nil(t, err)

	err = fakeClient.Create(context.Background(), crb)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	err = fi.EnableClusterRoleBinding(context.Background(), namespace, "my-serviceaccount")
	require.Nil(t, err)

	updatedCrb := &rbacv1.ClusterRoleBinding{}
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: crb.Name}, updatedCrb)

	require.Nil(t, err)

	require.Len(t, updatedCrb.Subjects, 1)
	require.Equal(t, "my-serviceaccount", updatedCrb.Subjects[0].Name)
	require.Equal(t, namespace, updatedCrb.Subjects[0].Namespace)
}

func TestFlagdContainerInjector_EnableClusterRoleBinding_ClusterRoleBindingNotFound(t *testing.T) {

	namespace, fakeClient := initEnableClusterroleBindingTestEnv()

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-serviceaccount",
			Namespace: namespace,
		},
	}

	err := fakeClient.Create(context.Background(), serviceAccount)
	require.Nil(t, err)

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
		Image:                     testImage,
		Tag:                       testTag,
	}

	err = fi.EnableClusterRoleBinding(context.Background(), namespace, "my-serviceaccount")
	require.NotNil(t, err)
}

func TestFlagdContainerInjector_EnableClusterRoleBinding_ServiceAccountNotFound(t *testing.T) {

	namespace, fakeClient := initEnableClusterroleBindingTestEnv()

	fi := &FlagdContainerInjector{
		Client:                    fakeClient,
		Logger:                    testr.New(t),
		FlagdProxyConfig:          getProxyConfig(),
		FlagdResourceRequirements: getResourceRequirements(),
	}

	err := fi.EnableClusterRoleBinding(context.Background(), namespace, "my-serviceaccount")
	require.NotNil(t, err)
}

func initEnableClusterroleBindingTestEnv() (string, client.WithWatch) {
	namespace := "my-namespace"

	_ = api.AddToScheme(scheme.Scheme)

	fakeClientBuilder := fake.NewClientBuilder().
		WithScheme(scheme.Scheme)

	fakeClient := fakeClientBuilder.Build()
	return namespace, fakeClient
}
