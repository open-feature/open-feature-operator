<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/openfeature-horizontal-white.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/openfeature-horizontal-black.svg">
  <img alt="OpenFeature Logo" src="./images/openfeature-horizontal-black.svg">
</picture>

![build](https://img.shields.io/github/workflow/status/open-feature/open-feature-operator/ci)
![goversion](https://img.shields.io/github/go-mod/go-version/open-feature/open-feature-operator/main)
![version](https://img.shields.io/badge/version-pre--alpha-green)
![status](https://img.shields.io/badge/status-not--for--production-red)

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagD](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagD server for feature flags in a variety of ways.

### Deploy the latest release

_Requires [cert manager](https://cert-manager.io/docs/installation/kubernetes/) installed_

<!---x-release-please-start-version-->

```
kubectl create namespace open-feature-operator-system
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.0/certificate.yaml
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.0/release.yaml
```

<!---x-release-please-end-->

### Architecture

As per the issue [here](https://github.com/open-feature/ofep/issues/1)

As per v0.1.1, the default sync provider has been optimized as per this OpenFeature Enhancement Proposal [issue](https://github.com/open-feature/ofep/blob/main/004-OFEP-kubernetes-sync-service.md).

High level architecture is as follows:

<img src="images/arch-0.png" width="700">

### Example

When wishing to leverage feature flagging within the local pod, the following steps are required:

1. Create a new feature flag custom resource.

_See [here](config/samples/crds/custom_provider.yaml) for additional custom resource parameters_

```
apiVersion: core.openfeature.dev/v1alpha1
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  featureFlagSpec: |
    {
      "flags": {
        "foo": {
          "state": "ENABLED",
          "variants": {
            "bar": "BAR",
            "baz": "BAZ"
          },
          "defaultVariant": "bar",
          "targeting": {}
        }
      }
    }
```

1. Reference the CR within the pod spec annotations

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
root@nginx:/# curl -X POST "localhost:8013/schema.v1.Service/ResolveString" -d '{"flagKey":"foo","context":{}}' -H "Content-Type: application/json"
{"value":"BAR","reason":"DEFAULT","variant":"bar"}
```

### Running the operator locally

#### Create a local cluster with cert manager and our operator

1.  Create a local cluster with MicroK8s or Kind (forward requests from your localhost:30000 to your cluster, see MicroK8s/Kind doc)
1.  `IMG=ghcr.io/open-feature/open-feature-operator:main make deploy-operator`

#### Run the example

1. Apply the end-to-end example: `kubectl -n open-feature-operator-system apply -f config/samples/end-to-end.yaml`
1. Visit `http://localhost:30000/`
1. Update the value of the `defaultVariant` field in the custom resource instance in `config/samples/end-to-end.yaml` and re-apply to update the flag value!
1. Visit `http://localhost:30000/` and see the change!

## Releases

This repo uses _Release Please_ to release packages. Release Please sets up a running PR that tracks all changes for the library components, and maintains the versions according to [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/), generated when [PRs are merged](https://github.com/amannn/action-semantic-pull-request). When Release Please's running PR is merged, any changed artifacts are published.
