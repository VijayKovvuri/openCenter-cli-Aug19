---
last_updated: 2026-09-25
id: cluster-lifecycle-map
title: "Explain the Cluster Lifecycle"
sidebar_label: Cluster Lifecycle
description: "Execution map for cluster initialization, configuration, validation, GitOps generation, provider bootstrap, day-2 operations, and destruction."
doc_type: explanation
audience: "contributors, maintainers, operators"
tags: [clusters, lifecycle, bootstrap, gitops, providers]
---
# Cluster lifecycle

`cmd/` selects an operation and translates flags. `internal/cluster` orchestrates lifecycle services; config, GitOps, cloud, security, and resilience packages perform bounded work.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| Init | lifecycle / `internal/cluster` | [`internal/cluster/init_service.go`](../../internal/cluster/init_service.go) → `InitService.Initialize`, `NewInitServiceWithConfigMgr` | `PathResolver`, defaults, v2 config, SOPS/Age key setup; [`cmd/cluster_init_integration_test.go`](../../cmd/cluster_init_integration_test.go) |
| Configure | guided config / `internal/cluster` | [`internal/cluster/configure_service.go`](../../internal/cluster/configure_service.go) → `ConfigureService.Configure` | orchestration registries, v2 persistence; [`cmd/cluster_configure_test.go`](../../cmd/cluster_configure_test.go) |
| Validate/doctor | readiness / `internal/cluster` | [`internal/cluster/validate_service.go`](../../internal/cluster/validate_service.go) → `ValidateService.Validate`; [`cmd/cluster_validate.go`](../../cmd/cluster_validate.go) → `newClusterValidateCmd` | v2 validator, shared validation, optional provider discovery; [`cmd/cluster_validate_integration_test.go`](../../cmd/cluster_validate_integration_test.go) |
| Generate | GitOps orchestration / `internal/cluster` | [`internal/cluster/setup_service.go`](../../internal/cluster/setup_service.go) → `SetupService.Setup`, `generateGitOpsManifestsWithPromotion` | v2 config, SOPS, `gitops.GenerateClusterTree`, optional OpenTofu; [`cmd/cluster_generate_render_integration_test.go`](../../cmd/cluster_generate_render_integration_test.go) |
| Deploy/bootstrap | lifecycle / `internal/cluster` | [`internal/cluster/bootstrap_service.go`](../../internal/cluster/bootstrap_service.go) → `BootstrapService.Bootstrap`, `buildBootstrapSteps` | provider-specific `lifecycleBootstrapProvider`, `security.CommandRunner`, persisted `bootstrap-state.json`; [`internal/cluster/bootstrap_provider_infra_test.go`](../../internal/cluster/bootstrap_provider_infra_test.go) |
| Destroy | lifecycle / `internal/cluster` | [`internal/cluster/destroy_service.go`](../../internal/cluster/destroy_service.go) → `DestroyInfrastructure`, `getDestroyProvider` | OpenTofu, Kind, or Magnum destroy provider; [`internal/cluster/destroy_service_test.go`](../../internal/cluster/destroy_service_test.go) |
| Day-2 service/pool/drift/backup/lock | command + bounded subsystems | [`cmd/cluster_service.go`](../../cmd/cluster_service.go), [`cmd/cluster_pool.go`](../../cmd/cluster_pool.go), [`cmd/cluster_drift.go`](../../cmd/cluster_drift.go), [`cmd/cluster_backup.go`](../../cmd/cluster_backup.go), [`cmd/cluster_lock.go`](../../cmd/cluster_lock.go) | These do not own generation or config loading; see [operations map](import-operations-and-resilience.md) |
| Import entry | adoption / `internal/importer` | [`cmd/cluster_import.go`](../../cmd/cluster_import.go) → scan/report/apply constructors | Separate entry into existing repositories; see [Import map](import-operations-and-resilience.md) |

## Actual paths

```text
cluster init
  -> cmd handler -> InitService.Initialize -> paths/default config/key material -> v2 persistence
cluster configure
  -> ConfigureService -> orchestration capability/provider handlers -> v2 save
cluster validate
  -> ValidateService.Validate -> load/readiness/provider/service/GitOps reports
cluster generate
  -> SetupService.Setup -> load + ValidateForGeneration
  -> gitops.GenerateClusterTree (see GitOps map)
cluster deploy
  -> BootstrapService.Bootstrap -> provider BuildSteps
  -> bootstrap state + sanitized external commands -> endpoint/readiness
cluster destroy
  -> DestroyService.DestroyInfrastructure -> destroy provider BuildSteps
```

`SetupService.Setup` resolves organization-aware paths, loads v2 config, requires schema `2.0` and a configured GitOps directory, then passes one staged generation request. `BootstrapService` selects provider steps, records statuses and supports `OnlyStep`/`FromStep`; it is not a renderer. Destroy has its own provider contract.

Provider routing is explicit: [`bootstrap_service.go`](../../internal/cluster/bootstrap_service.go) chooses OpenStack/VMware/Baremetal shared infrastructure steps, Kind steps, or Magnum steps. OpenTofu materialization is skipped for Kind and Magnum during generation. OpenStack provider/storage plan/apply are separate local/provider operations, not deploy or generate.

## Safe-change boundaries

- Keep command flag parsing in `cmd/` and lifecycle orchestration in `internal/cluster`; do not put provider API calls or template writes in handlers.
- Preserve bootstrap step IDs, ordering, persisted state version, resume semantics, sanitized command execution, and log paths.
- Preserve the distinction between `Validate`, `ValidateForGeneration`, and deployment validation; a successful config read is not proof that bootstrap is safe.
- Changes to `PathResolver`, v2 config, generation, or provider routing require checking init, validate, generate, deploy, destroy, and integration tests.
- Keep day-2 drift, backup, import, and locks out of the generation transaction.

## Related maps

[Config system](config-system.md) · [GitOps engine](gitops-engine.md) · [Providers](providers.md) · [OpenStack operations](openstack-provider-storage-operations.md) · [Import and resilience](import-operations-and-resilience.md)
