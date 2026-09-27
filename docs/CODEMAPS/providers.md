---
last_updated: 2026-09-25
id: providers-map
title: "Explain Provider Capability Boundaries"
sidebar_label: Providers
description: "Capability and routing map separating configuration acceptance, GitOps generation, lifecycle bootstrap, drift, and storage operations."
doc_type: explanation
audience: "contributors, maintainers, operators"
tags: [providers, openstack, magnum, vmware, baremetal, kind]
---
# Providers

Provider support is not one interface. Configuration validation, rendered infrastructure, lifecycle bootstrap/destroy, drift, OpenStack planning, and object storage provisioning have separate symbols and dependencies.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| CLI availability gate | command policy / `cmd` | [`cmd/provider_availability.go`](../../cmd/provider_availability.go) → `checkProviderAvailability` | Called by init/generate/deploy; command-surface and lifecycle tests in [`cmd/ga_command_surface_test.go`](../../cmd/ga_command_surface_test.go) |
| Config provider validation | v2 config / `internal/config/v2` | [`internal/config/v2/validator.go`](../../internal/config/v2/validator.go) → `ValidateProvider`; [`internal/config/v2/readiness.go`](../../internal/config/v2/readiness.go) → `ValidateReadiness` | Schema-valid names can exceed reachable CLI capabilities; [`internal/config/v2/provider_test.go`](../../internal/config/v2/provider_test.go) |
| Lifecycle bootstrap | cluster / `internal/cluster` | [`internal/cluster/bootstrap_provider.go`](../../internal/cluster/bootstrap_provider.go) → `lifecycleBootstrapProvider.BuildSteps`; [`internal/cluster/bootstrap_service.go`](../../internal/cluster/bootstrap_service.go) → provider selection | Security command runner, config, paths, persisted state; provider tests |
| OpenStack/VMware/Baremetal bootstrap | infrastructure provider | [`internal/cluster/bootstrap_provider_infra.go`](../../internal/cluster/bootstrap_provider_infra.go) → `newOpenStackBootstrapProvider`, `BuildSteps` | OpenTofu/Kubespray-style commands and provider environment |
| Kind bootstrap | local cloud provider | [`internal/cluster/kind_bootstrap_provider.go`](../../internal/cluster/kind_bootstrap_provider.go) → `newKindBootstrapProvider`, `BuildSteps`; [`internal/cloud/kind/provider.go`](../../internal/cloud/kind/provider.go) | Kind runtime and kubeconfig; [`cmd/kind_provider_workflow_test.go`](../../cmd/kind_provider_workflow_test.go) |
| Magnum bootstrap/destroy | managed cloud provider | [`internal/cluster/magnum_bootstrap_provider.go`](../../internal/cluster/magnum_bootstrap_provider.go) → `newMagnumBootstrapProvider`, `BuildSteps`; [`internal/cluster/magnum_destroy_provider.go`](../../internal/cluster/magnum_destroy_provider.go) | [`internal/cloud/magnum/provider.go`](../../internal/cloud/magnum/provider.go), existing cluster template; [`internal/cluster/magnum_lifecycle_test.go`](../../internal/cluster/magnum_lifecycle_test.go) |
| Drift capability | cloud registry / `internal/cloud` | [`internal/cloud/factory.go`](../../internal/cloud/factory.go) → `CloudProvider`, `CloudProviderFactory`, `NewCloudProviderFactory` | OpenStack/VMware implementations can register; lifecycle deploy providers are outside this factory; [`internal/cloud/factory_test.go`](../../internal/cloud/factory_test.go) |
| Provider plan | OpenStack operation | [`internal/cluster/provider/openstack/service.go`](../../internal/cluster/provider/openstack/service.go) → `Plan` | Typed candidate config + read-only `DiscoverySnapshot`; no remote mutation |
| Storage plan/apply | OpenStack operation | [`internal/cluster/storage/openstack/service.go`](../../internal/cluster/storage/openstack/service.go) → `Plan`, `Apply`; [`internal/cloud/openstack/storage.go`](../../internal/cloud/openstack/storage.go) → `StorageAdapter` | One service, remote container/credential actions, typed persistence/recovery |

## Capability matrix

| Provider | v2 config/readiness | Lifecycle path | Drift registry | OpenStack provider/storage ops |
|---|---:|---|---:|---:|
| OpenStack | yes | shared infrastructure provider | implementation-dependent | yes |
| VMware/vSphere | yes | shared infrastructure provider | provider implementation | no |
| Baremetal | yes | shared infrastructure provider with static-node checks | no registered cloud implementation | no |
| Kind | yes | Kind provider | no | no |
| Magnum | yes | dedicated Magnum API provider | no | no |
| AWS/GCP/Azure | schema/validator acceptance exists | CLI availability gate rejects | no registered implementation | no |

The final row is intentionally not user support: config acceptance or latent bootstrap code does not bypass [`checkProviderAvailability`](../../cmd/provider_availability.go). Treat unreachable branches as scaffolding until the command gate, provider implementation, and tests all make them reachable.

## Actual routing paths

```text
cluster init/generate/deploy
  -> checkProviderAvailability
  -> lifecycle service
       -> provider.BuildSteps
            -> sanitized external commands or cloud client
cluster drift
  -> CloudProviderFactory.GetProvider
  -> CloudProvider.GetCurrentState / DetectDrift / ReconcileDrift
cluster provider openstack plan/apply
  -> profile -> ProfileDiscovery.DiscoverWithOptions (reads only)
  -> provider/openstack.Plan -> validate -> optional local persistence
cluster service storage plan/apply
  -> storage/openstack.ValidateOptions
  -> StorageAdapter.Preflight -> Plan
  -> Apply: ensure container -> credential -> validate -> backup/atomic persist -> revoke old credential
```

## Safe-change boundaries

- Do not equate v2 provider strings with reachable deployment capability; update availability gates, routing, implementation, and tests together.
- Keep `CloudProvider` drift APIs separate from `lifecycleBootstrapProvider` and from OpenStack storage adapters.
- Bootstrap/destroy commands must use `security.CommandRunner`; provider clients must honor context and avoid leaking credentials.
- Provider plan is read-only against OpenStack. Storage apply is the only map path here that intentionally performs container/credential mutations and recovery journaling.
- Magnum owns its managed-cluster API lifecycle and does not acquire an OpenTofu step.

## Related maps

[Cluster lifecycle](cluster-lifecycle.md) · [OpenStack operations](openstack-provider-storage-operations.md) · [Config system](config-system.md) · [Import and resilience](import-operations-and-resilience.md)
