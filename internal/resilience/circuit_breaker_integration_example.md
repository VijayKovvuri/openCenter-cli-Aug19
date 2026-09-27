# Circuit breaker integration

The resilience package exposes a circuit breaker for operations that call an
external service. `internal/barbican/client.go` wraps each Barbican operation
with the breaker and then the retry handler:

```go
err := client.circuitBreaker.Call(ctx, func() error {
    return client.retryHandler.Do(ctx, func() error {
        return callOpenStackAPI()
    })
})
```

`NewCircuitBreaker` fills zero-valued configuration fields with these defaults:

| Setting | Default |
|---|---:|
| Consecutive failures before opening | 5 |
| Successes in half-open before closing | 2 |
| Open timeout | 60 seconds |
| Concurrent half-open requests | 1 |

The states are `StateClosed`, `StateOpen`, and `StateHalfOpen`. A closed
breaker opens after the failure threshold. After the timeout, one or more
permitted calls test recovery; a failure reopens the breaker and enough
successes close it. `GetState` reads the current state and `Reset` returns it
to closed.

The implementation is `circuit_breaker.go`. Unit and property coverage is in
`circuit_breaker_test.go` and `circuit_breaker_property_test.go`.
