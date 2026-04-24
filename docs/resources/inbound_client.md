---
page_title: "threexui_inbound_client Resource - 3x-ui"
subcategory: "Inbound"
description: |-
  Manages a client within an inbound in the 3x-ui panel.
---

# threexui_inbound_client (Resource)

Manages a client within an existing inbound. Clients are users that can connect through the inbound proxy.

## Example Usage

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
      target       = "www.apple.com:443"
      server_names = ["www.apple.com"]
    }
  }
}

resource "threexui_inbound_client" "user1" {
  inbound_id  = threexui_inbound.vless.id
  email       = "user1@example.com"
  enable      = true
  total_gb    = 10
  expiry_time = 1735689600000
  comment     = "Main account"
}
```

### Hysteria Client

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

resource "threexui_inbound_client" "hysteria_user" {
  inbound_id = threexui_inbound.hysteria.id
  email      = "hysteria-user@example.com"
  auth       = "my-secret-auth"
  enable     = true
}
```

## Argument Reference

- `inbound_id` (Required, Number, ForceNew) - The ID of the inbound this client belongs to. Changing this forces a new resource.
- `email` (Required, String) - Unique email identifier for the client. Required by 3x-ui to avoid database errors.
- `client_id` (Optional, String, ForceNew) - UUID for the client. Auto-generated if not provided. Changing this forces a new resource.
- `security` (Optional, String) - Security type.
- `password` (Optional, String, Sensitive) - Client password (used by trojan/shadowsocks).
- `flow` (Optional, String) - Flow control (e.g. `xtls-rprx-vision`).
- `auth` (Optional, String) - Auth password for Hysteria clients. Used as client identifier instead of UUID.
- `limit_ip` (Optional, Number) - Maximum concurrent connections.
- `total_gb` (Optional, Number) - Traffic limit in GB.
- `expiry_time` (Optional, Number) - Expiry time as Unix timestamp in milliseconds.
- `enable` (Optional, Boolean) - Whether the client is enabled.
- `tg_id` (Optional, Number) - Telegram user ID for bot notifications.
- `comment` (Optional, String) - Client description for administrative notes.
- `reset` (Optional, Number) - Traffic reset period in days. `0` means never (default).

## Read-Only Attributes

- `sub_id` (String) - Auto-generated subscription ID used for subscription URLs.

## Usage: Subscription URLs

When the panel subscription service is enabled (`threexui_panel_subscription`),
each client receives an auto-generated `sub_id`. Use it to build subscription
URLs that clients can import into v2rayN, Hiddify, Streisand, or any compatible app:

```hcl
resource "threexui_inbound_client" "user1" {
  inbound_id = threexui_inbound.vless.id
  email      = "user1@example.com"
  enable     = true
  comment    = "Main account"
}

output "user1_subscription_url" {
  value = "https://your-domain.com/sub/${threexui_inbound_client.user1.sub_id}"
}
```

For multiple clients:

```hcl
locals {
  clients = {
    user1 = threexui_inbound_client.user1
    user2 = threexui_inbound_client.user2
  }
}

output "subscription_urls" {
  value = { for name, client in local.clients : name => "https://your-domain.com/sub/${client.sub_id}" }
}
```

> **Note:** The subscription path (`/sub/`) and port (`2096`) depend on your
> `threexui_panel_subscription` configuration. Adjust the URL accordingly.

## Attribute Reference

All arguments are also exported as attributes.

## Import

Inbound clients can be imported using `inbound_id:client_id`:

```shell
terraform import threexui_inbound_client.example 1:550e8400-e29b-41d4-a716-446655440000
```
