# Contributing

Thank you for your interest in contributing to the Terraform Provider for 3x-ui!

## Prerequisites

- **Go 1.26+**
- **[Task](https://taskfile.dev/)** — task runner (`brew install go-task` / `go install github.com/go-task/task/v3/cmd/task@latest`)
- **[golangci-lint](https://golangci-lint.run/welcome/install/)**
- **[pre-commit](https://pre-commit.com/)** — git hooks framework (`pip install pre-commit`)
- **Docker** — for running the local 3x-ui environment
- **Terraform** — required for acceptance tests

## Local Setup

```bash
# Clone the repository
git clone https://github.com/batonogov/terraform-provider-threexui.git
cd terraform-provider-threexui

# Install pre-commit hooks
pre-commit install

# Build the provider
task build
```

## Common Commands

| Command | Description |
| --- | --- |
| `task build` | Build the provider binary |
| `task fmt` | Format Go code (gofmt) |
| `task vet` | Run `go vet` |
| `task lint` | Run golangci-lint |
| `task pre-commit` | Run all checks manually (fmt, vet, lint, build) |
| `task test:unit` | Run unit tests (no Docker or Terraform needed) |
| `task test:acc` | Run acceptance tests (starts Docker automatically) |
| `task test` | Run unit + acceptance tests |

## Testing

### Unit Tests

Unit tests do not require Docker or Terraform. Run them with:

```bash
task test:unit
```

### Acceptance Tests

Acceptance tests spin up a real 3x-ui instance via Docker Compose, run Terraform-based tests against it, and tear it down automatically:

```bash
task test:acc
```

This handles `docker compose up`, sets the required environment variables (`TF_ACC`, `THREEXUI_ENDPOINT`, etc.), runs the tests, and calls `docker compose down` on exit.

> **Note:** acceptance tests are **not** run in pre-commit hooks. Run them explicitly before submitting changes that affect provider logic.

### Writing Tests

- Unit tests go alongside provider code in `provider/` with a `_test.go` suffix.
- Acceptance test function names must start with `TestAcc`.
- Use `testAccProtoV6ProviderFactories()` and `ProtoV6ProviderFactories` in test cases.
- HCL configs in tests use typed blocks and attributes (not `jsonencode()`).

## Project Structure

All provider code lives in the `provider/` directory. See [CLAUDE.md](CLAUDE.md) for a detailed file-by-file breakdown.

Key patterns:

- **Typed blocks** — settings, stream_settings, sniffing, and xray resources use typed Terraform blocks (not raw JSON).
- **Three-layer conversion** — Typed Model ↔ Untyped Map ↔ JSON String.
- **Per-protocol blocks** — inbound settings and outbound settings are split by protocol (`vless_settings`, `trojan_settings`, etc.).
- **Singletons** — panel settings and xray settings resources use a fixed ID (`"settings"` or `"xray_version"`).

## 3x-ui Source Snapshots

The repository may contain `3x-ui-<version>/` directories (git-ignored) with 3x-ui source snapshots. These are used for diffing between versions when updating the provider to support a new 3x-ui release. See the "Updating 3x-ui Version" section in [CLAUDE.md](CLAUDE.md) for the full process.

## Submitting Changes

1. Create a feature branch from `main`.
2. Make your changes and ensure `task pre-commit` passes.
3. Run relevant tests (`task test:unit` at minimum; `task test:acc` if provider logic changed).
4. Commit using [Conventional Commits](https://www.conventionalcommits.org/) format:
   - `feat: add mixed inbound support`
   - `fix: handle empty client list`
   - `docs: update README examples`
   - `test: add round-trip tests for DNS`
   - `chore: bump golangci-lint version`
5. Open a pull request against `main`.

## Release Process

Releases are fully automated:

1. Commits land on `main` with conventional prefixes (`feat:`, `fix:`, `feat!:`, etc.).
2. [Release Please](https://github.com/googleapis/release-please) automatically creates/updates a Release PR with a version bump and changelog.
3. When the Release PR is merged, a git tag is created and [GoReleaser](https://goreleaser.com/) builds, signs (GPG), and publishes the GitHub Release.
4. The [Terraform Registry](https://registry.terraform.io/providers/batonogov/threexui/latest) picks up the new release automatically.

Maintainers only need to review and merge the Release PR — everything else is handled by CI.

## Code Style

- Run `task fmt` before committing (also enforced by pre-commit hooks).
- Follow existing patterns in the codebase — keep changes minimal and focused.
- Do not add features or refactor code beyond what the issue requires.

## Questions?

Open an issue or start a discussion on GitHub.
