# Documentation

This directory contains OpenFeature Operator documentation.

Interested in having OpenFeature Operator up and running within 5 minutes? Follow the quick start guide found below.

- [Quick Start](./quick_start.md)

Follow the detailed installation guide to deploy open feature operator to your local cluster, 

- [Installation](./installation.md)

## Configuration

Configuration of the deployed sidecars is handled through the `FlagSourceConfiguration` custom resources defined at`openfeature.dev/flagsourceconfiguration` annotation of the deployed `PodSpec`.

The relationship between the deployment and custom resources is highlighted in the diagram below,

```mermaid
flowchart TD
    A[Pod]-->|Annotation: openfeature.dev/flagsourceconfiguration| B[FlagSourceConfiguration CR]
    B--> |Flag source| C[FeatureFlagConfiguration CR]
    B--> |Flag source| D[HTTP sync]
    B--> |Flag source| E[Filepath sync]
    B--> |Flag source| F[GRPC sync]
    B--> |Flag source| G[flagd-proxy]
```

To configure and understand more,

- Deployment configurations: [Annotations](./annotations.md)
- Define flag sources for the deployment: [FlagSourceConfiguration](./flag_source_configuration.md)
- Define feature flags as custom resource: [FeatureFlagConfigurations](./feature_flag_configuration.md)

## Other Resources
- [Permissions](./permissions.md)
- [Concepts](./concepts.md)
- [Development Notes](./development_notes.md)
- [Threat model](./threat_model.md)
- [flagd Kube Proxy](./flagd_proxy.md)
- [API Reference](./crds.md)