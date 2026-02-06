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

  stream_settings {
    network  = "tcp"
    security = "reality"

    reality_settings {
      target = "www.apple.com:443"
    }
  }
}

resource "threexui_inbound_client" "user1" {
  inbound_id  = threexui_inbound.vless.id
  email       = "user1@example.com"
  enable      = true
  total_gb    = 10
  expiry_time = 1735689600000
}
```

## Argument Reference

- `inbound_id` (Required, Number, ForceNew) - The ID of the inbound this client belongs to. Changing this forces a new resource.
- `email` (Required, String) - Unique email identifier for the client. Required by 3x-ui to avoid database errors.
- `client_id` (Optional, String, ForceNew) - UUID for the client. Auto-generated if not provided. Changing this forces a new resource.
- `security` (Optional, String) - Security type.
- `password` (Optional, String, Sensitive) - Client password (used by trojan/shadowsocks).
- `flow` (Optional, String) - Flow control (e.g. `xtls-rprx-vision`).
- `limit_ip` (Optional, Number) - Maximum concurrent connections.
- `total_gb` (Optional, Number) - Traffic limit in GB.
- `expiry_time` (Optional, Number) - Expiry time as Unix timestamp in milliseconds.
- `enable` (Optional, Boolean) - Whether the client is enabled.
- `tg_id` (Optional, Number) - Telegram user ID for notifications.
- `sub_id` (Optional, String) - Subscription ID.
- `comment` (Optional, String) - Comment.
- `reset` (Optional, Number) - Traffic reset period.

## Attribute Reference

All arguments are also exported as attributes.

## Import

Inbound clients can be imported using `inbound_id:client_id`:

```shell
terraform import threexui_inbound_client.example 1:550e8400-e29b-41d4-a716-446655440000
```
