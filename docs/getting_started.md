# Getting Started

Once you [understand the basic concepts](./concepts.md) and [install the operator](./installation.md) you can follow this guide to deploy an example application demonstrating the operator.

## Quick start

### Deploy the demo app

To get started with the operator you can deploy our e2e example using the [cloud-native-demo](https://github.com/open-feature/cloud-native-demo), showcasing an integration with ArgoCD to control feature flags in a GitOps environment.  
Alternatively, you can deploy your own application following the guide below.

## Deploy your own application

### Deploy a `FeatureFlagConfiguration`

This `FeatureFlagConfiguration` is watched by the injected `flagd` container and used to construct its internal flag definitions state. If multiple configurations are supplied to `flagd` these states will be merged.

```yaml
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  featureFlagSpec:
    flags:
      foo:
        state: "ENABLED"
        variants:
          "bar": "BAR"
          "baz": "BAZ"
        defaultVariant: "bar"
        targeting: {}
```

### Reference the deployed FeatureFlagConfiguration within FlagSourceConfiguration.

The `FlagSourceConfiguration` defined below can be used to assign the `FeatureFlagConfiguration`, as well as any other configuration settings, to the injected sidecars. In this example, the port exposed by the injected container is also set.

```yaml
apiVersion: core.openfeature.dev/v1alpha3
kind: FlagSourceConfiguration
metadata:
  name: flagsourceconfiguration-sample
spec:
  sources:
  - source: featureflagconfiguration-sample
    provider: kubernetes
  port: 8080
```

### Reference the deployed FlagSourceConfiguration within a Deployment spec annotation.

In this example, a `Deployment` containing a `busybox-curl` container is created. In the configuration below, the `metadata.annotations` object contains the required annotations for the operator to correctly configure and inject the `flagd` sidecar into each deployed `Pod`. The documentation for these annotations can be found [here](./annotations.md).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-curl
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-busybox-curl-app
  template:
    metadata:
      labels:
        app: my-busybox-curl-app
      annotations:
        openfeature.dev/enabled: "true"
        openfeature.dev/flagsourceconfiguration: "default/flagsourceconfiguration-sample"
    spec:
      containers:
        - name: busybox
          image: yauritux/busybox-curl:latest
          ports:
            - containerPort: 80
          args:
            - sleep
            - "30000"
```

### Confirm that operator has injected the `flagd` sidecar

Once the `deployment.yaml` has been applied, our `Pod` should be created grouping 2 containers.
```sh
kubectl get pods -n default
```
Should give a similar output to the following
```sh
NAME                                                READY   STATUS              RESTARTS   AGE
busybox-curl-7bd5767999-spf7v                              0/2     ContainerCreating   0          2s
```
When the `Pod` is described, the injected sidecar has the following configuration:
```sh
kubectl describe pod busybox-curl-7bd5767999-spf7v
```
```yaml
  flagd:
    Image: ghcr.io/open-feature/flagd:v0.5.2
    Port: 8014/TCP
    Host Port: 0/TCP
    Args:
      start
      --sources
      - '[{"uri":"default/featureflagconfiguration-sample","provider":"kubernetes"}]'
    Environment:
      FLAGD_METRICS_PORT: 8014
```

Now that we have confirmed that the `flagd` sidecar has been injected and the configuration is correct, we can test the flag evaluation using `curl`.

> This is not the usual suggested best practice for evaluating flags in applications, typically a language specific `flagd` provider would be used in conjunction with the OpenFeature SDK, documentation can be found [here](https://github.com/open-feature/flagd/blob/main/docs/usage/flagd_providers.md).

```sh
kubectl exec -it busybox-curl-7bd5767999-spf7v sh
curl -X POST "localhost:8080/schema.v1.Service/ResolveString" -d '{"flagKey":"foo","context":{}}' -H "Content-Type: application/json"
```
output:
```sh
{"value":"BAR","reason":"STATIC","variant":"bar"}
```
