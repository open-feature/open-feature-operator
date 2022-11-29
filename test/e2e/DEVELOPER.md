# E2E Testing

This suite tests the end-to-end deployment of open-feature-operator by deploying an nginx reverse proxy and asserting that curls to the proxy elicit expected behaviour from the flagd sidecar created by open-feature-operator.

## Running on a Kind cluster

```shell
kind create cluster --config ./test/e2e/kind-cluster.yml
IMG=ghcr.io/open-feature/open-feature-operator:main make deploy-operator
IMG=ghcr.io/open-feature/open-feature-operator:main make e2e-test
```

## Running on a Kind cluster using a locally built image

```shell
kind create cluster --config ./test/e2e/kind-cluster.yml
kind load docker-image local-image-tag:latest
IMG=local-image-tag:latest make deploy-operator
IMG=local-image-tag:latest make e2e-test
```

