---
last_updated: 2026-09-25
id: rendering-ownership-and-secret-artifacts
title: "Explain Rendering Ownership and Secret Artifacts"
sidebar_label: Rendering Ownership
description: "Planner map for descriptor/catalog ownership, safe output paths, physical secret-artifact targets, and promotion exclusions."
doc_type: explanation
audience: "contributors, maintainers"
tags: [rendering, ownership, descriptors, catalog, secrets]
---
# Rendering ownership and secret artifacts

This map is intentionally narrower than [GitOps engine](gitops-engine.md): it answers **who owns each output and where a logical secret lands**, while the GitOps map answers **when the staged transaction executes**.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| Descriptor registry | descriptor subsystem / `internal/services/descriptors` | [`internal/services/descriptors/loader.go`](../../internal/services/descriptors/loader.go) → `LoadEmbedded`; [`internal/services/descriptors/types.go`](../../internal/services/descriptors/types.go) → `Registry` | Embedded descriptor metadata; [`internal/services/descriptors/loader_test.go`](../../internal/services/descriptors/loader_test.go) |
| Explicit action planning | GitOps planner / `internal/gitops` | [`internal/gitops/descriptor_renderer.go`](../../internal/gitops/descriptor_renderer.go) → `planClusterAppActions`, `validateClusterAppActions` | v2 config, descriptor registry, normalized output paths; [`internal/gitops/generator_test.go`](../../internal/gitops/generator_test.go) |
| Auto service rendering | immutable catalog / `internal/gitops` | [`internal/gitops/auto_descriptor.go`](../../internal/gitops/auto_descriptor.go) → `planAutoServiceActions`; [`internal/gitops/render_catalog.go`](../../internal/gitops/render_catalog.go) → `RenderCatalog`, `newBuiltInRenderCatalog` | Enabled typed service config, catalog entry, artifact list; [`internal/gitops/auto_descriptor_test.go`](../../internal/gitops/auto_descriptor_test.go) |
| Logical → physical secret mapping | neutral planner / `internal/secretartifacts` | [`internal/secretartifacts/planner.go`](../../internal/secretartifacts/planner.go) → `Plan`, `ValidateTargets` | Typed v2 secrets, `service_secrets`, service enablement; [`internal/secretartifacts/planner_test.go`](../../internal/secretartifacts/planner_test.go) |
| Secret ownership state | artifact state / `internal/secretartifacts` | [`internal/secretartifacts/state.go`](../../internal/secretartifacts/state.go) → `LoadOwnershipState`, ownership records | Consumed by secret sync, not the renderer; planner/state coverage in [`internal/secretartifacts/planner_test.go`](../../internal/secretartifacts/planner_test.go) |
| Tree ownership | promotion / `internal/gitops` | [`internal/gitops/ownership.go`](../../internal/gitops/ownership.go) → `repositoryClusterScopes`, `planRepositoryPromotion`, `applyGeneratedTreePlan` | Ledger hashes/modes, scope checks, custom/Flux exclusions |

## Actual planning path

```text
v2.Config
  -> LoadEmbedded descriptor registry
  -> secretartifacts.Plan
       -> normalize owner/service/key names
       -> group by physical target/path
       -> deterministic merge or conflict
       -> ValidateTargets
  -> planClusterAppActions
       -> explicit descriptor actions
       -> planAutoServiceActions
            -> RenderCatalog lookup
       -> validateClusterAppActions
  -> AtomicWriter workspace
  -> ownership promotion (GitOps map)
```

Descriptor ownership comes from declared roots/files and conditions. Auto rendering applies only to enabled, non-external services without an explicit descriptor and with a built-in catalog entry. `validateClusterAppActions` rejects empty, unsafe, escaping, or disagreement-prone output paths before an action can write.

`secretartifacts.Plan` is backend- and renderer-independent. It normalizes names, maps Grafana to the `kube-prometheus-stack` physical target, merges owners only when canonical keys agree, sorts results, and rejects missing/disabled explicit targets. `internal/secrets` later encrypts/materializes these artifacts; storage credential provisioning does not do so.

## Safe-change boundaries

- A new renderer must declare ownership, create a plan, pass coverage and containment checks, and render into the workspace; it must not write directly to the final overlay.
- Do not infer generator ownership from existing files. Existing unowned secret files and modified tracked generated files must block adoption.
- Preserve deterministic owner/key normalization, target enablement checks, artifact paths, and secret ownership state.
- Keep descriptor and catalog decisions compiled/testable; do not make config-selected renderer names or runtime plugins authoritative.
- Changes to service secrets require checking `planner_test.go`, secrets sync tests, and the relevant GitOps catalog/descriptor tests.

## Related maps

[GitOps engine](gitops-engine.md) · [Secrets management](secrets-management.md) · [Config system](config-system.md)
