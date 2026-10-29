package resources

import corev1 "k8s.io/api/core/v1"

type FlagdConfiguration struct {
	FlagdPort        int
	OFREPPort        int
	SyncPort         int
	ManagementPort   int
	DebugLogging     bool
	Image            string
	Tag              string
	ImagePullPolicy  corev1.PullPolicy
	ImagePullSecrets []string
	Labels           map[string]string
	Annotations      map[string]string

	OperatorNamespace      string
	OperatorDeploymentName string
}
