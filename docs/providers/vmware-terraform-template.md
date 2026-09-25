---
last_updated: 2026-09-25
id: vmware-terraform-template
title: "VMware Terraform Template"
sidebar_label: VMware Terraform Template
description: Reference for the VMware-specific Terraform template used to generate cluster infrastructure.
doc_type: reference
audience: "platform engineers, operators"
tags: [vmware, terraform, template, infrastructure]
---
# VMware Terraform Template

**Purpose:** For platform engineers, operators, documents the VMware-specific Terraform template used to generate cluster infrastructure.

Documentation for the VMware-specific Terraform template (`main-vmware.tf.tpl`).

## Table of Contents

* [Overview](#overview)
* [Template Selection](#template-selection)
* [Key Differences from Baremetal Template](#key-differences-from-baremetal-template)
* [Template Structure](#template-structure)
  * [Locals Block](#locals-block)
  * [Node Filtering Logic](#node-filtering-logic)
  * [Module Invocations](#module-invocations)
* [Configuration Mapping](#configuration-mapping)
  * [Input Configuration](#input-configuration)
  * [Generated Terraform](#generated-terraform)
* [Example Configurations](#example-configurations)
  * [Example Platform k8s-qa](#example-platform-k8s-qa)
  * [Example Sandbox k8s-sandbox](#example-sandbox-k8s-sandbox)
* [VMware-Specific Features](#vmware-specific-features)
  * [Network Interface Detection](#network-interface-detection)
  * [VRRP Configuration](#vrrp-configuration)
  * [SSH Key Path](#ssh-key-path)
* [Template Variables](#template-variables)
  * [Required Variables](#required-variables)
  * [Optional Variables](#optional-variables)
* [Validation](#validation)
* [Testing](#testing)
* [Troubleshooting](#troubleshooting)
  * [Wrong Template Selected](#wrong-template-selected)
  * [Nodes Not Appearing](#nodes-not-appearing)
  * [Template Rendering Errors](#template-rendering-errors)
* [Migration from Baremetal Template](#migration-from-baremetal-template)
* [Related Documentation](#related-documentation)

## Overview

The VMware template generates Terraform configuration for deploying Kubernetes on pre-provisioned VMware vSphere VMs. Unlike the OpenStack template, it does not provision infrastructure - VMs must already exist.

## Template Selection

The CLI automatically selects the correct template based on provider:

```go
switch provider {
case "vmware":
    mainTfTemplate = "main-vmware.tf.tpl"
case "baremetal":
    mainTfTemplate = "main-baremetal.tf.tpl"
default:
    mainTfTemplate = "main-default.tf.tpl"  // OpenStack, AWS
}
```

## Key Differences from Baremetal Template

| Feature | Baremetal Template | VMware Template |
| --- | --- | --- |
| Node Source | `Infrastructure.Compute.MasterNodes` | `Infrastructure.Compute.MasterNodes` and `WorkerNodes` |
| Node Filtering | Pre-filtered by role | Separate `master_nodes` and `worker_nodes` entries |
| Network Config | Generic | VMware-specific (ens192 default) |
| vCenter Info | Not included | Datacenter, datastore metadata |
| Documentation | Minimal | VMware-specific comments |

## Template Structure

### Locals Block

```hcl
locals {
  # Cluster identification
  cluster_name = "{{ .OpenCenter.Cluster.ClusterName }}"

  # Network configuration
  network_name    = "{{ .OpenCenter.Infrastructure.Cloud.VMware.Network }}"
  subnet_pods     = "{{ .OpenCenter.Cluster.Kubernetes.SubnetPods }}"
  subnet_services = "{{ .OpenCenter.Cluster.Kubernetes.SubnetServices }}"

  # VMware-specific settings
  address_bastion = "{{ .OpenCenter.Infrastructure.Bastion.Address }}"
  cni_iface       = "ens192"  # Override in network_plugin.calico when needed

  # Node definitions from the provider-agnostic compute configuration.
  # VMware uses pre-provisioned static nodes.
  master_nodes = .OpenCenter.Infrastructure.Compute.MasterNodes
  worker_nodes = .OpenCenter.Infrastructure.Compute.WorkerNodes
}
```

### Node Filtering Logic

The configuration separates pre-provisioned nodes by role:

```go
{{- range .OpenCenter.Infrastructure.Compute.MasterNodes }}
  {
    id           = "{{ .Name }}"
    name         = "{{ .Name }}"
    access_ip_v4 = "{{ .AccessIPv4 }}"
  },
{{- end }}
```

The node lists are provider-agnostic and are also used by baremetal deployments; VMware-specific metadata remains in `cloud.vmware`.

### Module Invocations

1. **kubespray-cluster**: Deploys Kubernetes using Ansible
2. **calico/cilium/kube-ovn**: Configures CNI plugin

No infrastructure provisioning modules (no `openstack-nova`).

## Configuration Mapping

### Input Configuration

```yaml
opencenter:
  infrastructure:
    provider: vmware
    cloud:
      vmware:
        vcenter_server: vcenter.example.com
        datacenter: Datacenter1
        datastore: datastore1
        network: VM Network
        template: ubuntu-24.04-template
    compute:
      master_nodes:
        - {name: k8s-qa-ord1-cp0, access_ip_v4: 172.26.0.11}
      worker_nodes:
        - {name: k8s-qa-ord1-wn0, access_ip_v4: 172.26.0.14}
```

### Generated Terraform

```hcl
locals {
  cluster_name = "k8s-qa"
  subnet_nodes = "172.26.0.0/24"

  master_nodes = [
    {
      id           = "k8s-qa-ord1-cp0"
      name         = "k8s-qa-ord1-cp0"
      access_ip_v4 = "172.26.0.11"
    }
  ]

  worker_nodes = [
    {
      id           = "k8s-qa-ord1-wn0"
      name         = "k8s-qa-ord1-wn0"
      access_ip_v4 = "172.26.0.14"
    }
  ]
}
```

## Example Configurations

### Example Platform k8s-qa

Configuration:

```yaml
opencenter:
  infrastructure:
    provider: vmware
    compute:
      master_nodes:
        - {name: k8s-qa-ord1-cp0, access_ip_v4: 172.26.0.11}
        - {name: k8s-qa-ord1-cp1, access_ip_v4: 172.26.0.12}
        - {name: k8s-qa-ord1-cp2, access_ip_v4: 172.26.0.13}
      worker_nodes:
        - {name: k8s-qa-ord1-wn0, access_ip_v4: 172.26.0.14}
        - {name: k8s-qa-ord1-wn1, access_ip_v4: 172.26.0.15}
        - {name: k8s-qa-ord1-wn2, access_ip_v4: 172.26.0.16}
```

Generated:

* 3 master nodes (172.26.0.11-13)
* 3 worker nodes (172.26.0.14-16)
* VRRP IP: 172.26.0.5
* Public API: 198.51.100.164

### Example Sandbox k8s-sandbox

Configuration:

```yaml
opencenter:
  infrastructure:
    provider: vmware
    compute:
      master_nodes:
        - {name: 3bk8s40, access_ip_v4: 192.168.12.20}
        - {name: 3bk8s41, access_ip_v4: 192.168.12.21}
        - {name: 3bk8s42, access_ip_v4: 192.168.12.22}
      worker_nodes:
        - {name: 3bk8s43, access_ip_v4: 192.168.12.23}
        - {name: 3bk8s44, access_ip_v4: 192.168.12.24}
        - {name: 3bk8s45, access_ip_v4: 192.168.12.25}
```

Generated:

* 3 master nodes (192.168.12.20-22)
* 3 worker nodes (192.168.12.23-25)
* The nodes are pre-provisioned; VMware deployment does not provision a bastion or VIP.

## VMware-Specific Features

### Network Interface Detection

Default interface for VMware VMs:

```hcl
cni_iface = "ens192"  # Standard VMware virtual NIC
```

Override in configuration:

```yaml
opencenter:
  cluster:
    kubernetes:
      network_plugin:
        calico:
          cni_iface: ens224  # Custom interface
```

### VRRP Configuration

VMware clusters use VRRP for HA API endpoint:

```hcl
vrrp_enabled = true
vrrp_ip      = "172.26.0.5"  # Internal VIP
k8s_api_ip   = "108.166.24.164"  # External/public IP
```

### SSH Key Path

VMware deployments use absolute SSH key paths:

```hcl
ssh_key_path = "/etc/openCenter/example-platform/secrets/ssh/k8s-qa-svc01m-ord1"
```

## Template Variables

### Required Variables

* `OpenCenter.Cluster.ClusterName`
* `OpenCenter.Infrastructure.Compute.MasterNodes[]` and `WorkerNodes[]`
  * `.Name` - Node hostname
  * `.AccessIPv4` - Node IP used for access
* `OpenCenter.Infrastructure.Bastion.Address`

### Optional Variables

* `OpenCenter.Infrastructure.Cloud.VMware.Network` - VMware network name
* `OpenCenter.Infrastructure.Cloud.VMware.Template` - Base VM template name
* `OpenCenter.Infrastructure.K8sAPIIP` - Public API IP (default: VRRP IP)
* `OpenCenter.Infrastructure.Networking.VRRPIP` - Internal VIP (default: .5 of subnet)
* `OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.CNIIface` - Network interface (default: ens192)

## Validation

The template validates:

* At least one master node exists
* At least one worker node exists
* All nodes have `name` and `access_ip_v4`

Validation happens in `internal/core/validation/validators/provider.go`.

## Testing

Test VMware configuration and generation:

```bash
# Initialize cluster
opencenter cluster init test-vmware --type vmware --org myorg

# Fill in the VMware provider settings
opencenter cluster configure myorg/test-vmware

# Setup (generates main.tf)
opencenter cluster generate myorg/test-vmware

# Verify generated main.tf
cat ~/.config/opencenter/clusters/myorg/test-vmware/gitops/infrastructure/clusters/test-vmware/main.tf
```

Expected main.tf structure:

* `locals` block with node definitions
* `module "kubespray-cluster"` invocation
* `module "calico"` invocation (if Calico enabled)
* No infrastructure provisioning modules

## Troubleshooting

### Wrong Template Selected

If baremetal template is used instead of VMware:

```bash
# Check provider in configuration
grep "provider:" ~/.config/opencenter/clusters/*/.*-config.yaml

# Should show: provider: vmware
```

### Nodes Not Appearing

If master_nodes or worker_nodes are empty:

```bash
# Check node configuration
yq '.opencenter.infrastructure.compute' config.yaml

# Check name and access_ip_v4 in each static node entry.
```

### Template Rendering Errors

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Re-run setup
opencenter cluster generate test-vmware

# Check for template errors in output
```

## Migration from Baremetal Template

To migrate existing baremetal clusters to VMware template:

1. Update provider:

   ```yaml
   infrastructure:
     provider: vmware  # was: baremetal
   ```
2. Move node definitions:

   ```yaml
   # Old (baremetal)
   infrastructure:
     compute:
       master_nodes: [...]
       worker_nodes: [...]

    # New (vmware) -- node lists remain under compute.
    infrastructure:
      compute:
        master_nodes:
          - {name: master-1, access_ip_v4: 172.26.0.11}
        worker_nodes:
          - {name: worker-1, access_ip_v4: 172.26.0.14}
   ```
3. Add VMware metadata (optional):

   ```yaml
   cloud:
     vmware:
       vcenter_server: vcenter.example.com
       datacenter: Datacenter1
       datastore: datastore1
   ```
4. Re-run setup:

   ```bash
   opencenter cluster generate <cluster-name>
   ```

## Related Documentation

* [VMware Provider Guide](./vmware.md)
* [VMware Quick Start](./vmware-quick-start.md)
* [Infrastructure Providers](README.md)
