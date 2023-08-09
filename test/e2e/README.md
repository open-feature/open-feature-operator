# E2E Testing

This suite tests the end-to-end operation of the open-feature-operator.

Tests are written with [kuttl](https://kuttl.dev/) and assertions are executed from a curl enabled Job.
Ngnix reverse proxy is used as the workload where flagd get injected using OFO annotations.

## Running and validating locally

It is recommended to run and validate e2e test locally before opening a pull request.

To run locally (commands are executed from the project root level),

1. Build the operator locally - `docker build . -t open-feature-operator-local:validate`
2. Create a kind - `kind create cluster --config ./test/e2e/kind-cluster.yml`
3. Load locally build operator image - `kind load docker-image open-feature-operator-local:validate --name kind`
4. Deploy Operator to kind cluster - `IMG=open-feature-operator-local:validate make deploy-operator`
5. Execute kuttl tests - `IMG=open-feature-operator-local:validate make e2e-test-kuttl`

