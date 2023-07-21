# Troubleshooting 

This section contain some common issues you can face while installing, operating the operator and possible solutions for them.

## Service account and custom resource access errors

When using `kubernetes` flag sync method, operator rely on K8s RBAC to grant injected flagd access to custom resources.
If your K8s cluster has permission restrictions or if you have cluster configurations as code which can override `ClusterRoleBinding` with a new rollout, then there can be operator errors.

For example, if you see error such as,

```sh
Error creating: admission webhook <WEBHOOK_NAME> denied the request: ServiceAccount <NAME> not found
```

```sh
User <SERVICE_ACCOUNT> cannot get resource <FLAG_CONFIGURATION_CR> in API group "core.openfeature.dev" in the namespace <NAMESPACE>
```

then, please check if you have correct `ClusterRoleBinding` configuration under `open-feature-operator-flagd-kubernetes-sync`.

> kubectl describe ClusterRoleBinding open-feature-operator-flagd-kubernetes-sync

And you must see your workload namespace listed there,

>ServiceAccount default <NAMESPACE>