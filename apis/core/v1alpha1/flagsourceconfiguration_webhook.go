package v1alpha1

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWebhookWithManager register webhook for FlagSourceConfiguration. Links conversion hub to server
func (r *FlagSourceConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
