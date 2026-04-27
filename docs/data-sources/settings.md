---
page_title: "threexui_settings Data Source - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Retrieves all panel settings from the 3x-ui panel.
---

# threexui_settings (Data Source)

Retrieves all current panel settings from the 3x-ui panel as a JSON string. Includes general, security, telegram, and subscription settings.

## Example Usage

```hcl
data "threexui_settings" "all" {}

output "panel_settings" {
  value     = data.threexui_settings.all.json
  sensitive = true
}
```

> **Upgrade note:** Since the provider release that marked `json` as sensitive, any `output` referencing `data.threexui_settings.<name>.json` must declare `sensitive = true`. Without it, `terraform plan` fails with `Output refers to sensitive values`.

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `json` (String, Sensitive) - All panel settings as a JSON string. Marked as sensitive because the payload contains Telegram bot tokens, LDAP passwords, 2FA secrets, and other panel credentials.
