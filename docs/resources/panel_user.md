---
page_title: "threexui_panel_user Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages the admin username and password for the 3x-ui panel.
---

# threexui_panel_user (Resource)

Manages the admin username and password for the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not revert the credentials.

~> **Warning:** After applying changes, keep the provider's steady-state `username` and `password` in sync with this resource. For first-run bootstrap of a fresh panel, the provider-level `bootstrap_username` and `bootstrap_password` arguments can authenticate with the initial credentials once, then this resource can rotate the panel to the steady-state credentials during the same apply.

## Example Usage

```hcl
resource "threexui_panel_user" "admin" {
  username = "myadmin"
  password = "s3cureP@ss"
}
```

## Argument Reference

- `username` (Required, String) - The desired admin username.
- `password` (Required, String, Sensitive) - The desired admin password.

## Attribute Reference

- `id` (String) - Always `"user"`.

## Import

```shell
terraform import threexui_panel_user.admin user
```
