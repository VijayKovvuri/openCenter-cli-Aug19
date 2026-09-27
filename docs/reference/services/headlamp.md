---
last_updated: 2026-09-25
id: service-headlamp
title: "Headlamp Dashboard"
sidebar_label: Headlamp
description: Headlamp service configuration, OIDC dependency rule, and catalog ownership.
doc_type: reference
audience: "platform engineers, operators"
tags: [dashboard, oidc, headlamp, services]
---

> **Evidence:** `internal/config/services/headlamp.go`, `internal/config/v2/defaults.go`, `internal/config/v2/config.go`, `internal/config/services/dependency_validator.go`, and `internal/gitops/render_catalog.go`.

## Configuration

The generated default is enabled in `headlamp`, with hostname `dashboard.<cluster FQDN>`. `HeadlampConfig` adds `hostname`, `oidc_issuer_url`, and `oidc_client_id` to `BaseConfig`. `HeadlampSecrets` declares `oidc_client_secret`.

```yaml
opencenter:
  services:
    headlamp:
      enabled: true
      namespace: headlamp
      hostname:
      oidc_issuer_url:
      oidc_client_id:
      adoption_mode: managed
      address_pool:
```

## Runtime and rendering

The config dependency graph records `keycloak` for Headlamp. A separate `ValidateHeadlampOIDC` check also describes the conditional case: when Headlamp is enabled and either OIDC field is set, `keycloak` must be enabled. The repository therefore contains both an unconditional graph entry and a conditional OIDC-specific check. The catalog uses base path `applications/base/services/headlamp` and a dedicated override-values renderer. No explicit service descriptor is present.

## Commands

```bash
opencenter cluster service enable headlamp
opencenter cluster service disable headlamp
opencenter cluster service options headlamp
```
