<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/openfeature-horizontal-white.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/openfeature-horizontal-black.svg">
  <img alt="OpenFeature Logo" src="./images/openfeature-horizontal-black.svg">
</picture>

![build](https://img.shields.io/github/actions/workflow/status/open-feature/open-feature-operator/pr-checks.yml?branch=main)
![goversion](https://img.shields.io/github/go-mod/go-version/open-feature/open-feature-operator/main)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/open-feature-operator)](https://artifacthub.io/packages/search?repo=open-feature-operator)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6615/badge)](https://bestpractices.coreinfrastructure.org/projects/6615)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopen-feature%2Fopen-feature-operator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopen-feature%2Fopen-feature-operator?ref=badge_shield)

## Get started

The OpenFeature Operator allows you to expose feature flags to your applications. 
It injects a [flagd](https://github.com/open-feature/flagd) sidecar into relevant pods and exposes gRPC and HTTP interfaces for flag evaluation.
To get started, follow the installation instructions in the [docs](./docs).

> [!NOTE]
> With version [v0.5.0](https://github.com/open-feature/open-feature-operator/releases/tag/v0.5.0), we have migrated 
> to API version `v1beta1`. Please check the [migration guide](./docs/v1beta_migration.md) to migrate from old configurations.

## Demos

- [Try the OpenFeature Operator locally on your machine](https://openfeature.dev/docs/tutorials/ofo)
- [Try the OpenFeature Operator in the Killercoda Playground (in browser)](https://killercoda.com/open-feature/scenario/openfeature-operator-demo)

## Changelog

See [CHANGELOG.md](https://github.com/open-feature/open-feature-operator/blob/main/CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to the OpenFeature project.

Our community meetings are held regularly and open to everyone, as well as other community channels.
Check the [OpenFeature community page]https://openfeature.dev/community/) for the links and participation guidelines.

Thanks so much to our contributors.

<a href="https://github.com/open-feature/flagd/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=open-feature/open-feature-operator" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
