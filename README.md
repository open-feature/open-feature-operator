<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/openfeature-horizontal-white.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/openfeature-horizontal-black.svg">
  <img alt="OpenFeature Logo" src="./images/openfeature-horizontal-black.svg">
</picture>

![build](https://img.shields.io/github/workflow/status/open-feature/open-feature-operator/ci)
![goversion](https://img.shields.io/github/go-mod/go-version/open-feature/open-feature-operator/main)
![version](https://img.shields.io/badge/version-pre--alpha-green)
![status](https://img.shields.io/badge/status-not--for--production-red)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/open-feature-operator)](https://artifacthub.io/packages/search?repo=open-feature-operator)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6615/badge)](https://bestpractices.coreinfrastructure.org/projects/6615)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopen-feature%2Fopen-feature-operator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopen-feature%2Fopen-feature-operator?ref=badge_shield)


## Get started

The OpenFeature Operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagD](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagD server for feature flags in a variety of ways. To get started, follow the installation instructions in the [docs](./docs).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to the OpenFeature project.

Our community meetings are held regularly and open to everyone. Check the [OpenFeature community calendar](https://calendar.google.com/calendar/u/0?cid=MHVhN2kxaGl2NWRoMThiMjd0b2FoNjM2NDRAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for specific dates and for the Zoom meeting links.

Thanks so much to our contributors.

<a href="https://github.com/open-feature/flagd/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=open-feature/open-feature-operator" />
</a>

Made with [contrib.rocks](https://contrib.rocks).