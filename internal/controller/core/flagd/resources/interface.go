package resources

import (
	"context"

	"github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IFlagdResource interface {
	GetResource(ctx context.Context, flagd *v1beta1.Flagd) (client.Object, error)
	AreObjectsEqual(o1 client.Object, o2 client.Object) bool
}
