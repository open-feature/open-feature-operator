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
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. 

If no namespace is provided, it is assumed that the custom resource is within the **same namespace** as the annotated pod.
If multiple CRs are provided, they are merged with the latest taking precedence. 

For example, in the scenario below, `config-B` will take priority in the merge, replacing duplicated values that are set in `config-A`.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/featureflagsource: "config-A, config-B"
```

### `openfeature.dev/inprocessconfiguration`

This annotation specifies the names of the `InProcessConfigurations` used to configure the injected environment variables to support flagd's [in-process evaluation mode](https://flagd.dev/architecture/#in-process-evaluation).
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. 

If no namespace is provided, it is assumed that the custom resource is within the **same namespace** as the annotated pod.
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

### `openfeature.dev/allowkubernetessync`
*This annotation is used INTERNALLY by the operator.*

This annotation is used to mark pods which should have their permissions backfilled in the event of an upgrade.
When the OFO manager pod is started, all `Service Accounts` of any `Pods` with this annotation set to `"true"` will be added to the `flagd-kubernetes-sync` `Cluster Role Binding`.
