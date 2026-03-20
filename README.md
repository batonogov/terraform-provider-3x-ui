# Terraform Provider for 3x-ui

[![CI](https://github.com/batonogov/terraform-provider-threexui/actions/workflows/ci.yml/badge.svg)](https://github.com/batonogov/terraform-provider-threexui/actions/workflows/ci.yml)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blueviolet)](https://registry.terraform.io/providers/batonogov/threexui/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Terraform provider for managing [3x-ui](https://github.com/MHSanaei/3x-ui) panel inbounds, clients, settings, and Xray configuration via its HTTP API.

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
      target       = "www.apple.com:443"
      server_names = ["www.apple.com"]
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
| `threexui_inbounds` | List of all inbounds (JSON) |
| `threexui_server_status` | Server status: CPU, memory, disk, uptime (JSON) |
| `threexui_settings` | All panel settings (JSON) |
| `threexui_xray_config` | Current Xray template (JSON) |
| `threexui_xray_versions` | Available Xray versions (list of strings) |
| `threexui_online_clients` | Currently online client emails |
| `threexui_client_traffics` | Client traffic statistics by email |

## Development

### Requirements

- Go 1.25+
- [Task](https://taskfile.dev/) - task runner
- [golangci-lint](https://golangci-lint.run/welcome/install/) - linter
- [pre-commit](https://pre-commit.com/) - git hooks framework
- Docker - for local 3x-ui environment

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

See [CONTRIBUTING.md](CONTRIBUTING.md) for local setup, testing, and submission guidelines.

## License

[MIT](LICENSE)
