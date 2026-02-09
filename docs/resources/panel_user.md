---
page_title: "threexui_panel_user Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages the admin username and password for the 3x-ui panel.
---

# threexui_panel_user (Resource)

Manages the admin username and password for the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not revert the credentials.

~> **Warning:** After applying changes, update the provider's `username` and `password` to match the new values, otherwise the provider will fail to authenticate on the next run.

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
