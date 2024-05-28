# Installation 

## Prerequisites

The OpenFeature Operator is a server that communicates with Kubernetes components within a cluster. As such, it requires a means of authorizing requests between peers. Cert manager handles authorization by adding certificates and certificate issuers as resource types in Kubernetes clusters. This simplifies the process of obtaining, renewing, and using those certificates.
The installation docs for cert manager can be found [here](https://cert-manager.io/docs/installation/kubernetes/).
Alternatively, running the commands below will install cert manager into the `cert-manager` namespace.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.3/cert-manager.yaml &&
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
```

## Helm

[Artifact hub](https://artifacthub.io/packages/helm/open-feature-operator/open-feature-operator)

Install the latest helm release:
```sh
helm repo add openfeature https://open-feature.github.io/open-feature-operator/ &&
helm repo update &&
helm upgrade --install openfeature openfeature/open-feature-operator
```

### Upgrading

```sh
helm upgrade --install openfeature openfeature/open-feature-operator
```

> [!NOTE]
> If you upgrade to OFO `v0.5.4` or higher while using a `flagd-proxy` provider, the instance of
`flagd-proxy` will be automatically upgraded to the latest supported version by the `open-feature-operator`.
The upgrade of `flagd-proxy` will also consider your current `FeatureFlagSource` configuration and adapt
the `flagd-proxy` Deployment accordingly.
If you are upgrading OFO to `v0.5.3` or lower, `flagd-proxy` (if present) won't be upgraded automatically.

#### Upgrading CRDs

CRDs are not upgraded automatically with helm (https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).
OpenFeature Operator's CRDs are templated, and can be updated apart from the operator itself by using helm's template functionality and piping the output to `kubectl`:

To install the CRDs:

```sh
helm template openfeature/open-feature-operator -s "templates/crds/*.yaml" | kubectl apply -f -
```

Keep in mind, you can set values as usual during this process:

```sh
helm template openfeature/open-feature-operator -s "templates/crds/*.yaml" --set defaultNamespace=myns | kubectl apply -f -
```

### Uninstall
```sh
helm uninstall openfeature
```

## kubectl
Apply the release yaml directly via kubectl
<!-- x-release-please-start-version -->
```sh
kubectl create namespace open-feature-operator-system &&
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.5.6/release.yaml
```
<!-- x-release-please-end -->
### Uninstall
<!-- x-release-please-start-version -->
```sh
kubectl delete -f https://github.com/open-feature/open-feature-operator/releases/download/v0.5.6/release.yaml &&
kubectl delete namespace open-feature-operator-system
```
<!-- x-release-please-end -->

## Release contents
- `FeatureFlag` `CustomResourceDefinition` (custom type that holds the configured state of feature flags).
- Standard kubernetes primitives (e.g. namespace, accounts, roles, bindings, configmaps).
- Operator controller manager service.
- Operator webhook service.
- Deployment with containers kube-rbac-proxy & manager.
- `MutatingWebhookConfiguration` (configures webhooks to call the webhook service).


## What's next ?

- Follow quick start guide to install custom resources and validate operator behavior: [Quick Start](./quick_start.md)
