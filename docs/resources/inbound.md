---
page_title: "threexui_inbound Resource - 3x-ui"
subcategory: "Inbound"
description: |-
  Manages an inbound in the 3x-ui panel.
---

# threexui_inbound (Resource)

Manages an inbound proxy in the 3x-ui panel. Supports protocols: vless, vmess, trojan, shadowsocks, http, socks, mixed, wireguard, and tunnel.

## Example Usage

### VLESS with Reality

```hcl
resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

  stream_settings = jsonencode({
    network  = "tcp"
    security = "reality"
    realitySettings = {
      dest        = "www.apple.com:443"
      serverNames = ["www.apple.com"]
    }
    tcpSettings = {
      acceptProxyProtocol = false
      header = { type = "none" }
    }
  })

  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls", "quic", "fakedns"]
  })
}
```

### VMess

```hcl
resource "threexui_inbound" "vmess" {
  port     = 10086
  protocol = "vmess"
  enable   = true
  remark   = "VMess"

  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
  })
}
```

### Shadowsocks

```hcl
resource "threexui_inbound" "ss" {
  port     = 8388
  protocol = "shadowsocks"
  enable   = true
  remark   = "Shadowsocks"

  settings = jsonencode({
    method   = "chacha20-ietf-poly1305"
    password = "my-password"
    network  = "tcp,udp"
  })
}
```

## Argument Reference

- `port` (Required, Number) - Port number for the inbound.
- `protocol` (Required, String) - Protocol type (e.g. `vless`, `vmess`, `trojan`, `shadowsocks`, `http`, `socks`, `mixed`, `wireguard`, `tunnel`).
- `enable` (Optional, Boolean) - Whether the inbound is enabled. Default is `true`.
- `remark` (Optional, String) - A label/name for the inbound.
- `listen` (Optional, String) - Listen address.
- `up` (Optional, Number) - Upload traffic counter in bytes.
- `down` (Optional, Number) - Download traffic counter in bytes.
- `total` (Optional, Number) - Total traffic limit in bytes.
- `expiry_time` (Optional, Number) - Expiry time as Unix timestamp in milliseconds.
- `traffic_reset` (Optional, String) - Traffic reset period. Default is `never`.
- `settings` (Optional, String) - Protocol-specific settings as a JSON string. Default settings are applied per protocol if not specified. Clients are managed separately via `threexui_inbound_client`.
- `stream_settings` (Optional, String) - Transport and security settings as a JSON string. Uses Xray stream settings format with camelCase keys (e.g. `realitySettings`, `tcpSettings`).
- `sniffing` (Optional, String) - Traffic sniffing settings as a JSON string (e.g. `enabled`, `destOverride`).

-> **Note:** `settings`, `stream_settings`, and `sniffing` use a subset diff suppressor -- only the keys you specify are compared during plan. Keys returned by the API but absent from your config are ignored.

## Attribute Reference

- `id` - The inbound ID (numeric).
- `all_time` (Number) - All-time traffic in bytes.
- `last_traffic_reset_time` (Number) - Last traffic reset timestamp.
- `tag` (String) - Auto-generated inbound tag.

## Import

Inbounds can be imported using their numeric ID:

```shell
terraform import threexui_inbound.example 1
```
