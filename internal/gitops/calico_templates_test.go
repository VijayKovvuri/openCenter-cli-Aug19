package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalicoAutodetectionModesUseOneNodeAddressKey(t *testing.T) {
	cases := []struct {
		name      string
		mode      string
		iface     string
		cidr      string
		wantKey   string
		wantValue any
	}{
		{name: "default", wantKey: "firstFound", wantValue: true},
		{name: "first-found", mode: "first-found", wantKey: "firstFound", wantValue: true},
		{name: "interface", mode: "interface", iface: "ens192", wantKey: "interface", wantValue: "ens192"},
		{name: "cidr", mode: "cidr", cidr: "10.0.0.0/8", wantKey: "cidrs", wantValue: []any{"10.0.0.0/8"}},
		{name: "normalized interface", mode: " INTERFACE ", iface: " ENS192 ", wantKey: "interface", wantValue: "ens192"},
		{name: "normalized cidr", mode: " CIDR ", cidr: " 10.0.0.0/8 ", wantKey: "cidrs", wantValue: []any{"10.0.0.0/8"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dst := t.TempDir()
			cfg := newDefault("calico-autodetect-" + strings.ReplaceAll(tc.name, "-", ""))
			cfg.OpenCenter.GitOps.Repository.LocalDir = dst
			calico := cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico
			calico.CalicoInterfaceAutodetect = tc.mode
			calico.CNIIface = tc.iface
			calico.AutodetectCIDR = tc.cidr

			require.NoError(t, RenderClusterApps(cfg))
			path := filepath.Join(dst, "applications", "overlays", cfg.ClusterName(), "services", "calico", "helm-values", "override_values.yaml")
			values := readYAMLMap(t, path)
			autodetection := mapAt(t, mapAt(t, mapAt(t, values, "installation"), "calicoNetwork"), "nodeAddressAutodetectionV4")

			require.Len(t, autodetection, 1)
			require.Equal(t, tc.wantValue, autodetection[tc.wantKey])
		})
	}
}

func TestCalicoTerraformModuleRequiresKubesprayInstallMethod(t *testing.T) {
	providers := []string{"openstack", "baremetal", "vmware"}
	modes := []struct {
		name          string
		enabled       bool
		installMethod string
		wantModule    bool
	}{
		{name: "enabled kubespray", enabled: true, installMethod: "kubespray", wantModule: true},
		{name: "enabled helm", enabled: true, installMethod: "helm", wantModule: false},
		{name: "disabled kubespray", enabled: false, installMethod: "kubespray", wantModule: false},
	}

	for _, provider := range providers {
		for _, mode := range modes {
			t.Run(provider+"/"+mode.name, func(t *testing.T) {
				dst := t.TempDir()
				cfg := mustNewGitOpsTestConfig("calico-module-gating", provider)
				cfg.OpenCenter.GitOps.Repository.LocalDir = dst
				calico := cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico
				calico.Enabled = mode.enabled
				calico.InstallMethod = mode.installMethod

				require.NoError(t, RenderInfrastructureCluster(cfg))
				content, err := os.ReadFile(filepath.Join(dst, "infrastructure", "clusters", cfg.ClusterName(), "main.tf"))
				require.NoError(t, err)
				hasModule := strings.Contains(string(content), `module "calico"`)
				require.Equal(t, mode.wantModule, hasModule)
			})
		}
	}
}
