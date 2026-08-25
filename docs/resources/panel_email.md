---
page_title: "threexui_panel_email Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages SMTP/email notification settings in the 3x-ui panel (3x-ui v3.4.0+).
---

# threexui_panel_email (Resource)

Manages the SMTP/email notification settings of the 3x-ui panel (3x-ui v3.4.0+). Older panels ignore these attributes.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

~> **Note:** Changing `smtp_enable`, `smtp_enabled_events`, `smtp_cpu` or `smtp_memory` triggers a **panel restart** and therefore brief panel downtime. Those four decide whether the CPU and memory alarm jobs are registered at all, and the panel makes that decision once at startup — without the restart the setting applies to the database and to Terraform state while no alarm job is running. The SMTP transport settings (`smtp_host`, `smtp_port`, `smtp_username`, `smtp_password`, `smtp_from`, …) are read per message and do **not** restart the panel. A restart fires only on an actual value change.

## Example Usage

```hcl
resource "threexui_panel_email" "settings" {
  smtp_enable          = true
  smtp_host            = "smtp.example.com"
  smtp_port            = 587
  smtp_username        = "alerts@example.com"
  smtp_password_wo     = "supersecret"
  smtp_password_wo_version = 1
  smtp_to              = "admin@example.com"
  smtp_encryption_type = "starttls"
  smtp_enabled_events  = "login,backup"
  smtp_cpu             = 80
  smtp_memory          = 90
}
```

## Argument Reference

- `smtp_enable` (Optional, Boolean) - Enable SMTP email notifications.
- `smtp_host` (Optional, String) - SMTP server host.
- `smtp_port` (Optional, Number) - SMTP server port (1-65535).
- `smtp_username` (Optional, String) - SMTP username.
- `smtp_password` (Optional, String, Sensitive) - SMTP password. Prefer `smtp_password_wo` on Terraform 1.11+ / OpenTofu 1.11+.
- `smtp_password_wo` (Optional, String, WriteOnly) - Write-only version of `smtp_password`. Not persisted in state. Terraform 1.11+ / OpenTofu 1.11+.
- `smtp_password_wo_version` (Optional, Number) - Increment to trigger re-send of `smtp_password_wo`. Must be set together with `smtp_password_wo`.
- `smtp_to` (Optional, String) - Comma-separated recipient email addresses.
- `smtp_from` (Optional, String) - SMTP From address (RFC 5322). Added in 3x-ui v3.6.0; ignored by older panels.
- `smtp_from_name` (Optional, String) - SMTP From display name. Added in 3x-ui v3.6.0; ignored by older panels.
- `smtp_encryption_type` (Optional, String) - SMTP encryption: `none`, `starttls`, or `tls`.
- `smtp_enabled_events` (Optional, String) - Comma-separated event types to send via email.
- `smtp_cpu` (Optional, Number) - CPU usage threshold (%) for email alerts (0-100).
- `smtp_memory` (Optional, Number) - Memory usage threshold (%) for email alerts (0-100).

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_email.settings settings
```
