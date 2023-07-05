# Concepts

This document covers some key concepts related to the OpenFeature operator (OFO). 

For general OpenFeature or feature flag concepts, see our [online documentation](https://openfeature.dev/docs/reference/intro).

## Architecture

The high level architecture of the operator is as follows:  
<p>
    <img src="../images/arch-0.png" width="650">
</p>

## Modes of flag syncs

- Kubernetes:  sync configuration which configures injected flagd sidecar instances to monitor the Kubernetes API 
  for changes in flag configuration custom resources (`FeatureFlagConfiguration`).
- filepath:  sync configuration which creates and mounts ConfigMap files from flag configuration custom  resources
  (`FeatureFlagConfiguration`) and configures injected flagd sidecar instances to monitor them.
- grpc: sync configuration which listen for flagd compatible grpc stream
- http: sync configuration which watch and periodically poll flagd compatible http endpoint
- [flagd-proxy](./flagd_proxy.md)

Each approach have their advantages and disadvantages. 

The kubernetes, grpc and flagd-proxy sync configuration has the advantage of providing near-realtime flag updates(on the order of seconds) to the flagd sidecar. 

For example, Kubernetes syncs require the flagd sidecar(and consequently the workload pod) to communicate with the 
Kubernetes API. This may violate the security or network policies of some organizations.

The `"filepath"` provider requires no such communication, but relies on the fact that [Kubernetes automatically updates mounted ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/#mounted-configmaps-are-updated-automatically). 
The disadvantage of this approach is that flag configuration updates may take as long as two minutes to propagate, depending on cluster configuration:

> "the total delay from the moment when the ConfigMap is updated to the moment when new keys are projected to the Pod can be as long as the kubelet sync period + cache propagation delay"

Consider your individual requirements and select the configuration most appropriate for your needs.