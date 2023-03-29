# Docs

This directory contains all OpenFeature Operator documentation, see table of contents below:

## Usage

Follow the documentation below to deploy the open feature operator to your local cluster, followed by a simple example app using `curl` to evaluate a static flag.

- [Installation](./installation.md)
- [Getting Started](./getting_started.md)

## Configuration

Configuration of the deployed sidecars is handled through the `FeatureFlagConfiguration` CRs defined in the `openfeature.dev/featureflagconfiguration` annotation of a deployed `PodSpec`. 
> Further configuration of the operator will be possible in the future, to help contribute [click here](https://github.com/open-feature/open-feature-operator/issues)

- [Annotations](./annotations.md)
- [FeatureFlagConfigurations](./feature_flag_configuration.md)
- [FlagSourceConfiguration](./flag_source_configuration.md)

## Other Resources
- [Architecture](./architecture.md)
- [Permissions](./permissions.md)
- [Development Notes](./development_notes.md)
- [flagd Kube Proxy](./kube_flagd_proxy.md)