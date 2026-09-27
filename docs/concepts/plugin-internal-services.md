---
last_updated: 2026-09-24
id: plugin-internal-services
title: "Plugin Internal Services"
sidebar_label: Plugin Internal Services
description: How openCenter CLI internal service plugin contracts relate to the live service configuration and rendering path.
doc_type: explanation
audience: "developers, platform engineers"
tags: [plugins, services, cert-manager, gitops, extensions]
---
# Plugin Internal Services

**Purpose:** For developers and platform engineers, explains the
`internal/services` plugin subsystem, its relationship to the live service
system, and the source-backed path for adding a platform service.

## Two different service systems

The repository contains two independent abstractions that both use the word
“service.” They must not be treated as one runtime plugin system.

The live path used by `opencenter cluster service ...` and `opencenter cluster
generate` is:

1. Typed service configuration in `internal/config/services`.
2. Type lookup through `internal/config/registry`.
3. Enforced enable/disable dependencies from
   `internal/config/services/dependency_validator.go`.
4. Embedded descriptors in `internal/services/descriptors` and the built-in
   render catalog in `internal/gitops/render_catalog.go`.
5. Direct GitOps planning and rendering in `internal/gitops`, called by
   `SetupService`.

The canonical source-grounded explanation of this live model is
[Platform Services Architecture](../reference/platform-services.md). The
rendering and ownership contract is covered by
[GitOps Workflow](gitops-workflow.md).

## What `internal/services` provides

`internal/services` defines a generic `ServicePlugin` contract with metadata,
validation, rendering, and status methods. It also defines a
`ServiceRegistry` with service definitions, dependency ordering, lifecycle
hooks, and validation-engine integration. Concrete implementations and
`RegisterBuiltInServices` live under `internal/services/plugins`.

Evidence:

* `internal/services/plugin.go`
* `internal/services/base_plugin.go`
* `internal/services/registry.go`
* `internal/services/plugins/registry.go`

This subsystem is self-contained and covered by its own tests, but it is not
the registry used by the live command path. Nothing in `cmd/` or the live
`cluster generate` pipeline calls `RegisterBuiltInServices` or constructs its
`ServiceRegistry`. Its dependency graph, plugin validation, status reporting,
and `ServicePlugin.Render` methods are therefore not claims about generated
output or enforced CLI behavior.

## `ServiceStage` is a supporting, unwired API

`internal/gitops/stages.ServiceStage` consumes an `internal/services.ServiceRegistry`
and a template registry. It can, when constructed by a caller, resolve enabled
services, order dependencies, run lifecycle hooks, evaluate template
conditions, render service templates, write through an atomic writer, validate
outputs, and produce a dry-run plan.

However, `NewServiceStage` is only constructed by its own tests; it is not in
the stage list used by `opencenter cluster generate`. The live command path
uses `SetupService` and the descriptor/render-catalog planners instead. Do not
document `ServiceStage` as the production service renderer.

Evidence:

* `internal/gitops/stages/service_stage.go`
* `internal/gitops/stages/service_stage_test.go`
* `internal/cluster/setup_service.go`
* `internal/gitops/auto_descriptor.go`
* `internal/gitops/render_catalog.go`

## Cert-manager as an example

Cert-manager has a live typed configuration in
`internal/config/services/cert_manager.go`, defaults in
`internal/config/v2/defaults.go`, and live rendering behavior owned by its
descriptor and `internal/gitops/cert_manager_renderer.go`. Its
`internal/services/plugins/cert_manager.go` object is part of the unwired
plugin subsystem described above; its `Render` callback is not what produces
the production cert-manager overlay.

For the complete cert-manager files, dynamic planning, Flux dependencies, and
SOPS behavior, use the canonical rendering explanation in
[Service Templates](services-templates.md) and the live service-system
overview in [Platform Services Architecture](../reference/platform-services.md).
This page intentionally does not duplicate those rendering details.

## Adding a live platform service

Adding a service to the production path is not accomplished by registering an
`internal/services` plugin alone. Follow the live contracts:

1. Add the typed configuration under `internal/config/services/` and register
   it with `internal/config/registry`.
2. Add its desired-state defaults in `internal/config/v2/defaults.go` when it
   should appear in newly initialized configurations.
3. Add only the enforced cross-service dependency rules needed by the CLI to
   `internal/config/services/dependency_validator.go`.
4. Add an embedded descriptor under `internal/services/descriptors/data/`, or
   add a `RenderSpec` to `internal/gitops/render_catalog.go` for the service's
   live GitOps output.
5. Add any service-specific validation in the live command/config validation
   path, then add rendering and configuration tests.

Use `internal/services/plugins` only when a caller explicitly needs that
separate plugin API. It does not make a service live in `cluster generate`.

## Related reading

* [Platform Services Architecture](../reference/platform-services.md)
* [Service Templates](services-templates.md)
* [Plugin External CLI](plugin-external-cli.md)
* [Adding Services](../contributing/adding-services.md)
* [GitOps Workflow](gitops-workflow.md)
