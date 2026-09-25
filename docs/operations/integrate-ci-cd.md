---
last_updated: 2026-09-25
id: integrate-ci-cd
title: "Integrate CI/CD"
sidebar_label: Integrate CI/CD
description: Provider-neutral guidance for integrating openCenter into CI/CD pipelines.
doc_type: how-to
audience: "devops engineers, developers"
tags: [ci-cd, automation, deployment, testing]
---
# Integrate CI/CD

**Role:** Provider-neutral guidance for integrating openCenter into CI/CD pipelines
for validation, deployment, testing, and cleanup.

This page describes portable pipeline stages and security controls. It does not
document this repository's own automation. For the authoritative workflow matrix,
runner contracts, permissions, tool pins, and release outputs in this repository,
see [GitHub Actions Workflows](../reference/github-actions-workflows.md).

## Prerequisites

* Access to a CI/CD platform and its protected environments
* A pinned, verified openCenter CLI available to the runner
* Credentials for the selected infrastructure provider, supplied by the CI secret store
* A Git repository containing the cluster configuration and any GitOps remote
* Network access from deployment runners to the provider APIs and cluster endpoints

## Pipeline stages

### Validate changes

Run configuration validation for pull requests and other change-review events.
Validation should use the same configuration directory and CLI version that the
deployment path uses. Keep this stage free of cloud credentials when possible.

### Generate and deploy

After a trusted change is merged:

1. Resolve the organization and cluster identifier.
2. Load credentials into a protected, runner-local environment.
3. Run `opencenter cluster validate`.
4. Run `opencenter cluster generate --force` when generated GitOps content must be refreshed.
5. Run `opencenter cluster deploy`.
6. Run smoke or integration checks against the resulting cluster.

Use a protected environment and an explicit approval step before production
deployment. Keep provider-specific discovery, credential setup, and network
configuration in the CI platform's deployment job rather than in review jobs.

### Test ephemeral environments

For disposable test clusters, create a unique organization/cluster identifier
per run, apply the test workload, run checks, and destroy the cluster in an
always-run cleanup step. Serialize runs when provider quotas or shared GitOps
resources make parallel execution unsafe.

### Promote through environments

Promote the same validated configuration or generated artifact from development
to staging and then production. Require successful checks at each boundary and
keep production credentials unavailable to lower-trust jobs. If a pipeline
rebuilds or regenerates between environments, record the exact CLI version and
source revision used for each promotion.

## One portable lifecycle example

Use this shell step in a protected deployment job for a disposable cluster. The
CI platform supplies the binary, configuration directory, credentials, and a
unique `CLUSTER_ID`.

```sh
set -eu

: "${CLUSTER_ID:?set a unique organization/cluster identifier}"
trap 'opencenter cluster destroy "$CLUSTER_ID" --force --break-lock --remove-files || true' EXIT

opencenter cluster validate "$CLUSTER_ID"
opencenter cluster generate "$CLUSTER_ID" --force
opencenter cluster deploy "$CLUSTER_ID" --break-lock
# Run provider-specific smoke tests here.
```

For a persistent environment, replace the destroy trap with an explicit
promotion and rollback policy. If the job creates the cluster, cleanup must run
even when validation, deployment, or tests fail.

## Credentials, artifacts, and isolation

* Install a pinned CLI from the platform's trusted artifact or tool cache, verify
  its checksum where available, and record `opencenter version` in the job log.
* Pass provider credentials through the CI secret store and write them only to a
  runner-local file with restrictive permissions. Never print or upload
  credentials, kubeconfigs, SOPS keys, OpenTofu state, or generated secret material.
* Keep pull-request validation separate from credentialed deployment. Use
  protected branches/environments and short-lived provider credentials for
  mutating jobs.
* Use unique cluster identifiers, set an overall timeout, and make cleanup
  idempotent.
* Store generated GitOps output as a reviewable artifact or commit only when that
  is part of the chosen promotion model; do not treat CI workspace state as
  durable state.

## Verification

Verify the integration by reviewing the CI job result, checking the cluster health
from the deployment runner, and inspecting the test reports. For a disposable
cluster, confirm that the cleanup step removed the cluster and temporary files.

```sh
opencenter cluster status "$CLUSTER_ID" --refresh
opencenter cluster env "$CLUSTER_ID"
```

## Troubleshooting

### CLI is not available

Install the pinned binary or tool package in the job image, verify its checksum,
and run `opencenter version` before invoking a cluster command.

### Authentication fails

Confirm that the provider credential file is present only in the protected job,
has restrictive permissions, and matches the provider selected by the cluster
configuration. Avoid embedding secret values in command arguments or logs.

### Deployment times out

Set a job timeout that covers infrastructure provisioning and reconciliation.
Retain a cleanup job or an `always`/`finally` step with permission to remove
temporary resources.

### A cluster already exists

Use a run identifier in the organization/cluster identifier and clean up only
that run's resources. Do not destroy a shared cluster as a recovery shortcut.

## Best practices

1. Validate before every deployment.
2. Keep review jobs free of infrastructure credentials.
3. Promote through development and staging before production.
4. Require approval for production mutation.
5. Use short-lived, least-privilege credentials.
6. Pin and verify the CLI and supporting tools.
7. Use unique names and explicit concurrency controls.
8. Always clean up ephemeral resources.
9. Keep secrets and state out of logs and artifacts.
10. Document the pipeline's environment, rollback, and ownership model.

## Related topics

* [Validate Configuration](validate-configuration.md) — configuration validation
* [Multi-Cluster Management](../getting-started/multi-cluster-setup.md) — manage multiple clusters
* [Configuration Lifecycle](../concepts/configuration-lifecycle.md) — configuration management
* [CLI Commands](../reference/cli-commands.md) — complete CLI reference
* [GitHub Actions Workflows](../reference/github-actions-workflows.md) — repository-specific workflow details
