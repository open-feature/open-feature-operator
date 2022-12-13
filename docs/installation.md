# Installation 


## Prerequisites

Cert manager is required by the operator, the installation docs can be found [here](https://cert-manager.io/docs/installation/kubernetes/). (see why here TODO).
Alternatively, running the commands below will install cert manager into the `cert-manager` namespace.

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
```

## kubectl
Apply the release yaml directly via kubectl
```
kubectl create namespace open-feature-operator-system
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.20/release.yaml
```
### Uninstall
```
kubectl delete -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.20/release.yaml
kubectl delete namespace open-feature-operator-system
```

## Helm

Add the chart repository to helm:
```
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
```
Install the OFO helm charts:
```
helm install ofo openfeature/ofo
```
### Uninstall
```
helm uninstall ofo
```

