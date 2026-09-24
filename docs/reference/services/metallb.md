---
id: service-metallb
title: "MetalLB"
sidebar_label: MetalLB
description: MetalLB IP address pool and L2 advertisement configuration for bare-metal LoadBalancer services.
doc_type: reference
audience: "platform engineers, operators"
tags: [networking, load-balancer, bare-metal, services]
---

> **Purpose:** For platform engineers, documents MetalLB configuration: IP address pools and L2 advertisements.

## Overview

MetalLB provides a `LoadBalancer`-type Service implementation for clusters that are not on a cloud provider with a built-in load balancer. It is disabled by default.

MetalLB requires two independent settings: select `metallb` as the load-balancer provider and enable the `metallb` service. Selecting the provider does not enable the service, and enabling the service does not select it as the provider.

## Configuration

```yaml
opencenter:
  infrastructure:
    networking:
      loadbalancer_provider: metallb
  services:
    metallb:
      enabled: true
      namespace: metallb-system     # default: metallb-system
      ip_address_pools:
        - name: public-pool
          addresses:
            - 192.168.1.240-192.168.1.254
          auto_assign: true
          avoid_buggy_ips: false
      l2_advertisements:
        - name: public-pool-l2
          type: l2
          ip_address_pools:
            - public-pool
          interfaces:
            - metal.105
            - mgmt.102
```

The interfaces list scopes announcements to those node interfaces; replace them with interfaces that exist on every node that should announce the pool. Both the provider selection and service enablement are required for MetalLB-backed `LoadBalancer` Services to work.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Whether MetalLB is deployed |
| `namespace` | string | `metallb-system` | Namespace for generated MetalLB resources |
| `ip_address_pools` | list | — | `IPAddressPool` definitions |
| `ip_address_pools[].name` | string | required | Pool identifier |
| `ip_address_pools[].addresses` | list of strings | required | IP ranges (CIDR or `start-end`) |
| `ip_address_pools[].auto_assign` | bool | `true` when omitted (`GetAutoAssign()`) | Automatically assign IPs from this pool |
| `ip_address_pools[].avoid_buggy_ips` | bool | `false` | Avoid `.0`/`.255` addresses |
| `l2_advertisements` | list | — | `L2Advertisement` definitions |
| `l2_advertisements[].name` | string | required | Advertisement identifier |
| `l2_advertisements[].type` | string | `l2` | Advertisement type; only `l2` is supported by configuration |
| `l2_advertisements[].ip_address_pools` | list of strings | all pools if empty | Pools this advertisement selects |
| `l2_advertisements[].interfaces` | list of strings | — | Node interfaces to advertise on |

If `l2_advertisements` is omitted entirely, no L2 advertisement is generated.

## Generated resources and custom resources

When the service is enabled, `opencenter cluster generate` renders the configured pools and L2 advertisements into the MetalLB service overlay:

* `ipaddresspool.yaml` contains the generated `IPAddressPool` resources when `ip_address_pools` is configured.
* `l2advertisement.yaml` contains the generated `L2Advertisement` resources when `l2_advertisements` includes supported L2 entries.

These generated files under `applications/overlays/<cluster>/services/metallb/` are CLI-owned and must not be hand-edited; change the cluster configuration and regenerate instead. L2 is the supported config-driven advertisement type. BGP resources, including `BGPPeer`, `BGPAdvertisement`, and `BFDProfile`, remain custom, user-owned manifests and belong under `applications/overlays/<cluster>/services/metallb/custom/`, which is preserved across regeneration.

## Dependencies

None enforced by `opencenter cluster service enable|disable`.

## Rendering

`metallb` has no dedicated YAML descriptor; it is rendered through the built-in render catalog.

## CLI commands

```bash
opencenter cluster service enable metallb
opencenter cluster service disable metallb
opencenter cluster service status
opencenter cluster service options metallb
```
