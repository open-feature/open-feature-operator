# E2E Testing

This suite tests the end-to-end operation of the open-feature-operator.

Tests are written with [kuttl](https://kuttl.dev/) and assertions are executed from a curl enabled Job.
Ngnix reverse proxy is used as the workload where flagd get injected using OFO annotations.

## Running and validating locally

It is recommended to run and validate e2e test locally before opening a pull request.

To run locally (commands are executed from the project root level),

1. Build the operator locally - `docker build . -t open-feature-operator-local:validate`
2. Create a kind cluster - `kind create cluster --config ./test/e2e/kind-cluster.yml --name e2e-tests`
3. Load locally build operator image - `kind load docker-image open-feature-operator-local:validate --name e2e-tests`
4. Deploy Operator to kind cluster - `IMG=open-feature-operator-local:validate make deploy-operator`
5. Execute kuttl tests - `IMG=open-feature-operator-local:validate make e2e-test-kuttl`

Alternatively, you can use `e2e-test-validate-local` Makefile rule to execute all above and cleanup the kind cluster,

> make e2e-test-validate-local

After the test run, make sure test status by validating kuttl output,

```text
--- PASS: kuttl (48.71s)
    --- PASS: kuttl/harness (0.00s)
        --- PASS: kuttl/harness/assets (0.01s)
        --- PASS: kuttl/harness/flagd-disabled (12.58s)
        --- PASS: kuttl/harness/inject-flagd (26.41s)
        --- PASS: kuttl/harness/fsconfig-file-sync (31.73s)
        --- PASS: kuttl/harness/fsconfig-k8s-sync (31.74s)
        --- PASS: kuttl/harness/fsconfig-flagd-proxy-sync (48.49s)
```

### Running individual tests

You can use kuttl command options to execute individual tests. Consider the example command below,

>$ kubectl kuttl test --start-kind=false ./test/e2e/kuttl --config=kuttl-test.yaml --test=flagd-disabled