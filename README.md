[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md)

# Terraform Provider for 3x-ui

> Manage [3x-ui](https://github.com/MHSanaei/3x-ui) inbounds, clients, panel settings, and Xray configuration as code — backup, migrate, and scale your VPN/proxy fleet without clicking through the panel.

[![CI](https://github.com/batonogov/terraform-provider-threexui/actions/workflows/ci.yml/badge.svg)](https://github.com/batonogov/terraform-provider-threexui/actions/workflows/ci.yml)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blueviolet)](https://registry.terraform.io/providers/batonogov/threexui/latest)
[![Latest Release](https://img.shields.io/github/v/release/batonogov/terraform-provider-threexui?sort=semver)](https://github.com/batonogov/terraform-provider-threexui/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/batonogov/terraform-provider-threexui)](https://goreportcard.com/report/github.com/batonogov/terraform-provider-threexui)
[![Go Version](https://img.shields.io/github/go-mod/go-version/batonogov/terraform-provider-threexui)](go.mod)
[![Last Commit](https://img.shields.io/github/last-commit/batonogov/terraform-provider-threexui)](https://github.com/batonogov/terraform-provider-threexui/commits/main)
[![Codecov](https://codecov.io/gh/batonogov/terraform-provider-threexui/branch/main/graph/badge.svg)](https://codecov.io/gh/batonogov/terraform-provider-threexui)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Why use it

Running 3x-ui in production means dozens of inbounds, hundreds of clients, and Xray configuration that is easy to break. With this provider you can:

- **Treat configuration as code** — your inbound list lives in git, every change is reviewed and versioned.
- **Migrate between servers** — re-create the same setup on a new VPS with one `terraform apply`.
- **Snapshot the panel** — `terraform state pull` is a full export of inbounds, clients, and settings.
- **Scale onboarding** — add 100 clients in a single PR instead of 100 panel clicks.
- **Plan before prod** — `terraform plan` shows exactly what will change before anything ships.

## Without vs with the provider

| Task | Panel UI | This provider |
| --- | --- | --- |
| Add 50 clients | 50 forms, ~30 seconds each | one `for_each`, one `apply` |
| Migrate to a new server | manual re-entry | `terraform apply` against the new endpoint |
| Audit who has access today | scroll the client list | `git log` on a `.tf` file |
| Roll back a misconfiguration | restore from a JSON backup | `git revert` + `terraform apply` |
| Sync staging ↔ production | export/import JSON, fix conflicts | shared module + per-environment vars |
| Rotate Reality keys on 10 hosts | open 10 panels, click each | one variable change, one apply |

## Quick Start

```hcl
terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

resource "threexui_inbound" "vless" {
  remark   = "VLESS Reality"
  port     = 443
  protocol = "vless"

  vless_settings {
    decryption = "none"
  }

  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "www.amazon.com:443"
      server_names = ["www.amazon.com"]
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

resource "threexui_inbound_client" "client_a" {
  inbound_id = threexui_inbound.vless.id
  email      = "client-a@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
```

## Compatibility

**Support policy:** the provider officially supports three 3x-ui minor lines: **2.9.x**, **3.0.x**, and **3.1.x** — every released patch in all three lines is exercised by the acceptance matrix on each push to `main` and every pull request.

| 3x-ui version | Status |
| --- | --- |
| v3.1.0 | Tested |
| v3.0.2 | Tested |
| v3.0.1 | Tested |
| v3.0.0 | Tested |
| v2.9.4 | Tested |
| v2.9.3 | Tested |
| v2.9.2 | Tested |
| v2.9.1 | Tested |
| v2.9.0 | Tested |

Newer protocol features are guarded with `requireMinVersion` and skip automatically on older versions, so the provider runs cleanly across the matrix without per-version forks.

## Examples

| Example | Description |
| --- | --- |
| [Provider with env config](examples/provider-env-config/) | Configure the provider using Terraform variables and `TF_VAR_*` environment variables |
| [Trojan inbound](examples/trojan-inbound/) | Trojan protocol with WebSocket transport |
| [Shadowsocks inbound](examples/shadowsocks-inbound/) | Shadowsocks with AEAD cipher |
| [Inbound with clients](examples/inbound-with-client/) | Complete workflow: inbound + multiple clients |
| [Multi-server fleet](examples/multi-server/) | Manage many 3x-ui hosts via a reusable module + `for_each` |
| [Import existing resources](examples/import-existing/) | Import existing 3x-ui resources into Terraform state |

## Guides

In-repo walkthroughs for common operational scenarios:

- [Backup-as-code](docs/guides/backup-as-code.md) — keep your full panel state in git, restore in seconds.
- [Migrating 3x-ui between servers](docs/guides/server-migration.md) — move an entire panel to a new VPS without re-typing anything.
- [Onboarding many clients at once](docs/guides/bulk-clients.md) — `for_each` patterns and CSV-driven onboarding.

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/batonogov/threexui/latest/docs).

### Resources

| Resource | Description |
| --- | --- |
| `threexui_inbound` | Inbound proxy (vless, vmess, trojan, shadowsocks, http, socks, mixed, wireguard, dokodemo-door) |
| `threexui_inbound_client` | Client within an inbound |
| `threexui_panel_general` | General panel settings |
| `threexui_panel_security` | Security settings (2FA) |
| `threexui_panel_user` | Admin credentials |
| `threexui_panel_telegram` | Telegram bot integration |
| `threexui_panel_subscription` | Subscription service settings |
| `threexui_xray_basics` | Basic Xray config (log, policy, api, stats) |
| `threexui_xray_dns` | DNS servers and hosts |
| `threexui_xray_routing` | Routing rules |
| `threexui_xray_balancers` | Load balancers |
| `threexui_xray_reverse` | Reverse proxy (bridges, portals) |
| `threexui_xray_outbounds` | Outbound connections |
| `threexui_xray_version` | Installed Xray core version |

### Data Sources

| Data Source | Description |
| --- | --- |
| `threexui_inbounds` | List of all inbounds (JSON, sensitive) |
| `threexui_server_status` | Server status: CPU, memory, disk, uptime (JSON) |
| `threexui_settings` | All panel settings (JSON, sensitive) |
| `threexui_xray_config` | Current Xray template (JSON, sensitive) |
| `threexui_xray_versions` | Available Xray versions (list of strings) |
| `threexui_online_clients` | Currently online client emails |
| `threexui_client_traffics` | Client traffic statistics by email |

## Security

The provider handles secrets the panel issues automatically (Reality `privateKey`, WireGuard `secretKey`, client UUIDs, Telegram bot tokens, LDAP passwords). All such fields are marked `Sensitive` and never logged in plaintext. See [SECURITY.md](SECURITY.md) for the full list and for guidance on protecting your Terraform state.

## Development

### Requirements

- Go (version pinned in [`go.mod`](go.mod))
- [Task](https://taskfile.dev/) — task runner
- [golangci-lint](https://golangci-lint.run/welcome/install/) — linter
- [pre-commit](https://pre-commit.com/) — git hooks framework
- Docker — for the local 3x-ui environment

### Commands

```bash
task build        # Build the provider
task fmt          # Format code (gofmt)
task vet          # Run go vet
task lint         # Run golangci-lint
task pre-commit   # Run all checks manually (fmt, vet, lint, build)
task test:unit    # Run unit tests (no Docker / Terraform needed)
task test:acc     # Run acceptance tests (starts docker compose)
task test         # Run unit + acceptance tests
```

### Local environment

```bash
# Start 3x-ui on localhost:2053
docker compose up -d

# Login: admin / admin

# Stop
docker compose down
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local setup, testing, and submission guidelines. Bug reports, feature requests, and pull requests are all welcome — and so are notes about which 3x-ui versions you run in production.

## Changelog

Releases follow [Conventional Commits](https://www.conventionalcommits.org/) and are published automatically. See [CHANGELOG.md](CHANGELOG.md) for the full version history.

## License

[MIT](LICENSE)
