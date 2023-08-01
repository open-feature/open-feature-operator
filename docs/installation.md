# Installation 

## Prerequisites

The OpenFeature Operator is a server that communicates with Kubernetes components within a cluster. As such, it requires a means of authorizing requests between peers. Cert manager handles authorization by adding certificates and certificate issuers as resource types in Kubernetes clusters. This simplifies the process of obtaining, renewing, and using those certificates.
The installation docs for cert manager can be found [here](https://cert-manager.io/docs/installation/kubernetes/).
Alternatively, running the commands below will install cert manager into the `cert-manager` namespace.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml &&
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

#### Upgrading CRDs

CRDs are not upgraded automatically with helm (https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).
OpenFeature Operator's CRDs are templated, and can be updated apart from the operator itself by using helm's template functionality and piping the output to `kubectl`:

```console
helm template openfeature/open-feature-operator -s templates/{CRD} | kubectl apply -f -
```

For the `featureflagconfigurations.core.openfeature.dev` CRD:

```sh
helm template openfeature/open-feature-operator -s templates/apiextensions.k8s.io_v1_customresourcedefinition_featureflagconfigurations.core.openfeature.dev.yaml | kubectl apply -f -
```

For the `flagsourceconfigurations.core.openfeature.dev` CRD:

```sh
helm template openfeature/open-feature-operator -s templates/apiextensions.k8s.io_v1_customresourcedefinition_flagsourceconfigurations.core.openfeature.dev.yaml | kubectl apply -f -
```

Keep in mind, you can set values as usual during this process:

```console
helm template openfeature/open-feature-operator -s templates/{CRD} --set defaultNamespace=myns | kubectl apply -f -
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
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.35/release.yaml
```
<!-- x-release-please-end -->
### Uninstall
<!-- x-release-please-start-version -->
```sh
kubectl delete -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.35/release.yaml &&
kubectl delete namespace open-feature-operator-system
```
<!-- x-release-please-end -->

## Release contents
- `FeatureFlagConfiguration` `CustomResourceDefinition` (custom type that holds the configured state of feature flags).
- Standard kubernetes primitives (e.g. namespace, accounts, roles, bindings, configmaps).
- Operator controller manager service.
- Operator webhook service.
- Deployment with containers kube-rbac-proxy & manager.
- `MutatingWebhookConfiguration` (configures webhooks to call the webhook service).


## What's next ?

- Follow quick start guide to install custom resources and validate operator behavior: [Quick Start](./quick_start.md)