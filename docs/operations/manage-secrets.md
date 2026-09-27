---
id: manage-secrets
title: Manage secrets
sidebar_label: Manage secrets
description: Generate keys and manage encrypted secret files with openCenter.
doc_type: how-to
audience: openCenter operators
tags: [secrets, security]
last_updated: 2026-09-25
---
# Manage secrets

The CLI exposes SOPS/Age key and file operations. Secret values should not be
placed in documentation or committed unencrypted.

## Generate and validate a key

```bash
opencenter secrets keys generate
opencenter secrets keys generate --key-file ./keys/cluster.txt
opencenter secrets keys validate
opencenter secrets keys validate --key-file ./keys/cluster.txt
```

`generate` updates SOPS configuration by default; use
`--update-sops-config=false` to disable that update. `validate` also accepts
`--config-file` for the SOPS configuration path.

## Encrypt, decrypt, and inspect files

```bash
opencenter secrets encrypt --path PATH
opencenter secrets decrypt --path PATH
opencenter secrets status --path PATH
```

The path is passed to the SOPS file operation; the command does not establish a
particular repository layout. `--no-backup` is available for encrypt/decrypt.

For cluster-generated manifests, synchronize configured values first:

```bash
opencenter secrets sync ORG/CLUSTER
opencenter secrets validate ORG/CLUSTER
```

`sync` supports `--services`, `--dry-run`, `--force`, and `--all`; `validate
--fix` can invoke synchronization for detected drift.

## Rotate and back up keys

```bash
opencenter secrets keys rotate --cluster ORG/CLUSTER --type age
opencenter secrets keys rotate --cluster ORG/CLUSTER --type ssh
opencenter secrets keys backup --backup-dir ./key-backups
opencenter secrets keys reconcile --cluster ORG/CLUSTER
```

Rotation accepts `--dry-run`; Age rotation also accepts `--complete`. Reconcile
is a preview unless `--apply` is supplied.

## Evidence

- Key commands: `cmd/secrets_keys.go`
- File operations: `cmd/secrets_sops.go`
- Cluster synchronization: `cmd/secrets_sync.go`, `cmd/secrets_validate.go`
- Configuration and key paths: `internal/sops/`, `internal/config/v2/`
