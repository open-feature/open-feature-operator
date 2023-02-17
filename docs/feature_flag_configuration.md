# Feature Flag Configuration

The `FeatureFlagConfiguration` version `v1alpha2` CRD defines a CR with the following example structure:

```yaml
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  syncProvider:
    name: filepath # kubernetes (default), filepath or http 
  featureFlagSpec:
    flags:
      foo:
        state: "ENABLED"
        variants:
          bar: "BAR"
          baz: "BAZ"
        defaultVariant: "bar"
```

Within the CRD there are 2 main objects, namely the `featureFlagSpec` and the `syncProvider` each offering a different set of configurations for the injected `flagd` sidecars.

## featureFlagSpec

The `featureFlagSpec` is an object representing the flag configurations themselves, the documentation for this object can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/flag_configuration.md).

## syncProvider

The `syncProvider` specifies how the flag configuration will be supplied to flagd. It contains 2 optional members, `name` and `httpSyncConfiguration`.

The default `syncProvider` is `kubernetes`, but can be configured globally at installation time by modifying the `defaultSyncProvider` parameter or by setting the `sidecarConfiguration.defaultSyncProvider` helm value (`helm upgrade -i ofo openfeature/open-feature-operator  --set  sidecarConfiguration.defaultSyncProvider=filepath`).

### name

| Value                  | Explanation                                                                                                                                                                                                                                                                                                |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "kubernetes" (default) | Instruct flagd to query the Kubernetes API directly for the flag configuration custom resource(s) specified in the workload annotation (`openfeature.dev/featureflagconfiguration`). This configuration requires the flag sidecar (and therefore the workload pod) to be able to access the Kubernetes API. |
| "filepath"             | Mounts the flag configuration custom resource(s) specified in the workload annotation (`openfeature.dev/featureflagconfiguration`) as volume mounted ConfigMaps, and configures flagd to watch them.                                                                                                       |
| "http"                 | Retrieves the flag configuration from the specified HTTP endpoint. Open Feature Operator does not automatically provide the specified configuration at this URL. Requires [`httpSyncConfiguration`](#httpSyncConfiguration) to be configured.                                                               |

### httpSyncConfiguration

httpSyncConfiguration must

| Value         | Explanation                                              |
| ------------- | -------------------------------------------------------- |
| "target"      | Target URL for flagd to poll for the flag configuration. |
| "bearerToken" | Authorization token to be included in HTTP request.      |
