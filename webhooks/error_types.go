package webhooks

import (
	goErr "errors"
	"net/http"
)

func (m *PodMutator) IsReady(_ *http.Request) error {
	if m.ready {
		return nil
	}
	return goErr.New("pod mutator is not ready")
}
