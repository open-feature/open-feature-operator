package v1alpha1

import ctrl "sigs.k8s.io/controller-runtime"

// SetupWebhookWithManager register webhook for FeatureFlagConfiguration. Links conversion hub to server
func (r *FeatureFlagConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
