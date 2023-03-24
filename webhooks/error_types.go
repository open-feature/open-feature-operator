package webhooks

import (
	goErr "errors"
	"net/http"
)

type kubeProxyDeferError struct{}

func (d *kubeProxyDeferError) Error() string {
	return "kube-flagd-proxy is not ready, deferring pod admission"
}

func (m *PodMutator) IsReady(_ *http.Request) error {
	if m.ready {
		return nil
	}
	return goErr.New("pod mutator is not ready")
}
