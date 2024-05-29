## Quick Start for OpenFeature Operator

This guide helps to get OpenFeature Operator up and running with steps. 
You can skip to [step 4](#4-create-namespace-for-custom-resources) if you already have an Operator installation.  

### Pre-requisite

- Kubernetes cluster OR Kubernetes runtime capability([Kind](https://kind.sigs.k8s.io/))

### Steps

#### 1. Create a K8s cluster (Optional)

```sh
kind create cluster -n kind
```

#### 2. Install cert-manager

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.3/cert-manager.yaml &&
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
```

> [!NOTE]
> Requirement of this dependency is explained in the [installation](./installation.md) guide.

#### 3. Install OpenFeature Operator

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
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.6.0/release.yaml
```
<!-- x-release-please-end -->

Next steps focus on adding feature flags, flag source configuration and a workload deployment

#### 4. Create namespace for custom resources

```sh
kubectl create ns flags
```

> [!NOTE]
> We use the namespace `flags` for flag related custom resources

#### 5. Install feature flags definition 

This is added as a custom resource of kind `FeatureFlag` in `flags` namespace

```sh
kubectl apply -n flags -f - <<EOF
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlag
metadata:
  name: sample-flags
spec:
  flagSpec:
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

#### 5. Install a source definition

This is added as a custom resource of kind `FeatureFlagSource` in `flags` namespace

```sh
kubectl apply -n flags -f - <<EOF
apiVersion: core.openfeature.dev/v1beta1
kind: FeatureFlagSource
metadata:
  name: feature-flag-source
spec:
  sources:
  - source: flags/sample-flags
    provider: kubernetes
  port: 8080
EOF
```

#### 6. Deploy sample workload 

Workload is deployed to namespace `workload`

```sh
kubectl create ns workload
```

The workload here is a simple busy box with curl support. Additionally, it contains OpenFeature Operator annotations.

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
        openfeature.dev/featureflagsource: "flags/feature-flag-source"
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

#### 7. Validate deployment & flag evaluation

First, obtain the pod name of the workload,

```sh
kubectl get pods -n workload
```

This will yield pod name of our workload. For example, `busybox-curl-784775c488-76cr9` as in below example output

```text
NAME                            READY   STATUS    RESTARTS      AGE
busybox-curl-784775c488-76cr9   2/2     Running     0           20h
```

_Optional_ - you can further validate flagd sidecar by describing the pod and validating flagd container,

```sh
kubectl describe pod -n workload busybox-curl-784775c488-76cr9
```

Now with the pod name, exec into the pod,

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

If you are facing errors or if things are not working, 

- See if our troubleshooting guide helps: [Troubleshooting](./troubleshoot.md)
- Reach us with a detailed issue: [Create issue](https://github.com/open-feature/open-feature-operator/issues/new)

### What's next ? 

- Learn more about core concepts behind operator: [concepts](./concepts.md)
- Lean more abour different feature flag sources supported: [FeatureFlagSource](./feature_flag_source.md)  
- Learn more about flagd flag definitions and configurations: [flag definition documentation](https://github.com/open-feature/flagd/blob/main/docs/configuration/flag_configuration.md)
- Read detailed installation instructions: [installation guide](./installation.md)
