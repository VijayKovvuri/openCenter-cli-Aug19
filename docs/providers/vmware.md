---
last_updated: 2026-09-25
id: vmware-provider-guide
title: "VMware Provider Guide"
sidebar_label: VMware Provider Guide
description: Complete guide for deploying openCenter clusters on VMware vSphere with pre-provisioned VMs.
doc_type: how-to
audience: "platform engineers, operators"
tags: [vmware, vsphere, deployment, csi]
---
# VMware Provider Guide

**Purpose:** For platform engineers, operators, shows how to deploying openCenter clusters on VMware vSphere with pre-provisioned VMs.

Guide for deploying openCenter clusters on VMware vSphere infrastructure with pre-provisioned VMs.

## Table of Contents

* [Overview](#overview)
* [Prerequisites](#prerequisites)
* [Architecture](#architecture)
* [Configuration](#configuration)
  * [Basic Configuration](#basic-configuration)
  * [Node Configuration](#node-configuration)
  * [vSphere Integration](#vsphere-integration)
* [Deployment](#deployment)
* [Storage](#storage)
* [Networking](#networking)
* [Limitations](#limitations)
* [Troubleshooting](#troubleshooting)

## Overview

The VMware provider enables openCenter cluster deployment on VMware vSphere infrastructure. VMs must be pre-provisioned - the provider treats VMware as baremetal, using Kubespray/Ansible to configure existing VMs rather than provisioning new ones.

Key characteristics:

* Requires pre-provisioned VMs with Ubuntu 24.04
* Uses Kubespray deployment method (Ansible-based)
* Supports vSphere CSI driver for persistent storage
* No automatic VM lifecycle management

When `opencenter cluster generate` runs for the `vmware` provider, it selects
the embedded `main-vmware.tf.tpl` template. The generated infrastructure
configuration passes the pre-provisioned node lists to the Kubespray module and
the selected CNI module; it does not contain an infrastructure-provisioning
module. VM creation, deletion, and resizing therefore remain outside
openCenter.

Provider-specific boundaries:

| Capability | VMware behavior |
| --- | --- |
| VM provisioning | Manual; VMs are pre-provisioned |
| Node scaling | Manual; update the static node definitions and infrastructure outside openCenter |
| Persistent storage | vSphere CSI |
| Load balancer | MetalLB; Octavia is disabled |

## Prerequisites

### Infrastructure Requirements

* VMware vSphere 7.0 or later
* Pre-provisioned Ubuntu 24.04 VMs (minimum 3 control plane + 2 worker nodes)
* VMs must have network connectivity to each other
* SSH access to all VMs from the deployment host (a bastion is optional)
* vCenter credentials (for CSI driver integration)

### VM Specifications

Control plane nodes (minimum):

* 4 vCPUs
* 8 GB RAM
* 40 GB disk

Worker nodes (minimum):

* 4 vCPUs
* 16 GB RAM
* 40 GB disk

### Network Requirements

* Static IP addresses for all nodes
* DNS resolution for all node hostnames
* Deployment host with SSH access to all nodes (a bastion is optional)
* Firewall rules allowing Kubernetes traffic (6443, 2379-2380, 10250-10252)

## Architecture

```
┌─────────────────────────────────────────────────────┐
│ vCenter Server                                      │
│  - Manages VMs                                      │
│  - Provides CSI driver integration                  │
└─────────────────────────────────────────────────────┘
                      │
                      │ API
                      ▼
┌─────────────────────────────────────────────────────┐
│ VMware Datacenter                                   │
│  ┌───────────────────────────────────────────────┐ │
│  │ Compute Cluster                               │ │
│  │  ┌─────────────────────────────────────────┐ │ │
│  │  │ Pre-provisioned VMs                     │ │ │
│  │  │  - master-1 (192.168.1.10)             │ │ │
│  │  │  - master-2 (192.168.1.11)             │ │ │
│  │  │  - master-3 (192.168.1.12)             │ │ │
│  │  │  - worker-1 (192.168.1.20)             │ │ │
│  │  │  - worker-2 (192.168.1.21)             │ │ │
│  │  └─────────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────┐ │
│  │ Datastore (Persistent Volumes)                │ │
│  └───────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Configuration

### Basic Configuration

Create a VMware cluster configuration, then open it for provider-specific values:

```bash
opencenter cluster init my-vmware-cluster --type vmware --org myorg
opencenter cluster configure myorg/my-vmware-cluster
```

Minimal configuration:

```yaml
schema_version: 2.0
opencenter:
  meta:
    name: my-vmware-cluster
    organization: myorg
    env: production
    region: on-premises
  infrastructure:
    provider: vmware
    os_version: "24"
    bastion:
      enabled: false
    cloud:
      vmware:
        vcenter_server: vcenter.example.com
        datacenter: Datacenter1
        datastore: datastore1
        cluster: Cluster1
        network: VM Network
        template: ubuntu-24.04-template
    compute:
      master_nodes:
        - {name: master-1.example.com, access_ip_v4: 192.168.1.10}
        - {name: master-2.example.com, access_ip_v4: 192.168.1.11}
        - {name: master-3.example.com, access_ip_v4: 192.168.1.12}
      worker_nodes:
        - {name: worker-1.example.com, access_ip_v4: 192.168.1.20}
        - {name: worker-2.example.com, access_ip_v4: 192.168.1.21}
    networking:
      subnet_nodes: 192.168.1.0/24
      allocation_pool_start: 192.168.1.100
      allocation_pool_end: 192.168.1.200
      loadbalancer_provider: metallb
      dns_zone_name: example.com
      dns_nameservers: [192.168.1.1]
      ntp_servers: [pool.ntp.org]
  cluster:
    cluster_name: my-vmware-cluster
    kubernetes:
      version: 1.33.5
  gitops:
    repository:
      local_dir: ./gitops-repo
opentofu:
  enabled: false
secrets:
  vsphere_csi:
    vcenter_host: vcenter.example.com
    username: administrator@vsphere.local
    password: "CHANGEME"  # Replace and encrypt with SOPS
    datacenters: Datacenter1
    insecure_flag: "false"
    port: "443"
```

### Node Configuration

Each pre-provisioned node is represented in `infrastructure.compute.master_nodes` or `worker_nodes` and requires:

```yaml
opencenter:
  infrastructure:
    compute:
      master_nodes:
        - name: master-1.example.com    # FQDN or hostname
          access_ip_v4: 192.168.1.10    # Static IP address used for access
          id: master-0                  # Optional stable identifier
```

Node roles:

* `master`: Control plane node (runs etcd, API server, scheduler, controller)
* `worker`: Worker node (runs application workloads)

The generated VMware template also passes the configured SSH user and key path
to Kubespray. For a deployment host that needs an explicit key, set them under
`infrastructure.ssh`:

```yaml
opencenter:
  infrastructure:
    ssh:
      user: ubuntu
      key_path: /path/to/my-vmware-cluster-key
```

### vSphere Integration

vSphere CSI driver configuration:

```yaml
opencenter:
  services:
    vsphere-csi:
      enabled: true
      image_repository: registry.k8s.io/csi-vsphere
      image_tag: v3.3.0

secrets:
  vsphere_csi:
    vcenter_host: vcenter.example.com
    username: administrator@vsphere.local
    password: "your-vcenter-password"  # Encrypt with SOPS
    datacenters: Datacenter1
    insecure_flag: "false"
    port: "443"
```

### Generated networking and HA defaults

The VMware infrastructure template uses these defaults unless the v2
configuration overrides them:

* Node network: `172.26.0.0/24`
* VRRP control-plane address: `172.26.0.5`, with VRRP enabled
* Kubernetes API address: `infrastructure.k8s_api_ip`, or the VRRP address when unset
* Calico interface: `ens192`
* OpenStack Octavia: disabled (`use_octavia = false`)

To override the Calico interface on VMs with a different NIC, configure the
interface explicitly:

```yaml
opencenter:
  cluster:
    kubernetes:
      network_plugin:
        calico:
          cni_iface: ens224
```

## Deployment

### Step 1: Initialize Cluster Configuration

```bash
opencenter cluster init my-vmware-cluster \
  --type vmware \
  --org myorg

# Expected output:
# ✓ Created cluster configuration
# ✓ Generated SSH keys
# ✓ Generated Age encryption keys
```

### Step 2: Configure Nodes

Edit the configuration file to add your pre-provisioned VMs:

```bash
# Configuration stored at:
# ~/.config/opencenter/clusters/myorg/.my-vmware-cluster-config.yaml

# Edit the compute.master_nodes and compute.worker_nodes sections with your VM details
```

### Step 3: Validate Configuration

```bash
opencenter cluster validate my-vmware-cluster

# Expected output:
# ✓ Schema validation passed
# ✓ Provider configuration valid
# ✓ Node configuration valid
# ✓ Network configuration valid
```

### Step 4: Setup GitOps Repository

```bash
opencenter cluster generate myorg/my-vmware-cluster

# Expected output:
# ✓ Created GitOps repository structure
# ✓ Generated Kubernetes manifests
# ✓ Generated Ansible inventory
# ✓ Encrypted secrets with SOPS
```

Inspect the generated `infrastructure/clusters/<cluster-name>/main.tf` before
deployment. For VMware it should contain the static `master_nodes` and
`worker_nodes` locals, a `kubespray-cluster` module, and the selected CNI
module; it should not contain a VM or other infrastructure-provisioning module.

### Step 5: Bootstrap Cluster

```bash
opencenter cluster deploy my-vmware-cluster

# This will:
# 1. Configure SSH access to all nodes
# 2. Install Kubernetes via Kubespray
# 3. Deploy FluxCD
# 4. Apply GitOps manifests
```

## Storage

### vSphere CSI Driver

The vSphere CSI driver provides dynamic persistent volume provisioning:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: vsphere-csi-sc
```

Storage classes:

```yaml
# Default storage class (created by vsphere-csi service)
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: vsphere-csi-sc
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: csi.vsphere.vmware.com
parameters:
  datastoreurl: "ds:///vmfs/volumes/datastore1/"
```

### Datastore Configuration

Specify datastore URL in service configuration:

```yaml
opencenter:
  services:
    vsphere-csi:
      enabled: true
      storage_class:
        default:
          datastore_url: "ds:///vmfs/volumes/1375553-datastore1/"
```

## Networking

### Network Plugin

VMware provider supports Calico CNI (default):

```yaml
opencenter:
  cluster:
    kubernetes:
      network_plugin:
        calico:
          enabled: true
          cni_iface: ens192  # Adjust to match your VM network interface
          encapsulation_type: VXLAN
```

### Load Balancer

For LoadBalancer services, use MetalLB:

```yaml
opencenter:
  services:
    metallb:
      enabled: true
      namespace: metallb-system
      ip_address_pools:
        - name: default
          addresses:
            - 192.168.1.100-192.168.1.110
      l2_advertisements:
        - name: default-l2
          ip_address_pools:
            - default
```

### Ingress

Gateway API with Istio or ingress-nginx:

```yaml
opencenter:
  services:
    gateway-api:
      enabled: true
    gateway:
      enabled: true
      hostname: "*.example.com"
```

## Limitations

### No Automatic Provisioning

* VMs must be pre-provisioned manually
* No automatic scaling (MachineDeployments not supported)
* Node lifecycle managed outside openCenter

### Deployment Method

* Only Kubespray deployment method supported
* Talos deployment not supported (requires cloud-init integration)
* Kamaji supported but requires manual worker node provisioning

### Infrastructure Management

* No Terraform/OpenTofu integration (opentofu.enabled: false)
* VM configuration changes require manual intervention
* No automated backup/restore of VMs

## Troubleshooting

### SSH Connection Issues

```bash
# Test SSH connectivity to all nodes
for node in master-1 master-2 master-3 worker-1 worker-2; do
  ssh ubuntu@${node}.example.com "hostname"
done

# Verify SSH key is configured
cat ~/.config/opencenter/clusters/myorg/secrets/ssh/my-vmware-cluster
```

### Node Not Joining Cluster

Check Kubespray logs:

```bash
# View Ansible playbook output
tail -f /var/log/opencenter/bootstrap.log

# Check node status
kubectl get nodes

# Verify kubelet is running on node
ssh ubuntu@worker-1.example.com "systemctl status kubelet"
```

### vSphere CSI Driver Issues

```bash
# Check CSI driver pods
kubectl get pods -n kube-system | grep vsphere-csi

# View CSI driver logs
kubectl logs -n kube-system deploy/vsphere-csi-controller

# Verify vCenter credentials
kubectl get secret vsphere-config-secret -n kube-system -o yaml
```

### Storage Provisioning Failures

```bash
# Check PVC status
kubectl get pvc

# View events
kubectl describe pvc my-pvc

# Verify datastore URL
kubectl get storageclass vsphere-csi-sc -o yaml
```

### Network Connectivity

```bash
# Test pod-to-pod connectivity
kubectl run test-pod --image=busybox --rm -it -- ping <pod-ip>

# Check Calico status
kubectl get pods -n calico-system

# Verify network interface
ssh ubuntu@worker-1.example.com "ip addr show ens192"
```

## Related Documentation

* [Kubespray Deployment Method](../getting-started/getting-started.md)
* [vSphere CSI Driver Configuration](../reference/platform-services.md)
* [Infrastructure Providers](README.md) (provider support boundaries)
* [Storage Configuration](../reference/platform-services.md)
