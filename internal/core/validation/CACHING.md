# Validation result cache

`ValidationEngine` creates a `ValidationCache` with a five-minute default TTL.
Entries are keyed by validator name and the SHA-256 hash of the JSON encoding
of the input. A cache hit returns an unexpired result without running that
requested validator.

```go
engine := validation.NewValidationEngineWithCache(10 * time.Minute)
result, err := engine.Validate(ctx, "cluster-name", value)

engine.InvalidateCache("cluster-name", oldValue)
engine.InvalidateAllCache("cluster-name")
engine.ClearCache()
removed := engine.CleanExpiredCache()
stats := engine.CacheStats()
```

Pass `0` to `NewValidationEngineWithCache` to disable caching. Expired entries
are ignored on reads and remain allocated until `CleanExpiredCache` or
`ClearCache` removes them. `GetCache` exposes the lower-level
`ValidationCache`, whose `Get`, `Set`, `Invalidate`, `InvalidateAll`, `Size`,
and `Stats` methods are concurrency-safe.

The cache implementation and coverage are in `cache.go`, `cache_test.go`, and
the engine tests.
