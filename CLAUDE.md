# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

Terraform provider for the [3x-ui](https://github.com/MHSanaei/3x-ui) panel (Go, terraform-plugin-framework).

## Project Structure

```text
provider/              — all provider code
  provider.go          — ThreeXUIProvider (framework): Metadata, Schema, Configure, Resources, DataSources
  client.go            — HTTP client for 3x-ui API (cookie auth, auto re-login)
  types.go             — Inbound, ClientTraffic, APIResponse, ParseJSONField
  resource_inbound.go  — threexui_inbound resource (CRUD, Reality, settings defaults)
  resource_inbound_client.go — threexui_inbound_client resource (mutex, UUID)
  resource_settings_tabs.go  — panel_general/security/telegram/subscription (typed attributes)
  resource_panel_user.go     — threexui_panel_user resource (admin credentials change)
  resource_xray_settings.go  — CRUD for xray_basics/dns/routing/balancers/reverse/outbounds (typed attributes)
  xray_basics_schema.go      — model, schema, expand/flatten for xray_basics (log, policy, api, stats)
  xray_dns_schema.go         — model, schema, expand/flatten for xray_dns (servers, hosts)
  xray_routing_schema.go     — model, schema, expand/flatten for xray_routing (rules)
  xray_balancers_schema.go   — model, schema, expand/flatten for xray_balancers
  xray_reverse_schema.go     — model, schema, expand/flatten for xray_reverse (bridges, portals)
  xray_outbounds_schema.go   — model, schema, expand/flatten for xray_outbounds (per-protocol settings)
  inbound_settings_schema.go      — model, schema, expand/flatten for per-protocol settings (vless, trojan, ss, http, socks, mixed, wg, dokodemo, hysteria)
  inbound_stream_settings_schema.go — model, schema, expand/flatten for stream_settings (tcp, ws, grpc, httpupgrade, xhttp, kcp, hysteria, reality, sockopt)
  inbound_sniffing_schema.go      — model, schema, expand/flatten for sniffing
  settings.go          — buildSettingsJSON(map[string]any), flattenSettings(string), expand/flatten clients/fallbacks/peers
  stream_settings.go   — buildStreamSettingsJSON(map[string]any), flattenStreamSettings(string), expand/flatten per-transport
  sniffing.go          — buildSniffingJSON(map[string]any), flattenSniffing(string)
  settings_helpers.go  — mergeSettings
  list_helpers.go      — typesListToAnySlice, typesListInt64ToAnySlice, anySliceToTypesList
  default_settings.go  — default settings per protocol, applyDefaultInboundSettings
  resource_xray_version.go     — threexui_xray_version resource (install/manage Xray core version)
  data_source_*.go     — data sources (inbounds, server_status, settings, xray_config, xray_versions, online_clients)
examples/              — example TF configs for manual testing
3x-ui-<version>/      — 3x-ui source snapshots (in .gitignore, for reference/diffing)
docker-compose.yaml    — 3x-ui on port 2053 (version via THREEXUI_VERSION env, default v2.9.2)
Taskfile.yml           — task build / test / fmt
.github/workflows/
  ci.yml               — lint, unit tests, acceptance tests, compatibility matrix (PR + push main)
  docs.yml             — docs/examples validation: terraform fmt, markdownlint, yamllint (PR + push main)
  release-please.yml   — Release Please + GoReleaser (conventional commits → semver tag → build + sign + publish)
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
| `threexui_inbounds` | List of all inbounds (JSON string) |
| `threexui_server_status` | Server status (JSON) |
| `threexui_xray_versions` | Available Xray versions (list of strings) |
| `threexui_xray_config` | Current Xray config (JSON) |
| `threexui_settings` | All panel settings (JSON) |
| `threexui_online_clients` | List of currently online client emails |
| `threexui_client_traffics` | Client traffic statistics by email |

## 3x-ui API (Key Endpoints)

- `POST /login` — authentication (form: username, password, twoFactorCode)
- `GET /panel/api/inbounds/list` — all inbounds
- `GET /panel/api/inbounds/get/:id` — single inbound
- `POST /panel/api/inbounds/add` — create (form-encoded)
- `POST /panel/api/inbounds/update/:id` — update
- `POST /panel/api/inbounds/del/:id` — delete
- `POST /panel/api/inbounds/addClient` — add client
- `POST /panel/api/inbounds/updateClient/:clientId` — update client
- `POST /panel/api/inbounds/:id/delClient/:clientId` — delete client
- `POST /panel/setting/all` — all settings
- `POST /panel/setting/update` — update settings (JSON body)
- `POST /panel/setting/updateUser` — change admin credentials (JSON: oldUsername, oldPassword, newUsername, newPassword)
- `POST /panel/xray` — Xray template (xraySetting)
- `POST /panel/xray/update` — update Xray template

Unauthenticated requests return 404 (not 401). The client performs auto re-login on 401/404.

## Key Code Details

### Framework (terraform-plugin-framework)

- Provider: `ThreeXUIProvider` implements `provider.Provider` (Metadata, Schema, Configure, Resources, DataSources)
- Factory: `New() provider.Provider`
- Resources implement `resource.Resource` + `resource.ResourceWithImportState`
- Data sources implement `datasource.DataSource`
- Models use `types.String`, `types.Int64`, `types.Bool` with `tfsdk:"..."` tags
- Plan modifiers: `stringplanmodifier.RequiresReplace()`, `int64planmodifier.RequiresReplace()`
- Defaults: `booldefault.StaticBool()`, `stringdefault.StaticString()`
- Import: `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`

### Inbound / Client

- `settings`, `stream_settings`, `sniffing` — JSON strings in the API, typed blocks in TF schema
- Three-layer conversion: Typed Model ↔ Untyped Map (expand/flatten*FromModel/*ToModel) ↔ JSON String (build*/flatten*)
- Per-protocol settings blocks: `vless_settings`, `trojan_settings`, `shadowsocks_settings`, `http_settings`, `socks_settings`, `mixed_settings`, `wireguard_settings`, `dokodemo_settings`, `hysteria_settings`
- stream_settings supports transports: tcp, ws, grpc, httpupgrade, xhttp, kcp, hysteria + reality, sockopt, external_proxy
- Sniffing supports `ips_excluded` and `domains_excluded` fields (added in 3x-ui 2.9.0)
- KCP: `congestion`, `read_buffer_size`, `write_buffer_size` replaced by `cwnd_multiplier`, `max_sending_window` (breaking, 2.9.0)
- WireGuard: `mtu` changed from int to list [v4, v6]; added `gateway` and `dns` list fields (breaking, 2.9.0)
- Hysteria: `auth` field on `threexui_inbound_client` used as client identifier (instead of UUID-based `id`)
- All `Optional+Computed` inbound attributes have `UseStateForUnknown` plan modifiers — prevents false drift (`known after apply`) after import
- `alignBlocksWithPlan` — prevents "was absent, but now present" errors for Optional blocks (Create/Read/Update); skipped during Import (detect: `state.Protocol.IsNull()`)
- `reality_settings.settings` — `SingleNestedAttribute` (not block) with `objectplanmodifier.UseStateForUnknown()`; preserves auto-generated values (public_key, fingerprint, etc.) from state when user omits the attribute
- `preserveInboundSettings` — on update, preserves clients and testseed from existing inbound
- `ensureRealityKeys` — auto-generates private/public key and short_ids
- `ensureInboundClientIDs` — auto-generates UUID for clients without id
- `applyDefaultInboundSettings` — default settings per protocol (vless: decryption=none, testseed)
- `inboundClientMu` — mutex for concurrent client operations
- `email` in `threexui_inbound_client` — **Required** (without email, 3x-ui crashes with SQL error when adding the next client)
- `isSubset` — standalone utility for JSON subset checking

### Panel Settings

- Settings resources are singletons (ID = `"settings"`), one instance per type
- Typed attributes (Optional + Computed + UseStateForUnknown) — each field is a separate attribute in the schema
- Per-resource models: `PanelGeneralModel`, `PanelSecurityModel`, `PanelTelegramModel`, `PanelSubscriptionModel`
- `settingsApplyTyped` / `settingsReadTyped` — shared CRUD logic (expand model → API → flatten → model)
- Delete only clears TF state, does **not** reset settings in the API
- Subscription resource performs double apply (workaround for 3x-ui bug: sub_json_enable not saved on first apply together with sub_enable); includes Clash/Mihomo fields: `sub_clash_enable`, `sub_clash_path`, `sub_clash_uri` (added in 2.9.0)
- Enabling 2FA — Warning added (partial support: TOTP code sent on initial login, but auto re-login fails when code expires)
- Changing `web_base_path` requires updating `base_path` in provider config — Warning added
- `panelSettingsNeedRestart` — keys: webListen, webDomain, webPort, webBasePath, webCertFile, webKeyFile, sessionMaxAge

### Panel User

- `threexui_panel_user` — singleton (ID = `"user"`), manages admin credentials
- Write-only: no API for reading username/password, Read is a no-op (state preserved)
- Create uses `r.client.username/password` as old credentials
- Update uses previous state as old credentials
- After successful UpdateUser, client updates its stored credentials for subsequent requests
- Delete only clears TF state, credentials on the panel are not reverted
- Warning reminds to update provider config after changing credentials

### Xray Settings

- Typed blocks (ListNestedBlock) — each resource has its own model and schema in `*_schema.go`
- Per-resource models: `XrayBasicsModel`, `XrayDNSModel`, `XrayRoutingModel`, `XrayBalancersModel`, `XrayReverseModel`, `XrayOutboundsModel`
- Two-layer conversion: typed model ↔ untyped map (expand/flatten) ↔ Xray JSON (build/flattenToMap)
- Xray resources work in 2 modes: merge root (`xray_basics`), set path (others)
- `xrayTemplateMu` — mutex for serializing read-modify-write on xray template (prevents race condition)
- `xrayApplyTyped` / `xrayReadSection` — shared CRUD logic
- CRUD: plan.Get → expand → build → xrayApplyTyped → xrayReadSection → flattenToMap → flatten → state.Set
- DNS servers: address-only → serialized as string in JSON, with extra fields → as object
- Outbound settings: per-protocol blocks (`freedom_settings`, `blackhole_settings`, ...) determined by `protocol` value; `freedom_settings` includes `ips_blocked` (list of string, added in 2.9.0)
- Policy levels: in Xray JSON map `{"0": {...}}`, in TF list `[{id=0, ...}]`
- Delete for xray resources only clears TF state, does not reset the xray config

## Commands

```bash
task build            # Build binary
task test:unit        # Run unit tests (no Docker / Terraform needed)
task test:acc         # Run acceptance tests (requires Docker)
task test:acc:compat  # Run all tests with version-aware skipping (THREEXUI_VERSION, default v2.9.2)
task test             # Run unit + acceptance tests
task fmt              # gofmt
task vet              # go vet
task lint             # golangci-lint
task pre-commit       # Run all pre-commit checks manually (fmt, vet, lint, build)

# Run a single test by name:
TF_ACC=1 THREEXUI_ENDPOINT=http://localhost:2053 THREEXUI_USERNAME=admin THREEXUI_PASSWORD=admin \
  go test ./provider -run TestAccInboundVLESS -count=1 -timeout 600s -v
```

## Pre-commit Hooks

Automatic pre-commit checks are configured:

- **go-fmt** — code formatting
- **go-vet** — static analysis
- **golangci-lint** — linter
- **go-build** — compilation check
- **markdownlint** — markdown linting (requires `markdownlint-cli2`)
- YAML/JSON checks, trailing whitespace, EOF

Acceptance tests are **not** run in pre-commit — use `task test:acc` explicitly.

Configuration files: `.pre-commit-config.yaml`, `.golangci.yml`

## Test Environment

```bash
task test              # Full cycle: docker up, acc tests (Terraform), docker down
docker compose up -d   # Start 3x-ui on localhost:2053
# Login: admin / admin
# Docker image defaults to webBasePath = / (NOT /panel/)
# Do not set THREEXUI_BASE_PATH

# Run all tests with version-aware skipping:
THREEXUI_VERSION=v2.8.9 task test:acc:compat

# Run all versions locally:
for v in v2.8.9 v2.8.10 v2.8.11 v2.9.0; do
  echo "=== Testing $v ===" && THREEXUI_VERSION=$v task test:acc:compat
done
```

### Version-Aware Test Skipping

Tests that use features introduced in specific 3x-ui versions call `requireMinVersion(t, "vX.Y.Z")`
at the start. When `THREEXUI_VERSION` env var is set (by `task test:acc:compat` or CI matrix),
tests requiring a newer version are automatically skipped via `t.Skip()`.

Version mapping:

- **v2.9.0+**: mixed protocol, WireGuard mtu as list/gateway/dns, sniffing ips\_excluded/domains\_excluded
- **v2.8.11+**: tunnel protocol, DNS enable\_parallel\_query/use\_system\_hosts

Tests without `requireMinVersion` run on all supported versions (v2.8.9+).
Helper: `provider/test_helpers.go` (`requireMinVersion` uses `golang.org/x/mod/semver`).

Acceptance tests use `terraform-plugin-testing`:

- `testAccProtoV6ProviderFactories()` — returns `map[string]func() (tfprotov6.ProviderServer, error)`
- `ProtoV6ProviderFactories` in TestCase (not `ProviderFactories`)
- HCL configs use typed blocks and attributes (not `jsonencode()`)

Acceptance tests require Terraform and environment variables for correct provider namespace:

- `TF_ACC_TERRAFORM_PATH` — absolute path to `terraform`
- `TF_ACC_PROVIDER_NAMESPACE=batonogov`
- `TF_ACC_PROVIDER_HOST=registry.terraform.io`

All of this is already configured in `Taskfile.yml` → `task test`.

## Releases

Flow: Conventional Commits → Release Please → GoReleaser → Terraform Registry.

1. Commits to `main` with prefixes `feat:`, `fix:`, `feat!:`, etc.
2. Release Please automatically creates/updates a Release PR (version + changelog)
3. Merging the Release PR → tag `v*` is created → GoReleaser runs in the same workflow
4. GoReleaser builds binaries, signs with GPG, publishes GitHub Release
5. Terraform Registry picks up the release

Note: GoReleaser runs as a dependent job inside `release-please.yml` (not a separate workflow), because tags created by `GITHUB_TOKEN` do not trigger other workflows.

Commits accumulate in the Release PR until merged — release only happens on PR merge.

## Updating 3x-ui Version

When a new 3x-ui version is released:

1. **Save source snapshots** — download and extract the source into `3x-ui-<version>/` directory:

   ```bash
   curl -sL https://github.com/MHSanaei/3x-ui/archive/refs/tags/v<VERSION>.tar.gz | tar xz
   mv 3x-ui-<VERSION> 3x-ui-<VERSION>/  # rename if needed (archive extracts as 3x-ui-<VERSION>)
   ```

2. **Diff sources** — compare with previous version: `diff -rq 3x-ui-<old> 3x-ui-<new> --exclude='.git'`, then inspect key files (API endpoints, models, services)
3. **Assess impact** — determine which changes affect the provider's API surface (new fields, changed formats, renamed endpoints)
4. **Update docker-compose.yaml** — bump the image tag to the new version
5. **Run tests** — `task test` (full cycle: docker up, acceptance tests, docker down)
6. **Adapt provider** — if API changes require it, update provider code, run `task build`, then `task test` again

## Development Workflow

Standard flow for working on issues:

1. **Issue** — pick an issue from `gh issue list`
2. **Code** — implement the fix/feature, run `task build`
3. **PR** — create a branch, commit, push, open a PR via `gh pr create`
4. **CI** — wait for the CI pipeline to pass (`gh pr checks <number> --watch`). If it fails, investigate logs (`gh run view <run_id> --log-failed`), fix, and push again
5. **Codex review** — run `codex review --base main` and address all findings
6. **Iterate** — repeat steps 2–5 until CI is green and codex review has zero remarks
7. **Done** — PR is ready for merge

## Core Principles

- **Always check 3x-ui sources** before making assumptions about API behavior. Source snapshots are in `3x-ui-<version>/` directories. Download new versions with `curl -sL https://github.com/MHSanaei/3x-ui/archive/refs/tags/v<VERSION>.tar.gz | tar xz`. Key files: `web/service/` (business logic), `web/controller/` (API endpoints), `web/entity/model/` (data models), `xray/` (xray config).
- Be pragmatic: understand the task first, then make the minimum necessary changes.
- Do not break backward compatibility without an explicit request.
- Preserve code style and project structure.
- Make targeted changes, avoid mass reformatting.
- Run `task build` after code changes.
- Be concise and to the point. Indicate which files were changed.
- When changing anything documented in CLAUDE.md (workflows, structure, conventions), update CLAUDE.md in the same commit/PR.
