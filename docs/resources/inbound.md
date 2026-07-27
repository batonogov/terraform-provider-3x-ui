---
page_title: "threexui_inbound Resource - 3x-ui"
subcategory: "Inbound"
description: |-
  Manages an inbound in the 3x-ui panel.
---

# threexui_inbound (Resource)

Manages an inbound proxy in the 3x-ui panel. Supports protocols: vless, vmess, trojan, shadowsocks, http, mixed, wireguard, tunnel, tun, hysteria, and mtproto. For older panels, the provider also preserves legacy `socks`, `dokodemo-door`, and `hysteria2` values from imported state; on 3x-ui v3.2.0+ use `mixed`, `tunnel`, and `hysteria` with `version = 2` for new configurations. TUN requires v3.2.7+, and MTProto requires v3.3.0+.

## Example Usage

### VLESS with Reality

```hcl
resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

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
    tcp_settings {
      accept_proxy_protocol = false
      header_type           = "none"
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}
```

### VMess

```hcl
resource "threexui_inbound" "vmess" {
  port     = 10086
  protocol = "vmess"
  enable   = true
  remark   = "VMess"

  stream_settings {
    network  = "tcp"
    security = "none"
  }
}
```

### Shadowsocks

```hcl
resource "threexui_inbound" "ss" {
  port     = 8388
  protocol = "shadowsocks"
  enable   = true
  remark   = "Shadowsocks"

  shadowsocks_settings {
    method   = "chacha20-ietf-poly1305"
    password = "my-password"
    network  = "tcp,udp"
  }
}
```

### VLESS with WebSocket

```hcl
resource "threexui_inbound" "ws" {
  port     = 8443
  protocol = "vless"
  enable   = true
  remark   = "VLESS WS"

  vless_settings {
    decryption = "none"
  }

  stream_settings {
    network  = "ws"
    security = "none"
    ws_settings {
      path = "/ws"
    }
  }
}
```

### HTTP Proxy with Authentication

```hcl
resource "threexui_inbound" "http" {
  port     = 8080
  protocol = "http"
  enable   = true
  remark   = "HTTP Proxy"

  http_settings {
    auth              = "password"
    allow_transparent = false
    account {
      user = "myuser"
      pass = "mypass"
    }
  }
}
```

### WireGuard

```hcl
resource "threexui_inbound" "wg" {
  port     = 51820
  protocol = "wireguard"
  enable   = true
  remark   = "WireGuard"

  wireguard_settings {
    mtu = [1420, 1280]
    peer {
      public_key  = "BASE64_PUBLIC_KEY"
      allowed_ips = ["10.0.0.2/32"]
    }
  }
}
```

### Hysteria

```hcl
resource "threexui_inbound" "hysteria" {
  port     = 8443
  protocol = "hysteria"
  enable   = true
  remark   = "Hysteria"

  hysteria_settings {
    version = 2
  }

  stream_settings {
    network  = "hysteria"
    security = "tls"
  }
}
```

### MTProto

```hcl
resource "threexui_inbound" "mtproto" {
  port     = 443
  protocol = "mtproto"
  enable   = true
  remark   = "MTProto FakeTLS"

  mtproto_settings {
    fake_tls_domain = "www.cloudflare.com"
    prefer_ip       = "prefer-ipv4"
  }
}
```

## Argument Reference

### Top-level

- `port` (Required, Number) - Port number for the inbound.
- `protocol` (Required, String) - Protocol type (`vless`, `vmess`, `trojan`, `shadowsocks`, `http`, `mixed`, `wireguard`, `tunnel`, `tun`, `hysteria`, `mtproto`). Legacy `socks` and `dokodemo-door` are available only on panels before 3x-ui v3.2.0; use `mixed` and `tunnel` on v3.2.0+. `tun` is a tunnel alias available on v3.2.7+, and `mtproto` is available on v3.3.0+.
- `enable` (Optional, Boolean) - Whether the inbound is enabled. Default is `true`.
- `remark` (Optional, String) - A label/name for the inbound.
- `listen` (Optional, String) - Listen address.
- `up` (Optional, Number) - Upload traffic counter in bytes.
- `down` (Optional, Number) - Download traffic counter in bytes.
- `total` (Optional, Number) - Total traffic limit in bytes.
- `expiry_time` (Optional, Number) - Expiry time as Unix timestamp in milliseconds.
- `traffic_reset` (Optional, String) - Traffic reset period. Default is `never`.
- `node_id` (Optional, Number) - 3x-ui v3 node ID for multi-node deployments. Leave unset for the local panel. Changing this value recreates the inbound because 3x-ui v3 does not support moving an existing inbound between nodes.
- `restart_xray` (Optional, Boolean) - Restart Xray core after create, update, or delete operations. Default is `false`.
- `sub_sort_index` (Optional, Number) - 1-based sort order of this inbound's links in subscription output (lower first; ties by id). Added in 3x-ui v3.3.1; ignored by older panels.
- `share_addr` (Optional, String) - Share address used in generated subscription links when share_addr_strategy is custom. Added in 3x-ui v3.3.1; ignored by older panels.
- `share_addr_strategy` (Optional, String) - Strategy for the share address in subscription links: `node` (inbound listen/node address), `listen`, or `custom` (uses `share_addr`). Added in 3x-ui v3.3.1; ignored by older panels.

### Per-protocol Settings Blocks

Use the block matching your `protocol`. Only one should be specified.

#### `vless_settings`

- `decryption` (Optional, String) - Decryption method (usually `none`).
- `encryption` (Optional, String) - Encryption method.
- `selected_auth` (Optional, String) - Selected authentication type.
- `fallback` (Optional, Block List) - Fallback destinations.
  - `name` (Optional, String)
  - `alpn` (Optional, String)
  - `path` (Optional, String)
  - `dest` (Optional, String)
  - `xver` (Optional, Number)

#### `trojan_settings`

- `fallback` (Optional, Block List) - Same structure as vless fallback.

#### `shadowsocks_settings`

- `method` (Optional, String) - Encryption method (e.g. `chacha20-ietf-poly1305`, `2022-blake3-aes-256-gcm`). On 3x-ui v2.9.3+ the legacy `aes-128-gcm`/`aes-256-gcm` ciphers were dropped from the xray user switch and silently route through Shadowsocks-2022; pick a chacha20 variant or a `2022-blake3-*` method to stay compatible across the matrix.
- `password` (Optional, String, Sensitive) - Password.
- `network` (Optional, String) - Network type (e.g. `tcp,udp`).
- `iv_check` (Optional, Boolean) - Enable IV check.

#### `http_settings`

- `auth` (Optional, String) - Auth type (e.g. `password`).
- `allow_transparent` (Optional, Boolean) - Allow transparent proxy.
- `account` (Optional, Block List) - Authentication accounts.
  - `user` (Optional, String)
  - `pass` (Optional, String)

#### `socks_settings`

- `auth` (Optional, String) - Auth type.
- `udp` (Optional, Boolean) - Enable UDP.
- `ip` (Optional, String) - IP address.
- `account` (Optional, Block List) - Same structure as http account.

#### `mixed_settings`

Settings for mixed (HTTP+SOCKS) proxy protocol.

- `auth` (Optional, String) - Authentication type (e.g. `password`, `noauth`).
- `udp` (Optional, Boolean) - Enable UDP support.
- `ip` (Optional, String) - IP address for UDP.
- `account` (Optional, Block List) - Authentication accounts. Same structure as http account.

#### `wireguard_settings`

- `mtu` (Optional, List of Number) - MTU values `[IPv4, IPv6]`.
- `secret_key` (Optional, String, Sensitive) - Secret key.
- `no_kernel_tun` (Optional, Boolean) - Disable kernel TUN.
- `gateway` (Optional, List of String) - Gateway addresses.
- `dns` (Optional, List of String) - DNS server addresses.
- `peer` (Optional, Block List) - WireGuard peers.
  - `private_key` (Optional, String, Sensitive)
  - `public_key` (Optional, String)
  - `pre_shared_key` (Optional, String, Sensitive)
  - `allowed_ips` (Optional, List of String)
  - `keep_alive` (Optional, Number)
- `clients` (Optional, Block List) - WireGuard multi-client peers (3x-ui v3.4.2+). Absent/empty on older panels. Each entry is one client device the server accepts, with its own keypair and traffic limits. Use EITHER `clients` OR the legacy `peer` for an inbound, not both — the panel treats them as separate models and populating both yields undefined behavior.
  - `private_key` (Optional, String, Sensitive)
  - `public_key` (Optional, String)
  - `pre_shared_key` (Optional, String, Sensitive)
  - `allowed_ips` (Optional, List of String)
  - `keep_alive` (Optional, Number)
  - `email` (Optional, String) - the panel requires a non-empty unique email (keys traffic counters on it); set it even though it is Optional in the schema.
  - `limit_ip` (Optional, Number)
  - `total_gb` (Optional, Number) - traffic quota in bytes (the field name mirrors the 3x-ui API).
  - `expiry_time` (Optional, Number) - Unix timestamp in milliseconds.
  - `enable` (Optional, Boolean)
  - `tg_id` (Optional, Number)
  - `sub_id` (Optional, String)
  - `comment` (Optional, String)
  - `reset` (Optional, Number) - traffic-counter reset period in days (0 = no periodic reset).

#### `dokodemo_settings`

Used for both `tunnel` and `dokodemo-door` protocols.

- `address` (Optional, String) - Target address.
- `port` (Optional, Number) - Target port.
- `rewrite_address` (Optional, String) - Tunnel rewrite address on 3x-ui v3.0.2+. Mirrored with `address` for older panel compatibility.
- `rewrite_port` (Optional, Number) - Tunnel rewrite port on 3x-ui v3.0.2+. Mirrored with `port` for older panel compatibility.
- `port_map` (Optional, Map of String) - Port mapping.
- `network` (Optional, String) - Network type.
- `allowed_network` (Optional, String) - Tunnel allowed network on 3x-ui v3.0.2+. Mirrored with `network` for older panel compatibility.
- `follow_redirect` (Optional, Boolean) - Follow redirect.

#### `hysteria_settings`

- `version` (Optional, Number) - Hysteria version (1 or 2, default 2).

#### `mtproto_settings` (Optional, Block)

Typed MTProto server settings available on 3x-ui v3.3.0+. On v3.5.0+, per-client FakeTLS `secret` and optional `ad_tag` values are managed by `threexui_inbound_client`; the panel heals their domain suffix from `fake_tls_domain`.

- `fake_tls_domain` (Optional+Computed, String) - FakeTLS domain. The panel default is `www.cloudflare.com`.
- `proxy_protocol_listener` (Optional+Computed, Boolean) - Enable the PROXY protocol listener.
- `prefer_ip` (Optional+Computed, String) - IP preference: `prefer-ipv6`, `prefer-ipv4`, `only-ipv6`, or `only-ipv4`.
- `debug` (Optional+Computed, Boolean) - Enable MTProto debug mode.
- `outbound_tag` (Optional+Computed, String) - Outbound tag used for routing.
- `route_through_xray` (Optional+Computed, Boolean) - Route MTProto traffic through Xray core.
- `route_xray_port` (Optional+Computed, Number) - Xray routing port.
- `public_ipv4` (Optional+Computed, String) - Public IPv4 address advertised by the service.
- `public_ipv6` (Optional+Computed, String) - Public IPv6 address advertised by the service.
- `domain_fronting` (Optional, Block List) - Domain-fronting listeners.

##### `domain_fronting` (Optional, Block List)

- `ip` (Optional+Computed, String) - Fronting IP address.
- `port` (Optional+Computed, Number) - Fronting port.
- `proxy_protocol` (Optional+Computed, Boolean) - Enable PROXY protocol for this listener.

### `stream_settings` Block

- `network` (Optional, String) - Transport network (`tcp`, `ws`, `grpc`, `httpupgrade`, `xhttp`, `kcp`, `hysteria`).
- `security` (Optional, String) - Security type (`none`, `tls`, `reality`).
- `tcp_settings` (Optional, Block) - TCP transport settings.
  - `accept_proxy_protocol` (Optional, Boolean)
  - `header_type` (Optional, String)
- `ws_settings` (Optional, Block) - WebSocket settings.
  - `path` (Optional, String)
  - `headers` (Optional, Map of String)
- `grpc_settings` (Optional, Block) - gRPC settings.
  - `service_name` (Optional, String)
  - `multi_mode` (Optional, Boolean)
  - `idle_timeout` (Optional, Number)
  - `health_check_timeout` (Optional, Number)
  - `permit_without_stream` (Optional, Boolean)
  - `initial_windows_size` (Optional, Number)
- `httpupgrade_settings` (Optional, Block) - HTTP Upgrade settings.
  - `path` (Optional, String)
  - `host` (Optional, String)
- `xhttp_settings` (Optional, Block) - XHTTP settings.
  - `path` (Optional, String)
  - `mode` (Optional, String)
  - `no_sse_header` (Optional, Boolean)
  - `keep_alive_interval` (Optional, Number)
  - `x_padding_bytes` (Optional, String) - xPadding bytes range (e.g. `100-1000`).
  - `x_padding_obfs_mode` (Optional, Boolean) - Enable xPadding obfuscation mode.
  - `x_padding_key` (Optional, String) - xPadding encryption key.
  - `x_padding_header` (Optional, String) - xPadding header name.
  - `x_padding_placement` (Optional, String) - xPadding placement (e.g. `header`, `body`).
  - `x_padding_method` (Optional, String) - xPadding method (e.g. `aes`).
- `kcp_settings` (Optional, Block) - mKCP settings.
  - `mtu` (Optional, Number)
  - `tti` (Optional, Number)
  - `uplink_capacity` (Optional, Number)
  - `downlink_capacity` (Optional, Number)
  - `cwnd_multiplier` (Optional, Number) - CWND multiplier.
  - `max_sending_window` (Optional, Number) - Maximum sending window size.
  - `header_type` (Optional, String)
- `hysteria_settings` (Optional, Block) - Hysteria transport settings.
  - `protocol` (Optional, String) - Hysteria transport protocol.
  - `version` (Optional, Number) - Hysteria version (default 2).
  - `auth` (Optional, String) - Hysteria auth string.
  - `udp_idle_timeout` (Optional, Number) - UDP idle timeout in seconds (default 60).
- `reality_settings` (Optional, Block) - Reality settings.
  - `show` (Optional, Boolean)
  - `xver` (Optional, Number)
  - `target` (Optional, String)
  - `server_names` (Optional, List of String)
  - `private_key` (Optional, String, Sensitive) - Auto-generated if not specified.
  - `min_client_ver` (Optional, String) - Minimum client Xray version the REALITY server accepts, as `major.minor.patch` (e.g. `26.3.27`). Leaving it unset is **not** the same as disabling the gate: Xray 26.7.x substitutes `26.3.27` for an empty value, rejecting clients that report an older or absent version (e.g. sing-box). Set `0.0.0` to remove the lower bound. Each component must be in `0-255`; an empty string is rejected.
  - `max_client_ver` (Optional, String) - Maximum client Xray version the REALITY server accepts, as `major.minor.patch`. Unset means no upper bound.
  - `max_timediff` (Optional, Number) - Maximum allowed time difference with the client, in milliseconds. `0` disables the check.
  - `short_ids` (Optional, List of String, Sensitive) - Auto-generated if not specified.
  - `mldsa65_seed` (Optional, String, Sensitive) - Private ML-DSA-65 seed material.
  - `settings` (Optional, Attribute) - Reality inner settings (client-side). Auto-populated if omitted.
    - `public_key` (Optional, String) - Auto-generated if not specified.
    - `fingerprint` (Optional, String)
    - `server_name` (Optional, String)
    - `spider_x` (Optional, String)
    - `mldsa65_verify` (Optional, String)
- `external_proxy` (Optional, Block List) - External proxy entries.
  - `dest` (Optional, String)
  - `port` (Optional, Number)
  - `remark` (Optional, String)
  - `force_tls` (Optional, String)
- `sockopt` (Optional, Block) - Socket options.
  - `mark` (Optional, Number)
  - `tcp_keep_alive_interval` (Optional, Number)
  - `tcp_no_delay` (Optional, Boolean)
  - `tfo_enable` (Optional, Boolean)
  - `tproxy` (Optional, String)
  - `domain_strategy` (Optional, String)

### `sniffing` Block

- `enabled` (Optional, Boolean) - Whether sniffing is enabled.
- `dest_override` (Optional, List of String) - Destination override protocols (e.g. `http`, `tls`, `quic`, `fakedns`).
- `metadata_only` (Optional, Boolean) - Only sniff metadata.
- `route_only` (Optional, Boolean) - Only use sniffing for routing.
- `ips_excluded` (Optional, List of String) - IPs/CIDRs excluded from sniffing (e.g. `geoip:private`).
- `domains_excluded` (Optional, List of String) - Domains excluded from sniffing (e.g. `domain:example.com`).

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

After import, you only need to specify the fields you want to manage in your configuration.
Server-populated fields (`show`, `xver`, `short_ids`, `fingerprint`, `public_key`, `private_key`,
`spider_x`, `metadata_only`, `route_only`, `encryption`, `selected_auth`, etc.) are automatically
preserved from the imported state — no need to copy them into your `.tf` file.

For example, a minimal configuration for an imported VLESS + Reality inbound:

```hcl
resource "threexui_inbound" "example" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "My VLESS"

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
```
