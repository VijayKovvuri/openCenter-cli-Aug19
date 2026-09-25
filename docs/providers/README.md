---
last_updated: 2026-09-25
id: providers
title: "Infrastructure Providers"
sidebar_label: Infrastructure Providers
description: Choose an openCenter infrastructure provider and find the provider-specific configuration and deployment guidance.
doc_type: reference
audience: "operators, platform engineers"
tags: [providers, openstack, magnum, vmware, baremetal, kind]
---
# Infrastructure Providers

Use this page to choose a provider before creating a cluster. The [Infrastructure Providers Reference](../reference/providers.md) has the detailed support matrix and drift-detection boundaries.

## Current cluster providers

| Provider | Use it for | Infrastructure model | Provider-specific guidance |
| --- | --- | --- | --- |
| OpenStack | Production private-cloud clusters | openCenter provisions infrastructure with OpenTofu and deploys Kubernetes with the supported bootstrap flow | [OpenStack first cluster](../getting-started/openstack-first-cluster.md) |
| Magnum | Managed Kubernetes on OpenStack | The existing Magnum cluster template owns image, network, and COE choices; openCenter creates, waits for, and deletes Magnum clusters | [Infrastructure Providers Reference](../reference/providers.md#magnum) |
| VMware | Existing vSphere estates | VMs are pre-provisioned; node definitions live under `infrastructure.compute.master_nodes` and `worker_nodes` | [VMware provider guide](vmware.md) |
| Baremetal | Physical hosts already managed by the operator | Hosts are pre-provisioned and described with static node definitions; openCenter does not provision cloud resources | [Infrastructure Providers Reference](../reference/providers.md) |
| Kind | Local development and CI | Kind runs a local cluster in Docker or Podman; it is not a production infrastructure target | [Kind local development](../getting-started/kind-local-development.md) |

The canonical provider names are `openstack`, `magnum`, `vmware`, `baremetal`, and `kind`. Existing `vsphere` values are accepted as a compatibility alias for `vmware`; use `vmware` in new configuration and documentation.

## Providers not available for cluster deployment

The configuration schema contains typed blocks for AWS, GCP, and Azure, but the CLI currently rejects those providers for cluster initialization, generation, and deployment as planned providers. AWS-backed service integrations do not make AWS a supported cluster provider.

## Configuration workflow

For a new cluster, either use the guided workflow or edit the generated v2 configuration:

```bash
opencenter cluster configure <cluster> --org <org> --type <provider>
# or
opencenter cluster init <cluster> --org <org> --type <provider>
opencenter cluster edit <org>/<cluster>
opencenter cluster validate <org>/<cluster>
opencenter cluster generate <org>/<cluster>
opencenter cluster deploy <org>/<cluster>
```

The exact required fields are provider-specific. OpenStack requires values such as `cloud.openstack.auth_url`, `region`, `project_id`, `image_id`, and application credentials. VMware requires `cloud.vmware.vcenter_server`, `datacenter`, `datastore`, `network`, and `template`, plus pre-provisioned nodes under `compute`. Magnum requires `cloud.magnum.auth_url`, `region`, `project_id`, application credentials, and `cluster_template`.

## Placeholders are not credentials

Initialization and full-template commands intentionally use placeholders so a configuration can be inspected before it is connected to real infrastructure. Replace every placeholder before generation or deployment. In particular, `CHANGEME`, values ending in `-placeholder`, and example values such as `your-project-id` or `vcenter.example.com` are not usable credentials or resource identifiers.

Run offline validation while editing and online validation when provider access is available:

```bash
opencenter cluster validate <org>/<cluster>
opencenter cluster validate <org>/<cluster> --validation online
```
