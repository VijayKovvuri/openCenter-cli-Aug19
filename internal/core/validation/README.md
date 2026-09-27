# Validation package

`internal/core/validation` provides a thread-safe registry and engine for
validators. A validator has a unique name, a priority, and a
`Validate(context.Context, interface{}) (*ValidationResult, error)` method.
Results contain errors, warnings, informational issues, and suggestions.

## Engine behavior

```go
engine := validation.NewValidationEngine()
engine.MustRegister(validation.NewValidatorFunc("cluster-name", validate))
result, err := engine.Validate(ctx, "cluster-name", value)
```

- `Validate` runs one registered validator.
- `ValidateAll` runs requested validators sequentially.
- `ValidateParallel` runs requested validators concurrently.
- `ValidateWithOptions` and the `...WithOptions` variants support
  `StopOnFirstError` and warning filtering.
- Security validators registered with `RegisterSecurityValidator` run before
  requested validators and cannot be skipped.
- Results are enhanced by the suggestion engine and cached by validator name
  plus a JSON/SHA-256 data hash. New engines use a five-minute cache;
  `NewValidationEngineWithCache(0)` disables it.

Requested validators are ordered by ascending priority for sequential and
parallel execution. Standard values are `PriorityHigh` (50),
`PriorityNormal` (100), and `PriorityLow` (200).

Implementation and tests are colocated in `engine.go`, `registry.go`,
`types.go`, `suggestions.go`, `cache.go`, and their test files. The
`validators/` subpackage contains built-in validators.
