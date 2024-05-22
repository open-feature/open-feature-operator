# Feature Flag In-process Configuration

The `FeatureFlagInProcessConfiguration` is a custom resource used to set up the
[configuration options](https://flagd.dev/providers/nodejs/?h=flagd_host#available-configuration-options)
for applications using `OpenFeature operator` with in-process evaluation mode enabled.

Below you can see a minimal example of `FeatureFlagInProcessConfiguration` resource

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagInProcessConfiguration
metadata:
  labels:
  name: featureflaginprocessconfiguration-sample
spec:
  port: 2424
  tls: true
  offlineFlagSourcePath: "my-path"
  cacheMaxSize: 11
  envVarPrefix: "my-prefix"
  envVars:
    - name: "name1"
      value: "val1"
    - name: "name2"
      value: "val2"
```

## How it works?

Similar to usage of [FeatureFlagSource](./feature_flag_source.md) configuration,
[annotations](./annotations.md#) are used to allow the injection of configuration data
into the annotated Pod.
The mutating webhook parses the annotations, retrieves the referenced `FeatureFlagInProcessConfiguration`
resources from the cluster and injects the data from the resource into all containers of the Pod via environment variables,
which are read by the application using the in-process feature flag evaluation.

## Merging of configurations

The value of `openfeature.dev/featureflaginprocessconfiguration` annotation is a comma separated list of values following one of two patterns: {NAME} or {NAMESPACE}/{NAME}.
If no namespace is provided, it is assumed that the CR is within the same namespace as the deployed pod, for example:

```yaml
metadata:
  annotations:
    openfeature.dev/enabled: "true"
    openfeature.dev/featureflaginprocessconfiguration: "inProcessConfig-A, inProcessConfig-B"
```

When multiple `FeatureFlagInProcessConfigurations` are provided, the custom resources are merged in runtime and the last `CR` takes precedence over the first, similarly how it's done for `FeatureFlagSource`.
In this example, 2 CRs are being used to set the injected configuration.

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagInProcessConfiguration
metadata:
    name: inProcessConfig-A
spec:
    port: 2424
    tls: true
    offlineFlagSourcePath: "my-path"
    cacheMaxSize: 11
    envVarPrefix: "my-prefix"
    envVars:
      - name: "name1"
        value: "val1"
      - name: "name2"
        value: "val2"
---
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagInProcessConfiguration
metadata:
    name: inProcessConfig-B
spec:
    envVarPrefix: "my-second-prefix"
    host: "my-host"
```

The resources are merged in runtime, which means that no changes are made to the `FeatureFlagInProcessConfiguration` resources
in the cluster, but the operator handles the merge and injection internally.

The resulting configuration will look like the following

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagInProcessConfiguration
metadata:
    name: internal
spec:
    port: 2424
    tls: true
    offlineFlagSourcePath: "my-path"
    cacheMaxSize: 11
    envVarPrefix: "my-seconf-prefix"
    host: "my-host"
    envVars:
      - name: "name1"
        value: "val1"
      - name: "name2"
        value: "val2"
```

This resulting resource is transformed into enviromant variables and injected into all containers
of the annotated Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    openfeature.dev/enabled: "true"
    openfeature.dev/featureflaginprocessconfiguration: "inProcessConfig-A, inProcessConfig-B"
  name: ofo-pod
spec:
  containers:
    - name: container1
      image: image1
      env:
        - name: my-second-prefix_name2
          value: val2
        - name: my-second-prefix_name1
          value: val1
        - name: my-second-prefix_HOST
          value: my-host
        - name: my-second-prefix_PORT
          value: "2424"
        - name: my-second-prefix_TLS
          value: "true"
        - name: my-second-prefix_OFFLINE_FLAG_SOURCE_PATH
          value: my-path
        - name: my-second-prefix_MAX_CACHE_SIZE
          value: "11"
        - name: my-second-prefix_RESOLVER
          value: in-process
    - name: container2
      image: image2
      env:
        - name: my-second-prefix_name2
          value: val2
        - name: my-second-prefix_name1
          value: val1
        - name: my-second-prefix_HOST
          value: my-host
        - name: my-second-prefix_PORT
          value: "2424"
        - name: my-second-prefix_TLS
          value: "true"
        - name: my-second-prefix_OFFLINE_FLAG_SOURCE_PATH
          value: my-path
        - name: my-second-prefix_MAX_CACHE_SIZE
          value: "11"
        - name: my-second-prefix_RESOLVER
          value: in-process
```
