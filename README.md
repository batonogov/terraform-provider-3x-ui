# Terraform/OpenTofu Provider for 3x-ui

A provider for managing [3x-ui](https://github.com/MHSanaei/3x-ui) panel inbounds, clients, settings, and Xray configuration via its HTTP API.

## Installation

```hcl
terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 0.1"
    }
  }
}
```

## Provider Configuration

```hcl
provider "threexui" {
  endpoint            = "http://localhost:2053"
  username            = "admin"
  password            = "admin"
  # base_path           = "/"           # optional, matches webBasePath in 3x-ui
  # insecure_skip_verify = true          # for self-signed HTTPS
  # request_timeout      = "30s"
}
```

## Resources

### `threexui_inbound`

Manages an inbound proxy. Supports protocols: `vless`, `vmess`, `trojan`, `shadowsocks`, `http`, `socks`, `mixed`, `wireguard`, `dokodemo-door`.

```hcl
resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
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
```

Key attributes:
- `remark` - Inbound label.
- `port` - Listen port.
- `protocol` - Protocol type.
- Per-protocol settings block (`vless_settings`, `shadowsocks_settings`, `http_settings`, etc.).
- `stream_settings` - Transport/security settings block (tcp, ws, grpc, httpupgrade, xhttp, kcp).
- `sniffing` - Sniffing settings block.

Note: Clients are managed exclusively via `threexui_inbound_client`.

### `threexui_inbound_client`

Manages a client within an existing inbound.

```hcl
resource "threexui_inbound_client" "client_a" {
  inbound_id = threexui_inbound.example.id
  email      = "client-a@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
```

Key attributes:
- `inbound_id` - Inbound ID.
- `email` - Client identifier (required).
- `enable` - Whether the client is enabled.
- `flow` - Flow for VLESS (`xtls-rprx-vision`, etc.).
- `expiry_time` - Expiry as Unix timestamp in milliseconds.
- `limit_ip` - IP limit.
- `total_gb` - Traffic limit.

### Panel Settings

| Resource | Description |
|---|---|
| `threexui_panel_general` | General panel settings (web server, LDAP, display preferences) |
| `threexui_panel_security` | Security settings (two-factor authentication) |
| `threexui_panel_user` | Admin username and password |
| `threexui_panel_telegram` | Telegram bot integration |
| `threexui_panel_subscription` | Subscription service settings |

### Xray Configuration

| Resource | Description |
|---|---|
| `threexui_xray_basics` | Basic Xray config (log, policy, api, stats) |
| `threexui_xray_dns` | DNS servers and hosts |
| `threexui_xray_routing` | Routing rules and domain strategy |
| `threexui_xray_balancers` | Load balancers |
| `threexui_xray_reverse` | Reverse proxy (bridges, portals) |
| `threexui_xray_outbounds` | Outbound connections (per-protocol settings) |

## Data Sources

| Data Source | Description |
|---|---|
| `threexui_inbounds` | List of all inbounds (JSON) |
| `threexui_server_status` | Server status: CPU, memory, disk, uptime (JSON) |
| `threexui_settings` | All panel settings (JSON) |
| `threexui_xray_config` | Current Xray template (JSON) |
| `threexui_xray_versions` | Available Xray versions (list of strings) |

## Import

```bash
# inbound
terraform import threexui_inbound.example 123

# inbound client: <inbound_id>:<client_id>
terraform import threexui_inbound_client.client_a 123:client-id
```

## Development

### Requirements

- Go 1.25+
- [Task](https://taskfile.dev/) - task runner
- [golangci-lint](https://golangci-lint.run/welcome/install/) - linter
- [pre-commit](https://pre-commit.com/) - git hooks framework
- Docker - for local 3x-ui environment

### Installing pre-commit hooks

```bash
# Install pre-commit
pip install pre-commit
# or via brew on macOS
brew install pre-commit

# Install git hooks
pre-commit install

# Run checks manually on all files
pre-commit run --all-files
```

### Development commands

```bash
task build        # Build the provider
task fmt          # Format code (gofmt)
task vet          # Run go vet
task lint         # Run golangci-lint
task pre-commit   # Run all checks manually (fmt, vet, lint, build)
task test         # Run acceptance tests (starts docker compose)
```

### Pre-commit checks

On every commit the following checks run automatically:
- **gofmt** - code formatting
- **go vet** - static analysis
- **golangci-lint** - linter
- **go build** - compilation check
- YAML/JSON file checks
- Trailing whitespace check

If checks fail, the commit is rejected. Fix the errors and try again.

### Local environment

```bash
# Start 3x-ui v2.8.9 on localhost:2053
docker compose up -d

# Login: admin / admin

# Stop
docker compose down
```
