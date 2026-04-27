---
page_title: "threexui_xray_outbounds Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray outbounds configuration in the 3x-ui panel.
---

# threexui_xray_outbounds (Resource)

Manages the outbounds section of the Xray template configuration. Uses a **set path** strategy -- the provided configuration completely replaces the `outbounds` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the outbounds configuration.

## Example Usage

```hcl
resource "threexui_xray_outbounds" "config" {
  outbound {
    tag      = "direct"
    protocol = "freedom"

    freedom_settings {
      domain_strategy = "AsIs"
    }
  }

  outbound {
    tag      = "blocked"
    protocol = "blackhole"

    blackhole_settings {
      response_type = "none"
    }
  }

  outbound {
    tag      = "proxy-wg"
    protocol = "wireguard"

    wireguard_settings {
      secret_key      = "your-secret-key"
      address         = ["10.0.0.2/32"]
      mtu             = 1420
      workers         = 2
      domain_strategy = "ForceIPv6v4"
      reserved        = [1, 2, 3]
      no_kernel_tun   = false

      peer {
        public_key  = "peer-public-key"
        endpoint    = "engage.cloudflareclient.com:2408"
        allowed_ips = ["0.0.0.0/0", "::/0"]
        keep_alive  = 30
      }
    }

    mux {
      enabled     = false
      concurrency = 8
    }
  }
}
```

## Argument Reference

### outbound (Block, Optional, List)

- `tag` (String, Optional) - Outbound tag name.
- `protocol` (String, Required) - Protocol type (`freedom`, `blackhole`, `dns`, `vmess`, `vless`, `trojan`, `shadowsocks`, `socks`, `http`, `wireguard`, `hysteria`, `hysteria2`).
- `send_through` (String, Optional) - Source IP address to bind.

#### mux (Block, Optional, Max: 1)

- `enabled` (Bool, Optional) - Enable mux.
- `concurrency` (Int, Optional) - Number of concurrent connections.
- `xudp_concurrency` (Int, Optional) - XUDP concurrency.
- `xudp_proxy_udp443` (String, Optional) - XUDP proxy for UDP 443.

### Per-protocol settings

Each outbound should have exactly one `*_settings` block matching its `protocol`.

#### freedom_settings (Block, Optional, Max: 1)

- `domain_strategy` (String, Optional) - Domain strategy (e.g. `AsIs`, `UseIP`).
- `redirect` (String, Optional) - Redirect address.
- `ips_blocked` (List of String, Optional) - List of IPs/CIDRs to block (e.g. `geoip:cn`).

##### fragment (Block, Optional, Max: 1)

- `packets` (String, Optional) - Packet fragmentation mode.
- `length` (String, Optional) - Fragment length range.
- `interval` (String, Optional) - Fragment interval range.

##### noises (Block, Optional, List)

- `type` (String, Optional) - Noise type.
- `packet` (String, Optional) - Noise packet content.
- `delay` (String, Optional) - Noise delay.

#### blackhole_settings (Block, Optional, Max: 1)

- `response_type` (String, Optional) - Response type (`none` or `http`).

#### dns_settings (Block, Optional, Max: 1)

- `network` (String, Optional) - Network type.
- `address` (String, Optional) - DNS server address.
- `port` (Int, Optional) - DNS server port.
- `non_ip_query` (String, Optional) - Non-IP query handling.
- `block_types` (List of Int, Optional) - DNS record types to block.

#### vmess_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - Server address.
- `port` (Int, Optional) - Server port.
- `id` (String, Optional) - VMess user ID (UUID).
- `security` (String, Optional) - Encryption method.

#### vless_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - Server address.
- `port` (Int, Optional) - Server port.
- `id` (String, Optional) - VLESS user ID (UUID).
- `flow` (String, Optional) - Flow control (e.g. `xtls-rprx-vision`).
- `encryption` (String, Optional) - Encryption method.

#### trojan_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - Server address.
- `port` (Int, Optional) - Server port.
- `password` (String, Optional, Sensitive) - Trojan password.

#### shadowsocks_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - Server address.
- `port` (Int, Optional) - Server port.
- `password` (String, Optional, Sensitive) - Shadowsocks password.
- `method` (String, Optional) - Encryption method (e.g. `2022-blake3-aes-128-gcm`).
- `uot` (Bool, Optional) - Enable UDP over TCP.
- `uot_version` (Int, Optional) - UoT version.

#### socks_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - SOCKS proxy address.
- `port` (Int, Optional) - SOCKS proxy port.
- `user` (String, Optional) - Username.
- `pass` (String, Optional, Sensitive) - Password.

#### http_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - HTTP proxy address.
- `port` (Int, Optional) - HTTP proxy port.
- `user` (String, Optional) - Username.
- `pass` (String, Optional, Sensitive) - Password.

#### wireguard_settings (Block, Optional, Max: 1)

- `mtu` (Int, Optional) - MTU size.
- `secret_key` (String, Optional, Sensitive) - WireGuard private key.
- `address` (List of String, Optional) - Local addresses (e.g. `["10.0.0.2/32"]`).
- `workers` (Int, Optional) - Number of worker threads.
- `domain_strategy` (String, Optional) - Domain strategy.
- `reserved` (List of Int, Optional) - Reserved bytes.
- `no_kernel_tun` (Bool, Optional) - Disable kernel TUN.

##### peer (Block, Optional, List)

- `public_key` (String, Optional) - Peer public key.
- `pre_shared_key` (String, Optional, Sensitive) - Pre-shared key.
- `allowed_ips` (List of String, Optional) - Allowed IP ranges.
- `endpoint` (String, Optional) - Peer endpoint (host:port).
- `keep_alive` (Int, Optional) - Keep-alive interval in seconds.

#### hysteria_settings (Block, Optional, Max: 1)

- `address` (String, Optional) - Server address.
- `port` (Int, Optional) - Server port.
- `version` (Int, Optional) - Hysteria version.

## Attribute Reference

- `id` - The resource identifier (`xray_outbounds`).

## Import

```shell
terraform import threexui_xray_outbounds.config xray_outbounds
```
