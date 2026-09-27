---
last_updated: 2026-09-25
id: service-cert-manager
title: "cert-manager"
sidebar_label: cert-manager
description: cert-manager service configuration, conditional credentials, and descriptor ownership.
doc_type: reference
audience: "platform engineers, operators"
tags: [cert-manager, tls, services]
---

> **Evidence:** `internal/config/services/cert_manager.go`, `internal/config/v2/defaults.go`, `internal/services/plugins/validators.go`, `internal/config/services/provider_registry.go`, `internal/config/services/secrets_validator.go`, and `internal/services/descriptors/data/service-cert-manager.yaml`.

## Configuration

The generated default is enabled in `cert-manager`. `CertManagerConfig` embeds `BaseConfig`.

```yaml
opencenter:
  services:
    cert-manager:
      enabled: true
      namespace: cert-manager
      letsencrypt_server:
      email:
      region:
      dns_zones: []
      create_cluster_issuer: false
      dns_provider:
      issuers:
        - name:
          type: letsencrypt
          server:
      adoption_mode: managed
      address_pool:
```

| Field | Default | Evidence |
|-------|---------|----------|
| `enabled` | `true` | `NewDefaultServiceConfig` |
| `namespace` | `cert-manager` | `NewDefaultServiceConfig` |
| `letsencrypt_server`, `email`, `region`, `dns_zones`, `issuers`, `dns_provider` | empty | `CertManagerConfig` |
| `create_cluster_issuer` | false in Go zero value; schema tag documents true | `CertManagerConfig` |
| `issuers[].name`, `type` | required fields | `CertIssuer` |
| `issuers[].server` | empty | `CertIssuer` |

The registered validator checks that a configured `letsencrypt_server` starts with `https://` and a configured `email` contains `@`. The provider registry records DNS provider choices and provider compatibility; it does not change the config type's fields.

## Secrets and conditional facts

`OpenCenterSecrets.CertManager` contains named AWS and Cloudflare credential maps plus legacy flat fields. The service secret mapping records conditional credentials for Route53, Cloudflare, Cloud DNS, and Azure DNS; Designate is documented in that mapping as using infrastructure credentials.

## Descriptor rendering

`service-cert-manager.yaml` declares `service: cert-manager`, root `services/cert-manager`, templates for its source and Flux files, and aggregate targets `services-fluxcd-aggregate` and `services-sources-aggregate`. It declares no `when` condition.

## Commands

```bash
opencenter cluster service enable cert-manager
opencenter cluster service disable cert-manager
opencenter cluster service options cert-manager
```
