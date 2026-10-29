# Feature Flag In-process Configuration

The `InProcessConfiguration` is a custom resource used to set up the
[configuration options](https://flagd.dev/providers/nodejs/?h=flagd_host#available-configuration-options)
for applications using `OpenFeature operator` with in-process evaluation mode enabled.

Below you can see a minimal example of `InProcessConfiguration` resource

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: InProcessConfiguration
metadata:
  labels:
  name: inprocessconfiguration-sample
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

## How does it work?

Similar to usage of [FeatureFlagSource](./feature_flag_source.md) configuration,
[annotations](./annotations.md#) are used to allow the injection of configuration data
into the annotated Pod.
The mutating webhook parses the annotations, retrieves the referenced `InProcessConfiguration` resources from the cluster and injects the data from the resource into all containers of the Pod via environment variables, which configure the provider in the workload to consume feature flag configuration from the available [sync implementation](https://flagd.dev/concepts/syncs/#grpc-sync) specified by the configuration.

## Merging of configurations

The value of `openfeature.dev/inprocessconfiguration` annotation is a comma separated list of values following one of two patterns: {NAME} or {OWN_NAMESPACE}/{NAME}.
If no namespace is provided, the CR is assumed to be in the same namespace as the deployed pod. Cross-namespace
references are rejected by the admission webhook; use a `ClusterInProcessConfiguration` to share configuration across
namespaces (see below). For example:

```yaml
metadata:
  annotations:
    openfeature.dev/enabled: "true"
    openfeature.dev/inprocessconfiguration: "inProcessConfig-A, inProcessConfig-B"
```

When multiple `InProcessConfigurations` are provided, the custom resources are merged in runtime and the last `CR` takes precedence over the first, similarly how it's done for `FeatureFlagSource`.
In this example, 2 CRs are being used to set the injected configuration.

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: InProcessConfiguration
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
kind: InProcessConfiguration
metadata:
    name: inProcessConfig-B
spec:
    envVarPrefix: "my-second-prefix"
    host: "my-host"
```

The resources are merged in runtime, which means that no changes are made to the `InProcessConfiguration` resources
in the cluster, but the operator handles the merge and injection internally.

The resulting configuration will look like the following

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: InProcessConfiguration
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

This resulting resource is transformed into environment variables and injected into all containers
of the annotated Pod
```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    openfeature.dev/enabled: "true"
    openfeature.dev/inprocessconfiguration: "inProcessConfig-A, inProcessConfig-B"
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

## Cluster-scoped InProcessConfiguration

A `ClusterInProcessConfiguration` is the cluster-scoped equivalent of an `InProcessConfiguration`, with an extra
`conditions` field. It is referenced via the `openfeature.dev/clusterinprocessconfiguration` annotation using bare
names. Use it when a cluster admin wants to publish shared in-process configuration to selected namespaces without
allowing cross-namespace references to namespaced resources.

`spec.conditions` works as on a [`ClusterFeatureFlagSource`](feature_flag_source.md#cluster-scoped-featureflagsource):
with no conditions, no namespace may use the resource.

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: ClusterInProcessConfiguration
metadata:
    name: shared-inprocess-config
spec:
    port: 2424
    tls: true
    envVarPrefix: "my-prefix"
    envVars:
      - name: "name1"
        value: "val1"
    conditions:
      - namespaceSelector: {}
```
