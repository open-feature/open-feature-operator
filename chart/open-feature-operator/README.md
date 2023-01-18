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
| `sidecarConfiguration.envVarPrefix`      | `FLAGD`  | Sets the prefix for all environment variables set in the injected sidecar |
| `sidecarConfiguration.port`      | 8013  | Sets the value of the `XXX_PORT` environment variable for the injected sidecar container. (`flagd` behavior: sets the port for `flagd` to listen on)|
| `sidecarConfiguration.metricsPort`      | 8014  | Sets the value of the `XXX_METRICS_PORT` environment variable for the injected sidecar container. (`flagd` behavior: sets the port for `flagd` serve metrics on)|
| `sidecarConfiguration.socketPath`      | `""`  | Sets the value of the `XXX_SOCKET_PATH` environment variable for the injected sidecar container. (`flagd` behavior: sets the socket path for `flagd` to listen on)|
| `sidecarConfiguration.repository`      | `ghcr.io/open-feature/flagd`  | Sets the image for the injected sidecar container|
| `sidecarConfiguration.tag`      | `main`  | Sets the version tag for the injected sidecar container |
| `sidecarConfiguration.providerArgs`      | `""`  | Used to append arguments to the sidecar startup command. This value is a comma separated string of key values separated by '=',
e.g. `key=value,key2=value2` results in the appending of `--sync-provider-args key=value --sync-provider-args key2=value2` |

## Changelog

See [CHANGELOG.md](https://github.com/open-feature/open-feature-operator/blob/main/CHANGELOG.md)
