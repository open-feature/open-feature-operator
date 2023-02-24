# Permissions 

The open feature operator uses the `open-feature-operator-controller-manager` service account, this service account contains the following `RoleBindings`:
- `open-feature-operator-leader-election-role` (role name: `leader-election-role`)
- `open-feature-operator-manager-role` (role name: `manager-role`)
- `open-feature-operator-proxy-role` (role name: `proxy-role`)
- `open-feature-operator-flagd-kubernetes-sync` (role name: `flagd-kubernetes-sync`)

### Leader Election Role

The `leader-election-role` provides the operator with the required permissions to perform leader election.
The definition of this role can be found [here](../config/rbac//leader_election_role.yaml)

| API Group      | Resource | Verbs |
| ----------- | ----------- | ----------- |
| -      | `ConfigMap`       | create, delete, get, list, patch, update, watch       |
| -      | `Event`       | create, patch, |
| `coordination.k8s.io`   | `Lease`        | create, delete, get, list, patch, update, watch       |


### Manager Role

The `manager-role` applies the rules described below, its definition can be found [here](../config/rbac/role.yaml). It provides the operator with sufficient permissions over the `core.openfeature.dev` resources, and the required permissions for injecting the `flagd` sidecar into appropriate pods. The `ConfigMap` permissions are needed to allow the mounting of `FeatureFlagConfiguration` resources for filepath syncs.

| API Group      | Resource | Verbs |
| ----------- | ----------- | ----------- |
| -      | `ConfigMap`       | create, delete, get, list, patch, update, watch       |
| -   | `Pod`        | create, delete, get, list, patch, update, watch       |
| -   | `ServiceAccount`        | get, list, watch       |
| `core.openfeature.dev`   | `FeatureFlagConfiguration`        | create, delete, get, list, patch, update, watch       |
| `core.openfeature.dev`   | `FeatureFlagConfiguration Finalizers`        | update  |
| `core.openfeature.dev`   | `FeatureFlagConfiguration Status`        | get, patch, update  |
| `rbac.authorization.k8s.io`   | `ClusterRoleBinding`    | get, list, update, watch  |

### Proxy Role

The `proxy-role` definition can be found [here](../config/rbac/auth_proxy_role.yaml)

| API Group      | Resource | Verbs |
| ----------- | ----------- | ----------- |
| `authentication.k8s.io`   | `Token Review`        | create       |
| `authentication.k8s.io`   | `Subject Access Review`        | create       |

### Flagd Kubernetes Sync

The `flagd-kubernetes-sync` role providers the permission to get, watch and list all `core.openfeature.dev` resources, permitting the kubernetes sync feature in injected `flagd` containers.
Its definition can be found [here](../config/rbac/flagd_kubernetes_sync_clusterrole.yaml). 
During startup the operator will backfill permissions to the `flagd-kubernetes-sync` cluster role binding from the current state of the cluster, adding all service accounts from pods with the `core.openfeature.dev/enabled` annotation set to `"true"`, preventing unexpected behavior during upgrades.

| API Group      | Resource | Verbs |
| ----------- | ----------- | ----------- |
| `core.openfeature.dev`   | `FlagSourceConfiguration`   | get, watch, list       |
| `core.openfeature.dev`   | `FeatureFlagConfiguration`  | get, watch, list       |

When a `Pod` has the `core.openfeature.dev/enabled` annotation value set to `"true"`, its `Service Account` is added as a subject for this role's `Role Binding`, granting it all required permissions for watching its associated `FeatureFlagConfigurations`. As a result `flagd` can provide real time events describing flag configuration changes.

