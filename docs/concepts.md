# Concepts

This document covers some key concepts related to the OpenFeature operator (OFO). For general OpenFeature or feature flag concepts, see our [online documentation](https://docs.openfeature.dev/docs/reference/intro).

## Architecture

The high level architecture of the operator is as follows:  
<p align="center">
    <img src="../images/arch-0.png" width="650">
</p>

## Modes of operation

OFO supports two primary modes of operation for supplying flag configurations to sidecar flagd instances:

- The `"kubernetes"` sync configuration (default), which configures injected flagd sidecar instances to monitor the Kubernetes API for changes in flag configuration custom resources (`featureflagconfiguration`).
- The `"filepath"` sync configuration, which creates and mounts ConfigMap files from flag configuration custom resources (`featureflagconfiguration`) and configures injected flagd sidecar instances to monitor them.

Both approaches have their advantages and disadvantages. The `"kubernetes"` sync configuration has the advantage of providing near-realtime flag updates (on the order of seconds) to the flagd sidecar, and therefore the associated workload. However, it also requires the flagd sidecar (and consequently the workload pod) to communicate with the Kubernetes API. This may violate the security or network policies of some organizations. The `"filepath"` provider requires no such communication, but relies on the fact that [Kubernetes automatically updates mounted ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/#mounted-configmaps-are-updated-automatically). The disadvantage of this approach is that flag configuration updates may take as long as two minutes to propagate, depending on cluster configuration:

> "the total delay from the moment when the ConfigMap is updated to the moment when new keys are projected to the Pod can be as long as the kubelet sync period + cache propagation delay"

Consider your individual requirements and select the configuration most appropriate for your needs. Note that the sync provider configuration to use can be configured globally and overridden per `featureflagconfiguration`. For details, see [the syncProvider documentation](./feature_flag_configuration.md#syncprovider).