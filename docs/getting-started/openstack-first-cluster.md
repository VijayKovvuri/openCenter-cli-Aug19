---
id: openstack-first-cluster
title: Configure an OpenStack cluster
sidebar_label: Configure an OpenStack cluster
description: Configure an OpenStack cluster file and review its provider selections with openCenter.
doc_type: tutorial
audience: openCenter users
tags: [openstack, configuration]
last_updated: 2026-09-25
---
# Configure an OpenStack cluster

This tutorial shows the repository's OpenStack configuration workflow. It does
not promise a cloud outcome: provider discovery needs a usable `clouds.yaml`
profile, while deployment needs credentials and a target cloud.

## 1. Initialize and inspect

```bash
opencenter cluster init demo --org my-org --type openstack
opencenter cluster describe my-org/demo
```

Edit the generated v2 file, or use the guided command:

Use the guided configuration workflow where interactive input is appropriate.
The command is exposed by the CLI; this page keeps the command matrix focused
on the file-based path below.

The configuration must contain the OpenStack fields required by readiness
validation, including `auth_url`, `region`, `project_id`, and `image_id`.
Application credential ID and secret are validated as a pair.

## 2. Plan provider selections

The provider plan performs discovery and reports a typed plan without writing the
cluster file or mutating OpenStack:

```bash
opencenter cluster provider openstack plan my-org/demo \
  --os-cloud my-profile
```

Selectors available in the command include `--image-id`, `--network-id`,
`--external-network-id`, `--subnet-id`, and `--availability-zone`. If the
generated infrastructure should own the internal network and subnet, use:

```bash
opencenter cluster provider openstack plan my-org/demo \
  --os-cloud my-profile --create-internal-network
```

That mode cannot be combined with `--network-id` or `--subnet-id`.

## 3. Apply the reviewed selections

Use the same selectors with apply:

```bash
opencenter cluster provider openstack apply my-org/demo \
  --os-cloud my-profile --image-id IMAGE_ID --network-id NETWORK_ID \
  --external-network-id EXTERNAL_NETWORK_ID --subnet-id SUBNET_ID
```

`--import-auth` and `--import-tls` import profile values when requested.
`--replace` permits replacing populated selections. `--yes` is required for
non-interactive confirmation; global `--dry-run` stops before the local write.

## 4. Validate and render

```bash
opencenter cluster validate my-org/demo
opencenter cluster generate my-org/demo
```

Use `--validation online` when provider or Git remote checks are wanted. Use
`--render-only` to render templates without the full repository setup flow.

## 5. Deploy

```bash
opencenter cluster deploy my-org/demo
```

Deployment runs the provider's configured steps and writes bootstrap logs and
resume state under the openCenter state directory. The command returns its log
path; use `--log` to select one explicitly. Use `--from-step` or `--restart`
after diagnosing a failed run.

## Evidence

- OpenStack command surface: `cmd/cluster_provider_openstack.go`
- Initialization and configuration: `cmd/cluster_init.go`,
  `cmd/cluster_configure.go`
- Readiness rules: `internal/config/v2/readiness.go`
- OpenStack lifecycle: `internal/cluster/bootstrap_provider_infra.go`
- Provider workflow examples: `tests/features/workflow.feature`
