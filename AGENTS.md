# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Cursor, Codex, Aider, etc.) when working with code in this repository. It follows the [AGENTS.md](https://agents.md) convention.

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
| `task test:acc:compat` | Acceptance tests vs a pinned `THREEXUI_VERSION` (compat matrix) |
| `task test:acc:postgres` | Acceptance tests with a PostgreSQL backend (docker compose `--profile postgres`) |
| `task pre-commit` | fmt + vet + lint + build |

**Watch CI status for a PR:**

```bash
gh pr checks <PR>          # one-shot
gh pr checks <PR> --watch  # live
```

Every PR must end with a **green CI + a clean Codecov report**. The
`ci.yml` test job uploads coverage to Codecov (`codecov/codecov-action`),
which then posts a **Patch coverage** comment. Review it as part of the PR:
see [Testing → Codecov patch coverage](#testing) below for the rule.

**Version-specific acc tests:**

```bash
THREEXUI_VERSION=v3.1.0 task test:acc:compat
```

**Single acceptance test:**

```bash
TF_ACC=1 THREEXUI_ENDPOINT=http://localhost:2053 \
  THREEXUI_USERNAME=admin THREEXUI_PASSWORD=admin \
  THREEXUI_VERSION=v3.3.1 \
  go test ./provider -run TestAccInboundReality -count=1 -timeout 600s -v
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
`inbound_client` does read-modify-write on the inbound's `settings.clients` array
under its own `inboundClientMu` mutex (a third RMW lock alongside `settingsMu` and
`xrayTemplateMu`). Both core resources accept `restart_xray = true` to restart
xray-core after create/update/delete (`POST /panel/api/server/restartXrayService`).

### Panel settings (singletons)

Five resources — `panel_general`, `panel_security`, `panel_telegram`,
`panel_email`, `panel_subscription` — share `resource_settings_tabs.go`.

- Each is a singleton with ID `"settings"`.
- Shared CRUD via `settingsApplyTyped` / `settingsReadTyped`.
- Delete only clears TF state — does not reset panel settings.
- `settingsMu` serializes read-modify-write (separate from `xrayTemplateMu`).
- `settingsSecrets` on Client remembers the last configured `tgBotToken` /
  `twoFactorToken` / `ldapPassword` / `smtpPassword` and replays them on partial updates when a GET
  returns an empty/redacted sentinel. Note: 3x-ui v3.0.2–v3.3.1 actually returns
  these secrets **raw** on `/setting/all`, so the replay path is defensive and
  mainly fires when the panel genuinely has no secret stored.
- Changing a restart-triggering key triggers a provider-initiated restart
  (`POST /setting/restartPanel`, SIGHUP; 3x-ui does **not** auto-restart) in two
  resources, each gated by the shared `panelSettingsNeedRestart` check followed by
  `SendRestart` + `WaitForReady`: `panel_general` (`webListen`, `webDomain`,
  `webPort`, `webBasePath`, `webCertFile`, `webKeyFile`, `sessionMaxAge`) and
  `panel_subscription` (the subscription-server binding keys `subEnable`, `subListen`,
  `subDomain`, `subPort`, `subPath`, `subCertFile`, `subKeyFile`). The 3x-ui sub
  server only (re)binds at startup, so without this restart a changed `sub_path`/
  `sub_port`/… does not take effect and the subscription URL 404s (#291). Link-
  generation fields (`subURI`, `subTitle`, …) are read per request and do **not**
  restart. For `webBasePath` the provider additionally calls `SetBasePath` +
  `WaitForReady` so subsequent requests target the new path.

### Write-only secret attributes (Terraform 1.11+ / OpenTofu 1.11+)

Five singleton secrets have write-only (`_wo`) alternatives following the AWS/Azure pattern.
Each secret gets three attributes: old `Sensitive` attr + `WriteOnly` attr + `_wo_version` trigger.

| Resource | Old attr | Write-only | Version trigger |
| --- | --- | --- | --- |
| `panel_user` | `password` | `password_wo` | `password_wo_version` |
| `panel_security` | `two_factor_token` | `two_factor_token_wo` | `two_factor_token_wo_version` |
| `panel_telegram` | `tg_bot_token` | `tg_bot_token_wo` | `tg_bot_token_wo_version` |
| `panel_email` | `smtp_password` | `smtp_password_wo` | `smtp_password_wo_version` |
| `panel_general` | `ldap_password` | `ldap_password_wo` | `ldap_password_wo_version` |
| `threexui_node` | `api_token`, `pinned_cert_sha256` | `api_token_wo`, `pinned_cert_sha256_wo` | `api_token_wo_version`, `pinned_cert_sha256_wo_version` |

- `PreferWriteOnlyAttribute` validator warns on TF >= 1.11 when using old attr.
- `int64validator.AlsoRequires` enforces that `*_wo_version` can only be set with `*_wo`.
- Write-only values read from `req.Config` (not plan/state — framework nulls them).
- Two resolution strategies:
  - **panel_user**: `password` is nulled in state when `password_wo` is used. Update
    falls back to provider credentials when state password is empty. No ModifyPlan
    needed because state has no prior password to conflict with.
  - **settings (security/telegram/email/general)**: `resolveXxxWO` copies `_wo` value into
    the old model field before `expand*()`. ModifyPlan marks the plain attr as Unknown
    on `woVersionTriggered` so Terraform accepts a new sensitive value from Apply.
    Needed because flatten reads the masked value back, which would otherwise be
    rejected as "inconsistent values for sensitive".
- Version trigger: `resolveXxxWOUpdate` only sends the `_wo` value when
  `woVersionTriggered(plan, state)` (version changes, or first use when state has none).

### Xray settings (two modes)

Seven xray resources share `resource_xray_settings.go`: `xray_basics`, `xray_dns`,
`xray_routing`, `xray_balancers`, `xray_reverse`, `xray_outbounds`,
`xray_observatory`. Each section has its own `*_schema.go` file.

| Mode | Resource | Behavior |
| --- | --- | --- |
| Merge root | `xray_basics` | Reads full template, merges changed fields, writes back |
| Set path | all others | Reads/writes a section by JSON path |

`xrayTemplateMu` serializes read-modify-write to prevent race conditions.

### Xray version (`resource_xray_version.go`)

Separate resource — singleton with ID `"xray_version"`. Calls `InstallXray` + polls
`waitForXrayVersion` (180×1s) until the version matches; if the panel still reports a
stale version after 30 attempts it re-issues `InstallXray` once (3x-ui v3.2.6–v3.2.7
sometimes silently drop the first install, #262). If the panel still reports
`"Unknown"` after `nudgeAfter` (60s) of polling, the provider issues a single
`RestartXrayService` (bounded by `nudgeTimeout` = 30s context) to prod the core
into reporting its version — this is a cheap local call (no GitHub download) and
is non-fatal: the nudge error is silently swallowed and polling continues. Read
treats `"Unknown"` as a soft-fail (Warning + preserved state). Delete is a no-op
with a warning (removing from state does NOT revert the installed version).

### Panel user (`resource_panel_user.go`)

`threexui_panel_user` — no read API exists. Read is a no-op, state is preserved
from plan. Create uses provider credentials as old credentials; Update uses
prior state credentials (falls back to provider credentials when state
password is empty, e.g. after using `password_wo`). See Write-only section
above for the `password` / `password_wo` lifecycle.

### Cluster node (`resource_node.go`)

`threexui_node` — manages a remote 3x-ui panel registered as a cluster node
(multi-node surface, `/panel/api/nodes/*`; available since v3.0.2, no legacy
fallback). Routes are gin subpaths: list `GET /list`, read `GET /get/:id`, create
`POST /add`, update `POST /update/:id`, delete `POST /del/:id` (POST, not DELETE).
Create does a form-POST `/add`, then re-reads `/get/:id` for observed state. The
**central panel probes the node for reachability (`ensureReachable`) before
persisting it**, so the node's web API must be reachable from the central panel
at apply time — there is no provider-side flag to bypass this (decided in #315).
Read does `GET /get/:id` and treats a missing node (the panel signals this as
HTTP 200 + `success:false` with a gorm "record not found" message, NOT HTTP 404
— see `util.go jsonMsgObj`; handled via `isAPIRecordNotFound`) as remove-from-
state. Update form-POSTs `/update/:id` then re-reads; the **3x-ui server
restarts the Xray core itself** when `outbound_tag` changes
(`controller/node.go:180-181`), so the provider does NOT call
`RestartXrayService` (unlike `threexui_inbound`'s `restart_xray = true`). Delete
POSTs `/del/:id`; 3x-ui refuses to delete a node that still owns inbounds
(DB-002, #314 R3) — surfaced as a clear error so the operator detaches the
inbounds first. Import is by numeric id (`ImportStatePassthroughID`). `api_token`
and `pinned_cert_sha256` are `Sensitive` (panel returns them raw, no redaction —
issue #314 R1) with write-only `_wo` variants following the settings-style
strategy (`api_token_wo`+`_wo_version`, `pinned_cert_sha256_wo`+`_wo_version`;
`resolveNodeSecretsWO`/`resolveNodeSecretsWOUpdate` + an **inlined** ModifyPlan
(not the shared `modifyPlanWOVersion` generic — the node has two write-only secrets
and the generic re-reads/overwrites `resp.Plan` on each call, which would erase the
first `Unknown` mark on the second). Schema lives in `node_schema.go`;
the typed `Node` model is shared with the `threexui_nodes` data source (`types.go`).

### Host group (`resource_host_group.go`)

`threexui_host_group` — manages a 3x-ui v3.5.0+ host group (bulk host management,
`/panel/api/hosts/*`; available since v3.5.0, no legacy fallback — older panels use
`/get/:id` with numeric id). One group = one `groupId`; routes are gin subpaths:
list aside, create `POST /add`, read `GET /get/:groupId`, update `POST /update/:groupId`,
delete `POST /del/:groupId` — all with a **JSON body** (`entity.HostGroup`), not
form-POST like inbounds/nodes. Create does `POST /add` (server generates
`groupId` via `random.NumLower(16)` when omitted), then re-reads `/get/:groupId`
for canonical state. Read treats a missing group (HTTP 200 + `success:false` with an
explicit `"host group not found"` message — **not** a gorm record-not-found like
nodes; detected via the separate `isHostGroupNotFound`, not `isAPIRecordNotFound`) as
remove-from-state. Update `POST /update/:groupId` (delete-then-recreate under transaction) then
re-read; delete `POST /del/:groupId`. Import is by `groupId`
(`ImportStatePassthroughID`). `inbound_ids` is Required (`min=1`), `remark` Required,
`port` 0–65535, `security`/`mihomo_ip_version` enum-validated. No sensitive fields →
no write-only variants. v3.5.0+ only.

### HTTP client (`client.go`)

Cookie auth, auto re-login on 401/404, CSRF token handling (enforced on all
`/panel/api/*` + login routes since ≥ v3.0.2; Bearer API-token auth bypasses it).
Supports bootstrap credentials for first-run panel setup (the panel auto-creates an
`admin`/`admin` user on an empty DB; bootstrap ordering differs by generation —
v2.9.x tries bootstrap first, v3.x tries primary first).
Provider attributes: `two_factor_code` (TOTP for 2FA login), `max_retries`
(default 1, max 10), `request_timeout`.
Three retry layers: 5xx on idempotent writes, read-after-write for SQLite
visibility lag, rate-limit retry for `GetXrayVersions`.
Two independent API-surface auto-detections (each probed once and cached):
(a) client API — probes `/panel/api/clients/list` to pick v3.1.0+
`/panel/api/clients/*` over `/panel/api/inbounds/*`; (b) settings/xray API — probes
`/panel/api/setting/all` to pick v3.3.0+ `/panel/api/setting/*` + `/panel/api/xray/*`
over `/panel/setting/*` + `/panel/xray/*`. Both fall back to legacy endpoints
mid-run via `markLegacy*` on 404.

### JSON format compatibility (`types.go`)

3x-ui v3.1.0 returns `settings`, `streamSettings`, `sniffing` as nested JSON
objects instead of escaped strings. Custom `UnmarshalJSON` on `Inbound`
normalises both formats to plain strings — rest of the code is unaffected.

### Data sources

Eight data sources: `inbounds`, `nodes`, `server_status`, `xray_versions`, `xray_config`,
`settings`, `online_clients`, `client_traffics`. All are read-only GET wrappers
that return the raw panel payload — none accept filter arguments. JSON attrs that
contain secrets (UUIDs, private keys) must be marked `Sensitive: true`.

`threexui_nodes` wraps `GET /panel/api/nodes/list` (3x-ui multi-node/cluster surface,
available since v3.0.2; no legacy fallback). The response is a **node tree**
including transitive sub-nodes (read-only projections with `Id == 0`,
`transitive == true`). The payload carries each node's `apiToken` and
`pinnedCertSha256` **raw** (no redaction layer), so the `nodes` attr is
`Sensitive: true`.

---

## Critical Gotchas

These are non-obvious constraints that have caused real bugs.

| Gotcha | Why it matters |
| --- | --- |
| `email` in `threexui_inbound_client` is **Required** | The `clients`/`client_traffics` tables key on it (`uniqueIndex; not null` / `gorm:"unique"`); the panel rejects an empty email with `"client email is required"` before the SQL layer |
| Docker `base_path` is `/` | Do **not** set `THREEXUI_BASE_PATH=/panel/` |
| Data-source JSON attrs need `Sensitive: true` | Payloads contain UUIDs, private keys, tokens |
| `Optional+Computed` attrs need `UseStateForUnknown` | Prevents false drift (`known after apply`) after import |
| `alignBlocksWithPlan` | Prevents "was absent, but now present" for Optional blocks; skipped during Import |
| `reality_settings.settings` is `SingleNestedAttribute` | Not a block. Uses `objectplanmodifier.UseStateForUnknown()` |
| Subscription resource does **double apply** | Workaround: `sub_json_enable` not saved on first apply with `sub_enable` |
| Policy levels: `{"0": {...}}` ↔ `[{id=0}]` | Xray JSON uses a map, TF uses a list |
| DNS servers: string vs object | Address-only → string; with extra fields → object |
| Protocol names are version-specific | `hysteria2` is a distinct protocol only on v3.0.2/v3.1.0; from v3.2.0 use `protocol = "hysteria"` with `streamSettings.version = 2`. `tunnel` requires ≥ v3.2.0; `tun`/`mtproto` require ≥ v3.3.0 |
| `xray_version` delete is a no-op | Removing from state does NOT revert the installed xray version |
| `web_base_path` change triggers panel restart | Must also update provider `base_path`; code auto-updates client |
| `panel_outbound` vs `panel_proxy` is version-specific | 3x-ui v3.3.1 **replaced** `panelProxy` (HTTP/SOCKS5 URL, present only in v3.2.0–v3.3.0; absent in v3.0.2/3.1.0) with `panelOutbound` (an Xray outbound/balancer tag). The provider's `panel_proxy` attr is `Deprecated` (provider-side annotation) and maps to the old field; gin silently ignores the unknown `panelProxy` form value on v3.3.1 |
| v3.4.0 AllSetting delta | 3x-ui v3.4.0 **added** 14 fields now managed by the provider: SMTP notifications (`smtp*`, 10 fields, `panel_email` resource, `smtpPassword` is sensitive + write-only), `tgEnabledEvents`/`tgMemory` (`panel_telegram`), `remarkTemplate`/`subHideSettings` (`panel_subscription`). v3.4.0 also **dropped** `remarkModel`/`subEmailInRemark`/`subShowInfo`/`tgBotLoginNotify` from `AllSetting`; the provider still sends them (backward compat for v3.1.x–v3.3.x) and they sit in `intentionallySkipped` in `drift_test.go` |
| v3.4.1 AllSetting delta | 3x-ui v3.4.1 **added** 2 subscription fields now managed by `panel_subscription`: `subIncyEnableRouting`/`subIncyRoutingRules` (Incy client routing injection). No fields were removed |
| v3.4.2 AllSetting + model delta | 3x-ui v3.4.2 **added** `ldapInsecureSkipVerify` to `AllSetting` (managed by `panel_general`). `Client`/`ClientModel` gained WireGuard multi-client fields (`privateKey`/`publicKey`/`allowedIPs`/`preSharedKey`/`keepAlive`); WG inbounds now have native `dns` + `clients[]` (single-peer inbounds migrate automatically). `ClientModel` also gained `resetUp`/`resetDown`. `updateUser` (password change) now requires a `twoFactorCode` form field when 2FA is enabled (`internal/web/controller/setting.go`); `updateSetting` requires it only when disabling 2FA — the provider auto-attaches `twoFactorCode` on both (see gotcha). xray-core bumped to v26.6.27. **Observatory/BurstObservatory** (`frontend/src/schemas/observatory.ts`) are new top-level xray template keys (`xray.ts:33-34`, opaque `z.unknown()`), auto-synced by the frontend with balancers (`balancer-helpers.ts:syncObservatories`) and absent from the default config; they are now exposed as the `threexui_xray_observatory` typed resource (`xray_observatory_schema.go`), which co-owns both the `observatory` and `burstObservatory` top-level keys — the panel's frontend `syncObservatories` only runs in the browser on UI-save, not on API-driven changes, so the two do not conflict at the API level (see gotcha) |
| v3.5.0 AllSetting + model delta | 3x-ui v3.5.0 **AllSetting is identical to v3.4.2** (0 field delta — `entity.go` `AllSetting` struct unchanged at 94 fields). `Client`/`ClientRecord` gained MTProto multi-client fields `secret` (FakeTLS, `json:"secret,omitempty"`) and `adTag` (`json:"adTag,omitempty"`) for the new `mtg-multi` sidecar; inbound-level MTProto `secret`/`adTag` are now vestigial (`StripMtprotoInboundSecret`/`StripMtprotoInboundAdTag`). xray-core bumped to **v26.7.11**. New top-level xray template keys: `env` (`map[string]string`, managed by `xray_basics`) and `outbounds[].targetStrategy` (domain-strategy enum, managed by `xray_outbounds`). **Shadowsocks** `method` `none`/`plain` and **vmess** `security` `none`/`zero` were removed in xray-core v26.7.11 — the panel auto-migrates them on startup (`migrateShadowsocksRemovedCiphers`/`migrateVmessRemovedSecurities` in `internal/database/db.go`), causing silent drift if a provider-managed inbound still carries them (see cipher gotcha). **ValidateOutboundConfig** now pre-rejects unencrypted public vless/trojan outbounds at apply (was a silent xray crash). **finalmask.tcp + REALITY** is rejected at save (XTLS/Xray-core#6453). **EnsureDnsServerRouting** (`internal/web/service/xray_setting_dns_routing.go`) auto-inserts managed `direct` allow-rules for private DNS servers before the `geoip:private` block on every xray-template save — the provider filters these (`isManagedDnsAllowRule`) to avoid permanent drift. New `entity.HostGroup` struct + `/panel/api/hosts/*` routes (bulk host management) are managed by the `threexui_host_group` resource. Balancer-to-balancer fallback reserves the `_bl_` tag prefix; `balancer.fallback_tag` + `balancer.strategy.settings` (expected/max_rtt/tolerance/baselines/costs) are now exposed on `xray_balancers`. WireGuard live updates no longer drop clients (`WireguardClientsToPeers` now runs in `GenXrayInboundConfig`) |
| MTProto multi-client on v3.5.0 (`mtg-multi`) | 3x-ui v3.5.0 moved MTProto to the `mtg-multi` engine: each client carries its own FakeTLS `secret` and optional `adTag` on `Client` (per-client), surfaced on `threexui_inbound_client` as `secret` (Sensitive) and `ad_tag` (RegexMatches `^[0-9a-fA-F]{32}$`). The panel heals per-client secrets on save via `HealMtprotoClientSecrets`, rebuilding the domain suffix from the inbound's `fakeTlsDomain` (default `www.cloudflare.com`) — so only the random 32-hex middle of `secret` must stay stable across applies. Inbound-level MTProto `secret`/`adTag` are vestigial. The provider has **no typed `mtproto_settings` block** on `threexui_inbound` (`fakeTlsDomain`/`domainFronting`/etc. remain a pre-existing gap); mtproto inbounds are created with `protocol = "mtproto"` and clients added via `threexui_inbound_client`. v3.5.0+ only — on v3.4.x these fields are absent from `Client` |
| Removed ciphers cause silent drift on v3.5.0 | xray-core v26.7.11 (3x-ui v3.5.0) **removed** shadowsocks `method` `none`/`plain` and vmess `security` `none`/`zero`. The panel runs `migrateShadowsocksRemovedCiphers` (→ `chacha20-ietf-poly1305`) and `migrateVmessRemovedSecurities` (→ `auto`) on **every startup** (`internal/database/db.go`, not seeder-gated, idempotent). A provider-managed inbound carrying one of these values applies cleanly, then drifts after the next panel restart: `terraform plan` shows state=`none`, panel=`chacha20-ietf-poly1305` (or `auto`). Avoid these values in `threexui_inbound` settings / `threexui_inbound_client.security` |
| WireGuard `workers` deprecated | xray-core v26.6.22 (3x-ui v3.4.0) removed the WireGuard `workers` field; `xray_outbound_wireguard` now marks it `DeprecationMessage` — accepted for backward compat, ignored by xray |
| Write-only attrs quirks | See "Write-only secret attributes" section above — read `_wo` from `req.Config`; ModifyPlan marks plain `Unknown` on version change; `panel_user` nulls state password instead |
| 2FA re-confirmation on v3.4.2 (provider auto-sends) | 3x-ui v3.4.2 made `/setting/updateUser` require a `twoFactorCode` whenever 2FA is enabled, and `/setting/update` require it when disabling 2FA. The provider's `Client.UpdateUser` and `Client.UpdateSettings` automatically attach `twoFactorCode` (from the provider `two_factor_code` config) to the payload when set — older panels (≤ v3.4.1) bind into `AllSetting` and ignore the extra key, so this is backward-compatible. `UpdateSettings` copies the caller's map first so the code never leaks into plan/state. No resource-level change is needed; users only need `two_factor_code` set when 2FA is on |
| WireGuard `peer` vs `clients` on v3.4.2 | On v3.4.2, single-peer WG inbounds auto-migrate upstream from `settings.peers[]` into `wireguard_settings.clients[]`. A provider-managed WG inbound declared with only `peer[]` may therefore surface `clients[]` populated and `peer[]` empty after the panel upgrades — the next `terraform plan` shows an unexpected diff (it self-heals on apply, since Read records the new state). Also: the legacy `peer` block is serialised with snake_case JSON keys while the v3.4.2 `clients[]` block uses camelCase (matching upstream `client_wireguard.go`); both arrays are Optional, so populate EITHER `peer` OR `clients` per inbound, never both (the panel treats them as separate models). The `clients[].email` attr is Optional in the schema but the panel requires a non-empty unique email (it keys traffic counters on it), mirroring the generic `threexui_inbound_client` `email` gotcha |
| WireGuard `clients[]` is managed by `threexui_inbound`, NOT `threexui_inbound_client` | Unlike vmess/vless/trojan/shadowsocks/hysteria — whose clients live in `settings.clients[]` and are owned by the separate `threexui_inbound_client` resource (the inbound settings layer deliberately **strips** `clients[]` for those protocols) — WireGuard multi-client peers live in `wireguard_settings.clients[]` and are managed by `threexui_inbound` itself. `buildSettingsJSON`/`flattenSettings` forward `clients[]` ONLY when `protocol == "wireguard"` (protocol-aware pass-through, #342). Adding a WireGuard client therefore means a `clients { ... }` block under `wireguard_settings`, not a `threexui_inbound_client` resource. Mixing the two surfaces on one WG inbound is unsupported. |
| xray-settings acc-tests restart xray-core | `TestAccXrayBasics`/`DNS`/`Routing`/`Balancers`/`Reverse`/`Outbounds` each apply a new config = xray restart; `version` becomes `"Unknown"` for ~30-90s after. Use `waitForXrayVersion(t, ctx, client)` poll-loop in any acc-test probing version after them (#280) |
| `TestAccXrayVersionDrift` drift sim skips on pickup failure | Step 2 simulates drift via `InstallXray` + bounded poll (3 installs × 6 polls). If the panel accepts the install but never swaps the binary (intermittent on GitHub-hosted runners, esp. downgrades), it `t.Skipf`s — same environment-issue category as #279/#280, NOT a provider regression. The poll budget is deliberately tight: the old 20-poll loop ran ~7.5 min and ate the whole test binary's `-timeout`, panic-killing unrelated tests (#306). Do NOT widen the budget back to 20. A `t.Deadline()` guard at the top of the test also skips cleanly when the preceding `TestAccXrayVersion`/`TestAccXrayVersionResource` tests already burned the binary budget via their 180s `waitForXrayVersion` stalls — otherwise the panic fires mid-Step-2 before drift's own skip path |
| `GetXrayVersions` hits GitHub anonymously | 3x-ui's `getXrayVersion` handler has no `Authorization`; 60 req/h/IP on shared runners. Use `skipOnUpstreamRateLimit(t, err)` / `testAccSkipOnXrayVersionsRateLimit(t)` to skip, not fail (#279) |
| PostgreSQL runner is slower than SQLite | Timing-sensitive flakes (xray restart windows, InstallXray pickup) surface on the `test:acc:postgres` job but not on `test:acc`. A green SQLite run does NOT prove PostgreSQL is green |
| Acceptance failure logs are captured before teardown | `task test:acc*` conditionally writes Docker logs to `.pi/artifacts/3xui-container.log` from a deferred command. Deferred commands run in reverse order, so keep the log-capture defer after the compose-down defer; CI uploads that file after the task has cleaned up containers. |

**Always check 3x-ui source snapshots** (`3x-ui-<version>/`) before assuming API behavior.
Key paths: `internal/database/model/`, `internal/web/service/`, `internal/web/controller/`, `internal/web/entity/`, `frontend/src/schemas/`, `xray/`.

> Note: 3x-ui reorganised its source tree under `internal/` (in v3.4.x snapshots the model/service/controller/entity files live under `internal/...`, not `database/...`/`web/...`). Older `3x-ui-<version>/` snapshots (≤ v3.3.x) still use the pre-`internal/` layout.

---

## Conventions

### Keeping AGENTS.md current

AGENTS.md is the only place that documents non-obvious behavior an agent cannot
derive from reading one file. Update it **in the same PR** when you:

- add or remove a resource, data source, or schema attribute;
- add a new mutex, ModifyPlan behavior, or read-modify-write path;
- change API endpoint usage, retry budgets, or version-gated behavior;
- hit a new non-obvious constraint that caused (or would cause) a real bug —
  add it to "Critical Gotchas".

Prefer citing the 3x-ui source snapshot (`3x-ui-<version>/`) as evidence for
API-behavior claims rather than restating them from memory.

### Working artifacts

Save scratch artifacts (subagent audit output, draft analysis, large intermediate
results) under `.pi/artifacts/`, not in the repo root. `.pi/` is git-ignored and
reserved for agent tooling; nothing under it is committed. This keeps audit
output reusable within a session without polluting the tree.

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
- **`context7.json`** (repo root) controls how [Context7](https://context7.com) indexes
  this project. It excludes the `3x-ui-<version>/` source snapshots (~1900 upstream
  files that would otherwise drown the provider's own API in search results), `CHANGELOG.md`,
  and build/cache artefacts, and exposes `rules` for coding agents. Update `excludeFolders`
  whenever a new snapshot is added (see "Adding a new 3x-ui version to CI").

### Testing

- Unit tests: `TestXxx` naming, table-driven where practical.
- Acceptance tests: `TestAccXxx`, `terraform-plugin-testing`,
  `ProtoV6ProviderFactories` (not `ProviderFactories`).
- Version-aware skipping: `requireMinVersion(t, "vX.Y.Z")` for features added in a
  specific 3x-ui version; `requireBelowVersion(t, "vX.Y.Z")` for features removed
  upstream (e.g. `hysteria2`, dropped in v3.2.0). Currently supported: **v3.1.x**,
  **v3.2.x**, **v3.3.x**, **v3.4.x**, **v3.5.x** (up to v3.5.0).
- Flaky test quarantine: `skipOnFlakyVersions(t, ...)` / `skipIfFlaky(t)` with
  `THREEXUI_SKIP_FLAKY` env var to skip known-broken upstream versions.
- Xray-version acc-tests use `waitForXrayVersion(t, ctx, client)` (180×1s retry on
  `ErrXrayVersionUnknown`) — covers both the cold-start race (handled by the
  `_wait-for-xray` + `_warm-xray-version-cache` Taskfile gates, the latter pre-warms
  3x-ui's 15-min GitHub-API cache) and mid-test restarts after `TestAccXray*`
  settings tests. **Do not** call `client.GetCurrentXrayVersion()` directly in
  acc-tests (#280).
- Any acc-test exercising `GetXrayVersions` (directly or via `threexui_xray_versions`
  data source / `xray_version` resource) must pre-flight with
  `testAccSkipOnXrayVersionsRateLimit(t)`, or wrap the error with
  `skipOnUpstreamRateLimit(t, err)`. Rate limit is an environment issue, skip —
  don't fail (#279).
- Protocol matrix test (`resource_inbound_matrix_test.go`): comprehensive
  create/update/import round-trip for every protocol.
- Destroy checks use `destroyVisibilityAttempts` (60 × 500 ms) to handle
  SQLite visibility lag after delete.
- **Codecov patch coverage is a PR gate.** The `ci.yml` test job uploads
  `coverage.out` to Codecov, which posts a *Patch coverage* comment on the PR
  (e.g. `⚠️ Patch coverage is 39% with 28 lines missing`). A red Codecov
  report must be resolved before merge — do not ignore it. Concretely:
  - New `provider/*.go` code is **unit-tested**, not only acc-tested. A data
    source's `Schema`/`Configure`/`Read`, a resource's CRUD methods, and new
    `Client` methods must each have a direct unit test. Acc tests don't count
    toward patch coverage and run only with `TF_ACC=1`.
  - If Codecov flags a line, either cover it with a unit test or, if the line
    is genuinely unreachable/defensive (e.g. a can't-happen error branch),
    document why inline. Do **not** merge with a red patch-coverage report
    just because CI is green — `task test:unit:coverage` reproduces the same
    numbers locally (`coverage.out`).
  - Install the Codecov GitHub app on the repo so uploads/comments are
    processed reliably (the bot warns: *"Please install the 'codecov app svg
    image'…"* when the app is missing).

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
8. **`AGENTS.md`** — update the "up to vX.Y.Z" version reference if the minor line changed
9. **`context7.json`** — add the new `3x-ui-<version>/` snapshot folder to `excludeFolders` so Context7 does not index the upstream source copy
