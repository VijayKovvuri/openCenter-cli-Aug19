---
last_updated: 2026-09-25
id: secrets-management-map
title: "Explain Secrets Management Boundaries"
sidebar_label: Secrets Management
description: "Execution map for logical secret planning, encrypted Kubernetes manifest synchronization, SOPS overlay encryption, and key lifecycle boundaries."
doc_type: explanation
audience: "contributors, maintainers, operators"
tags: [secrets, sops, age, encryption, ownership]
---
# Secrets management

There are three related but distinct flows: neutral secret-artifact planning, encrypted Kubernetes-manifest synchronization, and SOPS/Age file encryption. Backend CRUD (Barbican/file/SOPS) is a fourth command boundary.

## Feature → subsystem → symbol

| Feature | Subsystem / package | File → symbol | Dependencies and evidence |
|---|---|---|---|
| Command surface | Cobra / `cmd` | [`cmd/secrets.go`](../../cmd/secrets.go) → `NewSecretsCmd`; [`cmd/secrets_sync.go`](../../cmd/secrets_sync.go) → `newSecretsSyncCmd`, `runClusterSyncSecrets` | Backend CRUD, sync, validate, encrypt/decrypt/status, and `keys` are separate registrations; [`cmd/secrets_router_test.go`](../../cmd/secrets_router_test.go) |
| Logical artifact planning | neutral planner / `internal/secretartifacts` | [`internal/secretartifacts/planner.go`](../../internal/secretartifacts/planner.go) → `Plan`, `ValidateTargets` | v2 config, typed secret blocks, `service_secrets`; [`internal/secretartifacts/planner_test.go`](../../internal/secretartifacts/planner_test.go) |
| Manifest synchronization | secret manager / `internal/secrets` | [`internal/secrets/manager.go`](../../internal/secrets/manager.go) → `DefaultSecretsManager.SyncSecrets`, `syncServiceManifestOutcome` | artifact planner, SOPS manager, Age key, overlay lock, ownership state; [`internal/secrets/manager_sync_test.go`](../../internal/secrets/manager_sync_test.go) |
| Transaction rollback | secret manager / `internal/secrets` | [`internal/secrets/manager.go`](../../internal/secrets/manager.go) → mutation journal/reconcile methods; [`internal/secrets/rollback.go`](../../internal/secrets/rollback.go) → `RollbackManager` | snapshots, journal, hash state, rollback window; [`internal/secrets/rollback_test.go`](../../internal/secrets/rollback_test.go) |
| Key lifecycle | key subsystem / `internal/secrets` | [`internal/secrets/registry.go`](../../internal/secrets/registry.go) → `NewDefaultKeyRegistry`; [`internal/secrets/rotation.go`](../../internal/secrets/rotation.go), `revocation.go`, `reconcile.go` | SOPS/Age key manager and registry state; [`internal/secrets/rotation_test.go`](../../internal/secrets/rotation_test.go) |
| SOPS overlays | encryption / `internal/sops` | [`internal/sops/overlay_files.go`](../../internal/sops/overlay_files.go) → `overlayFilesToEncrypt`, `serviceOverrideValuesFilesToEncrypt`; [`internal/sops/manager.go`](../../internal/sops/manager.go) → `EncryptOverlayFiles`, `EncryptServiceOverrideValues` | provider-specific files and credential-bearing overrides; [`cmd/secrets_sops_test.go`](../../cmd/secrets_sops_test.go) |
| Security/audit | cross-cutting / `internal/security` | [`internal/security/credential_masker.go`](../../internal/security/credential_masker.go), [`internal/security/audit_logger.go`](../../internal/security/audit_logger.go) | Redaction, sanitized commands, HMAC audit events; security tests |

## Actual manifest synchronization path

```text
secrets sync
  -> initializeSecretsManager
  -> DefaultSecretsManager.SyncSecrets
       -> load v2 config
       -> secretartifacts.Plan
       -> resolve overlay + Age key
       -> acquire per-overlay lock (non-dry-run)
       -> load previous ownership
       -> refuse unsafe/unowned adoption
       -> encrypt/write each artifact
       -> journal and rollback on mutation failure
       -> reconcile stale artifacts
       -> persist ownership hashes/state and audit result
```

Artifacts are grouped by physical target path. Grafana targets `kube-prometheus-stack`; owners merge only when canonical keys and values do not conflict. The manager refuses an existing unowned or unsafe file rather than silently adopting it. Dry-run plans/reports without mutating files.

## Actual SOPS path

```text
SetupService or secrets encrypt
  -> SOPS manager
  -> overlayFilesToEncrypt / serviceOverrideValuesFilesToEncrypt
  -> Age/SOPS encrypt full selected files
  -> staged generation promotion or explicit file operation
```

The ordered compatibility set always includes Flux bootstrap/base-repo files, adds OpenStack or vSphere credential files by provider, and adds credential-bearing service override values. Missing generated files are skipped by callers. This file-level encryption is not `secrets sync` manifest generation.

OpenStack storage plan/apply creates/reuses remote credentials and writes typed config; it does not call `secretartifacts.Plan`, update secret ownership, or run SOPS. The credentials can be consumed by a later render/sync flow.

## Safe-change boundaries

- Keep logical planning independent of backend and renderer; preserve deterministic target paths, owner lists, canonical key conflict checks, and target enablement.
- Keep encrypted manifest sync transactional and ownership-aware; do not overwrite arbitrary existing files or bypass the overlay lock.
- Preserve SOPS file selection and full-file encryption for credential-bearing overrides; never place plaintext credentials in generated GitOps output.
- Keep backend CRUD, SOPS files, manifest sync, and key lifecycle as separate command and package boundaries.
- Any new service secret must be traced through typed config, artifact planning, SOPS selection if needed, renderer output, ownership state, and focused tests.

## Related maps

[Rendering ownership](rendering-ownership-and-secret-artifacts.md) · [GitOps engine](gitops-engine.md) · [OpenStack operations](openstack-provider-storage-operations.md) · [Config system](config-system.md) · [CLI commands](cli-commands.md)
