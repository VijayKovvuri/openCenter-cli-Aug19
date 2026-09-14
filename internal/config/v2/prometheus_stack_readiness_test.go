package v2

import (
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
)

func TestValidateReadinessKubePrometheusStackTeamsWebhookURL(t *testing.T) {
	for _, tt := range []struct {
		name       string
		webhookURL string
		wantIssue  bool
	}{
		{name: "optional", webhookURL: "", wantIssue: false},
		{name: "valid HTTPS", webhookURL: "https://example.webhook.office.com/webhookb2/abc", wantIssue: false},
		{name: "HTTP rejected", webhookURL: "http://example.webhook.office.com/webhookb2/abc", wantIssue: true},
		{name: "malformed rejected", webhookURL: "://not-a-url", wantIssue: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validReadinessConfig(t, "kind")
			cfg.OpenCenter.Services["kube-prometheus-stack"].(*services.PrometheusStackConfig).WebhookURL = tt.webhookURL

			report := ValidateReadiness(cfg)
			if tt.wantIssue {
				assertIssue(t, report, SeverityError, CategoryServices, "opencenter.services.kube-prometheus-stack.webhook_url")
				return
			}
			assertNoIssue(t, report, "opencenter.services.kube-prometheus-stack.webhook_url")
		})
	}
}
