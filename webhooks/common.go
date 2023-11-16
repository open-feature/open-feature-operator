package webhooks

import (
	"fmt"
	"strings"

	api "github.com/open-feature/open-feature-operator/apis/core/v1beta1"
	"github.com/open-feature/open-feature-operator/common/constant"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func OpenFeatureEnabledAnnotationIndex(o client.Object) []string {
	pod := o.(*corev1.Pod)
	if pod.ObjectMeta.Annotations == nil {
		return []string{
			"false",
		}
	}
	val, ok := pod.ObjectMeta.Annotations[fmt.Sprintf("openfeature.dev/%s", constant.AllowKubernetesSyncAnnotation)]
	if ok && val == "true" {
		return []string{
			"true",
		}
	}
	return []string{
		"false",
	}
}

func parseList(s string) []string {
	out := []string{}
	ss := strings.Split(s, ",")
	for i := 0; i < len(ss); i++ {
		newS := strings.TrimSpace(ss[i])
		if newS != "" { //function should not add empty values
			out = append(out, newS)
		}
	}
	return out
}

func containsK8sProvider(sources []api.Source) bool {
	for _, source := range sources {
		if source.Provider.IsKubernetes() {
			return true
		}
	}
	return false
}

func checkOFEnabled(annotations map[string]string) bool {
	val, ok := annotations[fmt.Sprintf("%s/%s", constant.OpenFeatureAnnotationPrefix, constant.EnabledAnnotation)]
	return ok && val == "true"
}
