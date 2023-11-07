package v1beta1

type SyncProviderType string

const (
	SyncProviderKubernetes SyncProviderType = "kubernetes"
	SyncProviderFilepath   SyncProviderType = "file"
	SyncProviderHttp       SyncProviderType = "http"
	SyncProviderGrpc       SyncProviderType = "grpc"
	SyncProviderFlagdProxy SyncProviderType = "flagd-proxy"
)
