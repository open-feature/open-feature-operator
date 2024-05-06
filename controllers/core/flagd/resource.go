package flagd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceReconciler struct {
	client.Client
	Log logr.Logger
}

func (r *ResourceReconciler) Reconcile(ctx context.Context, flagd *api.Flagd, obj client.Object, newObjFunc func() (client.Object, error), equalFunc func(client.Object, client.Object) bool) (*ctrl.Result, error) {
	exists := false
	existingObj := obj
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: flagd.Namespace,
		Name:      flagd.Name,
	}, existingObj)

	if err == nil {
		exists = true
	} else if err != nil && !errors.IsNotFound(err) {
		r.Log.Error(err, fmt.Sprintf("Failed to get flagd %s '%s/%s'", obj.GetObjectKind(), flagd.Namespace, flagd.Name))
		return &ctrl.Result{}, err
	}

	// check if the resource is managed by the operator.
	// if not, do not continue to not mess with anything user generated
	if !common.IsManagedByOFO(existingObj) {
		r.Log.Info(fmt.Sprintf("Found existing %s '%s/%s' that is not managed by OFO. Will not proceed.", obj.GetObjectKind(), flagd.Namespace, flagd.Name))
		return &ctrl.Result{}, nil
	}

	newObj, err := newObjFunc()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Could not create new flagd %s resource '%s/%s'", obj.GetObjectKind(), flagd.Namespace, flagd.Name))
		return &ctrl.Result{}, err
	}

	if exists && !equalFunc(existingObj, newObj) {
		if err := r.Client.Update(ctx, newObj); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to update Flagd %s '%s/%s'", obj.GetObjectKind(), flagd.Namespace, flagd.Name))
			return &ctrl.Result{}, err
		}
	} else {
		if err := r.Client.Create(ctx, newObj); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create Flagd %s '%s/%s'", obj.GetObjectKind(), flagd.Namespace, flagd.Name))
			return &ctrl.Result{}, err
		}
	}
	return nil, nil
}
