# Flag Source configuration

The injected sidecar is configured using the `FeatureFlagSource` custom resource definition. 
The `openfeature.dev/featureflagsource` annotation is used to assign Pods with their respective `FeatureFlagSource` custom resources.

A minimal example of a `FeatureFlagSource` is given below,

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
metadata:
  name: feature-flag-source
spec:
  sources:                        # flag sources for the injected flagd
    - source: flags/sample-flags  # FeatureFlag - namespace/name
      provider: kubernetes        # kubernetes flag source backed by FeatureFlag custom resource
  port: 8080                      # port of the flagd sidecar
```

## Feature flag sources

This section explains how to configure feature flag sources to injected flag sidecar.

`FeatureFlagSource` support multiple flag sources. Sources are configured as a list.
Supported sources and their configurations are listed below.

### kubernetes aka `FeatureFlag`

This is `FeatureFlag` custom resource backed flagd feature flag definition.
Read more about the custom resource at the dedicated documentation of [FeatureFlag](./feature_flag.md)

The following example of a `FeatureFlagSource` uses `kubernetes` as the `provider` type:

```yaml
sources:                        
  - source: flags/sample-flags  # FeatureFlag - namespace/custom_resource_name
    provider: kubernetes        # kubernetes flag source backed by FeatureFlag custom resource
```

### flagd-proxy

`flagd-proxy` is an alternative to direct resource access on `FeatureFlag` custom resources.
This source type is useful when there is a need for restricting workload permissions and/or to reduce k8s API load.

Read more about proxy approach to access kubernetes resources: [flagd-proxy](./flagd_proxy.md)

### file

In this mode, `FeatureFlag` custom resources are volume mounted to the injected flagd sidecar. 
flagd then source flag configurations from this volume.

For example, given `FeatureFlag` exist at `flags/sample-flags`, this source configuration look like below,

```yaml
sources:                        
  - source: flags/sample-flags
    provider: file          
```

### http

Feature flags can be sources from a http endpoint using provider type `http`,

```yaml
sources:
  - source: http://my-flag-source.json
    provider: http
    httpSyncBearerToken: token                  # optional bearer token for the http connection
    interval: 5                                 # optional interval in seconds for http requests
```

### grpc

Given below is an example configuration with provider type `grpc` and supported options, 

```yaml
sources:                        
  - source: my-flag-source:8080
    provider: grpc
    certPath: /certs/ca.cert                    # certificate for tls connectivity
    tls: true                                   # enforce tls connectivity
    providerID: flagd-weatherapp-sidecar        # identifier for this connection 
    selector: 'source=database,app=weatherapp'  # flag filtering options
```

### Azure Blob Storage

Given below is an example configuration with provider type `azblob` and supported options,

```yaml
sources:
  - source: azblob://my-bucket/test.json # my-bucket - container name
    provider: azblob
envVars:
  - name: AZURE_STORAGE_ACCOUNT
    value: <account_name>
  - name: AZURE_STORAGE_SAS_TOKEN
    value: <SAS token>
```

Alternative way to provide credentials is to use Kubernetes secrets, for example:

```yaml
sources:
  - source: azblob://my-bucket/test.json # my-bucket - container name
    provider: azblob
envVars:
  - name: AZURE_STORAGE_ACCOUNT
    valueFrom:
      secretKeyRef:
        name: my-secret
        key: account_name
  - name: AZURE_STORAGE_SAS_TOKEN
    valueFrom:
      secretKeyRef:
        name: my-secret
        key: sas_token
```

Other type of credentials for Azure Blob Storage are supported, for details (see [AZ credentials config](https://pkg.go.dev/gocloud.dev/blob/azureblob#hdr-URLs))

## Sidecar configurations

`FeatureFlagSource` provides configurations to the injected flagd sidecar.
Table given below is non-exhaustive list of overriding options. (see [full list](https://github.com/open-feature/open-feature-operator/blob/main/docs/crds.md#featureflagsourcespec))

| Configuration    | Explanation                   | Default                                        |
|------------------|-------------------------------|------------------------------------------------|
| port             | Flag evaluation endpoint port | 8013                                           |
| managementPort   | Management port               | 8014                                           |
| evaluator        | Evaluator to use              | json                                           |
| probesEnabled    | Enable/Disable health probes  | true                                           |
| otelCollectorUri | Otel exporter uri             |                                                |
| resources        | flagD resources               | operator sidecar-cpu-* and sidecar-ram-* flags |

## Merging of configurations

The annotation value is a comma separated list of values following one of two patterns: {NAME} or {NAMESPACE}/{NAME}. 
If no namespace is provided, it is assumed that the CR is within the same namespace as the deployed pod, for example:

```yaml
    metadata:
        namespace: test-ns
        annotations:
            openfeature.dev/enabled: "true"
            openfeature.dev/featureflagsource: "config-A, test-ns-2/config-B"
```

In this example, 2 CRs are being used to configure the injected container (by default the operator uses the `flagd:main` image), `config-A` (which is assumed to be in the namespace `test-ns`) and `config-B` from the `test-ns-2` namespace, with `config-B` taking precedence in the configuration merge.

The `FeatureFlagSource` version `v1beta1` CRD defines a CR with the following example structure.
The documentation for this CRD can be found
[here](crds.md#featureflagsource):

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
metadata:
    name: feature-flag-source-sample
spec:
    metricsPort: 8080
    port: 80
    evaluator: json
    image: my-custom-sidecar-image
    defaultSyncProvider: file
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
    otelCollectorUri: http://localhost:4317
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 200m
        memory: 256Mi
```

The relevant `FeatureFlagSources` are passed to the operator by setting the `openfeature.dev/featureflagsource` annotation, which provides the full configuration of the injected sidecar.

## Configuration Merging

When multiple `FeatureFlagSources` are provided, the configurations are merged. The last `CR` takes precedence over the first. 


```mermaid
flowchart LR
    FeatureFlagSource-values  -->|highest priority| environment-variables -->|lowest priority| defaults
```


An example of this behavior:
```
    metadata:
        annotations:
            openfeature.dev/enabled: "true"
            openfeature.dev/featureflagsource:"config-A, config-B"
```
Config-A:
```
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
metadata:
    name: config-A
spec:
    metricsPort: 8080
    tag: latest
```
Config-B:
```
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
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
