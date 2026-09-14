package gitops

import (
	"strings"
	"testing"

	configservices "github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	"github.com/stretchr/testify/require"
)

func TestKubePrometheusStackTeamsReceiverIsOptional(t *testing.T) {
	cfg := newDefault("teams-optional")
	stack := cfg.OpenCenter.Services["kube-prometheus-stack"].(*configservices.PrometheusStackConfig)
	stack.WebhookURL = ""

	values := renderOverrideValues(t, cfg, "kube-prometheus-stack")
	require.Contains(t, values, "receiver: \"null\"", "the root route must safely discard unmatched alerts")
	require.NotContains(t, values, "warning_alerts_receiver")
	require.NotContains(t, values, "msteamsv2_configs")
	require.NotContains(t, values, "webhook_url:")
}

func TestKubePrometheusStackTeamsReceiverUsesConfiguredHTTPSURL(t *testing.T) {
	cfg := newDefault("teams-configured")
	stack := cfg.OpenCenter.Services["kube-prometheus-stack"].(*configservices.PrometheusStackConfig)
	stack.WebhookURL = "https://example.webhook.office.com/webhookb2/abc?sig=def"

	values := renderOverrideValues(t, cfg, "kube-prometheus-stack")
	require.Contains(t, values, "receiver: warning_alerts_receiver")
	require.Contains(t, values, "msteamsv2_configs")
	require.Contains(t, values, `webhook_url: "https://example.webhook.office.com/webhookb2/abc?sig=def"`)
	require.False(t, strings.Contains(values, "webhook_url: <no value>"))
}
