# Installation 


## Prerequisites

The OpenFeature Operator is a server that communicates with Kubernetes components within a cluster. As such, it requires a means of authorizing requests between peers. Cert manager handles authorization by adding certificates and certificate issuers as resource types in Kubernetes clusters. This simplifies the process of obtaining, renewing, and using those certificates.
The installation docs for cert manager can be found [here](https://cert-manager.io/docs/installation/kubernetes/).
Alternatively, running the commands below will install cert manager into the `cert-manager` namespace.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
```

## Helm

[Artifact hub](https://artifacthub.io/packages/helm/open-feature-operator/ofo)

Install the latest helm release:
```sh
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
helm repo update
helm upgrade -i openfeature openfeature/open-feature-operator
```
### Uninstall
```sh
helm uninstall ofo
```

## kubectl
Apply the release yaml directly via kubectl
<!---x-release-please-start-version-->
```sh
kubectl create namespace open-feature-operator-system
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.21/release.yaml
```
<!---x-release-please-end-->
### Uninstall
<!---x-release-please-start-version-->
```sh
kubectl delete -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.21/release.yaml
kubectl delete namespace open-feature-operator-system
```
<!---x-release-please-end-->

## Release contents
- `FeatureFlagConfiguration` `CustomResourceDefinition` (custom type that holds the configured state of feature flags).
- Standard kubernetes primitives (e.g. namespace, accounts, roles, bindings, configmaps).
- Operator controller manager service.
- Operator webhook service.
- Deployment with containers kube-rbac-proxy & manager.
- `MutatingWebhookConfiguration` (configures webhooks to call the webhook service).
