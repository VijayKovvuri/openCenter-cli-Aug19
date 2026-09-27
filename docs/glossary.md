---
id: glossary
title: "Glossary"
sidebar_label: Glossary
description: Defensible definitions of terms used in the openCenter CLI documentation.
doc_type: reference
audience: "users, operators, contributors"
tags: [glossary, terminology, definitions]
last_updated: 2026-09-25
---

# Glossary

**Purpose:** For readers of the repository documentation, defines terms as they
are used by the CLI, its configuration, and its generated GitOps assets.

## A

- **Active cluster** — The current target for commands that accept an
  omitted cluster name. The CLI exposes `cluster use` and `cluster active` and
  also reads the `OPENCENTER_CLUSTER` setting. See [Environment Variables](reference/environment-variables.md).
- **Age** — A public-key encryption tool used by the CLI through SOPS key and
  secret workflows.
- **Audit signing key** — The key used by the CLI's HMAC-protected audit
  logger. The default paths are documented in [Audit Signing Key](reference/audit-key.md).

## B

- **Baremetal** — An infrastructure provider path for pre-provisioned
  machines. It is one of the provider values accepted by `cluster init`.
- **Blueprint** — A cluster's declarative configuration, stored in the
  organization-based cluster layout. See [File Locations](reference/file-locations.md).
- **Bootstrap** — The deployment workflow that provisions or connects the
  configured infrastructure, deploys Kubernetes, and performs the applicable
  GitOps setup.

## C

- **CNI (Container Network Interface)** — The networking interface used by
  Kubernetes container network plugins. The configuration validates Calico,
  Cilium, and Kube-OVN; Kind can use its built-in `kindnet` when its default
  CNI is not disabled.
- **Cobra** — The Go command framework used to register the built-in
  `opencenter` command tree and parse flags.
- **Control plane** — Kubernetes components that manage cluster state, such as
  the API server, scheduler, controller manager, and etcd.

## D

- **Diátaxis** — A documentation framework with four types: tutorial, how-to,
  reference, and explanation. This repository records the type in page
  frontmatter while retaining lifecycle directories.
- **Drift detection** — Comparison of provider-side infrastructure with the
  desired cluster state. The CLI's drift registry is separate from its
  lifecycle provider registry; see [Drift Detection](concepts/drift-detection.md).

## F–G

- **FluxCD** — The GitOps tooling represented by generated Flux resources and
  repository configuration in openCenter output.
- **GitOps** — An operating model in which desired infrastructure or workload
  configuration is versioned in Git and reconciled into the environment. See
  [GitOps Workflow](concepts/gitops-workflow.md).

## K

- **Kind** — Kubernetes in Docker, used by the CLI's local cluster provider
  workflow.
- **Kubespray** — Ansible-based Kubernetes deployment tooling used by the
  generated infrastructure path where the selected deployment method invokes
  it.
- **Kustomize** — Kubernetes manifest customization tooling used for generated
  overlays.
- **Kustomization** — A Flux custom resource that applies a Kustomize source;
  openCenter generates these resources for applicable services.

## M–O

- **Magnum** — OpenStack's managed-Kubernetes service. The `magnum` provider
  uses the Magnum API rather than the OpenTofu infrastructure path.
- **Mise** — The tool version manager and task runner configured in `.mise.toml`.
- **OpenTofu** — The infrastructure-as-code tool used by supported generated
  infrastructure workflows. The CLI can resolve an OpenTofu binary and has a
  Terraform compatibility path where configured.
- **Overlay** — A generated customization layer applied to a base set of
  Kubernetes manifests.

## P–R

- **Provider** — A named infrastructure or local-cluster lifecycle path. The
  current accepted paths are OpenStack, VMware, Baremetal, Kind, and Magnum;
  AWS, GCP, and Azure are rejected as unavailable.
- **Readiness validation** — Offline cross-field checks performed after schema
  validation by `cluster validate`. It checks configuration relationships and
  does not itself contact a cloud provider or Git remote. See [Validation Rules](reference/validation-rules.md).

## S–V

- **SOPS (Secrets OPerationS)** — A tool for encrypting structured files with
  Age or GPG keys. openCenter uses it for configured secret and overlay files.
- **Service plugin manifest** — The service metadata used for identity,
  dependencies, and validation. Rendering topology is described separately by
  service descriptors; see [Renderer Contract](contributing/rendering-contract.md).
- **Worker node** — A Kubernetes node that runs application workloads. Counts
  and pools are configuration-dependent; see [Default Values](reference/default-values.md).
- **vSphere** — VMware's virtualization platform. The canonical provider name
  in configuration is `vmware`; `vsphere` remains a compatibility alias. See
  [Provider Reference](reference/providers.md).

## Common acronyms

| Acronym | Meaning |
| --- | --- |
| API | Application Programming Interface |
| CLI | Command-Line Interface |
| CRD | Custom Resource Definition |
| CSI | Container Storage Interface |
| DNS | Domain Name System |
| HA | High Availability |
| IaC | Infrastructure as Code |
| OIDC | OpenID Connect |
| RBAC | Role-Based Access Control |
| SSH | Secure Shell |
| TLS | Transport Layer Security |
| VM | Virtual Machine |
| YAML | YAML Ain't Markup Language |
