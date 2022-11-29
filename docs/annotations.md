# Annotations

The following annotations are used by the operator to control the injection and define configuration of the flagd sidecar.

### `openfeature.dev/enabled`
When a value of `"true"` is provided, the operator will inject a flagd sidecar into the annotated pods.
```
    metadata:
    annotations:
        openfeature.dev/enabled: "true"
```

### `openfeature.dev/featureflagconfiguration`
This annotatation defines the FeatureFlagconfiguration CRD that is to be used by the flagd sidecar, only the name of the CRD should be passed, it is expected that the CRD is deployed to the same `namespace` as the pod.
```
    metadata:
    annotations:
        openfeature.dev/featureflagconfiguration: "demo"
```

### `openfeature.dev`
*This annotation is deprecated in favour of the `openfeature.dev/enabled` annotation and should no longer be used.* 

When a value of `"enabled"` is provided, the operator will inject a flagd sidecar into the annotated pods.
```
    metadata:
    annotations:
        openfeature.dev/enabled: "true"
