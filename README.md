# openCenter CLI

openCenter CLI turns a declarative YAML configuration into a Kubernetes cluster and a GitOps repository. It supports OpenStack, VMware, Baremetal, Magnum, and Kind workflows, with configuration validation, secrets management, and FluxCD/Kustomize integration.

See the [provider reference](docs/reference/providers.md) for the current support boundaries. AWS, GCP, and Azure are not supported cluster providers.

## Capabilities

- Define cluster infrastructure, Kubernetes settings, services, and secrets in one configuration.
- Validate configuration before deployment and generate a FluxCD-ready GitOps repository.
- Manage cluster lifecycle, platform services, worker pools, backups, drift checks, and secrets.
- Use SOPS Age encryption for secrets that live alongside configuration.

## Quick start

```bash
# Install tools
mise install

# Build CLI
mise run build

# Initialize cluster
./bin/opencenter cluster init my-cluster --org my-org

# Edit configuration
$EDITOR ~/.config/opencenter/clusters/blueprints/my-org/.my-cluster-config.yaml

# Validate
./bin/opencenter cluster validate my-cluster

# Generate GitOps repository
./bin/opencenter cluster generate my-cluster

# Deploy
./bin/opencenter cluster deploy my-cluster
```

Start with the [Getting Started guide](docs/getting-started/getting-started.md) for prerequisites and provider-specific guidance.

## Documentation

Use the [Documentation Home](docs/index.md) as the canonical index:

- [Getting Started](docs/getting-started/) — first cluster and provider walkthroughs.
- [Operations](docs/operations/) — deployment, configuration, secrets, upgrades, backups, and troubleshooting.
- [Reference](docs/reference/) — commands, configuration, providers, services, environment variables, and file locations.
- [Concepts](docs/concepts/) — architecture, GitOps, security, configuration, and provider explanations.
- [Contributing](docs/contributing/) — development setup, testing, code structure, and project changes.

Useful starting points include the [configuration schema](docs/reference/configuration-schema.md), [platform services](docs/reference/platform-services.md), [CLI command reference](docs/reference/cli-commands.md), and [file locations](docs/reference/file-locations.md).

## Development

See [Development Environment Setup](docs/contributing/development-setup.md) for the supported local toolchain and build/test workflow. Mise tasks are documented in the [Mise Tasks Reference](docs/reference/mise-tasks.md).

## Contributing

Contributions are welcome. Read the [Contributing Guide](docs/contributing/contributing.md) before opening a pull request.

## Support

- [Documentation](docs/index.md)
- [Security Policy](SECURITY.md)
- [GitHub Issues](https://github.com/opencenter-cloud/openCenter-cli/issues)
- [GitHub Discussions](https://github.com/opencenter-cloud/openCenter-cli/discussions)

## License

Licensed under the [Apache License 2.0](LICENSE).

## Related projects

- [openCenter-gitops-base](https://github.com/opencenter-cloud/openCenter-gitops-base) — platform services library.
- [openCenter-customer-app-example](https://github.com/opencenter-cloud/openCenter-customer-app-example) — reference application deployment patterns.
- [openCenter-AirGap](https://github.com/opencenter-cloud/openCenter-AirGap) — air-gapped deployment packaging.
- [opencenter-windows](https://github.com/opencenter-cloud/opencenter-windows) — Windows worker node support.
