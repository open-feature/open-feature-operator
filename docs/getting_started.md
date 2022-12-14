# Getting Started

Once you have [installed the operator](./installation.md) you can follow this guide to deploy an example application demonstrating the operator.

### Deploy a `FeatureFlagConfiguration`

This `FeatureFlagConfiguration` will be watched by the injected `flagd` container to create a twin of its flag definitions internally. If multiple configurations are supplied to `flagd` these states will be combined.

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
        defaultVariant: "bar",
        targeting: {}
```

### Reference the deployed FeatureFlagConfiguration within a Deployment spec annotation.

In this example we will be creating a `Deployment` containing an `busybox-curl` container, in the example below the `metadata.annotations` object contains the required annotations for the operator to correctly configure and inject the `flagd` sidecar into each deployed `Pod`. The documentation for these annotations can be found [here](./annotations.md).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-curl
  annotations:
    openfeature.dev/enabled: "true"
    openfeature.dev/featureflagconfiguration: "default/featureflagconfiguration-sample"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-busybox-curl-app
  template:
      metadata:
        labels:
          app: my-busybox-cur-app
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

Once the `deployment.yaml` has been applied, our `Pod` should be created and contain 2 containers.
```
kubectl get pods -n default
```
Should give a similar output to the following
```
NAME                                                READY   STATUS              RESTARTS   AGE
busybox-curl-7bd5767999-spf7v                              0/2     ContainerCreating   0          2s
```
If we describe our `Pod` we should should see that the injected sidecar has the following configuration:
```
kubectl describe pod busybox-curl-7bd5767999-spf7v
```
```yaml
  flagd:
    Image:         ghcr.io/open-feature/flagd:v0.2.5
    Port:          8014/TCP
    Host Port:     0/TCP
    Args:
      start
      --uri/
      core.openfeature.dev/default/featureflagconfiguration-sample
    Environment:
      FLAGD_METRICS_PORT:  8014
```

Now that we have confirmed that the flagd sidecar has been injected and the configuration is correct we can test out the flag evaluation using curl.

> This would not be the usual method for evaluating flags in our applications, instead a language specific flagd provider would be used in conjunction with the OpenFeature SDK, documentation can be found [here](https://github.com/open-feature/flagd/blob/main/docs/usage/flagd_providers.md).

```
kubectl exec -it busybox-curl-7bd5767999-spf7v sh
curl -X POST "localhost:8013/schema.v1.Service/ResolveString" -d '{"flagKey":"foo","context":{}}' -H "Content-Type: application/json"
```
output:
```
{"value":"BAR","reason":"DEFAULT","variant":"bar"}
```