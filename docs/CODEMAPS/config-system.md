---
last_updated: 2026-09-25
id: config-system-map
title: "Explain the Configuration System"
sidebar_label: Config System
description: "Execution map for v2 configuration loading, path resolution, references, defaults, validation, and persistence."
doc_type: explanation
audience: "contributors, maintainers, operators"
tags: [configuration, yaml, validation, defaults, paths]
---
# Configuration system

The authoritative cluster model is `internal/config/v2`. The top-level `internal/config` package still owns CLI settings and compatibility constructors; it is not a second configuration schema.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| Cluster identity and paths | path resolution / `internal/core/paths` | [`internal/core/paths/resolver.go`](../../internal/core/paths/resolver.go) → `PathResolver`, `NewPathResolverWithRoots`, `Resolve`, `ResolveWithFallback` | Organization-aware `ClusterPaths`; [`internal/core/paths/resolver_test.go`](../../internal/core/paths/resolver_test.go) |
| Native config load | v2 I/O / `internal/config/v2` | [`internal/config/v2/manager.go`](../../internal/config/v2/manager.go) → `ConfigurationManager.Load`, `loadFromCacheOrDisk`; [`internal/config/v2/io_handler.go`](../../internal/config/v2/io_handler.go) → `ConfigIOHandler` | PathResolver, cache, filesystem, `ConfigLoader`; [`internal/config/v2/io_handler_test.go`](../../internal/config/v2/io_handler_test.go) |
| Decode/normalize/resolve/hydrate/validate/freeze | v2 pipeline / `internal/config/v2` | [`internal/config/v2/loader.go`](../../internal/config/v2/loader.go) → `LoadFromBytes`, `normalize`, `resolveReferences`, `applyDefaults`, `validate`, `freeze` | `defaults.Hydrator`, `ReferenceResolver`, `Validator`; [`internal/config/v2/loader_test.go`](../../internal/config/v2/loader_test.go) |
| References | resolver / `internal/config/v2` | [`internal/config/v2/resolver.go`](../../internal/config/v2/resolver.go) → `ReferenceResolver.Resolve` | `${ref:}`, `${env:}`, `${file:}`; cycle/depth protection; [`internal/config/v2/resolver_test.go`](../../internal/config/v2/resolver_test.go) |
| Defaults | defaults / `internal/config/defaults` | [`internal/config/defaults/hydrator.go`](../../internal/config/defaults/hydrator.go) → `NewHydrator`; [`internal/config/defaults/registry.go`](../../internal/config/defaults/registry.go) → `NewRegistry` | Provider/region/service defaults without replacing explicit values; [`internal/config/defaults/hydrator_test.go`](../../internal/config/defaults/hydrator_test.go) |
| Validation | v2 validator + shared engine | [`internal/config/v2/validator.go`](../../internal/config/v2/validator.go) → `NewValidator`, `defaultValidator.Validate`; [`internal/core/validation/engine.go`](../../internal/core/validation/engine.go) → `ValidationEngine` | Schema, business, provider, deployment, service rules; [`internal/config/v2/validator_pool_test.go`](../../internal/config/v2/validator_pool_test.go) |
| Atomic save | v2 manager / `internal/config/v2` | [`internal/config/v2/manager.go`](../../internal/config/v2/manager.go) → `ConfigurationManager.Save`; [`internal/config/v2/io_handler.go`](../../internal/config/v2/io_handler.go) → `SaveConfig` | Public encode, validation, backup, atomic filesystem write, cache update |
| Schema/editor contract | schema / `internal/config/v2schema` | [`internal/config/v2schema/generator.go`](../../internal/config/v2schema/generator.go) → `Generate` | Generated schema reflects v2 public model; [`internal/config/v2schema/generator_test.go`](../../internal/config/v2schema/generator_test.go) |

## Actual load path

```text
cluster identifier
  -> PathResolver.Resolve / ResolveWithFallback
  -> ConfigurationManager.Load
       -> cache lookup
       -> ConfigIOHandler.LoadConfig
       -> ConfigLoader.LoadFromFile -> LoadFromBytes
            -> DecodePublicConfig
            -> normalize
            -> ReferenceResolver.Resolve
            -> defaults.Hydrator
            -> v2 Validator.Validate
            -> freeze
       -> cache validated model
```

`ConfigurationManager.LoadWithoutValidation` is an explicit test/repair boundary and must not replace `Load` in normal command paths. `Save` validates and serializes the public v2 representation before atomic persistence; the cache is an optimization, never a source of truth.

Consumers include [`internal/cluster`](../../internal/cluster), [`internal/gitops`](../../internal/gitops), [`internal/secretartifacts`](../../internal/secretartifacts), [`internal/secrets`](../../internal/secrets), [`internal/importer`](../../internal/importer), and [`internal/localdev`](../../internal/localdev). OpenStack provider plan/apply and storage apply operate on typed v2 values but add their own stale-file and recovery boundaries; see [OpenStack operations](openstack-provider-storage-operations.md).

## Safe-change boundaries

- Preserve public YAML field names, schema version `2.0`, reference order, default precedence, organization-aware paths, file permissions, and atomic-save behavior.
- Add service configuration through `internal/config/services`, registration/defaults, validation, and schema generation; do not add ad-hoc maps in command handlers.
- Do not treat `internal/config.ConfigManager`, caches, or legacy persistence helpers as authoritative for cluster data.
- Any change to path layout must account for lifecycle, import, secrets, localdev, and existing `PathResolver` tests.
- Provider/storage persistence must not bypass public encode/decode or source-byte checks.

## Related maps

[DI container](di-container.md) · [Cluster lifecycle](cluster-lifecycle.md) · [GitOps engine](gitops-engine.md) · [Secrets management](secrets-management.md) · [Import and resilience](import-operations-and-resilience.md)
