# OpenFeature Operator

## TL;DR
> This helm chart has a dependency on [cert manager](https://cert-manager.io/docs/installation/)
```
helm repo add openfeature https://open-feature.github.io/open-feature-operator/
helm repo update
helm upgrade -i openfeature openfeature/open-feature-operator
```

## Introduction

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagd](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagd server for feature flags in a variety of ways.
The full documentation for this project can be found here: [OpenFeature Operator](https://github.com/open-feature/open-feature-operator/tree/main/docs)

## Prerequisites

The OpenFeature Operator requires [cert manager](https://cert-manager.io/docs/installation/) to be installed on the target cluster.

## Install

To install/upgrade the chart with the release name `open-feature-operator`:
```
helm upgrade -i open-feature-operator openfeature/open-feature-operator
```
This installation will use the default helm configuration, described in the [configuration section](#configuration)
To overwrite these default values use the `--set` flag when calling `helm upgrade` or `helm install`, for example: 
```
helm upgrade -i open-feature-operator ./chart/open-feature-operator --set sidecarConfiguration.port=8080 --set controllerManager.kubeRbacProxy.resources.limits.cpu=400m
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
| Value       | Default     | Explanation                                                                                                                                                                                                                                               |
| ----------- | ----------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `sidecarConfiguration.envVarPrefix`      | `FLAGD`  | Sets the prefix for all environment variables set in the injected sidecar.                                                                                                                                                                                |
| `sidecarConfiguration.port`      | 8013  | Sets the value of the `XXX_PORT` environment variable for the injected sidecar container.                                                                                                                                                                 |
| `sidecarConfiguration.metricsPort`      | 8014  | Sets the value of the `XXX_METRICS_PORT` environment variable for the injected sidecar container.                                                                                                                                                         |
| `sidecarConfiguration.socketPath`      | `""`  | Sets the value of the `XXX_SOCKET_PATH` environment variable for the injected sidecar container.                                                                                                                                                          |
| `sidecarConfiguration.image.repository`      | `ghcr.io/open-feature/flagd`  | Sets the image for the injected sidecar container.                                                                                                                                                                                                        |
| `sidecarConfiguration.image.tag`      | current flagd version: `v0.3.4`  | Sets the version tag for the injected sidecar container.                                                                                                                                                                                                  |
| `sidecarConfiguration.providerArgs`      | `""`  | Used to append arguments to the sidecar startup command. This value is a comma separated string of key values separated by '=', e.g. `key=value,key2=value2` results in the appending of `--sync-provider-args key=value --sync-provider-args key2=value2` |
| `sidecarConfiguration.defaultSyncProvider`      | `kubernetes`  | Sets the value of the `XXX_SYNC_PROVIDER` environment variable for the injected sidecar container. There are 3 valid sync providers: `kubernetes`, `filepath` and `http`                                                                                  |
| `sidecarConfiguration.logFormat` | `json` | Sets the value of the `XXX_LOG_FORMAT` environment variable for the injected sidecar container.                                                                                                                                                                          |

### Operator resource configuration

| Value       | Default     |
| ----------- | ----------- |
| `defaultNamespace`      | `open-feature-operator`  | [INTERNAL USE ONLY] To override the namespace use the `--namespace` flag. This default is provided to ensure that the kustomize build charts in `/templates` deploy correctly when no `namespace` is provided via the `-n` flag.|
| `controllerManager.kubeRbacProxy.image.repository` | `gcr.io/kubebuilder/kube-rbac-proxy` |
| `controllerManager.kubeRbacProxy.image.tag` | `v0.13.1` |
| `controllerManager.kubeRbacProxy.resources.limits.cpu` | `500m` |
| `controllerManager.kubeRbacProxy.resources.limits.memory` | `128Mi` |
| `controllerManager.kubeRbacProxy.resources.requests.cpu` | `5m` |
| `controllerManager.kubeRbacProxy.resources.requests.memory` | `64Mi` |
| `controllerManager.manager.image.repository` | `ghcr.io/open-feature/open-feature-operator` |
| `controllerManager.manager.image.tag` | <!-- x-release-please-start-version --> `v0.2.28` <!-- x-release-please-end --> |
| `controllerManager.manager.resources.limits.cpu` | `500m` |
| `controllerManager.manager.resources.limits.memory` | `128Mi` |
| `controllerManager.manager.resources.requests.cpu` | `10m` |
| `controllerManager.manager.resources.requests.memory` | `64Mi` |
| `managerConfig.controllerManagerConfigYaml` | `1` |
| `managerConfig.replicas.health.healthProbeBindAddress` | `:8081` |
| `managerConfig.replicas.metrics.bindAddress` | `0.2.29.1:8080` |
| `managerConfig.replicas.webhook.port` | `9443` |

## Changelog

See [CHANGELOG.md](https://github.com/open-feature/open-feature-operator/blob/main/CHANGELOG.md)
