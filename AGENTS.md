# Repository Guidelines

This file is the source of truth for coding agents working in this repository.

## Purpose

Terraform provider for the [3x-ui](https://github.com/MHSanaei/3x-ui) panel,
implemented in Go with `terraform-plugin-framework`. The Go module is
`github.com/batonogov/terraform-provider-threexui`; the Terraform Registry
address is `batonogov/threexui`.

## Project Structure & Module Organization

```text
provider/              - all provider code
  provider.go          - ThreeXUIProvider (framework): Metadata, Schema, Configure, Resources, DataSources
  client.go            - HTTP client for 3x-ui API (cookie auth, auto re-login)
  types.go             - Inbound, ClientTraffic, APIResponse, ParseJSONField
  resource_inbound.go  - threexui_inbound resource (CRUD, Reality, settings defaults)
  resource_inbound_client.go - threexui_inbound_client resource (mutex, UUID)
  resource_settings_tabs.go  - panel_general/security/telegram/subscription (typed attributes)
  resource_panel_user.go     - threexui_panel_user resource (admin credentials change)
  resource_xray_settings.go  - CRUD for xray_basics/dns/routing/balancers/reverse/outbounds (typed attributes)
  xray_basics_schema.go      - model, schema, expand/flatten for xray_basics (log, policy, api, stats)
  xray_dns_schema.go         - model, schema, expand/flatten for xray_dns (servers, hosts)
  xray_routing_schema.go     - model, schema, expand/flatten for xray_routing (rules)
  xray_balancers_schema.go   - model, schema, expand/flatten for xray_balancers
  xray_reverse_schema.go     - model, schema, expand/flatten for xray_reverse (bridges, portals)
  xray_outbounds_schema.go   - model, schema, expand/flatten for xray_outbounds (per-protocol settings)
  inbound_settings_schema.go - model, schema, expand/flatten for per-protocol settings (vless, trojan, ss, http, socks, mixed, wg, dokodemo, hysteria)
  inbound_stream_settings_schema.go - model, schema, expand/flatten for stream_settings (tcp, ws, grpc, httpupgrade, xhttp, kcp, hysteria, reality, sockopt)
  inbound_sniffing_schema.go - model, schema, expand/flatten for sniffing
  settings.go          - buildSettingsJSON(map[string]any), flattenSettings(string), expand/flatten clients/fallbacks/peers
  stream_settings.go   - buildStreamSettingsJSON(map[string]any), flattenStreamSettings(string), expand/flatten per-transport
  sniffing.go          - buildSniffingJSON(map[string]any), flattenSniffing(string)
  settings_helpers.go  - mergeSettings
  list_helpers.go      - typesListToAnySlice, typesListInt64ToAnySlice, anySliceToTypesList
  default_settings.go  - default settings per protocol, applyDefaultInboundSettings
  resource_xray_version.go - threexui_xray_version resource (install/manage Xray core version)
  data_source_*.go     - data sources (inbounds, server_status, settings, xray_config, xray_versions, online_clients)
  testdata/            - round-trip fixtures for corpus_test.go; see provider/testdata/README.md to refresh
examples/              - example TF configs for manual testing
docs/
  index.md             - provider docs landing page (Terraform Registry)
  resources/           - per-resource Registry docs
  data-sources/        - per-data-source Registry docs
  guides/              - operational walkthroughs (backup-as-code, server-migration, bulk-clients)
README.md              - English README; localized in README.ru_RU.md, README.fa_IR.md, README.ar_EG.md, README.zh_CN.md, README.es_ES.md
3x-ui-<version>/       - 3x-ui source snapshots (in .gitignore, for reference/diffing)
docker-compose.yaml    - 3x-ui on port 2053 (version via THREEXUI_VERSION env, default v3.0.2)
Taskfile.yml           - task build / test / fmt
.github/workflows/
  ci.yml               - lint, unit tests, acceptance tests, compatibility matrix (PR + push main)
  docs.yml             - docs/examples validation: terraform fmt, markdownlint, yamllint; tfplugindocs validate for schema/docs drift (PR + push main)
  release-please.yml   - Release Please + GoReleaser (conventional commits -> semver tag -> build + sign + publish)
```

## Provider Resources

| Terraform Resource | File | Description |
| --- | --- | --- |
| `threexui_inbound` | resource_inbound.go + inbound_*_schema.go | Inbound (vless/vmess/trojan/ss/http/mixed/wg/tunnel/hysteria). Typed blocks for settings/stream_settings/sniffing |
| `threexui_inbound_client` | resource_inbound_client.go | Client within an inbound. Typed attributes |
| `threexui_panel_general` | resource_settings_tabs.go | Panel settings (web, LDAP). Typed attributes |
| `threexui_panel_security` | resource_settings_tabs.go | 2FA. Typed attributes |
| `threexui_panel_user` | resource_panel_user.go | Admin credentials change. Write-only (no read API) |
| `threexui_panel_telegram` | resource_settings_tabs.go | Telegram bot. Typed attributes |
| `threexui_panel_subscription` | resource_settings_tabs.go | Subscriptions. Typed attributes |
| `threexui_xray_basics` | resource_xray_settings.go + xray_basics_schema.go | Base Xray config (merge root). Typed blocks |
| `threexui_xray_dns` | resource_xray_settings.go + xray_dns_schema.go | DNS (set path). Typed blocks |
| `threexui_xray_routing` | resource_xray_settings.go + xray_routing_schema.go | Routing (set path). Typed blocks |
| `threexui_xray_balancers` | resource_xray_settings.go + xray_balancers_schema.go | Balancers (set path). Typed blocks |
| `threexui_xray_reverse` | resource_xray_settings.go + xray_reverse_schema.go | Reverse proxy (set path). Typed blocks |
| `threexui_xray_outbounds` | resource_xray_settings.go + xray_outbounds_schema.go | Outbounds (set path). Typed blocks |
| `threexui_xray_version` | resource_xray_version.go | Manage installed Xray core version. Singleton (ID = "xray_version") |

## Data Sources

| Terraform Data Source | Description |
| --- | --- |
| `threexui_inbounds` | List of all inbounds (JSON string, Sensitive) |
| `threexui_server_status` | Server status (JSON) |
| `threexui_xray_versions` | Available Xray versions (list of strings) |
| `threexui_xray_config` | Current Xray config (JSON, Sensitive) |
| `threexui_settings` | All panel settings (JSON, Sensitive) |
| `threexui_online_clients` | List of currently online client emails |
| `threexui_client_traffics` | Client traffic statistics by email |

> **Security note:** any data source that returns a raw JSON payload from the
> panel/Xray API (e.g. `inbounds`, `settings`, `xray_config`) MUST mark the JSON
> attribute `Sensitive: true`. The payloads contain client UUIDs, passwords,
> Reality `privateKey`, WireGuard `secretKey`, Telegram bot tokens, LDAP
> passwords. Comparable resource fields (`resource_inbound_client.go`,
> `resource_settings_tabs.go`, `xray_outbounds_schema.go`) already use
> `Sensitive: true`; the data source schema must mirror that.

## Build, Test, and Development Commands

Use `task` for all routine work:

- `task build` builds the local provider binary.
- `task fmt` runs `gofmt -w provider/*.go`.
- `task vet` runs `go vet ./...`.
- `task lint` runs `golangci-lint run`.
- `task test:unit` runs unit tests only.
- `task test:acc` starts Docker Compose and runs Terraform acceptance tests.
- `task test:acc:compat` runs acceptance tests against a selectable 3x-ui version via `THREEXUI_VERSION` (defaults to `v3.0.2`).
- `task test` runs both unit and acceptance suites.
- `task pre-commit` runs the Go pre-commit checks: fmt, vet, lint, and build. The actual `.pre-commit-config.yaml` also includes markdown and common file checks.

Other useful commands:

```bash
markdownlint-cli2     # Lint markdown (uses .markdownlint-cli2.yaml; localized READMEs are excluded by glob)

# Run a single test by name:
TF_ACC=1 THREEXUI_ENDPOINT=http://localhost:2053 THREEXUI_USERNAME=admin THREEXUI_PASSWORD=admin \
  go test ./provider -run TestAccInboundVLESS -count=1 -timeout 600s -v
```

## Coding Style & Naming Conventions

Follow standard Go formatting and keep code `gofmt`-clean. Use tabs for
indentation, exported identifiers in `CamelCase`, and Terraform schema field
names in `snake_case`. Match existing file patterns: new resources go in
`provider/resource_<name>.go`, data sources in `provider/data_source_<name>.go`,
and schema helpers in `provider/<area>_schema.go`. Keep changes focused; avoid
unrelated renames or broad reformatting.

## Testing Guidelines

Write table-driven unit tests where practical and keep them next to the code
they cover. Name tests `TestXxx`; reserve `TestAccXxx` for acceptance coverage.
Run `task test:unit` before every PR. Run `task test:acc` for API, state,
schema, or Docker-related changes; it expects Docker and Terraform locally and
uses the bundled `admin/admin` 3x-ui setup on `http://localhost:2053`.

## Commit & Pull Request Guidelines

History follows Conventional Commits: `feat:`, `fix:`, `docs:`, `ci:`, `test:`,
`chore:`. Keep subjects imperative and concise, for example
`fix: normalize inbound listen defaults`. PRs should include a short
problem/solution summary, linked issue when relevant, and the commands you ran.
Update `docs/` and `examples/` when resource behavior or schema changes.

## Documentation Conventions

- **README is localized** in 6 languages mirroring 3x-ui upstream (en, ru, fa, ar, zh, es). When changing user-facing copy in `README.md`, update all five `README.<locale>.md` files in the same PR. Keep the language-switcher line at the top of every file identical. Persian and Arabic READMEs wrap their body in `<div dir="rtl">`.
- **`docs/guides/`** holds operational walkthroughs (backup-as-code, server-migration, bulk-clients). Add a guide here when introducing a workflow that needs more than an `examples/` folder. Front matter (`page_title`, `subcategory: "Guides"`, `description`) is required for Terraform Registry rendering.
- **`SECURITY.md`** has a per-surface table of sensitive fields handled by the provider. When adding a new resource that handles secrets (passwords, private keys, tokens, UUIDs), add a row to that table; it is the canonical list referenced by the README and by the data-source security note above.

## 3x-ui API (Key Endpoints)

- `GET /csrf-token` - anonymous/session CSRF bootstrap for 3x-ui v3 unsafe requests
- `POST /login` - authentication (form: username, password, twoFactorCode, optional `_csrf`; v3 requires `X-CSRF-Token`)
- `GET /panel/csrf-token` - authenticated CSRF refresh for 3x-ui v3 SPA/API POSTs
- `GET /panel/api/inbounds/list` - all inbounds
- `GET /panel/api/inbounds/get/:id` - single inbound
- `POST /panel/api/inbounds/add` - create (form-encoded)
- `POST /panel/api/inbounds/update/:id` - update
- `POST /panel/api/inbounds/del/:id` - delete
- `POST /panel/api/inbounds/addClient` - add client
- `POST /panel/api/inbounds/updateClient/:clientId` - update client
- `POST /panel/api/inbounds/:id/delClient/:clientId` - delete client
- `GET /panel/api/nodes/list` + `/panel/api/nodes/*` - v3 multi-node API surface
- `POST /panel/setting/all` - all settings
- `POST /panel/setting/update` - update settings (JSON body)
- `GET /panel/setting/getApiToken` / `POST /panel/setting/regenerateApiToken` - v3 panel API token
- `POST /panel/setting/updateUser` - change admin credentials (JSON: oldUsername, oldPassword, newUsername, newPassword)
- `POST /panel/xray` - Xray template (xraySetting)
- `POST /panel/xray/update` - update Xray template

Unauthenticated `/panel/api` requests return 404 (not 401). The client performs
auto re-login on 401/404. 3x-ui v3 rejects unsafe methods with 403 when CSRF is
missing/stale; the client bootstraps `/csrf-token`, sends `X-CSRF-Token`, and
refreshes via `/panel/csrf-token` before retrying once.

### Provider authentication and bootstrap credentials

- Provider config defaults `username`/`password` to `admin`/`admin` when omitted.
- `bootstrap_username` and `bootstrap_password` are explicit opt-in credentials for first-run panel bootstrap. They must be configured together and non-empty; `bootstrap_password` is `Sensitive` and listed in `SECURITY.md`.
- `Client.LoginWithBootstrapCredentials` uses anonymous `GET /csrf-token` as the panel-generation signal. If a token is returned (3x-ui v3.x), the client tries steady-state `username`/`password` first and falls back to bootstrap credentials only on a panel login rejection. If no token is returned (3x-ui v2.9.x), the client tries bootstrap credentials first to avoid submitting and exposing the desired steady-state password through failed-login logs or Telegram login notifications.
- Bootstrap fallback is only for ordinary login rejection (`loginFailedError`). HTTP status errors, network errors, and CSRF bootstrap errors surface immediately and must not trigger a credential fallback.
- Intended workflow: configure provider `username`/`password` with the desired steady-state credentials, add `bootstrap_username`/`bootstrap_password` for the initial panel credentials, and manage `threexui_panel_user` in the same apply so the panel rotates to steady-state credentials. Coverage: `TestLoginWithBootstrapCredentials*`, `TestProviderSensitiveAttributes`, `TestAccPanelUserBootstrapCredentials`.

### Retry on transient 5xx (write endpoints)

- `Client.withRetry` - single retry policy. Wraps a request function with up to `maxRetries` additional attempts on `*HTTPStatusError` with code 5xx, fixed 500ms backoff, ctx-aware.
- `doFormRetryable` / `doJSONRetryable` - thin wrappers over `withRetry` that delegate to `doForm`/`doJSON`.
- Applied **only** to idempotent writes: `UpdateInbound`, `UpdateInboundClient`, `UpdateSettings`, `UpdateXrayTemplate`, `SetXrayOutboundTestURL`.
- **Not** applied via `withRetry` to: `AddInbound`, `AddInboundClient` (would duplicate), `UpdateUser` (stale creds).
- `DeleteInbound` has a custom retry-with-verify path (not `withRetry`): on 5xx it calls `inboundAbsent` (one `GetInbounds`). If the row is gone, returns success because the panel handler panicked after the SQLite delete had already committed; if the row is still present, retries the DELETE once. 4xx surfaces immediately. Rationale: `DelInbound` reads-then-deletes and errors on a missing row, so a naive `withRetry` would turn a successful-but-5xx delete into a failure (#161). Tests: `TestDeleteInboundReturnsSuccessIfRowAbsentAfter5xx`, `TestDeleteInboundRetriesOnce5xxIfRowStillPresent`, `TestDeleteInboundDoesNotRetryOn4xx`.
- `tflog.Warn` emitted on each retry with `operation`, `attempt`, `status_code`, `backoff`; operators can detect upstream flakiness instead of silent absorption.
- Configurable via `max_retries` provider attribute (default `1`, set to `0` to disable). Provider plumbs default into `ClientConfig.MaxRetries`; `Client.maxRetries` is the field used by `withRetry`.
- Composes with the 401/404 auto-relogin in `doRequest`: relogin happens inside a single `withRetry` attempt; only an HTTP 5xx surfaced from `decodeAPIResponse` triggers the outer retry.
- `HTTPStatusError` - error type returned by `decodeAPIResponse` when `resp.StatusCode >= 400` (both empty-body and non-JSON paths). `errors.As` is the supported way to inspect status.

### Retry on upstream GitHub API rate limit (GetXrayVersions)

Distinct from the 5xx retry above. The 3x-ui panel calls `api.github.com` for
Xray release metadata without authentication. On shared IPs the unauthenticated
rate limit (60 req/hr) is exhausted quickly, and the panel returns `success:false`
with a message containing "rate limit". This comes back as HTTP 200 (not 5xx),
so `withRetry` does not cover it.

- `isUpstreamRateLimitError(err)` - detects "rate limit" (case-insensitive) in the error message from `decodeAPIResponse`.
- `GetXrayVersions` retries up to `versionRetryAttempts` (4) times with exponential backoff (`versionRetryBaseBackoff` = 2s, doubling: 2s, 4s, 8s, 16s; ~30s worst case). Only retries on rate-limit errors; other errors surface immediately.
- This is a **production fix**, not just CI: any provider user behind a shared IP with multiple 3x-ui instances benefits.
- The 3x-ui `getXrayVersion` handler caches the GitHub API response for 15 minutes. Once the cache is warm, subsequent calls do not hit GitHub at all. The retry is primarily for the cold-start window.

### Read-after-write retry (post-write reads)

Distinct from the 5xx retry above. 3x-ui occasionally returns `success: true`
from a create/update endpoint while the underlying SQLite commit is not yet
visible to a follow-up GET (#157). The 5xx retry does not help because the
response is HTTP 200, just empty/missing the row. Use a separate
application-layer policy:

- `Client.WithReadAfterWriteRetry` - polls a caller-provided `func() (found bool, err error)` up to `readAfterWriteAttempts` (20) times with `readAfterWriteBackoff` (500ms) between attempts. A non-nil err aborts immediately; read failures are not retried, only the "not visible yet" condition is. Emits `tflog.Warn` per retry with `operation`, `attempt`, `max_attempts`, `backoff`.
- Applied to: `InboundResource.Create` (resolves the new row by `port` if `AddInbound` returned an empty obj), `InboundClientResource.Create`/`Update` (waits for the new client to appear in the inbound's settings JSON), `XrayVersionResource.waitForXrayVersion` (ignores `ErrXrayVersionUnknown` while xray is restarting).
- **Not** applied to plain `Read`; for an idle read, "row not present" is meaningful (resource was deleted out-of-band) and must be reported to Terraform immediately rather than retried.
- Test helpers `testAccCheckInboundDestroyed` / `testAccCheckInboundClientDestroyed` use a similar bounded poll (`destroyVisibilityAttempts x destroyVisibilityBackoff` = 60 x 500ms = 30s) for the inverse case: waiting for a successful DELETE to become invisible to a follow-up GET. Resource-side counterpart: `InboundResource.waitForInboundDeletion` (20 x 500ms = 10s) emits a Warning, not an Error, on exhaustion; the API has already accepted the DELETE, so leaving the resource in TF state would be the worse failure mode (#136, #161).

## Key Code Details

### Framework (terraform-plugin-framework)

- Provider: `ThreeXUIProvider` implements `provider.Provider` (Metadata, Schema, Configure, Resources, DataSources).
- Factory: `New() provider.Provider`.
- Resources implement `resource.Resource` + `resource.ResourceWithImportState`.
- Data sources implement `datasource.DataSource`.
- Models use `types.String`, `types.Int64`, `types.Bool` with `tfsdk:"..."` tags.
- Plan modifiers: `stringplanmodifier.RequiresReplace()`, `int64planmodifier.RequiresReplace()`.
- Defaults: `booldefault.StaticBool()`, `stringdefault.StaticString()`.
- Import: `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`.

### Inbound / Client

- `settings`, `stream_settings`, `sniffing` are JSON strings in the API and typed blocks in TF schema.
- Three-layer conversion: typed model <-> untyped map (expand/flatten*FromModel/*ToModel) <-> JSON string (build*/flatten*).
- Per-protocol settings blocks: `vless_settings`, `trojan_settings`, `shadowsocks_settings`, `http_settings`, `socks_settings`, `mixed_settings`, `wireguard_settings`, `dokodemo_settings`, `hysteria_settings`.
- `stream_settings` supports transports: tcp, ws, grpc, httpupgrade, xhttp, kcp, hysteria + reality, sockopt, external_proxy.
- Sniffing supports `ips_excluded` and `domains_excluded` fields (added in 3x-ui 2.9.0).
- KCP: `congestion`, `read_buffer_size`, `write_buffer_size` replaced by `cwnd_multiplier`, `max_sending_window` (breaking, 2.9.0).
- WireGuard: `mtu` changed from int to list [v4, v6]; added `gateway` and `dns` list fields (breaking, 2.9.0).
- Hysteria: `auth` field on `threexui_inbound_client` used as client identifier (instead of UUID-based `id`).
- All `Optional+Computed` inbound attributes have `UseStateForUnknown` plan modifiers to prevent false drift (`known after apply`) after import.
- `alignBlocksWithPlan` prevents "was absent, but now present" errors for Optional blocks (Create/Read/Update); skipped during Import (detect: `state.Protocol.IsNull()`).
- `reality_settings.settings` is a `SingleNestedAttribute` (not block) with `objectplanmodifier.UseStateForUnknown()`; preserves auto-generated values (public_key, fingerprint, etc.) from state when user omits the attribute.
- `preserveInboundSettings` preserves clients and testseed from existing inbound on update.
- `ensureRealityKeys` auto-generates private/public key and short_ids.
- `ensureInboundClientIDs` auto-generates UUID for clients without id.
- `applyDefaultInboundSettings` applies default settings per protocol (vless: decryption=none, testseed).
- `inboundClientMu` is the mutex for concurrent client operations.
- `email` in `threexui_inbound_client` is **Required**; without email, 3x-ui crashes with SQL error when adding the next client.
- `isSubset` is a standalone utility for JSON subset checking.

### Panel Settings

- Settings resources are singletons (ID = `"settings"`), one instance per type.
- Typed attributes (Optional + Computed + UseStateForUnknown) - each field is a separate attribute in the schema.
- Per-resource models: `PanelGeneralModel`, `PanelSecurityModel`, `PanelTelegramModel`, `PanelSubscriptionModel`.
- `settingsApplyTyped` / `settingsReadTyped` - shared CRUD logic (expand model -> API -> flatten -> model).
- Delete only clears TF state, does **not** reset settings in the API.
- Subscription resource performs double apply (workaround for 3x-ui bug: sub_json_enable not saved on first apply together with sub_enable); includes Clash/Mihomo fields: `sub_clash_enable`, `sub_clash_path`, `sub_clash_uri` (added in 2.9.0).
- Enabling 2FA has a Warning (partial support: TOTP code sent on initial login, but auto re-login fails when code expires).
- Changing `web_base_path` requires updating `base_path` in provider config; Warning added.
- `panelSettingsNeedRestart` keys: webListen, webDomain, webPort, webBasePath, webCertFile, webKeyFile, sessionMaxAge.

### Panel User

- `threexui_panel_user` is a singleton (ID = `"user"`), manages admin credentials.
- Write-only: no API for reading username/password, Read is a no-op (state preserved).
- Create uses `r.client.username/password` as old credentials. In the first-run bootstrap workflow, the provider may be authenticated with bootstrap credentials for the create, then this resource rotates the panel to the steady-state provider `username`/`password`.
- Update uses previous state as old credentials.
- After successful `UpdateUser`, client updates its stored credentials for subsequent requests.
- Delete only clears TF state; credentials on the panel are not reverted.
- Warning reminds to update provider config after changing credentials.

### Xray Settings

- Typed blocks (`ListNestedBlock`) - each resource has its own model and schema in `*_schema.go`.
- Per-resource models: `XrayBasicsModel`, `XrayDNSModel`, `XrayRoutingModel`, `XrayBalancersModel`, `XrayReverseModel`, `XrayOutboundsModel`.
- Two-layer conversion: typed model <-> untyped map (expand/flatten) <-> Xray JSON (build/flattenToMap).
- Xray resources work in 2 modes: merge root (`xray_basics`), set path (others).
- `xrayTemplateMu` serializes read-modify-write on xray template and prevents race conditions.
- `xrayApplyTyped` / `xrayReadSection` are shared CRUD logic.
- CRUD: plan.Get -> expand -> build -> xrayApplyTyped -> xrayReadSection -> flattenToMap -> flatten -> state.Set.
- DNS servers: address-only serializes as string in JSON; with extra fields serializes as object.
- Outbound settings: per-protocol blocks (`freedom_settings`, `blackhole_settings`, etc.) determined by `protocol` value; `freedom_settings` includes legacy `ips_blocked` (2.9.0) and `final_rule` (2.9.4+), `vless_settings` includes `reverse_tag` (2.9.4+).
- Policy levels: in Xray JSON map `{"0": {...}}`, in TF list `[{id=0, ...}]`.
- Delete for xray resources only clears TF state and does not reset the xray config.

## Pre-commit Hooks

Automatic pre-commit checks are configured:

- **go-fmt** - code formatting
- **go-vet** - static analysis
- **golangci-lint** - linter
- **go-build** - compilation check
- **markdownlint** - markdown linting (requires `markdownlint-cli2`)
- YAML/JSON checks, trailing whitespace, EOF

Acceptance tests are **not** run in pre-commit; use `task test:acc` explicitly.

Configuration files: `.pre-commit-config.yaml`, `.golangci.yml`,
`.markdownlint-cli2.yaml`.

`.markdownlint-cli2.yaml` notes:

- Glob list intentionally covers only `README.md`, `CONTRIBUTING.md`, `docs/**/*.md`. Localized READMEs (`README.<locale>.md`) are not linted because they mirror `README.md` structurally and a single lint pass is the source of truth.
- `first-line-heading: false` is set because every README starts with the language-switcher line, not a heading.

## Supply-Chain Pinning

All third-party code that runs in CI or pre-commit is pinned to a commit SHA,
not a floating tag. Tags are mutable: a maintainer can re-tag to point at
different code without any diff in workflow files. SHA pins make every
dependency change a reviewable diff.

**GitHub Actions** (`.github/workflows/*.yml`) - every `uses:` reference is
pinned as `<owner>/<repo>@<sha> # vN`. The trailing `# vN` comment is the format
Dependabot recognizes for major-version tracking; weekly Dependabot PRs bump
the SHA when the action publishes a new vN.x release. Same rule applies to
first-party `actions/*` for consistency.

**Pre-commit hooks** (`.pre-commit-config.yaml`) - external `repo:` references
use `rev: <sha>  # frozen: <tag>`. This is the format
`pre-commit autoupdate --freeze` produces. Bare `pre-commit autoupdate` will
un-pin; always use the `--freeze` flag locally, or update the SHA manually.

**Docker images** in `docker-compose.yaml` are intentionally NOT digest-pinned.
They run only as ephemeral test environments (3x-ui panel for acceptance tests),
never in published artifacts, and have no access to release secrets. The
residual risk is a tampered test signal, not a poisoned release. Maintaining
digests for all matrix versions manually outweighs the closed-off risk (#168).

**Go modules** are covered by `go.sum` hashing; no extra pinning needed. The
`go install ...@vX.Y.Z` references in workflows install specific versions whose
contents are verified by Go's module proxy.

When adding a new third-party action or pre-commit hook, resolve the SHA via:

```bash
gh api repos/<owner>/<repo>/git/refs/tags/<tag> --jq '.object.sha,.object.type'
# if type is "tag" (annotated), dereference:
gh api repos/<owner>/<repo>/git/tags/<sha> --jq '.object.sha'
```

## Test Environment

```bash
task test              # Full cycle: docker up, acc tests (Terraform), docker down
docker compose up -d   # Start 3x-ui on localhost:2053
# Login: admin / admin
# Docker image defaults to webBasePath = / (NOT /panel/)
# Do not set THREEXUI_BASE_PATH

# Run all tests with version-aware skipping:
THREEXUI_VERSION=v2.9.0 task test:acc:compat

# Run all versions locally:
for v in v2.9.0 v2.9.1 v2.9.2 v2.9.3 v2.9.4 v3.0.0 v3.0.1 v3.0.2; do
  echo "=== Testing $v ===" && THREEXUI_VERSION=$v task test:acc:compat
done
```

### Support Policy

The provider officially supports the **two latest 3x-ui minor lines**.
Currently that is **2.9.x** and **3.0.x**: every released patch in both lines
is in the CI `acceptance-matrix` and listed as `Tested` in the README
compatibility table.

When a new minor (e.g. `3.1.0`) is released:

1. Add the new minor's patches to `.github/workflows/ci.yml` `acceptance-matrix` and to the README compatibility tables (all 6 localized files).
2. **Drop the oldest supported line entirely** (matrix + README) so we keep exactly two minor lines.
3. Drop any `requireMinVersion(t, "v<dropped-line>...")` skip gates whose floor is no longer reachable, and prune the corresponding entry from the version mapping below.
4. Update the support-policy paragraph in all six READMEs to reflect the new pair.

Keep the policy line in README and the matrix entries in lockstep: the README
claim "Tested" must be backed by an actual CI matrix entry.

### Version-Aware Test Skipping

Tests that use features introduced in specific 3x-ui versions call
`requireMinVersion(t, "vX.Y.Z")` at the start. When `THREEXUI_VERSION` env var
is set (by `task test:acc:compat` or CI matrix), tests requiring a newer version
are automatically skipped via `t.Skip()`.

Version mapping:

- **v2.9.0+**: mixed protocol, WireGuard mtu as list/gateway/dns, sniffing `ips_excluded`/`domains_excluded`
- **v3.0.0+**: CSRF-protected unsafe requests, inbound `nodeId`, multi-node/API-token upstream surface

Tests without `requireMinVersion` run on all supported versions (v2.9.0+).
Helper: `provider/test_helpers.go` (`requireMinVersion` uses
`golang.org/x/mod/semver`).

Acceptance tests use `terraform-plugin-testing`:

- `testAccProtoV6ProviderFactories()` returns `map[string]func() (tfprotov6.ProviderServer, error)`.
- `ProtoV6ProviderFactories` in TestCase (not `ProviderFactories`).
- HCL configs use typed blocks and attributes (not `jsonencode()`).

Acceptance tests require Terraform and environment variables for correct
provider namespace:

- `TF_ACC_TERRAFORM_PATH` - absolute path to `terraform`
- `TF_ACC_PROVIDER_NAMESPACE=batonogov`
- `TF_ACC_PROVIDER_HOST=registry.terraform.io`

All of this is already configured in `Taskfile.yml` -> `task test`.

### Readiness Contract (acceptance suite)

The acceptance suite assumes 3x-ui is ready in **three stages**, all gated before
any test runs:

1. **Panel router up** - `docker-compose.yaml` declares a healthcheck that polls `/`. `docker compose up --wait` blocks until the healthcheck passes (max ~30s). Without this, `--wait` only waits for "container started", which is earlier than the gin router being ready. Do not poll `GET /login`: in v2 it is POST-only and in v3 login POST is CSRF-protected.
2. **Xray subsystem initialized** - `Taskfile.yml` `_wait-for-xray` runs after `compose up --wait` and polls `/panel/api/server/status` until `xray.state == "running"` (max 30s). Without this, tests like `TestAccXrayVersionDrift` start before xray reports its version and fail with bogus `ErrXrayVersionUnknown` (#161). Do NOT use `/panel/api/server/getXrayVersion` for this: that endpoint fetches the GitHub release list anonymously and intermittently rate-limits on shared CI runner IPs.
3. **Xray version cache pre-warmed** - `Taskfile.yml` `_warm-xray-version-cache` runs after `_wait-for-xray` and calls `/panel/api/server/getXrayVersion` with retries (5 attempts, exponential backoff 1s–16s). Once the 3x-ui internal cache is warm (15-min TTL), subsequent test calls never hit GitHub. Non-fatal: logs a warning on failure and proceeds (#184).

When adding new tests that touch xray-only state (templates, versions,
restart-required settings), assume all three gates have passed. Do NOT add per-test
sleeps.

### CI Flake Mitigation

Beyond the in-process retry budgets (`withRetry`, `WithReadAfterWriteRetry`,
`waitForInboundDeletion`, `destroyVisibilityAttempts`, `GetXrayVersions` rate-limit
retry), CI itself has three safety nets:

- **Per-job retry** - `acceptance-tests` and `acceptance-matrix` jobs in `.github/workflows/ci.yml` use `nick-fields/retry@v4` with `max_attempts: 2`. Catches the residual flake rate from GHCR pull jitter, one-off SQLite spikes, and runner contention. A green retry should be a no-op for code; if a retry consistently changes behavior, that is a real bug. Diff the two attempt logs.
- **Flaky test gate** - `skipIfFlaky(t, reason)` in `provider/test_helpers.go` skips when `THREEXUI_SKIP_FLAKY` env is set. Sub-day mitigation when a test starts firing falsely: gate it, push, file a follow-up. Quarantined tests must be tracked (#161 or follow-up); the gate is not a permanent home.
- **GitHub API rate-limit mitigation** - three layers addressing 3x-ui's unauthenticated GitHub API calls (#184): (1) provider-level `GetXrayVersions` retry with exponential backoff on rate-limit errors, (2) `_warm-xray-version-cache` Taskfile task pre-populates 3x-ui's 15-minute internal cache before tests, (3) `GITHUB_TOKEN` passed to the container via `docker-compose.yaml` env var for forward-compatibility with future 3x-ui versions that may use it for authenticated API calls.

### Docs/Schema Drift Check (tfplugindocs validate)

The `tfplugindocs-validate` job in `.github/workflows/docs.yml` runs
`tfplugindocs validate --provider-name threexui` to ensure every resource and
data source registered in the provider schema has a corresponding doc file
under `docs/resources/` or `docs/data-sources/`. This catches schema/docs
drift: a new resource without a doc file, a renamed resource whose doc was not
moved, or a stale doc for a removed resource.

Path triggers include `docs/**`, `examples/**`, `provider/**/*.go`, `go.mod`,
and `go.sum` so that both doc-only changes and schema-affecting code changes
are validated. When adding a new resource or data source, include the
corresponding doc file in the same PR; otherwise this job will fail.

## Releases

Flow: Conventional Commits -> Release Please -> GoReleaser -> Terraform
Registry.

1. Commits to `main` with prefixes `feat:`, `fix:`, `feat!:`, etc.
2. Release Please automatically creates/updates a Release PR (version + changelog).
3. Merging the Release PR creates a `v*` tag, then GoReleaser runs in the same workflow.
4. GoReleaser builds binaries, signs with GPG, and publishes a GitHub Release.
5. Terraform Registry picks up the release.

Note: GoReleaser runs as a dependent job inside `release-please.yml` (not a
separate workflow), because tags created by `GITHUB_TOKEN` do not trigger other
workflows.

Commits accumulate in the Release PR until merged; release only happens on PR
merge.

## Updating 3x-ui Version

When a new 3x-ui version is released:

1. **Save source snapshots** - download and extract the source into `3x-ui-<version>/` directory:

   ```bash
   curl -sL https://github.com/MHSanaei/3x-ui/archive/refs/tags/v<VERSION>.tar.gz | tar xz
   # Archive extracts as 3x-ui-<VERSION> (without the 'v' prefix); used as-is
   ```

2. **Diff sources** - compare with previous version: `diff -rq 3x-ui-<old> 3x-ui-<new> --exclude='.git'`, then inspect key files (API endpoints, models, services).
3. **Assess impact** - determine which changes affect the provider's API surface (new fields, changed formats, renamed endpoints).
4. **Update docker-compose.yaml** - bump the image tag to the new version.
5. **Run tests** - `task test` (full cycle: docker up, acceptance tests, docker down).
6. **Adapt provider** - if API changes require it, update provider code, run `task build`, then `task test` again.

## Development Workflow

Standard flow for working on issues:

1. **Issue** - pick an issue from `gh issue list`.
2. **Code** - implement the fix/feature, run `task build`.
3. **PR** - create a branch, commit, push, open a PR via `gh pr create`.
4. **CI** - wait for the CI pipeline to pass (`gh pr checks <number> --watch`). If it fails, investigate logs (`gh run view <run_id> --log-failed`), fix, and push again.
5. **Codex review** - run `codex review --base main` and address all findings.
6. **Iterate** - repeat steps 2-5 until CI is green and codex review has zero remarks.
7. **Done** - PR is ready for merge.

## Core Principles

- **Always check 3x-ui sources** before making assumptions about API behavior. Source snapshots are in `3x-ui-<version>/` directories. Download new versions with `curl -sL https://github.com/MHSanaei/3x-ui/archive/refs/tags/v<VERSION>.tar.gz | tar xz`. Key files: `web/service/` (business logic), `web/controller/` (API endpoints), `web/entity/model/` (data models), `xray/` (xray config).
- Be pragmatic: understand the task first, then make the minimum necessary changes.
- Do not break backward compatibility without an explicit request.
- Preserve code style and project structure.
- Make targeted changes, avoid mass reformatting.
- Run `task build` after code changes.
- Be concise and to the point. Indicate which files were changed.
- When changing anything documented in `AGENTS.md` (workflows, structure, conventions), update `AGENTS.md` in the same commit/PR.
