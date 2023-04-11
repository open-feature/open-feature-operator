package common

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnableClusterRoleBinding enables the open-feature-operator-flagd-kubernetes-sync cluster role binding for the given
// service account under the given namespace (required for kubernetes sync provider)
func EnableClusterRoleBinding(ctx context.Context, log logr.Logger, c client.Client, namespace, serviceAccountName string) error {
	serviceAccount := client.ObjectKey{
		Name:      serviceAccountName,
		Namespace: namespace,
	}
	if serviceAccountName == "" {
		serviceAccount.Name = "default"
	}
	// Check if the service account exists
	log.V(1).Info(fmt.Sprintf("Fetching serviceAccount: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
	sa := corev1.ServiceAccount{}
	if err := c.Get(ctx, serviceAccount, &sa); err != nil {
		log.V(1).Info(fmt.Sprintf("ServiceAccount not found: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
		return err
	}
	log.V(1).Info(fmt.Sprintf("Fetching clusterrolebinding: %s", constant.ClusterRoleBindingName))
	// Fetch service account if it exists
	crb := rbacv1.ClusterRoleBinding{}
	if err := c.Get(ctx, client.ObjectKey{Name: constant.ClusterRoleBindingName}, &crb); errors.IsNotFound(err) {
		log.V(1).Info(fmt.Sprintf("ClusterRoleBinding not found: %s", constant.ClusterRoleBindingName))
		return err
	}
	found := false
	for _, subject := range crb.Subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == serviceAccount.Name && subject.Namespace == serviceAccount.Namespace {
			log.V(1).Info(fmt.Sprintf("ClusterRoleBinding already exists for service account: %s/%s", serviceAccount.Namespace, serviceAccount.Name))
			found = true
		}
	}
	if !found {
		log.V(1).Info(fmt.Sprintf("Updating ClusterRoleBinding %s for service account: %s/%s", crb.Name,
			serviceAccount.Namespace, serviceAccount.Name))
		crb.Subjects = append(crb.Subjects, rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      serviceAccount.Name,
			Namespace: serviceAccount.Namespace,
		})
		if err := c.Update(ctx, &crb); err != nil {
			log.V(1).Info(fmt.Sprintf("Failed to update ClusterRoleBinding: %s", err.Error()))
			return err
		}
	}
	log.V(1).Info(fmt.Sprintf("Updated ClusterRoleBinding: %s", crb.Name))

	return nil
}
