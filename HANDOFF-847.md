## namespace isolation for FeatureFlagSource / InProcessConfiguration

Fixes #847 ([GHSA-398h-7f66-3h4p](https://github.com/open-feature/open-feature-operator/security/advisories/GHSA-398h-7f66-3h4p), CVE-2026-54495). Targets `0.10.0` (breaking). Namespace-scoped install (#795) is separate.

### Problem

A pod could reference another namespace's `FeatureFlagSource` / `InProcessConfiguration` via `<ns>/<name>` in its annotation, and the operator would inject that CR's contents (sync URIs, bearer tokens, etc.) into the pod. Separately, the `flagd-kubernetes-sync` ClusterRoleBinding gave any pod using the kubernetes sync provider cluster-wide read on `featureflagsources` / `featureflags`.

### Approach

- **New cluster-scoped CRDs**: `ClusterFeatureFlagSource` (`cffs`) and `ClusterInProcessConfiguration` (`cipc`). These are the sanctioned way for cluster admins to share config across namespaces.
- **Cross-namespace refs denied**: namespaced annotations accept `name` or `<own-ns>/name` (back-compat); `<other-ns>/name` is denied at admission with a message pointing at the cluster annotation.
- **kubernetes-sync RBAC scoped**: namespaced sources get a per-namespace `Role` + `RoleBinding`; only cluster-scoped sources use the cluster-wide binding.

### Annotations

| Annotation | Resolves to | Value |
|---|---|---|
| `openfeature.dev/featureflagsource` | `FeatureFlagSource` | `name` or `<own-ns>/name` |
| `openfeature.dev/clusterfeatureflagsource` (new) | `ClusterFeatureFlagSource` | `name` |
| `openfeature.dev/inprocessconfiguration` | `InProcessConfiguration` | `name` or `<own-ns>/name` |
| `openfeature.dev/clusterinprocessconfiguration` (new) | `ClusterInProcessConfiguration` | `name` |

When a pod sets both a namespaced and a cluster annotation, namespaced merges first and cluster second (cluster wins on conflict).

### Key decisions

- **Cluster specs embed the namespaced spec** (`json:",inline"`), so they share merge/env var logic and can diverge later without a breaking change.
- **CEL for shape validation** (first CEL use in the repo; both rules currently broken, see Known issues):
  - `ClusterFeatureFlagSource`: inner `sources[].source` must be `<ns>/<name>` (bare names are ambiguous at cluster scope).
  - `FeatureFlagSource`: inner `sources[].source` must be bare or same-namespace. Not expressible in CEL; needs to move to the injector/webhook.
  - `InProcessConfiguration` has no equivalent field, so no rule.
- **No injector-side source parsing** (needs revisiting, see Known issues): the plan was for CEL to reject bad shapes at apply time. `InjectFlagd` does take a `clusterScoped` flag, currently used only to pick the RBAC path.
- **Legacy pods keep cluster-wide sync access until restarted**: pods admitted before the upgrade only carry `openfeature.dev/allowkubernetessync=true`. `BackfillPermissions` treats them as cluster-scoped so an operator upgrade doesn't break running workloads. New admissions write `openfeature.dev/kubernetessyncscope`. **Consequence: the advisory isn't fully closed for a workload until its pods roll.** The migration doc covers this.
- **No `FeatureFlagSource` controller changes**: the webhook already blocks cross-namespace refs. The new `ClusterFeatureFlagSource` reconciler matches the namespaced one's restart-on-change behavior.

### Commits (squash-merge; split for review)

1. CRDs: types, CEL rules, editor/viewer ClusterRoles, manager RBAC markers
2. Webhook: `resolveNamespacedRef`, cluster getters, `ErrCrossNamespaceRef`, dispatch + merge order
3. kubernetes-sync RBAC: `EnableKubernetesSyncRBAC`, per-namespace Role/RoleBinding, backfill back-compat
4. `ClusterFeatureFlagSource` reconciler + index
5. Docs, samples, chainsaw e2e (cross-ns CEL denial; cluster-scoped k8s sync happy path), `docs/v0.10_migration.md`

### State / remaining

- Rebased on `main` (post-0.9.3); generated files refreshed (cluster CRD picked up #852 keepalive fields via the embed). Build and unit tests pass.
- Worth a reviewer look: the migration doc's guidance for rolling legacy-marker pods.

### Known issues (e2e and lint failing)

- `FeatureFlagSource` CEL rule doesn't compile: `self.metadata.namespace` isn't available in CRD validation rules. The same-namespace check on inner `sources[].source` needs to move to the injector/webhook (namespaced: reject `<other-ns>/name`; cluster: allow). Until then, that path is unprotected.
- `ClusterFeatureFlagSource` CEL rule exceeds the cost budget; needs `maxItems`/`maxLength` bounds or a cheaper expression.
- `PodMutator.Handle` exceeds `gocognit` (23 > 20); extract the dispatch.
- None of the e2e tests have run successfully yet; both CRDs are rejected at install, so all e2e failures cascade from the two CEL issues.

### Open design questions

| Topic | Alternative | This branch |
|---|---|---|
| Merge order | Cluster is the base; namespaced overrides (e.g. `host`) | Namespaced first; cluster wins on conflict |
| Cluster kind scoping | `spec.conditions`; absent = applies nowhere | No `conditions`; usable from any namespace |
| Cross-ns rollout | `--cross-namespace-refs=allow\|warn\|deny`, warn in 0.9.x first | Hard deny in 0.10.0 |
| Merge semantics | Pointer fields; lists merge by key | Existing merge logic |
| RBAC | `resourceNames`-limited, cleaned up when unused | Namespace-wide Role, not cleaned up |
| Bearer token | `httpSyncBearerTokenSecretRef` | Unchanged |

Two of these should be settled before 0.10.0:

- **Merge order**: the base-layer use case (shared spec, per-namespace `host`) needs namespaced to win, which is the reverse of what's here.
- **`conditions` default**: this is sharing scope, not install scope (#795); cluster kinds are admin-created and meant to be shared, so it's not a disclosure issue. But the default is hard to change later. Adding `conditions` after 0.10.0 is only non-breaking if "absent = all namespaces"; a safe "absent = nowhere" default would need to ship with the CRDs in 0.10.0.

### Out of scope

- Namespace-scoped install (#795); cluster CRDs won't apply there.
- A `ClusterFeatureFlag` CRD.
- Moving `featureflag_webhook.go` JSON Schema validation to CEL.
