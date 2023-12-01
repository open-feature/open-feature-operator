# Migrate to v1beta1

OpenFeature Operator version [v0.5.0](https://github.com/open-feature/open-feature-operator/releases/tag/v0.5.0) contain improved, stable CRDs which were upgraded to `v1beta1` api version.
This document shows what was changed and how you can upgrade your existing CRDs to be compatible with this new release.

## Renaming FeatureFlagConfiguration to FeatureFlag

Along with the API version upgrade to `v1beta1`, we have renamed `FeatureFlagConfiguration` to `FeatureFlag`, making the naming clearer for the CRD maintainers. 
In this CRD, spec section naming has changed. The older CRD used `featureFlagSpec` to define flag configurations.
In the new version, flags are configured under `flagSpec` element.

Consider below example with diff from old version to new version, 

```diff
- apiVersion: core.openfeature.dev/v1alpha3
+ apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagConfiguration
metadata:
  name: end-to-end
  labels:
    app: open-feature-demo
spec:
-  featureFlagSpec:
+  flagSpec 
    flags:
      new-welcome-message:
        state: ENABLED
        variants:
          'on': true
          'off': false
        defaultVariant: 'off'
```

## openfeature.dev annotations

There are several [annotations](./annotations.md) to control how OFO works on your workload.

With the upgrade to `v1beta1`, we no longer support deprecated `openfeature.dev/featureflagconfiguration` annotation.
Workloads which require feature flagging now need to use `openfeature.dev/featureflagsource` annotation, referring a [FeatureFlagSource](./feature_flag_source.md) CRD.

```diff
  annotations:
    openfeature.dev/enabled: "true"
-    openfeature.dev/featureflagconfiguration: "end-to-end"
+    openfeature.dev/featureflagsource: "end-to-end"
```

`FeatureFlagSource` provide more flexibility by allowing users to configure the injected flag with many options.
Consider below example for a `FeatureFlagSource` where flagd is instructed to use `FeatureFlag` CRD named `end-to-end` as its flag source

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
metadata:
  name: end-to-end
  labels:
    app: open-feature-demo
spec:
  sources:
    - source: end-to-end
      provider: kubernetes
```
