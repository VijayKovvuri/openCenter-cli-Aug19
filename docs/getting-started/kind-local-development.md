---
id: kind-local-development
title: Local development with Kind
sidebar_label: Local development with Kind
description: Configure and render a Kind cluster for local openCenter development.
doc_type: tutorial
audience: openCenter developers
tags: [kind, local-development]
last_updated: 2026-09-25
---
# Local development with Kind

The Kind provider is the repository's local-provider path. This tutorial
covers only the openCenter commands that create and render its configuration;
the checkout does not prove a particular Docker/Podman installation or service
set.

## Create the configuration

```bash
opencenter cluster init dev --org local --type kind
```

If openCenter should manage the CNI instead of Kind's default CNI, use the
provider-specific flag:

```bash
opencenter cluster init dev --org local --type kind \
  --kind-disable-default-cni
```

Inspect the generated paths and effective values:

```bash
opencenter cluster describe local/dev
```

## Validate and generate

```bash
opencenter cluster validate local/dev
opencenter cluster generate local/dev
```

Generation can be previewed with the global `--dry-run`, or limited to
template rendering with `--render-only`.

## Deploy with the selected container runtime

```bash
opencenter cluster deploy local/dev --container-runtime docker
```

The deploy flag accepts `docker` or `podman`. The provider implementation uses
the selected runtime for Kind operations; the repository does not establish
that either runtime is installed on a user's machine.

For a preview:

```bash
opencenter --dry-run cluster deploy local/dev --container-runtime docker
```

If a step fails, rerun deploy. `--restart` ignores saved state and
`--from-step <id>` starts at a named step. The exact available step IDs are
reported by the deploy plan and implementation, rather than being hard-coded
in this tutorial.

## Inspect the resulting cluster

Use the kubeconfig path shown by `cluster describe` or deploy output for live
inspection. Those checks require a running local cluster and tools outside this
repository's workflow fixture.

## Evidence

- Kind flags: `cmd/cluster_init.go`, `cmd/cluster_deploy.go`
- Provider validation: `internal/config/v2/readiness.go`
- Kind bootstrap implementation: `internal/cluster/kind_bootstrap_provider.go`
- Workflow limitation: `tests/features/workflow.feature:1-8`
