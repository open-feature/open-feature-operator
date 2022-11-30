# Annotations

The following annotations are used by the operator to control the injection and define configuration of the flagd sidecar.

### `openfeature.dev/enabled`
When a value of `"true"` is provided, the operator will inject a flagd sidecar into the annotated pods.  
Example: 
```
    metadata:
    annotations:
        openfeature.dev/enabled: "true"
```

### `openfeature.dev/featureflagconfiguration`
This annotation specifies the name of the FeatureFlagConfiguration used to configure the injected flagd sidecar, it is expected that the CR is deployed to the same `namespace` as the pod.  
Example:
```
    metadata:
    annotations:
        openfeature.dev/enabled: "true"
        openfeature.dev/featureflagconfiguration: "demo"
```

### `openfeature.dev`
*This annotation is deprecated in favour of the `openfeature.dev/enabled` annotation and should no longer be used.* 

When a value of `"enabled"` is provided, the operator will inject a flagd sidecar into the annotated pods.  
Example: 
```
    metadata:
    annotations:
        openfeature.dev: "enabled"
```
