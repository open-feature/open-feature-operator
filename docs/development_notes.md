# Development Notes

## Running the operator locally

The project `Makefile` defines a useful method for locally deploying the operator, allowing for the operator image to be defined:
```
IMG=ghcr.io/open-feature/open-feature-operator:main make deploy-operator
```

## Testing

Run `make test` to run the test suite. The controller integration tests use [envtest](https://book.kubebuilder.io/reference/envtest.html), this sets up and starts an instance of etcd and the Kubernetes API server, without kubelet, controller-manager or other components.
This provides means of asserting that the Kubernetes components reach the desired state without the overhead of using an actual cluster, keeping
test runtime and resource consumption down.

An e2e test suite can also be found in the [`/test/e2e`](../test/e2e/DEVELOPER.md) directory. These tests are run as part of the `pr-lint` github action, they work by deploying an nginx reverse proxy and asserting that curls to the proxy elicit expected behaviour from the flagd sidecar created by open-feature-operator.

## Releases

This repo uses _Release Please_ to release packages. Release Please sets up a running PR that tracks all changes for the library components, and maintains the versions according to [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/), generated when [PRs are merged](https://github.com/amannn/action-semantic-pull-request). When Release Please's running PR is merged, any changed artifacts are published.
