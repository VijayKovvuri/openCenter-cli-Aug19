package v2

import (
	"strings"
	"testing"
)

func TestKubesprayCloudInitTimeoutDefaultAndValidation(t *testing.T) {
	cfg, err := NewV2Default("timeout-default", "openstack")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}
	if got := cfg.Deployment.Kubespray.CloudInitTimeout; got != "10m" {
		t.Fatalf("CloudInitTimeout = %q, want 10m", got)
	}
	if got := cfg.Deployment.Kubespray.EffectiveCloudInitTimeout().String(); got != "10m0s" {
		t.Fatalf("EffectiveCloudInitTimeout() = %q, want 10m0s", got)
	}

	for _, value := range []string{"0s", "-1m", "not-a-duration"} {
		t.Run(value, func(t *testing.T) {
			cfg.Deployment.Kubespray.CloudInitTimeout = value
			err := NewValidator().ValidateDeployment(cfg)
			if err == nil || !strings.Contains(err.Error(), "cloud_init_timeout") {
				t.Fatalf("ValidateDeployment() error = %v, want cloud_init_timeout validation error", err)
			}
		})
	}
}
