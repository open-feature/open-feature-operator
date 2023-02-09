# Feature Flag Configuration

The `FeatureFlagConfiguration` version `v1alpha2` CRD defines the a CR with the following example structure:

```yaml
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
    featureFlagSpec: 
        flags: 
            foo: 
                state: "ENABLED"
                variants: 
                    bar: "BAR"
                    baz: "BAZ"
                defaultVariant: "bar"
```

Within the CRD there are 2 main objects, namely the `featureFlagSpec` and the `syncProvider` each offering a different set of configurations for the injected `flagd` sidecars.

## featureFlagSpec

The `featureFlagSpec` is an object representing the flag configurations themselves, the documentation for this object can be found [here](https://github.com/open-feature/flagd/blob/main/docs/configuration/flag_configuration.md).
