# OpenFeature Operator

## TL;DR
> This helm chart has a dependency on [cert manager](https://cert-manager.io/docs/installation/)
```
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
helm repo update
helm upgrade -i openfeature openfeature/open-feature-operator
```

## Introduction

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagD](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagD server for feature flags in a variety of ways.
The full documentation for this project can be found here: [OpenFeature Operator](https://github.com/open-feature/open-feature-operator/tree/main/docs)

## Prerequisites

The OpenFeature Operator requires [cert manager](https://cert-manager.io/docs/installation/) to be installed on the target cluster.

## Install

To install/upgrade the chart with the release name `open-feature-operator`:
```
helm upgrade -i open-feature-operator openfeature/open-feature-operator
```
This installation will use the default helm configuration, described in the [configuration section](#configuration)

## Uninstall

To uninstall the `open-feature-operator`:

```
helm uninstall open-feature-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
<a name="configuration"></a>

| Value       | Default     | Explanation |
| ----------- | ----------- | ----------- |
| `defaultNamespace`      | `open-feature-operator`  | [INTERNAL USE ONLY] To override the namespace use the `--namespace` flag. This default is provided to ensure that the kustomize build charts in `/templates` deploy correctly when no `namespace` is provided via the `-n` flag.|
