---
last_updated: 2026-09-25
id: service-openstack-ccm
title: "OpenStack Cloud Controller Manager"
sidebar_label: OpenStack CCM
description: OpenStack CCM service configuration and catalog ownership.
doc_type: reference
audience: "platform engineers, operators"
tags: [openstack, services]
---

> **Evidence:** `internal/config/services/default_services.go`, `internal/config/v2/defaults.go`, and `internal/gitops/render_catalog.go`.

## Configuration

`openstack-ccm` uses `DefaultServiceConfig`. The generated default is enabled in `openstack-ccm`.

```yaml
opencenter:
  services:
    openstack-ccm:
      enabled: true
      namespace: openstack-ccm
      adoption_mode: managed
      address_pool:
```

| Field | Default | Evidence |
|-------|---------|----------|
| `enabled` | `true` | `NewDefaultServiceConfig` |
| `namespace` | `openstack-ccm` | `NewDefaultServiceConfig` |
| common fields | — | `BaseConfig` |

## Rendering

The catalog entry uses namespace stage `openstack-ccm`, override values, base path `applications/base/services/openstack-ccm`, and override dependencies `sources` and `openstack-ccm-namespace`. No explicit service descriptor is present.

## Commands

```bash
opencenter cluster service enable openstack-ccm
opencenter cluster service disable openstack-ccm
opencenter cluster service options openstack-ccm
```
