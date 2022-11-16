## Guidelines

Welcome!

There are a few things to consider before contributing to open-feature-operator.

Firstly, there's [a code of conduct](https://github.com/open-feature/.github/blob/main/CODE_OF_CONDUCT.md).
TLDR: be respectful.

Any contributions are expected to include tests. These can be validated with `make test` or the automated github workflow will run them on PR creation.

The go version in the `go.mod` is the currently supported version of go.

Thanks! Issues and pull requests following these guidelines are welcome.

## Development

### FeatureFlagConfiguration custom resource definition versioning
Custom resource definitions support multiple versions. The kubebuilder framework exposes a system to seamlessly convert between versions (using a "hub and spoke" model) maintaining backwards compatibility. It does this by injecting conversion webhooks that call our defined convert functions. The hub version of the `FeatureFlagConfiguration` custom resource definition (the version to which all other versions are converted) is `v1alpha1`.
Follow [this tutorial](https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html) to implement a new version of the custom resource definition.
