package gitops

import (
	"path/filepath"
	"testing"

	configservices "github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	"github.com/stretchr/testify/require"
)

func TestKubePrometheusStackOverlayRendersRoutesWithPinnedChartBackends(t *testing.T) {
	cfg := newDefault("monitoring-routes")
	stack := cfg.OpenCenter.Services["kube-prometheus-stack"].(*configservices.PrometheusStackConfig)
	stack.PrometheusHostname = "prometheus.example.test"
	stack.AlertmanagerHostname = "alerts.example.test"
	stack.GrafanaHostname = "grafana.example.test"

	files, err := kubePrometheusStackOverlayFilesRenderer(cfg)
	require.NoError(t, err)

	for filename, expected := range map[string][]string{
		"prometheus-http-route.yaml":   {"name: prometheus-gateway-route", `"prometheus.example.test"`, "sectionName: prometheus-https", "name: kube-prometheus-stack-prometheus", "port: 9090"},
		"alertmanager-http-route.yaml": {"name: alertmanager-gateway-route", `"alerts.example.test"`, "sectionName: alertmanager-https", "name: kube-prometheus-stack-alertmanager", "port: 9093"},
		"grafana-http-route.yaml":      {"name: grafana-gateway-route", `"grafana.example.test"`, "sectionName: grafana-https", "name: kube-prometheus-stack-grafana", "port: 80"},
	} {
		content, found := files[filename]
		require.Truef(t, found, "missing generated %s", filename)
		require.Contains(t, content, "namespace: observability")
		require.Contains(t, content, "namespace: rackspace-system")
		for _, want := range expected {
			require.Containsf(t, content, want, "%s must contain %q", filename, want)
		}
	}
}

func TestKubePrometheusStackRoutesAreIncludedAndWaitForGatewayAPI(t *testing.T) {
	cfg := newDefault("monitoring-routes")
	cfg.OpenCenter.GitOps.Repository.LocalDir = t.TempDir()
	require.NoError(t, RenderClusterApps(cfg))

	base := filepath.Join(cfg.OpenCenter.GitOps.Repository.LocalDir, "applications", "overlays", cfg.ClusterName(), "services", "kube-prometheus-stack")
	for _, filename := range []string{"prometheus-http-route.yaml", "alertmanager-http-route.yaml", "grafana-http-route.yaml"} {
		require.FileExists(t, filepath.Join(base, filename))
	}
	kustomization := mustReadFile(t, filepath.Join(base, "kustomization.yaml"))
	for _, filename := range []string{"prometheus-http-route.yaml", "alertmanager-http-route.yaml", "grafana-http-route.yaml"} {
		require.Contains(t, kustomization, "- "+filename)
	}

	flux := mustReadFile(t, filepath.Join(cfg.OpenCenter.GitOps.Repository.LocalDir, "applications", "overlays", cfg.ClusterName(), "services", "fluxcd", "kube-prometheus-stack.yaml"))
	docs, err := decodeYAMLDocuments([]byte(flux))
	require.NoError(t, err)
	override := findFluxKustomization(t, docs, "kube-prometheus-stack-override")
	require.True(t, hasFluxDependency(t, override, "envoy-gateway-api-base"))
}
