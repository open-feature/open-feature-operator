# Flag Source configuration

The `FlagSourceConfiguration` version `v1alpha3` CRD defines the a CR with the following example structure:

```yaml
apiVersion: core.openfeature.dev/v1alpha3
kind: FlagSourceConfiguration
metadata:
    name: flagsourceconfiguration-sample
spec:
    metricsPort: 8080
    Port: 80
    evaluator: json
    image: my-custom-sidecar-image
    tag: main
    syncProviders:
    - source: namespace/name
      provider: kubernetes
    - source: namespace/name2
      provider: filepath
    - source: not-a-real-host.com
      provider: http
    envVars:
    - name: MY_ENV_VAR
      value: my-env-value
```

The relevant `FlagSourceConfigurations` are passed to the operator by setting the `openfeature.dev/flagsourceconfiguration` annotation, and is responsible for providing the full configuration of the injected sidecar.

## FlagSourceConfiguration Fields

| Field      | Behavior | Type | 
| ----------- | ----------- | ----------- |
| MetricsPort      | Defines the port for flagd to serve metrics on, defaults to 8013       | optional `int32`       |
| Port   | Defines the port for flagd to listen on, defaults to 8014        | optional `int32`        |
| SocketPath   | Defines the unix socket path to listen on        | optional `string`       |
| SyncProviderArgs   | String arguments passed to the sidecar on startup, flagd documentation can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/configuration.md)        | optional `array of strings`, key values separated by `=`, e.g `key=value`       |
| Image   | Allows for the flagd image to be overridden, defaults to `ghcr.io/open-feature/flagd`        | optional `string`       |
| Tag   |  Tag to be appended to the flagd image, defaults to `main`        | optional `string`       |
| SyncProviders   |  An array of objects defining configuration and sources for each sync provider to use within flagd, object documentation can be found below        | optional `array of objects`       |
| EnvVars   |  An array of environment variables to be applied to the sidecar, all names will be prepended with the EnvVarPrefix    | optional `array of environment variables`       |
| EnvVarPrefix   |  String value defining the prefix to be applied to all environment variables applied to the sidecar default FLAGD  | optional `string`       |

## SyncProvider Fields

| Field      | Behavior | Type | 
| ----------- | ----------- | ----------- |
| Source      | Defines the URI of the flag source, this can be either a `host:port` or the `namespace/name` of a `FeatureFlagConfiguration`       | `string`       |
| Provider      | Defines the provider to be used, can be set to `kubernetes`, `filepath` or `http`      | `string`       |
| HttpSyncBearerToken      | Defines the bearer token to be used with a `http` sync. Has no effect if `Provider` is not `http`      | optional `string`      |

Within the CRD there are 2 main objects, namely the `flagDSpec` and the `featureFlagSpec`, both offering a different set of configuration for the injected `flagd` sidecars.
