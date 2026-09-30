{{- /* Only render Calico GitOps manifests when install_method is "helm" (default) */}}
{{- if and .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Enabled (eq (.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.InstallMethod | default "helm") "helm") }}
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: calico
  namespace: tigera-operator
spec:
  interval: 30m
  chart:
    spec:
      chart: tigera-operator
      version: "{{ .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Version | default "3.29.2" }}"
      sourceRef:
        kind: HelmRepository
        name: projectcalico
        namespace: flux-system
  install:
    crds: CreateReplace
    remediation:
      retries: 3
  upgrade:
    crds: CreateReplace
    remediation:
      retries: 3
  valuesFrom:
    - kind: ConfigMap
      name: calico-values
      valuesKey: values.yaml
{{- end }}
