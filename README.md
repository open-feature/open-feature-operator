## open-feature-operator

![version](https://img.shields.io/badge/version-pre--alpha-green)
![status](https://img.shields.io/badge/status-not--for--production-red)

The open-feature-operator is a Kubernetes native operator that allows you to expose feature flags to your applications. It injects a [flagd](https://github.com/open-feature/flagd) sidecar into your pod and allows you to poll the flagd server for feature flags in a variety of ways.

### Architecture

As per the issue [here](https://github.com/open-feature/research/issues/1)
High level architecture is as follows:

<img src="images/arch-0.png" width="560">


### Example

When wishing to leverage featureflagging within the local pod, the following steps are required:

1. Create a new feature flag custom resource e.g.
```
apiVersion: core.openfeature.dev/v1alpha1
kind: FeatureFlagConfiguration
metadata:
  name: featureflagconfiguration-sample
spec:
  featureFlagSpec: |
    {
      "foo" : "bar"
    } 
```

2. Reference the CR within the pod spec annotations
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  annotations:
    openfeature.dev: "enabled"
    openfeature.dev/featureflagconfiguration: "featureflagconfiguration-sample"
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80
```

3. Example usage from host container

```
root@nginx:/# curl localhost:8080
{
  "foo" : "bar"
} 
```

 ### TODO
  
    * [ ] Implement feature flag reconciliation loop
    * [ ] Detect and update configmaps on change
    * [ ] Finalizers
    * [ ] Cleanup on deletion
