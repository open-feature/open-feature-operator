package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"github.com/open-feature/open-feature-operator/common/flagdinjector"
	"github.com/open-feature/open-feature-operator/common/flagdproxy"
	"github.com/open-feature/open-feature-operator/common/types"
	"github.com/open-feature/open-feature-operator/common/utils"
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
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;update,resourceNames=open-feature-operator-flagd-kubernetes-sync;

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

	if shouldUseSidecar(annotations) {
		if code, err := m.handleRPCConfiguration(ctx, req, annotations, pod); err != nil {
			if code == http.StatusForbidden {
				return admission.Denied(err.Error())
			} else {
				return admission.Errored(code, err)
			}
		}
	} else if shouldUseInProcess(annotations) { // use in-process evaluation
		if code, err := m.handleInProcessConfiguration(ctx, req, annotations, pod); err != nil {
			return admission.Errored(code, err)
		}
	} else {
		return admission.Denied("cannot mutate pods without a 'featureflagsource' or 'inprocessconfiguration' annotation as openfeature.dev/enabled annotation is present with a value true")
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (m *PodMutator) handleInProcessConfiguration(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (int32, error) {
	inProcessConfigurationSpec, code, err := m.createFSInProcessConfigSpec(ctx, req, annotations, pod)
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
	featureFlagSourceSpec, code, err := m.createFSConfigSpec(ctx, req, annotations, pod)
	if err != nil {
		return code, err
	}

	// Check for the correct clusterrolebinding for the pod if we use the Kubernetes mode
	if containsK8sProvider(featureFlagSourceSpec.Sources) {
		if err := m.FlagdInjector.EnableClusterRoleBinding(ctx, pod.Namespace, pod.Spec.ServiceAccountName); err != nil {
			return http.StatusForbidden, err
		}
	}

	if err := m.FlagdInjector.InjectFlagd(ctx, &pod.ObjectMeta, &pod.Spec, featureFlagSourceSpec); err != nil {
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
func (m *PodMutator) createFSConfigSpec(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (*api.FeatureFlagSourceSpec, int32, error) {
	// Check configuration
	fscNames := []string{}
	val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.FeatureFlagSourceAnnotation)]
	if ok {
		fscNames = parseList(val)
	}

	featureFlagSourceSpec := NewFeatureFlagSourceSpec(m.Env)

	for _, fscName := range fscNames {
		ns, name := utils.ParseAnnotation(fscName, req.Namespace)

		fc, err := m.getFeatureFlagSource(ctx, ns, name)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("FeatureFlagSource could not be retrieved for %s in namespace %s: %s", fscName, req.Namespace, err.Error()))
			if k8serrors.IsNotFound(err) {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, err
		}
		featureFlagSourceSpec.Merge(&fc.Spec)
	}

	return featureFlagSourceSpec, 0, nil
}

// nolint:dupl
func (m *PodMutator) createFSInProcessConfigSpec(ctx context.Context, req admission.Request, annotations map[string]string, pod *corev1.Pod) (*api.InProcessConfigurationSpec, int32, error) {
	// Check configuration
	fscNames := []string{}
	val, ok := annotations[fmt.Sprintf("%s/%s", common.OpenFeatureAnnotationPrefix, common.InProcessConfigurationAnnotation)]
	if ok {
		fscNames = parseList(val)
	}

	featureFlagSourceSpec := NewInProcessConfigurationSpec(m.Env)

	for _, fscName := range fscNames {
		ns, name := utils.ParseAnnotation(fscName, req.Namespace)

		fc, err := m.getInProcessConfiguration(ctx, ns, name)
		if err != nil {
			m.Log.V(1).Info(fmt.Sprintf("InProcessConfiguration could not be retrieved for %s in namespace %s: %s", fscName, req.Namespace, err.Error()))
			if k8serrors.IsNotFound(err) {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, err
		}
		featureFlagSourceSpec.Merge(&fc.Spec)
	}

	return featureFlagSourceSpec, 0, nil
}

// BackfillPermissions recovers the state of the flagd-kubernetes-sync role binding in the event of upgrade
func (m *PodMutator) BackfillPermissions(ctx context.Context) error {
	defer func() {
		m.ready = true
	}()
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

		// add each new service account to the flagd-kubernetes-sync role binding
		for _, pod := range podList.Items {
			m.Log.V(1).Info(fmt.Sprintf("backfilling permissions for pod %s/%s", pod.Namespace, pod.Name))
			if err := m.FlagdInjector.EnableClusterRoleBinding(ctx, pod.Namespace, pod.Spec.ServiceAccountName); err != nil {
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
	return errors.New("unable to backfill permissions for the flagd-kubernetes-sync role binding: timeout")
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
