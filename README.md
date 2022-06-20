## open-feature-operator

![build](https://img.shields.io/github/workflow/status/open-feature/open-feature-operator/ci)
![goversion](https://img.shields.io/github/go-mod/go-version/open-feature/open-feature-operator/main)
![version](https://img.shields.io/badge/version-pre--alpha-green)
![status](https://img.shields.io/badge/status-not--for--production-red)

The open-feature-operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagd](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagd server for feature flags in a variety of ways.

### Deploy the latest release

```
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.0.4/certificate.yaml
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.0.4/release.yaml
```

### Architecture

As per the issue [here](https://github.com/open-feature/research/issues/1)
High level architecture is as follows:

<img src="images/arch-0.png" width="560">

### Example

When wishing to leverage featureflagging within the local pod, the following steps are required:

1. Create a new feature flag custom resource e.g.

```
apiVersion: core.openfeature.dev/v1alpha1
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  featureFlagSpec: |
    {
      "stringFlags": {
        "foo": {
          "state": "enabled",
          "variants": {
            "bar": "BAR",
            "baz": "BAZ"
          },
          "defaultVariant": "bar",
          "rules": []
        }
      }
    }
```

2. Reference the CR within the pod spec annotations

```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  annotations:
    openfeature.dev: "enabled"
    openfeature.dev/featureflagconfiguration: "featureflagconfiguration-sample"
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80
```

3. Example usage from host container

```
root@nginx:/# curl -X POST localhost:8080/flags/foo/resolve/string?default-value=default
{"reason":"DEFAULT","value":"BAR"}
```

### Running the operator locally

#### Pull the "schemas" submodule

1. `git submodule update --init --recursive`

Note: you can update the submodule with `git submodule update --recursive --remote`

#### Create a local cluster with cert manager and our operator

1.  Create a local cluster with MicroK8s or Kind (forward requests from your localhost:30000 to your cluster, see MicroK8s/Kind doc)
1.  `kubectl create ns 'open-feature-operator-system'`
1.  `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml`
1.  `kubectl apply -f config/webhook/certificate.yaml`
1.  `IMG=ghcr.io/open-feature/open-feature-operator:main make deploy`

#### Run the example

1. Apply the end-to-end example: `kubectl apply -f config/samples/end-to-end.yaml`
1. Visit `http://localhost:30000/`
1. Update the value of the `defaultVariant` field in the custom resource instance in `config/samples/end-to-end.yaml` and re-apply to update the flag value!
1. Visit `http://localhost:30000/` and see the change!
