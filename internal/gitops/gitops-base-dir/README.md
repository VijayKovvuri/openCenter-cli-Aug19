# Generated GitOps repository

This directory is the base layout copied into a cluster's generated GitOps
workspace. It contains:

- `applications/` — application manifests and environment overlays.
- `infrastructure/` — cluster-specific infrastructure configuration.

The CLI owns generated files. Change the cluster configuration and regenerate
with:

```bash
opencenter cluster generate <organization>/<cluster> --render-only
```

Review generated changes before committing them to the GitOps repository.
