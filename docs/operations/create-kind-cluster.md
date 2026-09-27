---
id: create-kind-cluster
title: Create a Kind cluster
sidebar_label: Create a Kind cluster
description: Create and deploy a Kind cluster through the openCenter command path.
doc_type: how-to
audience: openCenter operators
tags: [kind, clusters]
last_updated: 2026-09-25
---
# Create a Kind cluster

This is the short command path for a local Kind configuration. It intentionally
does not claim a fixed node count, service set, host IP, or bootstrap duration;
those values come from the generated configuration and local runtime.

```bash
opencenter cluster init local-demo --org local --type kind
opencenter cluster validate local/local-demo
opencenter cluster generate local/local-demo
opencenter cluster deploy local/local-demo --container-runtime docker
```

Use `--container-runtime podman` when Podman is the selected runtime. The flag
is accepted only for the Kind deployment path.

## Inspect and resume

```bash
opencenter cluster describe local/local-demo
opencenter cluster deploy local/local-demo --dry-run
opencenter cluster deploy local/local-demo --restart
opencenter cluster deploy local/local-demo --from-step STEP_ID
```

After deployment has produced a kubeconfig, use its reported path for live
inspection with tools outside this CLI page.

## Cleanup

```bash
opencenter cluster destroy local/local-demo --force
```

`--remove-files` additionally removes local configuration and GitOps files;
`--skip-infrastructure --remove-files` skips infrastructure destruction.

## Evidence

- Kind flags: `cmd/cluster_init.go`, `cmd/cluster_deploy.go`
- Destroy flags: `cmd/cluster_destroy.go`
- Kind lifecycle: `internal/cluster/kind_bootstrap_provider.go`
- Workflow limitation: `tests/features/workflow.feature:1-8`
