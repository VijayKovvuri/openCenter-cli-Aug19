# Suggestion engine

`SuggestionEngine` enriches `ValidationResult` issues with actionable text.
The default engine registers typo and context rules; callers can add a
`SuggestionRule` through `ValidationEngine.AddSuggestionRule` or directly on
the suggestion engine.

The typo rule compares a value with the supplied `valid_values` context using
case-insensitive Levenshtein distance and returns up to three close matches.
The context rule uses field names and validation codes to suggest common fixes
for formats such as email, URL, CIDR, IP address, ports, names, versions, and
boolean or numeric values. Suggestions are deduplicated and sorted for stable
output.

```go
engine := validation.NewSuggestionEngine()
result := validation.NewValidationResult()
result.AddError("provider", "unknown provider")
engine.EnhanceResult(result, map[string]interface{}{
    "valid_values": []string{"openstack", "vsphere", "kind"},
})
```

The implementation is in `suggestions.go`; unit, demo, and example coverage
is in `suggestions_test.go` and the adjacent test files.
