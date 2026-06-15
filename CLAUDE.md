# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Terraform provider for the [3x-ui](https://github.com/MHSanaei/3x-ui) panel.
Go (version pinned in [`go.mod`](go.mod)) with `terraform-plugin-framework`. Module: `github.com/batonogov/terraform-provider-threexui`.
Registry: `batonogov/threexui`. All provider code lives in `provider/`.
Releases are automated: conventional commits on `main` → Release Please PR → GoReleaser (GPG-signed) → Terraform Registry.

Provider config attributes: `endpoint`, `username`, `password`, `base_path`,
`bootstrap_username`/`bootstrap_password` (first-run setup), `two_factor_code` (TOTP),
`insecure_skip_verify`, `request_timeout`, `max_retries`.
Env vars: `THREEXUI_ENDPOINT`, `THREEXUI_USERNAME`, `THREEXUI_PASSWORD`,
`THREEXUI_BASE_PATH`, `THREEXUI_INSECURE_SKIP_VERIFY`, `THREEXUI_REQUEST_TIMEOUT`,
`THREEXUI_MAX_RETRIES`. Bootstrap and 2FA attributes have **no** env-var fallback.

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

**Watch CI status for a PR:**

```bash
gh pr checks <PR>          # one-shot
gh pr checks <PR> --watch  # live
```

**Version-specific acc tests:**

```bash
THREEXUI_VERSION=v3.1.0 task test:acc:compat
```

**Single acceptance test:**

```bash
TF_ACC=1 THREEXUI_ENDPOINT=http://localhost:2053 \
  THREEXUI_USERNAME=admin THREEXUI_PASSWORD=admin \
  THREEXUI_VERSION=v3.3.1 \
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

### Write-only secret attributes (Terraform 1.11+ / OpenTofu 1.11+)

Four singleton secrets have write-only (`_wo`) alternatives following the AWS/Azure pattern.
Each secret gets three attributes: old `Sensitive` attr + `WriteOnly` attr + `_wo_version` trigger.

| Resource | Old attr | Write-only | Version trigger |
| --- | --- | --- | --- |
| `panel_user` | `password` | `password_wo` | `password_wo_version` |
| `panel_security` | `two_factor_token` | `two_factor_token_wo` | `two_factor_token_wo_version` |
| `panel_telegram` | `tg_bot_token` | `tg_bot_token_wo` | `tg_bot_token_wo_version` |
| `panel_general` | `ldap_password` | `ldap_password_wo` | `ldap_password_wo_version` |

- `PreferWriteOnlyAttribute` validator warns on TF >= 1.11 when using old attr.
- `int64validator.AlsoRequires` enforces that `*_wo_version` can only be set with `*_wo`.
- Write-only values read from `req.Config` (not plan/state — framework nulls them).
- Two resolution strategies:
  - **panel_user**: `password` is nulled in state when `password_wo` is used. Update
    falls back to provider credentials when state password is empty. No ModifyPlan
    needed because state has no prior password to conflict with.
  - **settings (security/telegram/general)**: `resolveXxxWO` copies `_wo` value into
    the old model field before `expand*()`. ModifyPlan marks the plain attr as Unknown
    on `woVersionTriggered` so Terraform accepts a new sensitive value from Apply.
    Needed because flatten reads the masked value back, which would otherwise be
    rejected as "inconsistent values for sensitive".
- Version trigger: `resolveXxxWOUpdate` only sends the `_wo` value when
  `woVersionTriggered(plan, state)` (version changes, or first use when state has none).

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

### Panel user (`resource_panel_user.go`)

`threexui_panel_user` — no read API exists. Read is a no-op, state is preserved
from plan. Create uses provider credentials as old credentials; Update uses
prior state credentials (falls back to provider credentials when state
password is empty, e.g. after using `password_wo`). See Write-only section
above for the `password` / `password_wo` lifecycle.

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

### github-cache (offline GitHub for xray-version tests)

`ci/github-cache/` is a local MITM forward proxy (`docker-compose.yaml`, issue #285)
that makes the xray-version acceptance tests (`TestAccXrayVersion`,
`TestAccXrayVersionResource`, `TestAccXrayVersionDrift`) fully offline and
deterministic. The panel's outbound HTTPS to `api.github.com` and `github.com`
is routed through the proxy via `HTTPS_PROXY` (honoured because
`netproxy.NewHTTPClient` returns a plain client when `panel_outbound` is unset).

- The proxy MITMs the two GitHub hosts with a CA generated at runtime and
  trusted by the `3xui` container via `SSL_CERT_FILE`; it serves vendored
  fixtures from `ci/github-cache/fixtures/` (release list + per-version
  `Xray-linux-64.zip`). Unknown GitHub paths 404 (fail loud).
- Non-GitHub hosts (`ident.me`, `ipify`, …) are **transparently tunnelled**
  (passthrough) so the panel's unrelated outbound traffic is unaffected.
- Only **amd64** zips are committed (CI = `ubuntu-latest`); arm64 is opt-in via
  `ARCHES="linux-64 linux-arm64-v8a" scripts/fetch-xray-fixtures.sh`.
- When a new 3x-ui image bundles a different Xray-core build, add the version
  to `VERSIONS` in `scripts/fetch-xray-fixtures.sh` and re-run it.
- The proxy is a separate Go module (`ci/github-cache/go.mod`) — CI-only,
  never shipped. Its correctness is guaranteed by the Docker build (compile)
  plus `go test ./...` in that module.

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
| Write-only attrs quirks | See "Write-only secret attributes" section above — read `_wo` from `req.Config`; ModifyPlan marks plain `Unknown` on version change; `panel_user` nulls state password instead |
| xray-settings acc-tests restart xray-core | `TestAccXrayBasics`/`DNS`/`Routing`/`Balancers`/`Reverse`/`Outbounds` each apply a new config = xray restart; `version` becomes `"Unknown"` for ~30-90s after. Use `waitForXrayVersion(t, ctx, client)` poll-loop in any acc-test probing version after them (#280) — the github-cache does NOT fix restart races |
| `GetXrayVersions` / `InstallXray` reach GitHub | Served offline by the **github-cache** proxy (`docker-compose.yaml`, issue #285). `skipOnUpstreamRateLimit(t, err)` / `testAccSkipOnXrayVersionsRateLimit(t)` remain as a defense for when the cache is bypassed (e.g. `panel_outbound` active) |
| PostgreSQL runner is slower than SQLite | Timing-sensitive flakes (xray restart windows, InstallXray pickup) surface on the `test:acc:postgres` job but not on `test:acc`. A green SQLite run does NOT prove PostgreSQL is green |

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

- **README** is localized in **7 languages** (en, ru, fa, ar, zh, es, tr). When changing
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
  specific 3x-ui versions. Currently supported: **v3.1.x**, **v3.2.x**, **v3.3.x** (up to v3.3.1).
- Flaky test quarantine: `skipOnFlakyVersions(t, ...)` / `skipIfFlaky(t)` with
  `THREEXUI_SKIP_FLAKY` env var to skip known-broken upstream versions.
- Xray-version acc-tests use `waitForXrayVersion(t, ctx, client)` (90×1s retry on
  `ErrXrayVersionUnknown`) — covers both the cold-start race (handled by the
  `_wait-for-xray` Taskfile gate) and mid-test restarts after `TestAccXray*` settings
  tests. **Do not** call `client.GetCurrentXrayVersion()` directly in acc-tests (#280).
- Any acc-test exercising `GetXrayVersions` (directly or via `threexui_xray_versions`
  data source / `xray_version` resource) must pre-flight with
  `testAccSkipOnXrayVersionsRateLimit(t)`, or wrap the error with
  `skipOnUpstreamRateLimit(t, err)`. Rate limit is an environment issue, skip —
  don't fail (#279).
- Protocol matrix test (`resource_inbound_matrix_test.go`): comprehensive
  create/update/import round-trip for every protocol.
- Destroy checks use `destroyVisibilityAttempts` (60 × 500 ms) to handle
  SQLite visibility lag after delete.

### Pre-commit

`pre-commit install` sets up hooks: `task fmt`, `task vet`, `task lint`, `task build`,
`markdownlint-cli2` (markdown files), plus standard checks (trailing whitespace,
end-of-file fixer, YAML/JSON validation, large files, merge conflicts).
Run manually: `pre-commit run --all-files`.

### Supply chain

CI lint job also runs `govulncheck ./...` before fmt/vet/lint.
All CI actions pinned to commit SHA (`@<sha> # vN`).
Pre-commit hooks use `--freeze` format (`rev: <sha>  # frozen: <tag>`).

### Adding a new 3x-ui version to CI

`compat-versions.json` is the single source of truth for tested versions.
CI and flake-tracking workflows read the matrix dynamically from this file.
`scripts/sync-versions.sh check` validates that README tables, docker-compose,
and Taskfile defaults are in sync.

When a new 3x-ui version is released:

1. **`compat-versions.json`** — add the new version to the `versions` array, update `default_version` if it's the latest patch
2. **`scripts/sync-versions.sh fix`** — auto-updates README tables, docker-compose default, and Taskfile defaults
3. **`scripts/sync-versions.sh check`** — verify all surfaces are in sync
4. **Source snapshot** — copy the 3x-ui source to `3x-ui-<version>/` (drift tests use the latest snapshot)
5. **`provider/drift_test.go`** — if fields were added/removed from upstream structs, update `intentionallySkipped` and known-field maps
6. **`provider/testdata/upstream_contract.json`** — update `all_setting_fields`, `protocols_go_model`, `version` to match the latest 3x-ui release (used by drift tests when no local snapshot is present)
7. **`docs/`** — if provider schema changed, update the corresponding resource/data-source doc
8. **`CLAUDE.md`** — update the "up to vX.Y.Z" version reference if the minor line changed
