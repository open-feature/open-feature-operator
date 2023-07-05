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

### `openfeature.dev/flagsourceconfiguration`
This annotation specifies the names of the FlagSourceConfigurations used to configure the injected flagd sidecar.
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. 

If no namespace is provided, it is assumed that the custom resource is within the **same namespace** as the annotated pod.
If multiple CRs are provided, they are merged with the latest taking precedence. 

For example, in the scenario below, `config-B` will take priority in the merge, replacing duplicated values that are set in `config-A`.

Example:
```yaml
  metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/flagsourceconfiguration: "config-A, config-B"
```

## Deprecated annotations

Given below are references to **deprecated** annotations used by previous versions of the operator.

### `openfeature.dev/allowkubernetessync`
*This annotation is used internally by the operator.*  
This annotation is used to mark pods which should have their permissions backfilled in the event of an upgrade. When the OFO manager pod is started, all `Service Accounts` of any `Pods` with this annotation set to `"true"` will be added to the `flagd-kubernetes-sync` `Cluster Role Binding`.


### `openfeature.dev/featureflagconfiguration`
*This annotation is DEPRECATED in favour of the `openfeature.dev/flagsourceconfiguration` annotation and should no longer be used.* 

This annotation specifies the names of the FeatureFlagConfigurations used to configure the injected flagd sidecar.
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. 
If no namespace is provided it is assumed that the CR is within the same namespace as the deployed pod.
Example:
```yaml
    metadata:
    annotations:
      openfeature.dev/enabled: "true"
      openfeature.dev/featureflagconfiguration: "demo, test/demo-2"
```

### `openfeature.dev`
*This annotation is DEPRECATED in favour of the `openfeature.dev/enabled` annotation and should no longer be used.* 

When a value of `"enabled"` is provided, the operator will inject a flagd sidecar into the annotated pods.  
Example: 
```yaml
    metadata:
    annotations:
      openfeature.dev: "enabled"
```
