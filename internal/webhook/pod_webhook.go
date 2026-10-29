package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/api/core/v1beta1"
	apicommon "github.com/open-feature/open-feature-operator/api/core/v1beta1/common"
	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/open-feature/open-feature-operator/internal/common/flagdinjector"
	"github.com/open-feature/open-feature-operator/internal/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/internal/common/types"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NOTE: RBAC not needed here.

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.openfeature.dev,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=inprocessconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clusterfeatureflagsources,verbs=get;list;watch
//+kubebuilder:rbac:groups=core.openfeature.dev,resources=clusterinprocessconfigurations,verbs=get;list;watch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;update,resourceNames=open-feature-operator-flagd-kubernetes-sync;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch

// PodMutator annotates Pods
type PodMutator struct {
	Client           client.Client
	decoder          admission.Decoder
	Log              logr.Logger
	ready            bool
	FlagdProxyConfig *flagdproxy.FlagdProxyConfiguration
	FlagdInjector    flagdinjector.IFlagdContainerInjector
	Env              types.EnvConfig
}

// Handle injects the flagd sidecar (if the prerequisites are all met)
//
//nolint:gocyclo
func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	defer func() {
		if err := recover(); err != nil {
			admission.Errored(http.StatusInternalServerError, fmt.Errorf("%v", err))
		}
	}()
	pod := &corev1.Pod{}
	err := m.decoder.Decode(req, pod)

	// Fixes an issue with admission webhook on older k8s versions
	// See: https://github.com/open-feature/open-feature-operator/issues/500
	if pod.Namespace == "" {
		pod.Namespace = req.Namespace
	}

	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	annotations := pod.GetAnnotations()
	// Check enablement
	if !checkOFEnabled(annotations) {
		m.Log.V(2).Info(`openfeature.dev/enabled annotation is not set to "true"`)
		return admission.Allowed("OpenFeature is disabled")
	}

	// Check if the pod is static or orphaned
	if len(pod.GetOwnerReferences()) == 0 {
		return admission.Denied("static or orphaned pods cannot be mutated")
	}

	if shouldUseSidecar(annotations) || shouldUseClusterFeatureFlagSource(annotations) {
		if code, err := m.handleRPCConfiguration(ctx, req, annotations, pod); err != nil {
			if code == http.StatusForbidden {
				return admission.Denied(err.Error())
			} else {
				return admission.Errored(code, err)
			}
		}
	} else if shouldUseInProcess(annotations) || shouldUseClusterInProcessConfiguration(annotations) { // use in-process evaluation
		if code, err := m.handleInProcessConfiguration(ctx, req, annotations, pod); err != nil {
			if code == http.StatusForbidden {
				return admission.Denied(err.Error())
			}
			return admission.Errored(code, err)
		}
	} else {
		return admission.Denied("cannot mutate pods without a 'featureflagsource', 'clusterfeatureflagsource', 'inprocessconfiguration', or 'clusterinprocessconfiguration' annotation as openfeature.dev/enabled annotation is present with a value true")
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (m *PodMutator) handleInProcessConfiguration(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (int32, error) {
	inProcessConfigurationSpec, code, err := m.createFSInProcessConfigSpec(ctx, annotations, pod)
	if err != nil {
		return code, err
	}

	envVars := inProcessConfigurationSpec.ToEnvVars()
	for i := 0; i < len(pod.Spec.Containers); i++ {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, envVars...)
	}
	return 0, nil
}

func (m *PodMutator) handleRPCConfiguration(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (int32, error) {
	// merge any provided flagd specs
	featureFlagSourceSpec, clusterScoped, code, err := m.createFSConfigSpec(ctx, annotations, pod)
	if err != nil {
		return code, err
	}

	// Check for the correct binding for the pod if we use the Kubernetes mode
	if containsK8sProvider(featureFlagSourceSpec.Sources) {
		if err := m.FlagdInjector.EnableKubernetesSyncRBAC(ctx, pod.Namespace, pod.Spec.ServiceAccountName, clusterScoped); err != nil {
			return http.StatusForbidden, err
		}
	}

	if err := m.FlagdInjector.InjectFlagd(ctx, &pod.ObjectMeta, &pod.Spec, featureFlagSourceSpec, clusterScoped); err != nil {
		if errors.Is(err, common.ErrFlagdProxyNotReady) {
			return http.StatusForbidden, err
		}
		//test
		m.Log.Error(err, "unable to inject flagd sidecar")
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

// nolint:dupl
func (m *PodMutator) createFSConfigSpec(ctx context.Context, annotations map[string]string, pod *corev1.Pod) (*api.FeatureFlagSourceSpec, bool, int32, error) {
	featureFlagSourceSpec := NewFeatureFlagSourceSpec(m.Env)
	defaultProvider := featureFlagSourceSpec.DefaultSyncProvider
	clusterScoped := false

	// Namespaced FeatureFlagSources; merged first so cluster-scoped entries can override on conflict.
	if val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation)]; ok {
		for _, token := range parseList(val) {
			spec, code, err := m.namespacedSourceSpec(ctx, token, pod.Namespace, defaultProvider)
			if err != nil {
				return nil, false, code, err
			}
			featureFlagSourceSpec.Merge(spec)
		}
	}

	// Cluster-scoped ClusterFeatureFlagSources; merged after namespaced so cluster-admin intent wins on conflict.
	if val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterFeatureFlagSourceAnnotation)]; ok {
		for _, name := range parseList(val) {
			spec, code, err := m.clusterSourceSpec(ctx, name, defaultProvider)
			if err != nil {
				return nil, false, code, err
			}
			featureFlagSourceSpec.Merge(spec)
			clusterScoped = true
		}
	}

	return featureFlagSourceSpec, clusterScoped, 0, nil
}

func (m *PodMutator) namespacedSourceSpec(ctx context.Context, token, podNS string, defaultProvider apicommon.SyncProviderType) (*api.FeatureFlagSourceSpec, int32, error) {
	ns, name, err := resolveNamespacedRef(token, podNS)
	if err != nil {
		return nil, http.StatusForbidden, err
	}
	fc, err := m.getFeatureFlagSource(ctx, ns, name)
	if err != nil {
		return nil, m.logFetchError(err, "FeatureFlagSource", token, podNS), err
	}
	common.PinSourceProviders(&fc.Spec, defaultProvider)
	if err := common.ValidateNamespacedSources(fc.Spec.Sources, ns); err != nil {
		return nil, http.StatusForbidden, fmt.Errorf("FeatureFlagSource %s/%s: %w", ns, name, err)
	}
	return &fc.Spec, 0, nil
}

func (m *PodMutator) clusterSourceSpec(ctx context.Context, name string, defaultProvider apicommon.SyncProviderType) (*api.FeatureFlagSourceSpec, int32, error) {
	cffs, err := m.getClusterFeatureFlagSource(ctx, name)
	if err != nil {
		return nil, m.logFetchError(err, "ClusterFeatureFlagSource", name, ""), err
	}
	common.PinSourceProviders(&cffs.Spec.FeatureFlagSourceSpec, defaultProvider)
	if err := common.ValidateClusterSources(cffs.Spec.Sources); err != nil {
		return nil, http.StatusForbidden, fmt.Errorf("ClusterFeatureFlagSource %s: %w", name, err)
	}
	return &cffs.Spec.FeatureFlagSourceSpec, 0, nil
}

// nolint:dupl
func (m *PodMutator) createFSInProcessConfigSpec(ctx context.Context, annotations map[string]string, pod *corev1.Pod) (*api.InProcessConfigurationSpec, int32, error) {
	inProcessSpec := NewInProcessConfigurationSpec(m.Env)

	// Namespaced InProcessConfigurations; merged first so cluster-scoped entries can override on conflict.
	if val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation)]; ok {
		for _, token := range parseList(val) {
			ns, name, err := resolveNamespacedRef(token, pod.Namespace)
			if err != nil {
				return nil, http.StatusForbidden, err
			}
			fc, err := m.getInProcessConfiguration(ctx, ns, name)
			if err != nil {
				return nil, m.logFetchError(err, "InProcessConfiguration", token, pod.Namespace), err
			}
			inProcessSpec.Merge(&fc.Spec)
		}
	}

	// Cluster-scoped ClusterInProcessConfigurations; merged after namespaced.
	if val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.ClusterInProcessConfigurationAnnotation)]; ok {
		for _, name := range parseList(val) {
			cipc, err := m.getClusterInProcessConfiguration(ctx, name)
			if err != nil {
				return nil, m.logFetchError(err, "ClusterInProcessConfiguration", name, ""), err
			}
			inProcessSpec.Merge(&cipc.Spec.InProcessConfigurationSpec)
		}
	}

	return inProcessSpec, 0, nil
}

func (m *PodMutator) logFetchError(err error, kind, ref, namespace string) int32 {
	if namespace == "" {
		m.Log.V(1).Info(fmt.Sprintf("%s could not be retrieved for %s: %s", kind, ref, err.Error()))
	} else {
		m.Log.V(1).Info(fmt.Sprintf("%s could not be retrieved for %s in namespace %s: %s", kind, ref, namespace, err.Error()))
	}
	if k8serrors.IsNotFound(err) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// BackfillPermissions recovers kubernetes-sync bindings after operator restart.
// Legacy pods without the scope marker (from <0.10.0) are treated as cluster-scoped
// for back-compat; the advisory only closes for those on next pod restart.
func (m *PodMutator) BackfillPermissions(ctx context.Context) error {
	defer func() {
		m.ready = true
	}()
	scopeAnnotationKey := fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.KubernetesSyncScopeAnnotation)
	for i := 0; i < 5; i++ {
		// fetch all pods with the fmt.Sprintf("%s/%s", OpenFeatureAnnotationPath, EnabledAnnotation) annotation set to "true"
		podList := &corev1.PodList{}
		err := m.Client.List(ctx, podList, client.MatchingFields{
			fmt.Sprintf("%s/%s", common.PodOpenFeatureAnnotationPath, common.AllowKubernetesSyncAnnotation): "true",
		})
		if err != nil {
			if !errors.Is(err, &cache.ErrCacheNotStarted{}) {
				return err
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// restore the correct binding for each pod; default (legacy) = cluster-scoped
		for _, pod := range podList.Items {
			clusterScoped := true
			if scope, ok := pod.Annotations[scopeAnnotationKey]; ok {
				clusterScoped = scope != common.KubernetesSyncScopeNamespace
			}
			m.Log.V(1).Info(fmt.Sprintf("backfilling permissions for pod %s/%s (clusterScoped=%t)", pod.Namespace, pod.Name, clusterScoped))
			if err := m.FlagdInjector.EnableKubernetesSyncRBAC(ctx, pod.Namespace, pod.Spec.ServiceAccountName, clusterScoped); err != nil {
				m.Log.Error(
					err,
					fmt.Sprintf("unable backfill permissions for pod %s/%s", pod.Namespace, pod.Name),
					"webhook",
					fmt.Sprintf("%s/%s", common.PodOpenFeatureAnnotationPath, common.AllowKubernetesSyncAnnotation),
				)
			}
		}
		return nil
	}
	return errors.New("unable to backfill permissions for the flagd-kubernetes-sync bindings: timeout")
}

// InjectDecoder injects the decoder.
func (m *PodMutator) InjectDecoder(d admission.Decoder) error {
	m.decoder = d
	return nil
}

func (m *PodMutator) IsReady(_ *http.Request) error {
	if m.ready {
		return nil
	}
	return errors.New("pod mutator is not ready")
}
