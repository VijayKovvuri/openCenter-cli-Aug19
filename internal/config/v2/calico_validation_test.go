package v2

import (
	"strings"
	"testing"
)

func TestResolveCalicoInterfaceAutodetect(t *testing.T) {
	tests := []struct {
		name       string
		config     CalicoConfig
		want       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "default first found",
			config: CalicoConfig{
				CNIIface:       "stale-interface",
				AutodetectCIDR: "not-a-cidr",
			},
			want: CalicoInterfaceAutodetectFirstFound,
		},
		{
			name: "first found compatible spelling",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: "first_found",
			},
			want: CalicoInterfaceAutodetectFirstFound,
		},
		{
			name: "interface",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: CalicoInterfaceAutodetectInterface,
				CNIIface:                  "ens192",
				AutodetectCIDR:            "stale-cidr",
			},
			want: CalicoInterfaceAutodetectInterface,
		},
		{
			name: "cidr",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: CalicoInterfaceAutodetectCIDR,
				CNIIface:                  "stale-interface",
				AutodetectCIDR:            "192.168.10.0/24",
			},
			want: CalicoInterfaceAutodetectCIDR,
		},
		{
			name: "invalid mode",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: "route",
			},
			wantErr:    true,
			wantErrMsg: "calico_interface_autodetect must be one of",
		},
		{
			name: "interface requires cni iface",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: CalicoInterfaceAutodetectInterface,
			},
			wantErr:    true,
			wantErrMsg: "cni_iface is required",
		},
		{
			name: "cidr requires IPv4 CIDR",
			config: CalicoConfig{
				CalicoInterfaceAutodetect: CalicoInterfaceAutodetectCIDR,
				AutodetectCIDR:            "2001:db8::/64",
			},
			wantErr:    true,
			wantErrMsg: "autodetect_cidr must be a valid IPv4 CIDR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := tt.config
			got, err := ResolveCalicoInterfaceAutodetect(&tt.config)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Fatalf("ResolveCalicoInterfaceAutodetect() error = %v, want substring %q", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveCalicoInterfaceAutodetect() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ResolveCalicoInterfaceAutodetect() = %q, want %q", got, tt.want)
			}
			if tt.config != before {
				t.Fatalf("ResolveCalicoInterfaceAutodetect() modified the configuration: got %#v, want %#v", tt.config, before)
			}
		})
	}
}

func TestValidateReadinessCalicoInterfaceAutodetect(t *testing.T) {
	tests := []struct {
		name string
		set  func(*CalicoConfig)
		path string
	}{
		{
			name: "default",
			set:  func(config *CalicoConfig) {},
		},
		{
			name: "interface",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = CalicoInterfaceAutodetectInterface
				config.CNIIface = "ens192"
			},
		},
		{
			name: "cidr",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = CalicoInterfaceAutodetectCIDR
				config.AutodetectCIDR = "10.2.128.0/22"
			},
		},
		{
			name: "first found",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = CalicoInterfaceAutodetectFirstFound
				config.CNIIface = "stale-interface"
				config.AutodetectCIDR = "stale-cidr"
			},
		},
		{
			name: "invalid mode",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = "route"
			},
			path: "opencenter.cluster.kubernetes.network_plugin.calico.calico_interface_autodetect",
		},
		{
			name: "missing interface",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = CalicoInterfaceAutodetectInterface
			},
			path: "opencenter.cluster.kubernetes.network_plugin.calico.cni_iface",
		},
		{
			name: "missing cidr",
			set: func(config *CalicoConfig) {
				config.CalicoInterfaceAutodetect = CalicoInterfaceAutodetectCIDR
			},
			path: "opencenter.cluster.kubernetes.network_plugin.calico.autodetect_cidr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validReadinessConfig(t, "openstack")
			calico := config.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico
			tt.set(calico)
			report := ValidateReadiness(config)
			if tt.path == "" {
				assertNoIssue(t, report, "opencenter.cluster.kubernetes.network_plugin.calico.calico_interface_autodetect")
				assertNoIssue(t, report, "opencenter.cluster.kubernetes.network_plugin.calico.cni_iface")
				assertNoIssue(t, report, "opencenter.cluster.kubernetes.network_plugin.calico.autodetect_cidr")
				return
			}
			assertIssue(t, report, SeverityError, CategorySchema, tt.path)
		})
	}
}
