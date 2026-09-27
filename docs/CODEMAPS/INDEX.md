---
last_updated: 2026-09-25
id: codemaps-index
title: "openCenter CLI Architecture Maps"
sidebar_label: Architecture Maps
description: "Canonical topic maps for the openCenter CLI execution paths, ownership boundaries, symbols, dependencies, and test evidence."
doc_type: explanation
audience: "contributors, maintainers, code-oriented agents"
tags: [architecture, codemaps, cli, gitops, providers, secrets]
---
# openCenter CLI architecture maps

This directory is the canonical codemap family. Each map is organized by a user-visible feature, then traces **subsystem → package/module → file → symbol → dependency**. Links point to repository source or tests; statements about behavior are anchored to those paths rather than to archival documentation.

These are topic maps, not package-level inventories. Keep a topic here only when it owns a distinct execution path or change boundary. The GitOps engine and rendering ownership maps intentionally remain separate: the former describes the staged command transaction, while the latter describes action ownership and secret-artifact planning.

## Production execution spine

```text
main.main
  -> di.SetupContainer (outer shutdown/context bridge)
  -> cmd.ExecuteWithContext
       -> initializeApp -> di.NewApp
       -> di.NewAppContainer -> context(AppKey, ContainerKey)
       -> NewBuiltinRootCmd tree
       -> plugins.LoadExternalPlugins (production only)
       -> Cobra ExecuteContext
            -> cmd handler -> internal service/package -> filesystem/API/process
```

The live generation path is:

```text
cluster generate
  -> cmd.runClusterGenerate
  -> di.App.SetupService.Setup
  -> config/path resolution and v2 load/validation
  -> gitops.GenerateClusterTree
  -> staged copy/render -> optional tofu -> SOPS override encryption
  -> manifest validation -> ownership preflight -> promotion
```

## Map index and disposition

| Topic map | Disposition | Start at feature | Primary evidence |
|---|---|---|---|
| [CLI commands](cli-commands.md) | Retained; canonical command-surface map | Built-in Cobra tree and plugin boundary | [`cmd/root.go`](../../cmd/root.go), [`cmd/cluster.go`](../../cmd/cluster.go), [`cmd/ga_command_surface_test.go`](../../cmd/ga_command_surface_test.go) |
| [Config system](config-system.md) | Retained; configuration data-flow map | Load, normalize, resolve, hydrate, validate, save | [`internal/config/v2/loader.go`](../../internal/config/v2/loader.go), [`internal/config/v2/manager.go`](../../internal/config/v2/manager.go), [`internal/config/v2/loader_test.go`](../../internal/config/v2/loader_test.go) |
| [DI container](di-container.md) | Retained; typed wiring map | Process startup and command dependencies | [`internal/di/app.go`](../../internal/di/app.go), [`cmd/root_container_test.go`](../../cmd/root_container_test.go) |
| [Cluster lifecycle](cluster-lifecycle.md) | Retained; end-to-end lifecycle map | init/configure/validate/generate/deploy/destroy | [`internal/cluster/setup_service.go`](../../internal/cluster/setup_service.go), [`internal/cluster/bootstrap_service.go`](../../internal/cluster/bootstrap_service.go), [`cmd/cluster_generate_integration_test.go`](../../cmd/cluster_generate_integration_test.go) |
| [GitOps engine](gitops-engine.md) | Retained; staged generation transaction | Render one private tree and promote it | [`internal/gitops/copy.go`](../../internal/gitops/copy.go), [`internal/gitops/ownership.go`](../../internal/gitops/ownership.go), [`cmd/cluster_generate_render_integration_test.go`](../../cmd/cluster_generate_render_integration_test.go) |
| [Rendering ownership and secret artifacts](rendering-ownership-and-secret-artifacts.md) | Retained; planner/ownership boundary, not a duplicate of GitOps flow | Decide output owner and physical secret target | [`internal/gitops/descriptor_renderer.go`](../../internal/gitops/descriptor_renderer.go), [`internal/secretartifacts/planner.go`](../../internal/secretartifacts/planner.go), [`internal/secretartifacts/planner_test.go`](../../internal/secretartifacts/planner_test.go) |
| [Secrets management](secrets-management.md) | Retained; encrypted manifest, SOPS, and key boundary | Sync manifests or encrypt overlays | [`internal/secrets/manager.go`](../../internal/secrets/manager.go), [`internal/sops/manager.go`](../../internal/sops/manager.go), [`cmd/secrets_sync.go`](../../cmd/secrets_sync.go) |
| [Providers](providers.md) | Retained; capability/routing map | Distinguish config support, bootstrap, drift, and local providers | [`cmd/provider_availability.go`](../../cmd/provider_availability.go), [`internal/cloud/factory.go`](../../internal/cloud/factory.go), [`internal/cluster/bootstrap_service.go`](../../internal/cluster/bootstrap_service.go) |
| [OpenStack provider and storage operations](openstack-provider-storage-operations.md) | Retained; explicit plan/apply map | Read-only provider planning or one-service storage apply | [`cmd/cluster_provider_openstack.go`](../../cmd/cluster_provider_openstack.go), [`cmd/cluster_service_storage.go`](../../cmd/cluster_service_storage.go), package tests |
| [Import, operations, and resilience](import-operations-and-resilience.md) | Retained; adoption/day-2 safety map | Import, drift, backup, locks, retry, circuits | [`internal/importer/scanner.go`](../../internal/importer/scanner.go), [`internal/operations/backup_manager.go`](../../internal/operations/backup_manager.go), [`internal/resilience/lock_manager.go`](../../internal/resilience/lock_manager.go) |
| [Runtime extensions and local development](runtime-extensions-and-local-development.md) | Retained; separate executable/extension map | External plugins or `opencenter-local` | [`internal/plugins/loader.go`](../../internal/plugins/loader.go), [`cmd/opencenter-local/main.go`](../../cmd/opencenter-local/main.go), localdev tests |

No map was deleted or merged: history shows this family was deliberately consolidated in commit [`2852943`](https://github.com/opencenter-cloud/opencenter-cli/commit/2852943), and the current topics have non-overlapping change boundaries.

## Safe-change rules shared by the family

- Preserve command names, hidden-command behavior, serialized v2 YAML, organization-aware paths, ownership ledgers, plugin executable names, embedded asset names, and exit-code semantics unless the change explicitly includes compatibility work.
- Treat `internal/config/v2` as the authoritative cluster-config pipeline; caches and the legacy container are adapters, not alternate sources of truth.
- Do not write generated output directly to the live GitOps tree; use the staged transaction and ownership preflight.
- Keep provider discovery, lifecycle bootstrap, drift, storage provisioning, SOPS encryption, and Kubernetes secret synchronization as separate capabilities.
- When a symbol or path changes, update the topic map and its linked tests in the same change.
