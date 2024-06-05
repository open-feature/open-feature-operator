package resources

type FlagdConfiguration struct {
	FlagdPort        int
	OFREPPort        int
	SyncPort         int
	ManagementPort   int
	DebugLogging     bool
	Image            string
	Tag              string
	ImagePullSecrets []string

	OperatorNamespace      string
	OperatorDeploymentName string
}
