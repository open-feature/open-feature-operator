# OpenFeature Operator

## TL;DR

> This helm chart has a dependency on [cert manager](https://cert-manager.io/docs/installation/)

```
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
helm repo update
helm upgrade --install open-feature-operator openfeature/open-feature-operator
```

## Introduction

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagd](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagd server for feature flags in a variety of ways.
The full documentation for this project can be found here: [OpenFeature Operator](https://github.com/open-feature/open-feature-operator/tree/main/docs)

## Prerequisites

The OpenFeature Operator requires [cert manager](https://cert-manager.io/docs/installation/) to be installed on the target cluster.

## Install

To install the chart with the release name `open-feature-operator`:

```
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
helm repo update
helm upgrade --install open-feature-operator openfeature/open-feature-operator
```

This installation will use the default helm configuration, described in the [configuration section](#configuration)
To overwrite these default values use the `--set` flag when calling `helm upgrade` or `helm install`, for example:

```
helm upgrade -i open-feature-operator ./chart/open-feature-operator --set sidecarConfiguration.port=8080 --set controllerManager.kubeRbacProxy.resources.limits.cpu=400m
```

## Upgrade

To install the chart with the release name `open-feature-operator`:

```sh
helm repo update
helm upgrade --install open-feature-operator openfeature/open-feature-operator
```

#### Upgrade CRDs

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

## Uninstall

To uninstall the `open-feature-operator`:

```
helm uninstall open-feature-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

<a name="configuration"></a>

### Sidecar configuration

| Value                                      | Default                         | Explanation                                                                                                                                                                                                                                                |
| ------------------------------------------ | ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `sidecarConfiguration.envVarPrefix`        | `FLAGD`                         | Sets the prefix for all environment variables set in the injected sidecar.                                                                                                                                                                                 |
| `sidecarConfiguration.port`                | 8013                            | Sets the value of the `XXX_PORT` environment variable for the injected sidecar container.                                                                                                                                                                  |
| `sidecarConfiguration.metricsPort`         | 8014                            | Sets the value of the `XXX_METRICS_PORT` environment variable for the injected sidecar container.                                                                                                                                                          |
| `sidecarConfiguration.socketPath`          | `""`                            | Sets the value of the `XXX_SOCKET_PATH` environment variable for the injected sidecar container.                                                                                                                                                           |
| `sidecarConfiguration.image.repository`    | `ghcr.io/open-feature/flagd`    | Sets the image for the injected sidecar container.                                                                                                                                                                                                         |
| `sidecarConfiguration.image.tag`           | current flagd version: `v0.4.5` | Sets the version tag for the injected sidecar container.                                                                                                                                                                                                   |
| `sidecarConfiguration.providerArgs`        | `""`                            | Used to append arguments to the sidecar startup command. This value is a comma separated string of key values separated by '=', e.g. `key=value,key2=value2` results in the appending of `--sync-provider-args key=value --sync-provider-args key2=value2` |
| `sidecarConfiguration.defaultSyncProvider` | `kubernetes`                    | Sets the value of the `XXX_SYNC_PROVIDER` environment variable for the injected sidecar container. There are 3 valid sync providers: `kubernetes`, `filepath` and `http`                                                                                   |
| `sidecarConfiguration.logFormat`           | `json`                          | Sets the value of the `XXX_LOG_FORMAT` environment variable for the injected sidecar container. There are 2 valid log formats: `json` and `console`                                                                                                        |
| `sidecarConfiguration.evaluator`           | `json`                          | Sets the value of the `XXX_EVALUATOR` environment variable for the injected sidecar container.                                                                                                                                                             |
| `sidecarConfiguration.probesEnabled`       | `true`                          | Enable or Disable Liveness and Readiness probes of the flagd sidecar. When enabled, HTTP probes( paths - `/readyz`, `/healthz`) are set with an initial delay of 5 seconds                                                                                 |

### Operator resource configuration

| Value                                                       | Default                                      | Explanation                                                                                                                                                                                                                      |
| ----------------------------------------------------------- | -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `defaultNamespace`                                          | `open-feature-operator`                      | [INTERNAL USE ONLY] To override the namespace use the `--namespace` flag. This default is provided to ensure that the kustomize build charts in `/templates` deploy correctly when no `namespace` is provided via the `-n` flag. |
| `controllerManager.kubeRbacProxy.image.repository`          | `gcr.io/kubebuilder/kube-rbac-proxy`         |                                                                                                                                                                                                                                  |
| `controllerManager.kubeRbacProxy.image.tag`                 | `v0.13.1`                                    |                                                                                                                                                                                                                                  |
| `controllerManager.kubeRbacProxy.resources.limits.cpu`      | `500m`                                       |                                                                                                                                                                                                                                  |
| `controllerManager.kubeRbacProxy.resources.limits.memory`   | `128Mi`                                      |                                                                                                                                                                                                                                  |
| `controllerManager.kubeRbacProxy.resources.requests.cpu`    | `5m`                                         |                                                                                                                                                                                                                                  |
| `controllerManager.kubeRbacProxy.resources.requests.memory` | `64Mi`                                       |                                                                                                                                                                                                                                  |
| `controllerManager.manager.image.repository`                | `ghcr.io/open-feature/open-feature-operator` |                                                                                                                                                                                                                                  |
| `controllerManager.manager.image.tag`                       | `v0.2.31` <!-- x-release-please-version -->  |                                                                                                                                                                                                                                  |
| `controllerManager.manager.resources.limits.cpu`            | `500m`                                       |                                                                                                                                                                                                                                  |
| `controllerManager.manager.resources.limits.memory`         | `128Mi`                                      |                                                                                                                                                                                                                                  |
| `controllerManager.manager.resources.requests.cpu`          | `10m`                                        |                                                                                                                                                                                                                                  |
| `controllerManager.manager.resources.requests.memory`       | `64Mi`                                       |                                                                                                                                                                                                                                  |
| `managerConfig.controllerManagerConfigYaml`                 | `1`                                          |                                                                                                                                                                                                                                  |
| `managerConfig.replicas.health.healthProbeBindAddress`      | `:8081`                                      |                                                                                                                                                                                                                                  |
| `managerConfig.replicas.metrics.bindAddress`                | `127.0.0.1:8080`                             |                                                                                                                                                                                                                                  |
| `managerConfig.replicas.webhook.port`                       | `9443`                                       |                                                                                                                                                                                                                                  |

## Changelog

See [CHANGELOG.md](https://github.com/open-feature/open-feature-operator/blob/main/CHANGELOG.md)
