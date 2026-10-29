# Annotations

The following annotations are used by the operator to control the injection and define configuration of the flagd sidecar.

### `openfeature.dev/enabled`

When a value of `"true"` is provided, the operator will inject a flagd sidecar into the annotated pods.  

Example: 
```yaml
    metadata:
    annotations:
      openfeature.dev/enabled: "true"
```

### `openfeature.dev/featureflagsource`

This annotation specifies the names of the `FeatureFlagSources` used to configure the injected flagd sidecar.
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {OWN_NAMESPACE}/{NAME}.

The referenced `FeatureFlagSource` must live in the **same namespace** as the annotated pod. A bare {NAME}
resolves to the pod's namespace; the explicit {OWN_NAMESPACE}/{NAME} form is accepted only when the namespace
matches the pod's own. Cross-namespace references (a namespace other than the pod's) are **rejected** by the
admission webhook. To share a source across namespaces, use a `ClusterFeatureFlagSource` (see below).
If multiple CRs are provided, they are merged with the latest taking precedence.

For example, in the scenario below, `config-B` will take priority in the merge, replacing duplicated values that are set in `config-A`.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/featureflagsource: "config-A, config-B"
```

### `openfeature.dev/clusterfeatureflagsource`

This annotation specifies the names of the cluster-scoped `ClusterFeatureFlagSources` used to configure the injected
flagd sidecar. It is the sanctioned way to share a sidecar configuration across namespaces; a cluster admin creates
the `ClusterFeatureFlagSource`, and pods in the namespaces its `spec.conditions` select may reference it.

The annotation value is a comma separated list of bare names only (cluster-scoped objects have no namespace).
Cluster and namespaced annotations may be combined on the same pod. Cluster resources are merged first as a shared
base, then namespaced ones, so a `FeatureFlagSource` or `InProcessConfiguration` wins on conflicting keys (for example,
a per-namespace `host`). The same applies to `openfeature.dev/clusterinprocessconfiguration`. A field set to its default
value does not override, so a namespaced resource cannot reset a cluster value back to the default.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/clusterfeatureflagsource: "shared-config"
```

### `openfeature.dev/inprocessconfiguration`

This annotation specifies the names of the `InProcessConfigurations` used to configure the injected environment variables to support flagd's [in-process evaluation mode](https://flagd.dev/architecture/#in-process-evaluation).
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {OWN_NAMESPACE}/{NAME}.

The same-namespace rule and cross-namespace rejection described for `openfeature.dev/featureflagsource` apply here too.
If multiple CRs are provided, they are merged with the latest taking precedence.

Users should not combine `openfeature.dev/inprocessconfiguration` and `openfeature.dev/featureflagsource` annotations
for the same pod. If this happens `openfeature.dev/featureflagsource` will take precedence.

For example, in the scenario below, `inProcessConfig-B` will take priority in the merge, replacing duplicated values that are set in `inProcessConfig-A`.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/inprocessconfiguration: "inProcessConfig-A, inProcessConfig-B"
```

### `openfeature.dev/clusterinprocessconfiguration`

This annotation specifies the names of the cluster-scoped `ClusterInProcessConfigurations` used to configure the
injected in-process environment variables. It is the cross-namespace counterpart to
`openfeature.dev/inprocessconfiguration`. The annotation value is a comma separated list of bare names only.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/clusterinprocessconfiguration: "shared-inprocess-config"
```

### `openfeature.dev/allowkubernetessync`
*This annotation is used INTERNALLY by the operator.*

This annotation is used to mark pods which should have their permissions backfilled in the event of an upgrade.
When the OFO manager pod is started, all `Service Accounts` of any `Pods` with this annotation set to `"true"` will have their kubernetes-sync read permissions restored.

### `openfeature.dev/kubernetessyncscope`
*This annotation is used INTERNALLY by the operator.*

Set alongside `allowkubernetessync` on pods that use the kubernetes sync provider. Its value (`namespace` or `cluster`)
records whether the pod's read access was granted via a per-namespace `Role`/`RoleBinding` (namespaced source) or the
operator-wide `ClusterRoleBinding` (cluster-scoped source). `BackfillPermissions` reads it after an operator restart to
restore the correct binding. Pods admitted before 0.10.0 lack this marker and are treated as cluster-scoped until they
are recreated.
