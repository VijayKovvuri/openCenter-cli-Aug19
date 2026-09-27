---
last_updated: 2026-09-25
id: default-values
title: "Default Values"
sidebar_label: Default Values
description: What opencenter cluster init actually populates before you touch the config file, verified against internal/config/v2/defaults.go.
doc_type: reference
audience: "operators, developers"
tags: [defaults, configuration, v2, reference]
---
# Default Values

**Purpose:** For operators and developers, documents the default values `opencenter cluster init` (via `v2.NewV2Default`) actually writes into a new cluster configuration, verified against `internal/config/v2/defaults.go`. Use this page and the live `opencenter cluster init` output as the sources of truth for these values.

## Kubernetes / cluster defaults (non-Kind providers)

| Field | Default |
| --- | --- |
| `opencenter.cluster.kubernetes.version` | `1.35.4` |
| `opencenter.cluster.kubernetes.api_port` | `443` |
| `opencenter.cluster.kubernetes.subnet_pods` | `10.42.0.0/16` |
| `opencenter.cluster.kubernetes.subnet_services` | `10.43.0.0/16` |
| Network plugin | Calico, version `3.31.6`, `install_method: helm` |
| `opencenter.infrastructure.os_version` | `24` |
| `opencenter.infrastructure.compute.master_count` | `3` |
| `opencenter.infrastructure.compute.worker_count` | `3` |
| `opencenter.infrastructure.storage.default_storage_class` | Computed by `defaultStorageClass(provider, region)`: a provider/region lookup table's value if one exists; otherwise `csi-cinder-sc-delete` for OpenStack, `vsphere-csi` for VMware, or a generic default for anything else. |
| `deployment.method` | `kubespray` |
| `deployment.kubespray.version` | `2.31.0` |

Optional CSI storage-plugin defaults, when present, are seeded with `enabled: false` and a pinned version: AWS EBS CSI `1.37.0`, Azure Disk CSI `1.30.0`, Ceph CSI `3.11.0`, GCP Compute CSI `1.13.0`, NetApp Trident `24.06.0`. OpenStack Cinder CSI and VMware vSphere CSI default to `enabled: true` at versions `1.30.0` and `3.3.0` respectively when those providers are selected.

## Kind provider defaults

Kind uses its own, separate default constants (not the table above):

| Field | Default |
| --- | --- |
| Kubernetes version | `1.35.0` |
| API port | `6443` |
| Control plane count | `1` |
| Worker count | `2` |
| Pod subnet | `10.244.0.0/16` |
| Service subnet | `10.96.0.0/16` |
| Default CNI | Kind's built-in `kindnet`; managed Calico is disabled unless `--kind-disable-default-cni` is used. |

## Cluster-init defaults from `config.yaml` (`loadCLIDefaults`)

`v2.NewV2Default` reads `<config-dir>/config.yaml` (or `$OPENCENTER_CONFIG_DIR/config.yaml`) for `cluster_defaults`. This is separate from the CLI settings file, which is `<config-dir>/settings.yaml`. The values currently consumed for a new cluster are `organization`, `provider`, `region`, `environment`, `gitops_auth_method`, and `ssh_authorized_keys`.

The loader also parses `base_domain`, `admin_email`, `kubernetes_version`, `cni`, and `ssh_user`, but those fields are not currently applied by `NewV2Default`; setting them in `config.yaml` does not change the corresponding generated values. Verify against `internal/config/v2/defaults.go` when relying on this compatibility file.

## Default platform-service namespaces

Every built-in service gets a default namespace from `NewDefaultServiceConfig` in `internal/config/v2/defaults.go` (used unless the config explicitly overrides `namespace`) -- see the verified table in [Kind Cluster Verification](../contributing/kind-cluster-verification.md#verified-default-namespaces) rather than duplicating it here, since that page cross-checks it against real FluxCD dependency wiring too.

## How defaults combine with everything else

Defaults are used when `InitService` creates a new v2 configuration. Explicit initialization options and dotted overrides are then applied by the init path. The generic `DefaultConfigurationMerger` described in [Configuration Precedence](configuration-precedence.md) is a separate configuration-file merging facility, not a description of the complete `cluster init` pipeline.

## Cross-references

* [Configuration Schema Reference](configuration-schema.md) -- full field structure.
* [Configuration Precedence](configuration-precedence.md) -- generic configuration-file merging and CLI-tool path resolution.
* [Cluster Init Details](../contributing/cluster-init-details.md) -- the exact code path that builds these defaults.
