# kube-flagd-proxy

> The flagd kube proxy is currently in an experimental state

The `kube-flagd-proxy` is a pub/sub for watching configuration changes in `FeatureFlagConfiguration` CRs without the requiring additional cluster wide permissions in the client pod. In order for a pod to have the required permissions to watch a `FeatureFlagConfiguration` CR in the default implementation, it must have its service account appended to the `flagd-kubernetes-sync` role binding, the details for this role can be found [here](./permissions.md). In some use cases this may not be favorable, in these scenarios the alternative `kube-flagd-proxy` implementation may be used.  

The `kube-flagd-proxy` bypasses the widespread permissions issue by acting as the single source of truth for subscribed flagd instances, broadcasting configuration changes to all subscribed pods via gRPC streams. For each requested `FeatureFlagConfiguration` a new ISync implementation is started, and closed once there are no longer any listeners. this results in only one set of resources requiring the `flagd-kubernetes-sync` permissions, tightening the restrictions on all other pods.

## Architecture

The diagram below describes the high level architecture and implementation of the `kube-flagd-proxy`

<p align="center">
    <img src="../images/kube-flagd-proxy-arch.png" width="95%">
</p>

The `kube-flagd-proxy` is only deployed once the reconcile loop for a `FlagSourceConfiguration` is run with a CR containing the provider `"kube-flagd-proxy"` in its source array.

## Implementation

Update the end to end test in `/config/samples/end-to-end.yaml` to use the `"kube-flagd-proxy"` provider, the source should be a `namespace/name`.

```diff
apiVersion: core.openfeature.dev/v1alpha2
kind: FlagSourceConfiguration
metadata:
  name: end-to-end
  namespace: open-feature-demo
spec:
  sources:
-  - source: open-feature-demo/end-to-end
-    provider: kubernetes
+  - source: open-feature-demo/end-to-end
+    provider: kube-flagd-proxy
```

Deploy the end-to-end demo, this will result in the deployment of the `kube-flagd-proxy` and the required configuration set to the injected flagd sidecar. The end result will be identical to the original end-to-end demo, however the `open-feature-demo-sa` will not be added to the `flagd-kubernetes-sync` role binding.

```sh
kubectl apply -f config/samples/end-to-end.yaml
```

## Configuration

The current implementation of the `kube-flagd-proxy` allows for a set of basic configurations.

| Environment variable | Behavior |
| ---------------------- | -------------------------|
| KUBE_PROXY_IMAGE | Allows for the default kube-flagd-proxy image to be overwritten |
| KUBE_PROXY_TAG | Allows for the default kube-flagd-proxy tag to be overwritten |
| KUBE_PROXY_PORT | Allows the default port of `8015` to eb overwritten  |
| KUBE_PROXY_METRICS_PORT | Allows the default metrics port of `8016` to eb overwritten  |
| KUBE_PROXY_DEBUG_LOGGING | Defaults to `"false"`, allows for the `--debug` flag to be set on the `kube-flagd-proxy` container |
