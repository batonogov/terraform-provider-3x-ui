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

  stream_settings {
    network  = "tcp"
    security = "reality"

    reality_settings {
      target = "www.apple.com:443"
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

  settings {
    method   = "chacha20-ietf-poly1305"
    password = "my-password"
    network  = "tcp,udp"
  }
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

### settings

Optional block. Protocol-specific settings. At most 1 block.

- `decryption` (Optional, String, Sensitive) - Decryption method (used by vless).
- `encryption` (Optional, String, Sensitive) - Encryption method.
- `selected_auth` (Optional, String) - Selected authentication type for vless.
- `method` (Optional, String) - Encryption method (used by shadowsocks).
- `password` (Optional, String) - Password (used by shadowsocks/trojan).
- `network` (Optional, String) - Network type (e.g. `tcp,udp`).
- `iv_check` (Optional, Boolean) - IV check for shadowsocks.
- `allow_transparent` (Optional, Boolean) - Allow transparent proxy (used by http).
- `auth` (Optional, String) - Auth type for http/socks (e.g. `password`).
- `udp` (Optional, Boolean) - Enable UDP (used by socks).
- `ip` (Optional, String) - IP address (used by socks).
- `address` (Optional, String) - Address for tunnel.
- `port` (Optional, Number) - Port for tunnel.
- `port_map` (Optional, Map of String) - Port mapping for tunnel.
- `follow_redirect` (Optional, Boolean) - Follow redirect for dokodemo-door.
- `mtu` (Optional, Number) - MTU for wireguard.
- `secret_key` (Optional, String) - Secret key for wireguard.
- `no_kernel_tun` (Optional, Boolean) - Disable kernel TUN for wireguard.
- `name` (Optional, String) - Name.
- `user_level` (Optional, Number) - User level.
- `testseed` (Optional, List of Number) - Test seed values (auto-managed).

#### accounts

Optional list block inside `settings`. Used for http/socks auth.

- `user` (Optional, String) - Username.
- `pass` (Optional, String) - Password.

#### fallbacks

Optional list block inside `settings`. Used for vless/trojan fallbacks.

- `name` (Optional, String) - Fallback name.
- `alpn` (Optional, String) - ALPN value.
- `path` (Optional, String) - Path.
- `dest` (Optional, String) - Destination.
- `xver` (Optional, Number) - PROXY protocol version.

#### peers

Optional list block inside `settings`. Used for wireguard.

- `private_key` (Optional, String) - Peer private key.
- `public_key` (Optional, String) - Peer public key.
- `pre_shared_key` (Optional, String) - Pre-shared key.
- `allowed_ips` (Optional, List of String) - Allowed IPs.
- `keep_alive` (Optional, Number) - Keep alive interval in seconds.

### stream_settings

Optional block. Transport/security settings. At most 1 block.

- `network` (Optional, String) - Transport protocol (e.g. `tcp`).
- `security` (Optional, String) - Security type (e.g. `none`, `reality`).

#### external_proxy

Optional list block inside `stream_settings`.

- `dest` (Optional, String) - Destination.
- `port` (Optional, Number) - Port.
- `remark` (Optional, String) - Remark.
- `force_tls` (Optional, String) - Force TLS mode.

#### reality_settings

Optional block inside `stream_settings`. At most 1 block. Keys and short IDs are auto-generated if not provided.

- `show` (Optional, Boolean) - Show option.
- `xver` (Optional, Number) - PROXY protocol version.
- `target` (Optional, String) - Reality target (e.g. `www.apple.com:443`). Auto-set if empty.
- `server_names` (Optional, List of String) - Server names for SNI. Auto-set from target if empty.
- `private_key` (Optional, String, Sensitive) - X25519 private key. Auto-generated if empty.
- `min_client_ver` (Optional, String) - Minimum client version.
- `max_client_ver` (Optional, String) - Maximum client version.
- `max_timediff` (Optional, Number) - Maximum time difference.
- `short_ids` (Optional, List of String) - Short IDs. Auto-generated if empty.
- `mldsa65_seed` (Optional, String) - ML-DSA-65 seed.

##### settings (inside reality_settings)

Optional block. At most 1 block.

- `public_key` (Optional, String, Sensitive) - X25519 public key. Auto-generated if empty.
- `fingerprint` (Optional, String) - TLS fingerprint.
- `server_name` (Optional, String) - Server name.
- `spider_x` (Optional, String) - SpiderX path.
- `mldsa65_verify` (Optional, String) - ML-DSA-65 verify key.

#### tcp_settings

Optional block inside `stream_settings`. At most 1 block.

- `accept_proxy_protocol` (Optional, Boolean) - Accept PROXY protocol.

##### header (inside tcp_settings)

Optional block. At most 1 block.

- `type` (Optional, String) - Header type.

### sniffing

Optional block. Traffic sniffing settings. At most 1 block.

- `enabled` (Optional, Boolean) - Enable sniffing.
- `dest_override` (Optional, List of String) - Destination override protocols (e.g. `http`, `tls`, `quic`, `fakedns`).
- `metadata_only` (Optional, Boolean) - Sniff metadata only.
- `route_only` (Optional, Boolean) - Route only without replacing destination.

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
