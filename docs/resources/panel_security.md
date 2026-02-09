---
page_title: "threexui_panel_security Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages security settings (2FA) in the 3x-ui panel.
---

# threexui_panel_security (Resource)

Manages the security settings of the 3x-ui panel, specifically two-factor authentication.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

~> **Warning:** Enabling `two_factor_enable` will block the provider from authenticating to the panel, as the provider does not support 2FA codes during login.

## Example Usage

```hcl
resource "threexui_panel_security" "settings" {
  two_factor_enable = false
}
```

## Argument Reference

- `two_factor_enable` (Optional, Boolean) - Enable two-factor authentication.
- `two_factor_token` (Optional, String, Sensitive) - Two-factor authentication token/secret.

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_security.settings settings
```
