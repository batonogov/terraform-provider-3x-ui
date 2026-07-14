---
page_title: "threexui_panel_subscription Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages subscription settings in the 3x-ui panel.
---

# threexui_panel_subscription (Resource)

Manages the subscription service settings of the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

~> **Note:** Changing any subscription **server-binding** field — `sub_enable`, `sub_listen`, `sub_domain`, `sub_port`, `sub_path`, `sub_cert_file`, `sub_key_file` — triggers a panel restart and therefore brief panel downtime. The 3x-ui subscription server is only (re)initialised at panel startup, so changing whether or where it listens needs a restart to take effect; without it the subscription URL keeps returning 404 until the panel is restarted manually. Link-generation and display fields (for example `sub_uri`, `sub_title`, `sub_json_enable`, `sub_encrypt`, `sub_show_info`) are read on every request and do **not** trigger a restart.

~> **Note:** When `sub_port` (default `2096`) differs from the main panel port and the panel runs behind a reverse proxy, the proxy must be configured to forward subscription path requests to the subscription port. Without this, subscription URLs will return 404.

**Caddy** example:

```caddy
handle /sub/* {
    reverse_proxy 3x-ui:2096
}
```

**Nginx** example:

```nginx
location /sub/ {
    proxy_pass http://3x-ui:2096;
}
```

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
- `sub_incy_enable_routing` (Optional, Boolean) - Enable routing injection for the Incy subscription client (3x-ui v3.4.1+).
- `sub_incy_routing_rules` (Optional, String) - Incy routing deep-link injected into the subscription body (3x-ui v3.4.1+).

### Server

- `sub_listen` (Optional, String) - Listen address.
- `sub_port` (Optional, Number) - Subscription server port.
- `sub_path` (Optional, String) - Subscription URL path.
- `sub_domain` (Optional, String) - Subscription domain.
- `sub_cert_file` (Optional, String) - TLS certificate file path.
- `sub_key_file` (Optional, String) - TLS key file path.
- `sub_updates` (Optional, Number) - Update interval in hours.
- `sub_encrypt` (Optional, Boolean) - Encrypt subscription data.
- `sub_show_info` (Optional, Boolean, Deprecated) - Show info in subscription. **Deprecated:** Removed from 3x-ui v3.4.0; accepted but has no effect on v3.4.0+ panels.
- `sub_email_in_remark` (Optional, Boolean, Deprecated) - Include the client email in subscription profile names. Default is `true` on 3x-ui v3.0.2+. **Deprecated:** Removed from 3x-ui v3.4.0; accepted but has no effect on v3.4.0+ panels.
- `sub_theme_dir` (Optional, String) - Absolute path to a folder containing a custom subscription page template. Added in 3x-ui v3.3.0; ignored by older panels.
- `remark_template` (Optional, String) - Subscription remark template (`{{VAR}}` tokens rendered per client, e.g. Jalali date/transport/status). Added in 3x-ui v3.4.0; ignored by older panels.
- `sub_hide_settings` (Optional, Boolean) - Hide server settings in happ subscription (Happ only). Added in 3x-ui v3.4.0; ignored by older panels.

### URI

- `sub_uri` (Optional, String) - Subscription URI.
- `sub_json_path` (Optional, String) - JSON subscription path.
- `sub_json_uri` (Optional, String) - JSON subscription URI.
- `sub_json_fragment` (Optional, String) - JSON fragment settings. The expected format depends on the 3x-ui version:
  - **v2.9.2+:** only the fragment parameters object, e.g. `{"packets":"tlshello","length":"100-200","interval":"10-20","maxSplit":"300-400"}`.
  - **v2.9.1 and earlier:** full outbound object with `tag`, `protocol`, `settings`, and `streamSettings`.
  - **Deprecated in 3x-ui v3.2.8** — replaced by `sub_clash_enable_routing`.
- `sub_json_noises` (Optional, String) - JSON noise settings. The expected format depends on the 3x-ui version:
  - **v2.9.2+:** only the noises array, e.g. `[{"type":"rand","packet":"10-20","delay":"10-16","applyTo":"ip"}]`.
  - **v2.9.1 and earlier:** full outbound object with `tag`, `protocol`, `settings`, and `streamSettings`.
  - **Deprecated in 3x-ui v3.2.8** — replaced by `sub_clash_rules`.
- `sub_json_mux` (Optional, String) - JSON mux settings.
- `sub_json_rules` (Optional, String) - JSON rules.

### Clash / Mihomo

- `sub_clash_enable` (Optional, Boolean) - Enable Clash/Mihomo subscription endpoint.
- `sub_clash_path` (Optional, String) - Path for Clash/Mihomo subscription endpoint.
- `sub_clash_uri` (Optional, String) - Clash/Mihomo subscription server URI.
- `sub_clash_enable_routing` (Optional, Boolean) - Enable global routing rules for Clash/Mihomo subscriptions. Available since 3x-ui v3.2.8.
- `sub_clash_rules` (Optional, String) - Clash/Mihomo global routing rules. Available since 3x-ui v3.2.8.
- `sub_json_final_mask` (Optional, String) - JSON subscription global finalmask (tcp/udp masks and quicParams). Available since 3x-ui v3.2.8.

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_subscription.settings settings
```
