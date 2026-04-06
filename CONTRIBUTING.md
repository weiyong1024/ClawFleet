# Contributing to ClawFleet

Thanks for your interest in contributing! ClawFleet is an open-source project and we welcome contributions of all kinds.

## Ways to Contribute

- **Bug reports** — found something broken? [Open an issue](https://github.com/clawfleet/ClawFleet/issues/new?labels=bug)
- **Feature requests** — have an idea? [Open an issue](https://github.com/clawfleet/ClawFleet/issues/new?labels=enhancement)
- **Code** — pick an issue labeled [`good first issue`](https://github.com/clawfleet/ClawFleet/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) or [`help wanted`](https://github.com/clawfleet/ClawFleet/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
- **Documentation** — improve the [Wiki](https://github.com/clawfleet/ClawFleet/wiki), fix typos, add examples
- **Testing** — try ClawFleet on your setup and report what works or doesn't

## Development Setup

```bash
# Clone the repo
git clone https://github.com/clawfleet/ClawFleet.git
cd ClawFleet

# Install dependencies
go mod tidy

# Build
make build

# Run tests
make test
```

## Pull Request Process

1. Fork the repo and create a branch from `main`
2. Make your changes — keep the PR focused on one thing
3. Run `make test` and `make vet` before submitting
4. Open a PR with a clear description of what and why
5. A maintainer will review and provide feedback

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Add comments only where the logic isn't self-evident
- Don't add features, abstractions, or cleanup beyond the scope of your PR

## Architecture

See [CLAUDE.md](CLAUDE.md) for the full architecture overview. Key rule: product layer (web/, cli/) → infrastructure layer (container/, state/, port/, config/) — never reverse the dependency.

## Community

- [Discord](https://discord.gg/b5ZSRyrqbt) — ask questions, discuss ideas
- [Issues](https://github.com/clawfleet/ClawFleet/issues) — report bugs, request features

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
