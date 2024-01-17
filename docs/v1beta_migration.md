# Migrate to v1beta1

OpenFeature Operator version [v0.5.0](https://github.com/open-feature/open-feature-operator/releases/tag/v0.5.0) contain improved, stable CRDs which were upgraded to `v1beta1` api version.
This document shows what was changed and how you can upgrade your existing CRDs to be compatible with this new release.

## Renaming FeatureFlagConfiguration to FeatureFlag

Along with the API version upgrade to `v1beta1`, we have renamed `FeatureFlagConfiguration` to `FeatureFlag`, making the naming clearer for the custom resource maintainers. 
In this CRD, spec section naming has changed. The older CRD used `featureFlagSpec` to define flag configurations.
In the new version, flags are configured under `flagSpec` element.

Consider below example with diff from old version to new version, 

```diff
- apiVersion: core.openfeature.dev/v1alpha3
- kind: FeatureFlagConfiguration
+ apiVersion: core.openfeature.dev/v1beta1
+ kind: FeatureFlag
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

## openfeature.dev annotations & FlagSourceConfiguration renaming

There are two [annotations](./annotations.md) to control how OFO works on your workload.

With the upgrade to `v1beta1`, we no longer support deprecated `openfeature.dev/featureflagconfiguration` annotation.
Workloads which require feature flagging now need to use `openfeature.dev/featureflagsource` annotation, referring a [FeatureFlagSource](./feature_flag_source.md) CRD.
This is the CRD previously named `FlagSourceConfiguration`.

```diff
  annotations:
    openfeature.dev/enabled: "true"
-    openfeature.dev/featureflagconfiguration: "end-to-end"
+    openfeature.dev/featureflagsource: "end-to-end"
```

`FeatureFlagSource` provide more flexibility by allowing users to configure the injected flag with many options.
Consider below example for a `FeatureFlagSource` where flagd is instructed to use `FeatureFlag` custom resource named `end-to-end` as its flag source

```diff
- apiVersion: core.openfeature.dev/v1alpha3
- kind: FlagSourceConfiguration
+ apiVersion: core.openfeature.dev/v1beta1
+ kind: FeatureFlagSource
metadata:
  name: end-to-end
  labels:
    app: open-feature-demo
spec:
  sources:
    - source: end-to-end
      provider: kubernetes
```

## Migration on helm

The new operator no longer support older API versions. Because of this, you need to plan your upgrade carefully.

We recommend following migration steps,

1. Remove all the old custom resources while running the older version of the operator
2. Update the operator to the latest version
3. Install upgraded custom resources
4. Update annotation of your workloads to the latest supported version

If you have used `flagd-proxy` provider, then you have to upgrade the image used by the `flagd-proxy` deployment.
For this, please edit the deployment of `flagd-proxy` to version [v0.3.1](https://github.com/open-feature/flagd/pkgs/container/flagd-proxy/152333134?tag=v0.3.1) or above.

**Note:**
Since OFO version `v0.5.4`, `flagd-proxy` pod (if present) will be upgraded automatically to the
to the latest supported version by `open-feature-operator`.
For more information see the [upgrade section](./installation.md#upgrading).
