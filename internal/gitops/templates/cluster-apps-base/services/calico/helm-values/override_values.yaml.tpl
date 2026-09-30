{{- /* Only render Calico GitOps manifests when install_method is "helm" (default) */}}
{{- if and .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Enabled (eq (.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.InstallMethod | default "helm") "helm") }}
{{- $calico := .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico }}
{{- $autodetectMode := $calico.CalicoInterfaceAutodetect | default "first-found" | trim | lower }}
installation:
  enabled: true
  kubernetesProvider: ""
  calicoNetwork:
    bgp: Disabled
    ipPools:
      - cidr: "{{ .OpenCenter.Cluster.Kubernetes.SubnetPods | default "10.42.0.0/16" }}"
        encapsulation: "{{ if eq (.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.VXLANMode | default "Always") "Always" }}VXLAN{{ else if eq .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.VXLANMode "CrossSubnet" }}VXLANCrossSubnet{{ else if eq (.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.IPIPMode | default "") "Always" }}IPIP{{ else if eq .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.IPIPMode "CrossSubnet" }}IPIPCrossSubnet{{ else }}VXLAN{{ end }}"
        natOutgoing: Enabled
        nodeSelector: all()
    nodeAddressAutodetectionV4:
      {{- if eq $autodetectMode "interface" }}
      interface: "{{ $calico.CNIIface | default "" | trim | lower }}"
      {{- else if eq $autodetectMode "cidr" }}
      cidrs:
        - "{{ $calico.AutodetectCIDR | default "" | trim | lower }}"
      {{- else }}
      firstFound: true
      {{- end }}
    {{- if gt (.OpenCenter.Infrastructure.Compute.WorkerCountWindows | default 0) 0 }}
    windowsDataplane: HNS
    {{- else }}
    windowsDataplane: Disabled
    {{- end }}
  serviceCIDRs:
    - "{{ .OpenCenter.Cluster.Kubernetes.SubnetServices | default "10.43.0.0/16" }}"

{{- if .OpenCenter.Services.calico.KubeAPIServer }}
kubernetesServiceEndpoint:
  host: "{{ .OpenCenter.Services.calico.KubeAPIServer }}"
  port: "443"
{{- end }}
{{- end }}
