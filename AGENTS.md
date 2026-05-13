# Repository Guidelines

## Project Structure & Module Organization

This repository is a single Go module for the `github.com/batonogov/terraform-provider-threexui` Terraform provider; the Terraform Registry address is `batonogov/threexui`. Entry point code lives in `main.go`. Nearly all provider logic is under `provider/`: `resource_*.go` files implement Terraform resources, `data_source_*.go` files implement data sources, and `*_schema.go` files hold typed schema/expand/flatten helpers. Unit and acceptance tests sit beside the code as `*_test.go`; acceptance coverage is typically grouped in `acc_*_test.go`. Registry docs live in `docs/`, example Terraform configs in `examples/`, and local 3x-ui test infrastructure in `docker-compose.yaml`.

## Build, Test, and Development Commands

Use `task` for all routine work:

- `task build` builds the local provider binary.
- `task fmt` runs `gofmt -w provider/*.go`.
- `task vet` runs `go vet ./...`.
- `task lint` runs `golangci-lint run`.
- `task test:unit` runs unit tests only.
- `task test:acc` starts Docker Compose and runs Terraform acceptance tests.
- `task test:acc:compat` runs acceptance tests against a selectable 3x-ui version via `THREEXUI_VERSION` (defaults to `v3.0.1`).
- `task test` runs both unit and acceptance suites.
- `task pre-commit` runs the Go pre-commit checks: fmt, vet, lint, and build. The actual `.pre-commit-config.yaml` also includes markdown and common file checks.

## Coding Style & Naming Conventions

Follow standard Go formatting and keep code `gofmt`-clean. Use tabs for indentation, exported identifiers in `CamelCase`, and Terraform schema field names in `snake_case`. Match existing file patterns: new resources go in `provider/resource_<name>.go`, data sources in `provider/data_source_<name>.go`, and schema helpers in `provider/<area>_schema.go`. Keep changes focused; avoid unrelated renames or broad reformatting.

## Testing Guidelines

Write table-driven unit tests where practical and keep them next to the code they cover. Name tests `TestXxx`; reserve `TestAccXxx` for acceptance coverage. Run `task test:unit` before every PR. Run `task test:acc` for API, state, schema, or Docker-related changes; it expects Docker and Terraform locally and uses the bundled `admin/admin` 3x-ui setup on `http://localhost:2053`.

## Commit & Pull Request Guidelines

History follows Conventional Commits: `feat:`, `fix:`, `docs:`, `ci:`, `test:`, `chore:`. Keep subjects imperative and concise, for example `fix: normalize inbound listen defaults`. PRs should include a short problem/solution summary, linked issue when relevant, and the commands you ran. Update `docs/` and `examples/` when resource behavior or schema changes.
