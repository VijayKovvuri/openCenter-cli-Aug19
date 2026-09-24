---
last_updated: 2026-09-24
id: service-mimir
title: "Grafana Mimir"
sidebar_label: Mimir
description: Long-term metrics storage, local RustFS integration, and deployment limits.
doc_type: reference
audience: "platform engineers, operators"
tags: [monitoring, metrics, mimir, rustfs, services]
---

> **Purpose:** For platform engineers and operators, documents the Mimir service's configuration surface.

## Overview

Mimir provides long-term metrics storage. For local development, its current
deployment uses the S3-compatible RustFS storage profile. Mimir does not expose
a service-level `storage_type`, `s3_endpoint`, or filesystem opt-out in the
current deployment; do not add `storage_type: filesystem` or `storage_type: none`
to the Mimir service.

Select RustFS through the infrastructure storage profile (and enable Longhorn,
as required by that profile):

```yaml
opencenter:
  infrastructure:
    storage:
      profile:
        lifecycle: non-production
        pvc_provider: longhorn
        object_storage_provider: rustfs
  services:
    longhorn:
      enabled: true
    mimir:
      enabled: true
```

## Configuration

```yaml
opencenter:
  services:
    mimir:
      enabled: false               # default: false
      namespace: observability      # default: observability
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Whether Mimir is deployed |
| `namespace` | string | `observability` | Namespace for Mimir resources |

## Secrets

The legacy `secrets.mimir.swift_application_credential_secret` field remains
in the schema for compatibility, but the local RustFS profile supplies the
S3-compatible storage integration and does not require that external secret.

## Dependencies

None enforced by `opencenter cluster service enable|disable`.

## Rendering

`mimir` has no dedicated YAML descriptor; it is rendered through the built-in render catalog with extra rendering-order dependencies on the observability namespace/sources and an override dependency on `sources`. Enabling `mimir` also causes the generated `services/sources/kustomization.yaml.tpl` to include the shared `opencenter-observability` source.

## CLI commands

```bash
opencenter cluster service enable mimir
opencenter cluster service disable mimir
opencenter cluster service status
opencenter cluster service options mimir
```
