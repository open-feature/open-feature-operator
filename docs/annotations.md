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

### `openfeature.dev/featureflagconfiguration`
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


### `openfeature.dev/flagsourceconfiguration`
This annotation specifies the names of the FlagSourceConfigurations used to configure the injected flagd sidecar.
The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. 
If no namespace is provided it is assumed that the CR is within the same namespace as the deployed pod.
If multiple CRs are provided, they are merged with the latest taking precedence, for example, in the scenario below, `config-B` will take priority in the merge, replacing duplicated values that are set in `config-A`.

Example:
```
    metadata:
    annotations:
        openfeature.dev/enabled: "true"
        openfeature.dev/featureflagconfiguration: "demo, test/demo-2"
        openfeature.dev/flagsourceconfiguration:"config-A, config-B"`
```

### `openfeature.dev`
*This annotation is deprecated in favour of the `openfeature.dev/enabled` annotation and should no longer be used.* 

When a value of `"enabled"` is provided, the operator will inject a flagd sidecar into the annotated pods.  
Example: 
```yaml
    metadata:
    annotations:
        openfeature.dev: "enabled"
```
