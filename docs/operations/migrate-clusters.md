---
id: migrate-clusters
title: Migrate the legacy layout
sidebar_label: Migrate the legacy layout
description: Preview and apply the filesystem layout migration for a legacy openCenter configuration.
doc_type: how-to
audience: openCenter operators
tags: [migration, configuration]
last_updated: 2026-09-25
---
# Migrate the legacy layout

The supported migration command moves a legacy mixed organization directory
into separate GitOps, state, and secrets locations. It is a filesystem layout
migration, not a provider-to-provider cluster migration.

## Preview

```bash
opencenter cluster migrate-layout --org ORG --dry-run
```

The command requires a legacy Git repository and legacy layout markers. It
prints planned `MOVE` operations and config path rewrites without changing
files.

## Apply

```bash
opencenter cluster migrate-layout --org ORG
```

Use `--force` to overwrite existing destinations. Review collisions before
using it.

## Move hand-authored overlay files

For one cluster, custom overlay migration is a dry run unless `--apply` is
given:

```bash
opencenter cluster migrate-layout --custom --org ORG --cluster CLUSTER
opencenter cluster migrate-layout --custom --org ORG --cluster CLUSTER --apply
```

The implementation maps eligible hand-authored service files into service
`custom/` directories and updates custom kustomizations. It reports tracked,
modified, unknown, refused, and symlinked files rather than claiming that all
files can be moved automatically.

## Evidence

- Command contract and flags: `cmd/cluster_migrate_layout.go`
- Migration tests: `cmd/cluster_migrate_layout_test.go`
- Secure path model: `internal/core/paths/`
