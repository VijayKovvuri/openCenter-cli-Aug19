---
last_updated: 2026-09-25
id: service-gateway
title: "Gateway"
sidebar_label: Gateway
description: Gateway service configuration, listener fields, and catalog rendering.
doc_type: reference
audience: "platform engineers, operators"
tags: [networking, gateway, services]
---

> **Evidence:** `internal/config/services/gateway.go`, `internal/config/v2/defaults.go`, `internal/services/plugins/registry.go`, and `internal/gitops/render_catalog.go`.

## Configuration

`GatewayConfig` embeds `BaseConfig`, adds Gateway identity fields, optional bring-your-own TLS settings, and listeners. The generated default is enabled in `gateway`.

```yaml
opencenter:
  services:
    gateway:
      enabled: true
      namespace: gateway
      gateway_name:
      gateway_namespace:
      gateway_class:
      default_issuer:
      tls:
        wildcard_secret_name:
        secret_namespace:
        per_listener_secrets: {}
      listeners:
        - name:
          port: 443
          protocol: HTTPS
          hostname:
          tls_secret_name:
      adoption_mode: managed
      address_pool:
```

| Field | Default | Evidence |
|-------|---------|----------|
| `enabled` | `true` | `NewDefaultServiceConfig` |
| `namespace` | `gateway` | `NewDefaultServiceConfig` |
| `gateway_name`, `gateway_namespace`, `gateway_class`, `default_issuer` | empty in config type | `GatewayConfig` |
| `tls.*` | absent | `GatewayTLSConfig` |
| `listeners[].name`, `port`, `protocol` | required by schema tags | `GatewayListener` |
| `listeners[].hostname`, `tls_secret_name` | empty | `GatewayListener` |

## Runtime TLS behavior

`GatewayConfig.IsBYO` is true when a wildcard or per-listener secret is configured; the generator then omits the cert-manager annotation. `TLSSecretFor` resolves a per-listener secret, then the wildcard, then its supplied built-in default. `secret_namespace` is returned separately and is empty when unset.

## Rendering and dependencies

The plugin registry records `gateway-api`; the catalog records `envoy-gateway-api-base`. The catalog marks Gateway as single-stage and names `namespace.yaml`, `gateway-class.yaml`, `gateway.yaml`, and `envoy-proxy-config.yaml` as generated resources. No service descriptor file names `gateway`.

## Commands

```bash
opencenter cluster service enable gateway
opencenter cluster service disable gateway
opencenter cluster service options gateway
```
