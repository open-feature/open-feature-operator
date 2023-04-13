package constant

import corev1 "k8s.io/api/core/v1"

const (
	FlagDImagePullPolicy          corev1.PullPolicy = "Always"
	Namespace                                       = "open-feature-operator-system"
	ClusterRoleBindingName        string            = "open-feature-operator-flagd-kubernetes-sync"
	AllowKubernetesSyncAnnotation                   = "allowkubernetessync"
	OpenFeatureAnnotationPrefix                     = "openfeature.dev"
	SourceConfigParam                               = "--sources"
	ProbeReadiness                                  = "/readyz"
	ProbeLiveness                                   = "/healthz"
	ProbeInitialDelay                               = 5
)
