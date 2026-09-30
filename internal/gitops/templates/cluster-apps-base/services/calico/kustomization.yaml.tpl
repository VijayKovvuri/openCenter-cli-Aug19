{{- /* Only render Calico GitOps manifests when install_method is "helm" (default) */}}
{{- if and .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Enabled (eq (.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.InstallMethod | default "helm") "helm") }}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: tigera-operator
configMapGenerator:
  - name: calico-values
    files:
      - values.yaml=helm-values/override_values.yaml
    options:
      disableNameSuffixHash: true
resources:
  - namespace.yaml
  - helmrepository.yaml
  - helmrelease.yaml
{{- end }}
