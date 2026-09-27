---
last_updated: 2026-09-25
id: import-operations-and-resilience
title: "Explain Import Operations and Resilience"
sidebar_label: Import and Resilience
description: "Execution map for GitOps repository import, drift and backup operations, locks, retries, circuit breakers, and safe mutation boundaries."
doc_type: explanation
audience: "contributors, maintainers, operators"
tags: [import, operations, drift, backup, resilience]
---
# Import, operations, and resilience

These are adoption and day-2 capabilities. They consume the validated config/path contracts but do not own primary config loading, GitOps rendering, or lifecycle bootstrap.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| Import scan | importer / `internal/importer` | [`internal/importer/scanner.go`](../../internal/importer/scanner.go) → `NewScanner`, `Scanner.ScanRepo`, `scanCluster` | README/legacy/GitOps sources, detectors, v2 defaults, evidence/confidence/conflicts; [`cmd/cluster_import_test.go`](../../cmd/cluster_import_test.go) |
| Import artifact/report | importer / `internal/importer` | [`internal/importer/store.go`](../../internal/importer/store.go) → `ArtifactStore`; [`internal/importer/report.go`](../../internal/importer/report.go) → `RenderScanResult` | persisted scan result, text/JSON/YAML report |
| Import apply | importer write plan | [`internal/importer/write_plan.go`](../../internal/importer/write_plan.go) → `PrepareClusterWritePlan`, `SelectApprovedFields`, `ApplyClusterWritePlan` | PathResolver, public YAML patch, backup, explicit confirmation; [`cmd/cluster_import.go`](../../cmd/cluster_import.go) |
| Drift | operations/cloud | [`internal/operations/drift_detector.go`](../../internal/operations/drift_detector.go) → `DriftDetector`, `NewDriftDetector`; [`internal/cloud/factory.go`](../../internal/cloud/factory.go) → `CloudProviderFactory` | desired v2 config vs provider state; [`internal/operations/drift_detector_property_test.go`](../../internal/operations/drift_detector_property_test.go) |
| Backup | operations | [`internal/operations/backup_manager.go`](../../internal/operations/backup_manager.go) → `BackupManager`, `CreateBackup`, `RestoreBackup` | PathResolver/filesystem, tar/gzip, checksums, optional AES-256-GCM; [`internal/operations/backup_manager_test.go`](../../internal/operations/backup_manager_test.go) |
| Locking | resilience | [`internal/resilience/lock_manager.go`](../../internal/resilience/lock_manager.go) → `LockManager`, `NewLockManager`, `Acquire` | file or Redis backend, TTL/metadata/force-break; [`internal/resilience/lock_manager_test.go`](../../internal/resilience/lock_manager_test.go) |
| Retry | resilience | [`internal/resilience/retry.go`](../../internal/resilience/retry.go) → `RetryHandler`, `NewRetryHandler`, `DoWithResult` | bounded exponential backoff, jitter, context, retry policy; [`internal/resilience/retry_test.go`](../../internal/resilience/retry_test.go) |
| Circuit breaking | resilience | [`internal/resilience/circuit_breaker.go`](../../internal/resilience/circuit_breaker.go) → `CircuitBreaker`, `NewCircuitBreaker`, `Call` | closed/open/half-open thresholds and timeout; [`internal/resilience/circuit_breaker_test.go`](../../internal/resilience/circuit_breaker_test.go) |

## Actual import path

```text
cluster import scan --repo
  -> cmd scanner setup
  -> Scanner.ScanRepo
       -> discover clusters/readme/legacy configs
       -> NewV2Default + disable services
       -> detectors infer provider/topology/services
       -> attach evidence, confidence, conflicts, skipped fields
       -> ArtifactStore.Save
cluster import report --repo
  -> ArtifactStore.LoadLatest -> RenderScanResult
cluster import apply --repo
  -> LoadLatest -> PrepareClusterWritePlan per cluster
  -> SelectApprovedFields (protected/conflicted/low confidence skipped)
  -> show diff/confirm -> ApplyClusterWritePlan
       -> backup existing config -> write approved public YAML
```

Import does not silently overwrite config. The write plan either creates a new config after confirmation or patches only approved YAML paths; source evidence and skipped fields remain in the artifact/report.

## Actual day-2/resilience path

```text
drift
  -> load desired v2 config -> CloudProviderFactory.GetProvider
  -> provider.GetCurrentState -> DetectDrift -> optional Reconcile/Schedule
backup
  -> resolve paths -> collect config/keys/GitOps/tofu state
  -> archive -> checksum -> optional encrypt -> restore verifies before extract
mutating operation
  -> LockManager.Acquire(resource, ttl)
  -> RetryHandler.Do / CircuitBreaker.Call around transient dependency work
  -> release lock; persist operation-specific state where required
```

Locks coordinate ownership, bootstrap state coordinates step progress, retry handles transient attempts, and circuit breakers fail fast. None replaces another.

## Safe-change boundaries

- Keep import inference evidence-based and conservative; protected/ambiguous fields must remain reviewable.
- Apply patches through the write-plan/public YAML path and preserve backups, permissions, and path validation.
- Keep drift provider interfaces separate from deploy providers; do not make a report silently reconcile.
- Preserve backup integrity checks and encrypted key material handling.
- Every mutating path must define its lock scope, cancellation behavior, retry policy, and circuit-breaker policy rather than adding blanket retries.

## Related maps

[Cluster lifecycle](cluster-lifecycle.md) · [Providers](providers.md) · [Config system](config-system.md) · [Runtime extensions](runtime-extensions-and-local-development.md)
