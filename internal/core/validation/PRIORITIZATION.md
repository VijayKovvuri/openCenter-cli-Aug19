# Validator priority

Validators expose `Priority() int`; lower values run first. The package
defines three conventional values:

| Constant | Value | Intended use |
|---|---:|---|
| `PriorityHigh` | 50 | Fast format and structural checks |
| `PriorityNormal` | 100 | Standard business and cross-field checks |
| `PriorityLow` | 200 | I/O, network, or other slower checks |

`ValidateAll` sorts requested validators by priority before sequential
execution. `ValidateParallel` sorts before launching the requested validators;
the calls still run concurrently, so priority is launch order rather than a
completion-order guarantee. Security validators always run sequentially before
the requested set.

Use `NewValidatorFunc` for normal priority or
`NewValidatorFuncWithPriority` for an explicit value. The ordering behavior is
covered by `prioritization_test.go`.
