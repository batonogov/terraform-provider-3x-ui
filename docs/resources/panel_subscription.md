---
page_title: "threexui_panel_subscription Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages subscription settings in the 3x-ui panel.
---

# threexui_panel_subscription (Resource)

Manages the subscription service settings of the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

## Example Usage

```hcl
resource "threexui_panel_subscription" "settings" {
  sub_enable      = true
  sub_json_enable = true
  sub_port        = 2096
  sub_path        = "/sub/"
  sub_domain      = "sub.example.com"
}
```

## Argument Reference

### General

- `sub_enable` (Optional, Boolean) - Enable subscription service.
- `sub_json_enable` (Optional, Boolean) - Enable JSON subscription format.
- `sub_title` (Optional, String) - Subscription title.
- `sub_support_url` (Optional, String) - Support URL shown to clients.
- `sub_profile_url` (Optional, String) - Profile URL.
- `sub_announce` (Optional, String) - Announcement text.

### Routing

- `sub_enable_routing` (Optional, Boolean) - Enable routing in subscriptions.
- `sub_routing_rules` (Optional, String) - Routing rules for subscriptions.

### Server

- `sub_listen` (Optional, String) - Listen address.
- `sub_port` (Optional, Number) - Subscription server port.
- `sub_path` (Optional, String) - Subscription URL path.
- `sub_domain` (Optional, String) - Subscription domain.
- `sub_cert_file` (Optional, String) - TLS certificate file path.
- `sub_key_file` (Optional, String) - TLS key file path.
- `sub_updates` (Optional, Number) - Update interval in hours.
- `sub_encrypt` (Optional, Boolean) - Encrypt subscription data.
- `sub_show_info` (Optional, Boolean) - Show info in subscription.

### URI

- `sub_uri` (Optional, String) - Subscription URI.
- `sub_json_path` (Optional, String) - JSON subscription path.
- `sub_json_uri` (Optional, String) - JSON subscription URI.
- `sub_json_fragment` (Optional, String) - JSON fragment settings.
- `sub_json_noises` (Optional, String) - JSON noise settings.
- `sub_json_mux` (Optional, String) - JSON mux settings.
- `sub_json_rules` (Optional, String) - JSON rules.

### Clash / Mihomo

- `sub_clash_enable` (Optional, Boolean) - Enable Clash/Mihomo subscription endpoint.
- `sub_clash_path` (Optional, String) - Path for Clash/Mihomo subscription endpoint.
- `sub_clash_uri` (Optional, String) - Clash/Mihomo subscription server URI.

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_subscription.settings settings
```
