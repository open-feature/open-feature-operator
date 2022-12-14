# OpenFeature Operator

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagD](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagD server for feature flags in a variety of ways.
The documentation for this project can be found here: [OpenFeature Operator](https://github.com/open-feature/open-feature-operator)

## Prerequisites

the OpenFeature Operator requires cert manager ot be installed on the target cluster.


## Values

| Value       | Default     | Explanation |
| ----------- | ----------- | ----------- |
| `defaultNamespace`      | `open-feature-operator`  | [INTERNAL USE ONLY] To override the namespace use the `--namespace` flag. This default is provided to ensure that the kustomize build charts in `/templates` deploy correctly when no `namespace` is provided via the `-n` flag.|
