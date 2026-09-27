---
last_updated: 2026-09-25
id: service-metallb
title: "MetalLB"
sidebar_label: MetalLB
description: MetalLB service configuration and generated pool and L2 advertisement files.
doc_type: reference
audience: "platform engineers, operators"
tags: [networking, load-balancer, metallb, services]
---

> **Evidence:** `internal/config/services/metallb.go`, `internal/config/v2/defaults.go`, and `internal/gitops/render_catalog.go`.

## Configuration

The generated default is disabled in `metallb-system`. `MetalLBConfig` adds IP address pools and L2 advertisements to `BaseConfig`.

```yaml
opencenter:
  services:
    metallb:
      enabled: false
      namespace: metallb-system
      ip_address_pools:
        - name:
          addresses: []
          default: false
          auto_assign: true
          avoid_buggy_ips: false
      l2_advertisements:
        - name:
          type: l2
          ip_address_pools: []
          interfaces: []
```

| Field | Default | Evidence |
|-------|---------|----------|
| `enabled` | `false` | `NewDefaultServiceConfig` |
| `namespace` | `metallb-system` | `NewDefaultServiceConfig` |
| `ip_address_pools[].name`, `addresses` | required fields | `IPAddressPool` |
| `ip_address_pools[].default` | false | `IPAddressPool.Default` |
| `ip_address_pools[].auto_assign` | true when omitted | `GetAutoAssign` |
| `ip_address_pools[].avoid_buggy_ips` | false | `IPAddressPool` |
| `l2_advertisements[].name` | required | `L2Advertisement` |
| `l2_advertisements[].type` | `l2` when omitted | `GetType` |
| `l2_advertisements[].ip_address_pools`, `interfaces` | empty | `L2Advertisement` |

`DefaultPoolName` selects the explicitly marked pool, otherwise the first pool, otherwise an empty string.

## Rendering

The built-in catalog uses an overlay-files renderer for MetalLB. The renderer names `ipaddresspool.yaml` and `l2advertisement.yaml` as generated outputs. No explicit service descriptor is present.

## Dependencies

No dependency is recorded for MetalLB in the plugin registry or catalog.

## Commands

```bash
opencenter cluster service enable metallb
opencenter cluster service disable metallb
opencenter cluster service options metallb
```
