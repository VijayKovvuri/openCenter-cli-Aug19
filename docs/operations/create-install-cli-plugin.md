---
id: create-install-cli-plugin
title: Discover CLI plugins
sidebar_label: Discover CLI plugins
description: Discover and inspect external CLI plugins available to openCenter.
doc_type: how-to
audience: openCenter operators
tags: [plugins, cli]
last_updated: 2026-09-25
---
# Discover CLI plugins

The built-in CLI can list external plugins discovered on the host:

```bash
opencenter plugins list
opencenter plugins list --output json
opencenter plugins list --output yaml
```

Discovery order is `OPENCENTER_PLUGINS_DIR`, `<config-dir>/plugins`, then
`PATH`. The list is read-only and reports each discovered plugin's name, path,
status, and message.

This checkout does not provide a built-in command that creates, installs, or
publishes a plugin. Build and installation conventions are therefore outside
this page's evidence boundary.

## Evidence

- Plugin command: `cmd/plugins.go`
- Discovery implementation: `internal/plugins/`
