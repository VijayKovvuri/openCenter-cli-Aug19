# Golden template fixtures

This directory contains expected outputs for
`TestTemplateRenderingGolden` and its custom-function and edge-case variants.
The test compares rendered templates with these files; it does not discover
goldens automatically.

To verify the fixtures:

```bash
go test ./internal/template -run TestTemplateRenderingGolden
```

When a template change intentionally changes output, review the rendered diff
and regenerate the affected files with:

```bash
go test ./internal/template -run TestTemplateRenderingGolden -update-golden
```

The test cases in `../engine_golden_test.go` are the source of truth for the
fixture names and input data.
