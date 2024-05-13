package types

// SourceConfig is a 1:1 mapping for flagd SourceConfig type. JSON encoded SourceConfig becomes a startup argument for
// flagd
// NOTE - once we have API stability, make this a dependency at flagd to avoid duplication
type SourceConfig struct {
	URI      string `json:"uri"`
	Provider string `json:"provider"`

	BearerToken string `json:"bearerToken,omitempty"`
	CertPath    string `json:"certPath,omitempty"`
	TLS         bool   `json:"tls,omitempty"`
	ProviderID  string `json:"providerID,omitempty"`
	Selector    string `json:"selector,omitempty"`
	Interval    uint32 `json:"interval,omitempty"`
}
