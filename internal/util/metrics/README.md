# Metrics package

`internal/util/metrics` records in-process operation timings. A
`MetricsCollector` is enabled by default, is safe for concurrent use, and
stores measurements in memory until cleared.

Supported metric types are `template_render`, `config_build`,
`gitops_generation`, `validation`, and `migration`. Each `Metric` records a
name, duration, timestamp, success flag, optional error, and metadata.

```go
collector := metrics.NewMetricsCollector()
timer := collector.NewTimer(metrics.MetricTypeTemplateRender, "cluster.tmpl")
timer.WithMetadata("provider", "openstack")
timer.Stop()

summary := collector.GetSummary()
collector.SetEnabled(false)
```

Collectors expose `RecordMetric`, `GetMetrics`, `GetMetricsByType`, `Clear`,
and `GetSummary`. Summaries include counts, success/failure totals, total,
average, minimum, maximum, and P50/P95/P99 durations. `StopWithError` records
a failed measurement. The package also provides a global collector and
`RecordTemplateRender`, `RecordConfigBuild`, and `RecordGitOpsGeneration`
helpers.

Template rendering records through the global collector in
`internal/template/engine.go`. The implementation and tests are in this
directory.
