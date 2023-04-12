# Flag Source configuration

The injected sidecar is configured using the `FlagSourceConfiguration` CRD, the `openfeature.dev/flagsourceconfiguration` annotation is used to assign `Pods` with their respective `FlagSourceConfiguration` CRs. The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. If no namespace is provided, it is assumed that the CR is within the same namespace as the deployed pod, for example:
```
    metadata:
        namespace: test-ns
        annotations:
            openfeature.dev/enabled: "true"
            openfeature.dev/flagsourceconfiguration:"config-A, test-ns-2/config-B"
```
In this example, 2 CRs are being used to configure the injected container (by default the operator uses the `flagd:main` image), `config-A` (which is assumed to be in the namespace `test-ns`) and `config-B` from the `test-ns-2` namespace, with `config-B` taking precedence in the configuration merge.

The `FlagSourceConfiguration` version `v1alpha3` CRD defines a CR with the following example structure:

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
    defaultSyncProvider: filepath
    tag: main
    sources:
    - source: namespace/name
      provider: kubernetes
    - source: namespace/name2
    - source: not-a-real-host.com
      provider: http
    envVars:
    - name: MY_ENV_VAR
      value: my-env-value
    probesEnabled: true
    debugLogging: false
```

The relevant `FlagSourceConfigurations` are passed to the operator by setting the `openfeature.dev/flagsourceconfiguration` annotation, and is responsible for providing the full configuration of the injected sidecar.

## FlagSourceConfiguration Fields

| Field               | Behavior                                                                                                                                                                                    | Type                                                                      | Default                      | 
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|------------------------------|
| MetricsPort         | Defines the port for flagd to serve metrics on                                                                                                                                              | optional `int32`                                                          | `8013`                       |
| Port                | Defines the port for flagd to listen on                                                                                                                                                     | optional `int32`                                                          | `8014`                       |
| SocketPath          | Defines the unix socket path to listen on                                                                                                                                                   | optional `string`                                                         | `""`                         |
| SyncProviderArgs    | String arguments passed to the sidecar on startup, flagd documentation can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/configuration.md)             | optional `array of strings`, key values separated by `=`, e.g `key=value` | `""`                         | 
| Image               | Allows for the sidecar image to be overridden                                                                                                                                               | optional `string`                                                         | `ghcr.io/open-feature/flagd` | 
| Tag                 | Tag to be appended to the sidecar image                                                                                                                                                     | optional `string`                                                         | `main`                       |
| Sources             | An array of objects defining configuration and sources for each sync provider to use within flagd, documentation of the object is directly below this table                                 | optional `array of objects`                                               | `[]`                         |
| EnvVars             | An array of environment variables to be applied to the sidecar, all names become prepended with the EnvVarPrefix                                                                            | optional `array of environment variables`                                 | `[]`                         | 
| EnvVarPrefix        | String value defining the prefix to be applied to all environment variables applied to the sidecar                                                                                          | optional `string`                                                         | `FLAGD`                      | 
| DefaultSyncProvider | Defines the default provider to be used, can be set to `kubernetes`, `filepath`, `http` or `flagd-proxy` (experimental).                                                                                                  | optional `string`                                                         | `kubernetes`                 | 
| RolloutOnChange     | When set to true the operator will trigger a restart of any `Deployments` within the `FlagSourceConfiguration` reconcile loop, updating the injected sidecar with the latest configuration. | optional `boolean`                                                        | `false`                      | 
| ProbesEnabled       | Enable or disable Liveness and Readiness probes of the flagd sidecar. When enabled, HTTP probes( paths - `/readyz`, `/healthz`) are set with an initial delay of 5 seconds                  | optional `boolean`                                                        | `true`                       |       
| DebugLogging        | Enable or disable `--debug` flag of flagd sidecar                                                                                                                                             | optional `boolean`                                                        | `false`                      |

## Source Fields

| Field               | Behavior                                                                                                                                            | Type              | 
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|
| Source              | Defines the URI of the flag source, this can be either a `host:port` or the `namespace/name` of a `FeatureFlagConfiguration`                        | `string`          |
| Provider            | Defines the provider to be used, can be set to `kubernetes`, `filepath`, `http(s)` or `grpc(s)`. If not provided the default sync provider is used. | optional `string` |
| HttpSyncBearerToken | Defines the bearer token to be used with a `http` sync. Has no effect if `Provider` is not `http`                                                   | optional `string` |
| TLS                 | Enable/Disable secure TLS connectivity. Currently used only by GRPC sync                                                                            | optional `string` |
| CertPath            | Defines the certificate path to be used by grpc TLS connectivity. Has no effect on other `Provider` types                                           | optional `string` |
| ProviderID          | Defines the identifier for grpc connection. Has no effect on other `Provider` types                                                                 | optional `string` |
| Selector            | Defines the flag configuration selection criteria for grpc connection. Has no effect on other `Provider` types                                      | optional `string` |

> The flagd-proxy provider type is experimental, documentation can be found [here](./flagd_proxy.md)

## Configuration Merging

When multiple `FlagSourceConfigurations` are provided, the configurations are merged. The last `CR` takes precedence over the first, with any configuration from the deprecated `FlagDSpec` field of the `FeatureFlagConfiguration` CRD taking the lowest priority. 


```mermaid
flowchart LR
    FlagSourceConfiguration-values  -->|highest priority| environment-variables -->|lowest priority| defaults
```


An example of this behavior:
```
    metadata:
        annotations:
            openfeature.dev/enabled: "true"
            openfeature.dev/flagsourceconfiguration:"config-A, config-B"
```
Config-A:
```
apiVersion: core.openfeature.dev/v1alpha2
kind: FlagSourceConfiguration
metadata:
    name: config-A
spec:
    metricsPort: 8080
    tag: latest
```
Config-B:
```
apiVersion: core.openfeature.dev/v1alpha2
kind: FlagSourceConfiguration
metadata:
    name: config-B
spec:
    port: 8000
    tag: main
```
Results in the following configuration:
```
spec:
    metricsPort: 8080
    port: 8000
    tag: main
```