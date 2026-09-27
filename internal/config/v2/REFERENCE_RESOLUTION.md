# v2 reference resolution

`ReferenceResolver` resolves references in the mutable v2 configuration
before defaults and validation are applied. Supported forms are:

```text
${env:VARIABLE_NAME}
${file:/path/to/file}
${ref:path.to.value}
```

Environment variables must be set and non-empty. File contents are read with
the configured filesystem, trimmed, and cached for the lifetime of the
resolver. Config-path references resolve YAML-tagged struct fields, string-key
maps, and slice indexes, for example:

```yaml
opencenter:
  meta:
    name: example
  cluster:
    cluster_name: "${ref:opencenter.meta.name}"
```

References can be combined in one string and can occur in nested structs,
maps, slices, and arrays. The resolver detects cycles and rejects traversal
deeper than ten levels. Errors identify missing environment variables, unreadable
files, unresolved paths, cycles, or excessive depth.

The implementation is `resolver.go`; focused and property-based coverage is in
`resolver_test.go` and `resolver_property_test.go`.
