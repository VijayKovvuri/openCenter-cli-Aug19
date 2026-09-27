# v2 configuration package

`internal/config/v2` defines the `schema_version: "2.0"` cluster
configuration model, loader, resolver, defaults, provider/deployment
validation, readiness checks, and serialization helpers.

## Structure

- `config.go`, `cluster.go`, `infrastructure.go`, `deployment.go`, and
  `services.go` define the configuration domains.
- `loader.go` and `manager.go` load and manage configurations.
- `resolver.go` resolves `${ref:...}`, `${env:...}`, and `${file:...}` values.
- `defaults.go`, `provider.go`, and `deployment_validator.go` apply defaults
  and enforce provider/deployment constraints.
- `readiness.go` and the generation/storage validation files protect the
  generation and service-storage contracts.

## Reference resolution

References are resolved recursively in strings, maps, slices, and structs.
Environment and file values are cached per resolver; file values are trimmed.
Config-path references support nested fields, map keys, and slice indexes, and
the resolver rejects missing values, cycles, and recursion deeper than ten
levels. See `REFERENCE_RESOLUTION.md` for the supported syntax.

## Validation and tests

Provider, deployment, service, storage, readiness, property, and serialization
tests live beside the implementation. Run the package tests with:

```bash
go test ./internal/config/v2/...
```

When changing the schema, update the corresponding Go types and tests rather
than documenting a separate implementation summary here.
