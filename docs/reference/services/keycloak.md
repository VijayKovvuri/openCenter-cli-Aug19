---
last_updated: 2026-09-25
id: service-keycloak
title: "Keycloak"
sidebar_label: Keycloak
description: Keycloak service configuration, runtime validation, dependencies, and descriptor conditions.
doc_type: reference
audience: "platform engineers, operators"
tags: [keycloak, identity, services]
---

> **Evidence:** `internal/config/services/keycloak.go`, `internal/config/v2/defaults.go`, `internal/services/plugins/keycloak.go`, `internal/config/services/dependency_validator.go`, `internal/services/plugins/registry.go`, and `internal/services/descriptors/data/service-keycloak.yaml`.

## Configuration

The generated default is enabled in `keycloak`; its generated hostname is `auth.<cluster FQDN>`. `KeycloakConfig` embeds `BaseConfig` and contains access, realm, runtime, resource, scaling, database, observability, TLS, backup, and SMTP fields.

| Group | Fields |
|-------|--------|
| Access | `hostname`, `frontend_url` |
| Realm | `realm`, `client_id`, `realm_import_enabled`, `realm_groups`, `realm_admin_email` |
| Runtime | `start_optimized`, `cache_enabled`, `cache_stack` |
| Resources | `resource_requests_cpu`, `resource_requests_memory`, `resource_limits_cpu`, `resource_limits_memory` |
| Scaling | `instances`, `min_replicas`, `max_replicas` |
| Database | `database_host`, `database_port`, `database_name`, `database_user`, `db_pool_min_size`, `db_pool_initial_size`, `db_pool_max_size` |
| Observability | `metrics_enabled`, `event_metrics_enabled`, `health_enabled`, `log_level`, `log_format` |
| TLS/backup | `tls_secret_name`, `tls_enabled`, `backup_enabled`, `backup_schedule` |
| SMTP | `smtp_host`, `smtp_port`, `smtp_from`, `smtp_starttls` |

The config type's documented defaults include `client_id: opencenter`, `realm_import_enabled: true`, `start_optimized: false`, `cache_enabled: true`, `cache_stack: kubernetes`, resource values, scaling values, `database_port: 5432`, pool values, observability booleans, log values, TLS values, backup values, and `smtp_port: 587`/`smtp_starttls: true`. `defaults.go` additionally materializes `resource_requests_cpu: 500m`, `resource_limits_cpu: 2`, and `instances: 3` in the generated default. Other fields are empty unless configured.

## Runtime validation and dependencies

The service plugin validates HTTP(S) `frontend_url`, optimized startup with at least two instances, `min_replicas <= max_replicas`, and `db_pool_min_size <= db_pool_max_size`. The registered validation engine also validates log level and format and allows cache stacks `kubernetes` and `ispn`.

The config dependency graph records `olm` and `postgres-operator` for Keycloak. The plugin registry separately registers `cert-manager` as a plugin dependency. These are distinct repository registries; this page does not merge their enforcement paths.

`KeycloakSecrets` declares `client_secret` and `admin_password`, but `secrets_validator.go` explicitly has no Keycloak mapping. Treat those declarations as a secrets model fact, not as a documented required-secret rule.

## Descriptor rendering

`service-keycloak.yaml` owns `services/keycloak` and aggregates into the services Flux and sources aggregates. The backup CronJob template is included when `opencenter.services.keycloak.backup_enabled` is true; the HPA template is included when `max_replicas` exists. The descriptor also lists source/config and Flux templates.

## Commands

```bash
opencenter cluster service enable keycloak
opencenter cluster service disable keycloak
opencenter cluster service options keycloak
```
