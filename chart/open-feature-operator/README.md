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

This installation will use the default helm configuration, described in the [Configuration section](#configuration)
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

For the `featureflags.core.openfeature.dev` CRD:

```sh
helm template openfeature/open-feature-operator -s templates/apiextensions.k8s.io_v1_customresourcedefinition_featureflags.core.openfeature.dev.yaml | kubectl apply -f -
```

For the `featureflagsources.core.openfeature.dev` CRD:

```sh
helm template openfeature/open-feature-operator -s templates/apiextensions.k8s.io_v1_customresourcedefinition_featureflagsources.core.openfeature.dev.yaml | kubectl apply -f -
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

### Global

| Name               | Description                                                                                                                                                                                                  | Value                          |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------ |
| `defaultNamespace` | To override the namespace use the `--namespace` flag. This default is provided to ensure that the kustomize build charts in `/templates` deploy correctly when no `namespace` is provided via the `-n` flag. | `open-feature-operator-system` |

### Sidecar configuration

| Name                                       | Description                                                                                                                                                                                                                                                 | Value                        |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| `sidecarConfiguration.port`                | Sets the value of the `XXX_PORT` environment variable for the injected sidecar.                                                                                                                                                                             | `8013`                       |
| `sidecarConfiguration.managementPort`      | Sets the value of the `XXX_MANAGEMENT_PORT` environment variable for the injected sidecar.                                                                                                                                                                  | `8014`                       |
| `sidecarConfiguration.socketPath`          | Sets the value of the `XXX_SOCKET_PATH` environment variable for the injected sidecar.                                                                                                                                                                      | `""`                         |
| `sidecarConfiguration.image.repository`    | Sets the image for the injected sidecar.                                                                                                                                                                                                                    | `ghcr.io/open-feature/flagd` |
| `sidecarConfiguration.image.tag`           | Sets the version tag for the injected sidecar.                                                                                                                                                                                                              | `v0.7.2`                     |
| `sidecarConfiguration.providerArgs`        | Used to append arguments to the sidecar startup command. This value is a comma separated string of key values separated by '=', e.g. `key=value,key2=value2` results in the appending of `--sync-provider-args key=value --sync-provider-args key2=value2`. | `""`                         |
| `sidecarConfiguration.envVarPrefix`        | Sets the prefix for all environment variables set in the injected sidecar.                                                                                                                                                                                  | `FLAGD`                      |
| `sidecarConfiguration.defaultSyncProvider` | Sets the value of the `XXX_SYNC_PROVIDER` environment variable for the injected sidecar container. There are 4 valid sync providers: `kubernetes`, `grpc`, `file` and `http`.                                                                               | `kubernetes`                 |
| `sidecarConfiguration.evaluator`           | Sets the value of the `XXX_EVALUATOR` environment variable for the injected sidecar container.                                                                                                                                                              | `json`                       |
| `sidecarConfiguration.logFormat`           | Sets the value of the `XXX_LOG_FORMAT` environment variable for the injected sidecar container. There are 2 valid log formats: `json` and `console`.                                                                                                        | `json`                       |
| `sidecarConfiguration.probesEnabled`       | Enable or Disable Liveness and Readiness probes of the flagd sidecar. When enabled, HTTP probes( paths - `/readyz`, `/healthz`) are set with an initial delay of 5 seconds.                                                                                 | `true`                       |
| `sidecarConfiguration.debugLogging`        | Controls the addition of the `--debug` flag to the container startup arguments.                                                                                                                                                                             | `false`                      |
| `sidecarConfiguration.otelCollectorUri`    | Otel exporter uri.                                                                                                                                                                                                                                          | `""`                         |
| `sidecarConfiguration.resources`           | Override resources of the flagd sidecar.                                                                                                                                                                                                                    | `{}`                         |

### Flagd-proxy configuration

| Name                                       | Description                                                                     | Value                              |
| ------------------------------------------ | ------------------------------------------------------------------------------- | ---------------------------------- |
| `flagdProxyConfiguration.port`             | Sets the port to expose the sync API on.                                        | `8015`                             |
| `flagdProxyConfiguration.managementPort`   | Sets the port to expose the management API on.                                  | `8016`                             |
| `flagdProxyConfiguration.image.repository` | Sets the image for the flagd-proxy deployment.                                  | `ghcr.io/open-feature/flagd-proxy` |
| `flagdProxyConfiguration.image.tag`        | Sets the tag for the flagd-proxy deployment.                                    | `v0.3.2`                           |
| `flagdProxyConfiguration.debugLogging`     | Controls the addition of the `--debug` flag to the container startup arguments. | `false`                            |

### Operator resource configuration

| Name                                                                      | Description                                              | Value                                        |
| ------------------------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------- |
| `controllerManager.kubeRbacProxy.image.repository`                        | Sets the image for the kube-rbac-proxy.                  | `gcr.io/kubebuilder/kube-rbac-proxy`         |
| `controllerManager.kubeRbacProxy.image.tag`                               | Sets the version tag for the kube-rbac-proxy.            | `v0.14.1`                                    |
| `controllerManager.kubeRbacProxy.resources.limits.cpu`                    | Sets cpu resource limits for kube-rbac-proxy.            | `500m`                                       |
| `controllerManager.kubeRbacProxy.resources.limits.memory`                 | Sets memory resource limits for kube-rbac-proxy.         | `128Mi`                                      |
| `controllerManager.kubeRbacProxy.resources.requests.cpu`                  | Sets cpu resource requests for kube-rbac-proxy.          | `5m`                                         |
| `controllerManager.kubeRbacProxy.resources.requests.memory`               | Sets memory resource requests for kube-rbac-proxy.       | `64Mi`                                       |
| `controllerManager.manager.image.repository`                              | Sets the image for the operator.                         | `ghcr.io/open-feature/open-feature-operator` |
| `controllerManager.manager.image.tag`                                     | Sets the version tag for the operator.                   | `v0.5.3`                                     |
| `controllerManager.manager.resources.limits.cpu`                          | Sets cpu resource limits for operator.                   | `500m`                                       |
| `controllerManager.manager.resources.limits.memory`                       | Sets memory resource limits for operator.                | `128Mi`                                      |
| `controllerManager.manager.resources.requests.cpu`                        | Sets cpu resource requests for operator.                 | `10m`                                        |
| `controllerManager.manager.resources.requests.memory`                     | Sets memory resource requests for operator.              | `64Mi`                                       |
| `controllerManager.replicas`                                              | Sets number of replicas of the OpenFeature operator pod. | `1`                                          |
| `managerConfig.controllerManagerConfigYaml.health.healthProbeBindAddress` | Sets the bind address for health probes.                 | `:8081`                                      |
| `managerConfig.controllerManagerConfigYaml.metrics.bindAddress`           | Sets the bind address for metrics.                       | `127.0.0.1:8080`                             |
| `managerConfig.controllerManagerConfigYaml.webhook.port`                  | Sets the bind address for webhook.                       | `9443`                                       |
