package types

type FlagdProxyDeferError struct{}

func (d *FlagdProxyDeferError) Error() string {
	return "flagd-proxy is not ready, deferring pod admission"
}
