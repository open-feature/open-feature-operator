# Flagd Configuration

The injected flagd sidecar is configured using the `FlagSourceConfiguration` CRD, the `openfeature.dev/flagsourceconfiguration` annotation is used to assign `Pods` with their respective `FlagSourceConfiguration` CRs. The annotation value is a comma separated list of values following one of 2 patterns: {NAME} or {NAMESPACE}/{NAME}. If no namespace is provided, it is assumed that the CR is within the same namespace as the deployed pod, for example:
```
    metadata:
        namespace: test-ns
        annotations:
            openfeature.dev/enabled: "true"
            openfeature.dev/flagsourceconfiguration:"config-A, test-ns-2/config-B"
```
In this example, 2 CRs are being used to configure the injected flagd container, `config-A` (which is assumed to be in the namespace `test-ns`) and `config-B` from the `test-ns-2` namespace, with `config-B` taking precedence in the configuration merge.

## FlagSourceConfiguration

The flagd configuration CRD contains the following fields:

| Field      | Behavior | Type | 
| ----------- | ----------- | ----------- |
| MetricsPort      | Defines the port for flagd to serve metrics on, defaults to 8013       | `int32`       |
| Port   | Defines the port for flagd to listen on, defaults to 8014        | `int32`        |
| SocketPath   | Defines the unix socket path to listen on        | `string`       |
| SyncProviderArgs   | String arguments passed to flagd on startup, documentation can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/configuration.md)        | `array of strings`, key values separated by `=`, e.g `key=value`       |
| Evaluator   | Sets the flagd evaluator, defaults to `json`        | `string`       |
| Image   | Allows for the flagd image to be overridden, defaults to `ghcr.io/open-feature/flagd`        | `string`       |
| Tag   |  Tag to be appended to the flagd image, defaults to `main`        | `string`       |

## Configuration Merging

When multiple `FlagSourceConfigurations` are passed the configurations are merged, the last `CR` will take precedence over the first, with any configuration from the deprecated `FlagDSpec` field of the `FeatureFlagConfiguration` CRD taking the lowest priority. 
An example of this behavior can be found below:
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
    name: test-configuration
spec:
    metricsPort: 8080
    tag: latest
```
Config-B:
```
apiVersion: core.openfeature.dev/v1alpha2
kind: FlagSourceConfiguration
metadata:
    name: test-configuration
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