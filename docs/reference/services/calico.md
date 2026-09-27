---
last_updated: 2026-09-25
id: service-calico
title: "Calico"
sidebar_label: Calico
description: Calico service configuration, generated default, and descriptor ownership.
doc_type: reference
audience: "platform engineers, operators"
tags: [networking, cni, calico, services]
---

> **Evidence:** `internal/config/services/calico.go`, `internal/config/v2/defaults.go`, `internal/services/plugins/calico.go`, and `internal/services/descriptors/data/service-calico.yaml`.

## Configuration

`CalicoConfig` embeds `BaseConfig` and adds `kube_api_server`. The generated default is enabled in `calico-system`.

```yaml
opencenter:
  services:
    calico:
      enabled: true
      namespace: calico-system
      kube_api_server:
      adoption_mode: managed
      address_pool:
```

| Field | Default | Evidence |
|-------|---------|----------|
| `enabled` | `true` | `NewDefaultServiceConfig` |
| `namespace` | `calico-system` | `NewDefaultServiceConfig` |
| `kube_api_server` | empty | `CalicoConfig.KubeAPIServer` |
| common fields | — | `BaseConfig`: `adoption_mode`, `source`, `image`, and `address_pool` |

## Runtime behavior

`CalicoPlugin.Validate` only checks that the value is `*CalicoConfig`. Its plugin `Render` method is a no-op; descriptor rendering is separate.

## Descriptor rendering

`service-calico.yaml` declares `service: calico` and one root, `services/calico`. It declares no conditional files or aggregate targets.

## Commands

```bash
opencenter cluster service enable calico
opencenter cluster service disable calico
opencenter cluster service options calico
```
