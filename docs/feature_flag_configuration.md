# Feature Flag Configuration

The `FeatureFlagConfiguration` version `v1alpha2` CRD defines the a CR with the following example structure:

```yaml
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  flagDSpec:
    envs:
      - name: FLAGD_PORT
        value: "8080"
  featureFlagSpec:
    flags:
      foo:
        state: "ENABLED"
        variants:
          bar: "BAR"
          baz: "BAZ"
        defaultVariant: "bar"
```

Within the CRD there are 3 main objects, namely the `flagDSpec`, the `featureFlagSpec`, and the `syncProvider`, each offering a different set of configurations for the injected `flagd` sidecars.

## featureFlagSpec

The `featureFlagSpec` is an object representing the flag configurations themselves, the documentation for this object can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/flag_configuration.md).

## flagDSpec (deprecated, see [FlagSourceConfiguration](./flagd_configuration.md#flagsourceconfiguration))

The `flagDSpec` has one property called `envs` which contains a list of environment variables used to override the default start command values in `flagd`, the documentation for this configuration can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/configuration.md). These environment variables are also made available in the workload `Pod` for simplified configuration.

## syncProvider

The `syncProvider` specifies how the flag configuration will be supplied to flagd. It contains 2 optional members, `name` and `httpSyncConfiguration`.

The default `syncProvider` is `kubernetes`, but can be configured globally at installation time via by modifying the `defaultSyncProvider` parameter or by setting the `sidecarConfiguration.defaultSyncProvider` helm value (`helm upgrade -i ofo openfeature/open-feature-operator  --set  sidecarConfiguration.defaultSyncProvider=filepath`).

### name

| Value                  | Explanation                                                                                                                                                                                                                                                                                                |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "kubernetes" (default) | Instruct flagd to query the Kubernetes API directly for the flag configuration custom resource(s) specified in the workload annotation (`openfeature.dev/featureflagconfiguration`). This configuration requires the flag sidecar (and therefor the workload pod) to be able to access the Kubernetes API. |
| "filepath"             | Mounts the flag configuration custom resource(s) specified in the workload annotation (`openfeature.dev/featureflagconfiguration`) as volume mounted ConfigMaps, and configures flagd to watch them.                                                                                                       |
| "http"                 | Retrives the flag configuration from the specified HTTP endpoint. Open Feature Operator does not automatically provide the specified configuration at this URL. Requires [`httpSyncConfiguration`](#httpSyncConfiguration) to be configured.                                                               |

### httpSyncConfiguration

httpSyncConfiguration must

| Value         | Explanation                                              |
| ------------- | -------------------------------------------------------- |
| "target"      | Target URL for flagd to poll for the flag configuration. |
| "bearerToken" | Authorization token to be included in HTTP request.      |
