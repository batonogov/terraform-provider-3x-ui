# Terraform/OpenTofu Provider for 3x-ui

A provider for managing [3x-ui](https://github.com/MHSanaei/3x-ui) panel inbounds, clients, settings, and Xray configuration via its HTTP API.

## Provider Configuration

```hcl
provider "threexui" {
  endpoint            = "http://localhost:2053"
  username            = "admin"
  password            = "admin"
  # base_path           = "/"           # optional
  # insecure_skip_verify = true          # for self-signed HTTPS
  # request_timeout      = "30s"
}
```

## Resources

### `threexui_inbound`

```hcl
resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
  protocol = "vless"

  stream_settings = jsonencode({
    network  = "tcp"
    security = "reality"
    realitySettings = {
      dest        = "www.apple.com:443"
      serverNames = ["www.apple.com"]
    }
  })

  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls", "quic", "fakedns"]
  })
}
```

Key attributes:
- `remark` - Inbound label.
- `port` - Listen port.
- `protocol` - Protocol (`vless`, `vmess`, `trojan`, `shadowsocks`, ...).
- `settings` - Protocol settings as a JSON string (clients excluded).
- `stream_settings` - Transport/security settings as a JSON string.
- `sniffing` - Sniffing settings as a JSON string.

Note:
- `settings`, `stream_settings`, and `sniffing` are **JSON strings**, not nested blocks. Use `jsonencode()`.
- Clients are managed exclusively via `threexui_inbound_client`.

### `threexui_inbound_client`

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
- `security` / `password` - Used for some protocols (sensitive).

## Import

```bash
# inbound
terraform import threexui_inbound.example 123

# inbound client: <inbound_id>:<client_id>
terraform import threexui_inbound_client.client_a 123:client-id
```

## Development

### Requirements

- Go 1.21+
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
