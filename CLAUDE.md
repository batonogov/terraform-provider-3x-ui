# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Terraform provider for the [3x-ui](https://github.com/MHSanaei/3x-ui) panel.
Go with `terraform-plugin-framework`. Module: `github.com/batonogov/terraform-provider-threexui`.
Registry: `batonogov/threexui`. All provider code lives in `provider/`.

Provider config attributes: `endpoint`, `username`, `password`, `base_path`,
`bootstrap_username`/`bootstrap_password` (first-run setup), `two_factor_code` (TOTP),
`insecure_skip_verify`, `request_timeout`, `max_retries`.

---

## Commands

| Command | Description |
| --- | --- |
| `task build` | Build provider binary |
| `task fmt` | `gofmt -w provider/*.go` |
| `task vet` | `go vet ./...` |
| `task lint` | `golangci-lint run` |
| `task test` | Unit + acceptance tests |
| `task test:unit` | Unit tests (no Docker/Terraform needed) |
| `task test:unit:coverage` | Unit tests with coverage (`coverage.out`) |
| `task test:acc` | Acceptance tests (Docker lifecycle included) |
| `task pre-commit` | fmt + vet + lint + build |

**Version-specific acc tests:**

```bash
THREEXUI_VERSION=v2.9.0 task test:acc:compat
```

**Single acceptance test:**

```bash
TF_ACC=1 THREEXUI_ENDPOINT=http://localhost:2053 \
  THREEXUI_USERNAME=admin THREEXUI_PASSWORD=admin \
  go test ./provider -run TestAccInboundVLESS -count=1 -timeout 600s -v
```

---

## Architecture

### Three-layer conversion (inbound)

`threexui_inbound` and `threexui_inbound_client` are the core resources.
`settings`, `stream_settings`, `sniffing` are JSON strings in the 3x-ui API but typed
blocks in the Terraform schema. Conversion flows through three layers:

```text
typed model  (structs with types.String / types.Int64 / types.Bool)
  ↔  untyped map  (map[string]any — expand* / flatten* functions)
  ↔  JSON string  (build* / flatten* functions)
```

Schema helpers: `inbound_*_schema.go`, `settings.go`, `stream_settings.go`, `sniffing.go`.

### Panel settings (singletons)

Four resources — `panel_general`, `panel_security`, `panel_telegram`,
`panel_subscription` — share `resource_settings_tabs.go`.

- Each is a singleton with ID `"settings"`.
- Shared CRUD via `settingsApplyTyped` / `settingsReadTyped`.
- Delete only clears TF state — does not reset panel settings.
- `settingsMu` serializes read-modify-write (separate from `xrayTemplateMu`).
- `settingsSecrets` on Client caches secret values from GET responses (3x-ui masks them)
  so they can be replayed in subsequent PUT requests.
- Changing `web_base_path` in `panel_general` triggers a panel restart and updates
  the provider client's base path via `SetBasePath` + `WaitForReady`.

### Xray settings (two modes)

Six xray resources share `resource_xray_settings.go`: `xray_basics`, `xray_dns`,
`xray_routing`, `xray_balancers`, `xray_reverse`, `xray_outbounds`. Each section
has its own `*_schema.go` file.

| Mode | Resource | Behavior |
| --- | --- | --- |
| Merge root | `xray_basics` | Reads full template, merges changed fields, writes back |
| Set path | all others | Reads/writes a section by JSON path |

`xrayTemplateMu` serializes read-modify-write to prevent race conditions.

### Xray version (`resource_xray_version.go`)

Separate resource — singleton with ID `"xray_version"`. Calls `InstallXray` + polls
`waitForXrayVersion` until the version matches. Delete is a no-op with a warning
(removing from state does NOT revert the installed version).

### Panel user (write-only)

`threexui_panel_user` — no read API exists. Read is a no-op, state is preserved
from plan. Create uses provider credentials as old credentials; Update uses
previous state.

### HTTP client (`client.go`)

Cookie auth, auto re-login on 401/404, CSRF handling for 3x-ui v3.
Supports bootstrap credentials for first-run panel setup.
Provider attributes: `two_factor_code` (TOTP for 2FA login), `max_retries`
(default 1, max 10), `request_timeout`.
Three retry layers: 5xx on idempotent writes, read-after-write for SQLite
visibility lag, rate-limit retry for `GetXrayVersions`.
Client API auto-detection: probes `/panel/api/clients/list` to detect v3.1.0+
and falls back to old `/panel/api/inbounds/*` endpoints on older versions.

### JSON format compatibility (`types.go`)

3x-ui v3.1.0 returns `settings`, `streamSettings`, `sniffing` as nested JSON
objects instead of escaped strings. Custom `UnmarshalJSON` on `Inbound`
normalises both formats to plain strings — rest of the code is unaffected.

### Data sources

Seven data sources: `inbounds`, `server_status`, `xray_versions`, `xray_config`,
`settings`, `online_clients`, `client_traffics`. All are read-only GET wrappers
except `inbounds` which supports filtering by protocol/tag. JSON attrs that
contain secrets (UUIDs, private keys) must be marked `Sensitive: true`.

---

## Critical Gotchas

These are non-obvious constraints that have caused real bugs.

| Gotcha | Why it matters |
| --- | --- |
| `email` in `threexui_inbound_client` is **Required** | Without it, 3x-ui crashes with SQL error on next client add |
| Docker `base_path` is `/` | Do **not** set `THREEXUI_BASE_PATH=/panel/` |
| Data-source JSON attrs need `Sensitive: true` | Payloads contain UUIDs, private keys, tokens |
| `Optional+Computed` attrs need `UseStateForUnknown` | Prevents false drift (`known after apply`) after import |
| `alignBlocksWithPlan` | Prevents "was absent, but now present" for Optional blocks; skipped during Import |
| `reality_settings.settings` is `SingleNestedAttribute` | Not a block. Uses `objectplanmodifier.UseStateForUnknown()` |
| Subscription resource does **double apply** | Workaround: `sub_json_enable` not saved on first apply with `sub_enable` |
| Policy levels: `{"0": {...}}` ↔ `[{id=0}]` | Xray JSON uses a map, TF uses a list |
| DNS servers: string vs object | Address-only → string; with extra fields → object |
| `xray_version` delete is a no-op | Removing from state does NOT revert the installed xray version |
| `web_base_path` change triggers panel restart | Must also update provider `base_path`; code auto-updates client |

**Always check 3x-ui source snapshots** (`3x-ui-<version>/`) before assuming API behavior.
Key paths: `database/model/`, `web/service/`, `web/controller/`, `web/entity/`, `frontend/src/schemas/`, `xray/`.

---

## Conventions

### Commits

Conventional Commits: `feat:`, `fix:`, `docs:`, `ci:`, `test:`, `chore:`.
Imperative mood, concise subjects.

### File naming

| Pattern | Example |
| --- | --- |
| Resources | `provider/resource_<name>.go` |
| Data sources | `provider/data_source_<name>.go` |
| Schema helpers | `provider/<area>_schema.go` |

### Documentation

- **README** is localized in **6 languages** (en, ru, fa, ar, zh, es). When changing
  `README.md`, update all `README.<locale>.md` files in the same PR. Persian/Arabic
  wrap body in `<div dir="rtl">`.
- **SECURITY.md** tracks sensitive fields — add a row when adding resources that
  handle secrets.
- **`docs/guides/`** for operational walkthroughs needing more than an `examples/` folder.

### Testing

- Unit tests: `TestXxx` naming, table-driven where practical.
- Acceptance tests: `TestAccXxx`, `terraform-plugin-testing`,
  `ProtoV6ProviderFactories` (not `ProviderFactories`).
- Version-aware skipping: `requireMinVersion(t, "vX.Y.Z")` for features from
  specific 3x-ui versions. Currently supported: **v2.9.x**, **v3.0.x**, **v3.1.x**, **v3.2.x** (up to v3.2.5).
- Flaky test quarantine: `skipOnFlakyVersions(t, ...)` / `skipIfFlaky(t)` with
  `THREEXUI_SKIP_FLAKY` env var to skip known-broken upstream versions.
- Protocol matrix test (`resource_inbound_matrix_test.go`): comprehensive
  create/update/import round-trip for every protocol.
- Destroy checks use `destroyVisibilityAttempts` (60 × 500 ms) to handle
  SQLite visibility lag after delete.

### Supply chain

All CI actions pinned to commit SHA (`@<sha> # vN`).
Pre-commit hooks use `--freeze` format (`rev: <sha>  # frozen: <tag>`).
