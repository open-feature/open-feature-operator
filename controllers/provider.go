package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/open-feature/open-feature-operator/apis/core/v1alpha1"
	"github.com/open-feature/open-feature-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rootFileSyncMountPath         = "/etc/flagd"
	AllowKubernetesSyncAnnotation = "allowkubernetessync"
	OpenFeatureAnnotationPrefix   = "openfeature.dev"
)

func HandleSourcesProviders(
	ctx context.Context, log logr.Logger, c client.Client, flagSourceConfig *v1alpha1.FlagSourceConfigurationSpec, ns, serviceAccountNameSpace, serviceAccountName string,
	ownerReferences []metav1.OwnerReference, podSpec *corev1.PodSpec, meta metav1.ObjectMeta, sidecar *corev1.Container,
) error {
	for _, source := range flagSourceConfig.Sources {
		if source.Provider == "" {
			source.Provider = flagSourceConfig.DefaultSyncProvider
		}
		switch {
		case source.Provider.IsFilepath():
			if err := handleFilepathProvider(ctx, log, c, ns, ownerReferences, podSpec, sidecar, source); err != nil {
				return fmt.Errorf("handleFilepathProvider: %w", err)
			}
		case source.Provider.IsKubernetes():
			if err := handleKubernetesProvider(ctx, log, c, ns, serviceAccountNameSpace, serviceAccountName, meta, sidecar, source); err != nil {
				return fmt.Errorf("handleKubernetesProvider: %w", err)
			}
		case source.Provider.IsHttp():
			handleHttpProvider(sidecar, source)
		default:
			return fmt.Errorf("unrecognized sync provider in config: %s", source.Provider)
		}
	}

	return nil
}

func handleFilepathProvider(
	ctx context.Context, log logr.Logger, c client.Client, ns string, ownerReferences []metav1.OwnerReference,
	podSpec *corev1.PodSpec, sidecar *corev1.Container, source v1alpha1.Source,
) error {
	// create config map
	ns, n := ParseAnnotation(source.Source, ns)
	cm := corev1.ConfigMap{}
	if err := c.Get(ctx, client.ObjectKey{Name: n, Namespace: ns}, &cm); errors.IsNotFound(err) {
		err := CreateConfigMap(ctx, log, c, ns, n, ownerReferences)
		if err != nil {
			log.Error(err, "create config map %s")
			return err
		}
	}

	// Add reference of the owner
	if !SharedOwnership(ownerReferences, cm.OwnerReferences) {
		reference := ownerReferences[0]
		reference.Controller = utils.FalseVal()
		cm.OwnerReferences = append(cm.OwnerReferences, reference)
		err := c.Update(ctx, &cm)
		if err != nil {
			log.Error(err, "update owner reference for %s")
		}
	}
	// mount configmap
	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: n,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: n,
				},
			},
		},
	})
	mountPath := fmt.Sprintf("%s/%s", rootFileSyncMountPath, v1alpha1.FeatureFlagConfigurationId(ns, n))
	sidecar.VolumeMounts = append(sidecar.VolumeMounts, corev1.VolumeMount{
		Name: n,
		// create a directory mount per featureFlag spec
		// file mounts will not work
		MountPath: mountPath,
	})
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		fmt.Sprintf("file:%s/%s",
			mountPath,
			v1alpha1.FeatureFlagConfigurationConfigMapKey(ns, n),
		),
	)
	return nil
}

func handleKubernetesProvider(
	ctx context.Context, log logr.Logger, c client.Client, ns, serviceAccountNameSpace, serviceAccountName string, meta metav1.ObjectMeta, sidecar *corev1.Container, source v1alpha1.Source,
) error {
	ns, n := ParseAnnotation(source.Source, ns)
	// ensure that the FeatureFlagConfiguration exists
	ff := FeatureFlag(ctx, c, ns, n)
	if ff.Name == "" {
		return fmt.Errorf("feature flag configuration %s/%s not found", ns, n)
	}
	// add permissions to pod
	if err := EnableClusterRoleBinding(ctx, log, c, serviceAccountNameSpace, serviceAccountName); err != nil {
		return err
	}
	// mark with annotation (required to backfill permissions if they are dropped)
	if meta.Annotations == nil {
		return fmt.Errorf("meta annotations is nil")
	}
	meta.Annotations[fmt.Sprintf("%s/%s", OpenFeatureAnnotationPrefix, AllowKubernetesSyncAnnotation)] = "true"
	// append args
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		fmt.Sprintf(
			"core.openfeature.dev/%s/%s",
			ns,
			n,
		),
	)
	return nil
}

func handleHttpProvider(sidecar *corev1.Container, source v1alpha1.Source) {
	// append args
	sidecar.Args = append(
		sidecar.Args,
		"--uri",
		source.Source,
	)
	if source.HttpSyncBearerToken != "" {
		sidecar.Args = append(
			sidecar.Args,
			"--bearer-token",
			source.HttpSyncBearerToken,
		)
	}
}
