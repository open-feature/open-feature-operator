/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-feature/open-feature-operator/internal/common"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type recordingFieldIndexer struct {
	object client.Object
	field  string
}

func (r *recordingFieldIndexer) IndexField(
	_ context.Context,
	object client.Object,
	field string,
	_ client.IndexerFunc,
) error {
	r.object = object
	r.field = field
	return nil
}

func TestIndexPodAllowKubernetesSyncUsesPodAnnotationPath(t *testing.T) {
	indexer := &recordingFieldIndexer{}

	err := indexPodAllowKubernetesSync(context.Background(), indexer)
	require.NoError(t, err)

	require.IsType(t, &corev1.Pod{}, indexer.object)

	expectedPodPath := fmt.Sprintf(
		"%s/%s",
		common.PodOpenFeatureAnnotationPath,
		common.AllowKubernetesSyncAnnotation,
	)
	deploymentPath := fmt.Sprintf(
		"%s/%s",
		common.OpenFeatureAnnotationPath,
		common.AllowKubernetesSyncAnnotation,
	)

	require.Equal(t, "metadata.annotations.openfeature.dev/allowkubernetessync", expectedPodPath)
	require.Equal(t, expectedPodPath, indexer.field)
	require.NotEqual(t, deploymentPath, indexer.field)
}
