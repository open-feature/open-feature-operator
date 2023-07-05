## Quick Start for OpenFeature Operator

### Pre-requisite

- Kubernetes cluster OR Kubernetes runtime capability([Kind](https://kind.sigs.k8s.io/))

### Steps

1. Create a K8s cluster (Optional)

```sh
kind create cluster -n kind
```

2. Install cert-manager

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml &&
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
```

Note - requirement of this dependency is explained in [Installation](./installation.md) guide

3. Install OpenFeature Operator

#### Helm based installation

```sh
helm repo add openfeature https://open-feature.github.io/open-feature-operator/ &&
helm repo update &&
helm upgrade --install openfeature openfeature/open-feature-operator
```

#### Kubectl based installation

<!-- x-release-please-start-version -->
```sh
kubectl create namespace open-feature-operator-system &&
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.34/release.yaml
```
<!-- x-release-please-end -->

Next steps focus on adding feature flags, flag source configuration and a workload deployment

4. Create namespace for custom resources

```sh
kubectl create ns flags
```

5. Install feature flags definition 

This is added as a custom resource of kind `FeatureFlagConfiguration` in `flags` namespace

```sh
kubectl apply -n flags -f - <<EOF
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: sample-flags
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
EOF
```

5. Install a source definition

This is added as a custom resource of kind `FlagSourceConfiguration` in `flags` namespace

```sh
kubectl apply -n flags -f - <<EOF
apiVersion: core.openfeature.dev/v1alpha3
kind: FlagSourceConfiguration
metadata:
  name: flag-source-configuration
spec:
  sources:
  - source: flags/sample-flags
    provider: kubernetes
  port: 8080
EOF
```

6. Deploy sample workload 

Workload is deployed to namespace `workload`

```sh
kubectl create ns workload
```

Workload here is a simple busy box with curl support. Additionally, it contains OpenFeature Operator annotations.

```sh
kubectl apply -n workload -f - <<EOF
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
        # here are the annotations for OpenFeature Operator
        openfeature.dev/enabled: "true"
        openfeature.dev/flagsourceconfiguration: "flags/flag-source-configuration"
    spec:
      containers:
        - name: busybox
          image: yauritux/busybox-curl:latest
          ports:
            - containerPort: 80
          args:
            - sleep
            - "30000"
EOF
```

7. Validate flag evaluation

First, obtain the pod name of the workload,

```sh
kubectl get pods -n workload
```

This will yield pod name of our workload. For example, `busybox-curl-784775c488-76cr9`

Now with the name, exec into the pod,

```sh
kubectl exec  --stdin --tty -n workload <POD_NAME> -- /bin/sh
```

Use the following curl command from the exec shell to evaluate a feature flag,

```sh
curl --location 'http://localhost:8080/schema.v1.Service/ResolveString' --header 'Content-Type: application/json' --data '{ "flagKey":"foo"}'
```

The output should be the following,

`{"value":"BAR", "reason":"STATIC", "variant":"bar"}`

This response is produced from flagd feature provider sidecar deployment, controlled by the operator and shows how 
operator pattern works end to end.