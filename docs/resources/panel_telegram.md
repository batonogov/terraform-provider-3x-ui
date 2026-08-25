---
page_title: "threexui_panel_telegram Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages Telegram bot settings in the 3x-ui panel.
---

# threexui_panel_telegram (Resource)

Manages the Telegram bot integration settings of the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

~> **Note:** Changing `tg_bot_enable`, `tg_run_time`, `tg_enabled_events`, `tg_cpu` or `tg_memory` triggers a **panel restart** and therefore brief panel downtime. The panel decides once at startup whether to register the periodic stats report and the CPU/memory alarm jobs, and on what schedule, so without the restart the change lands in the database and in Terraform state while the running panel keeps the old schedule — or never starts the job at all. The bot process itself is hot-reloaded, so `tg_bot_token`, `tg_bot_chat_id` and `tg_bot_api_server` take effect immediately and do **not** restart the panel. A restart fires only on an actual value change.

## Example Usage

```hcl
resource "threexui_panel_telegram" "settings" {
  tg_bot_enable       = true
  tg_bot_token        = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
  tg_bot_chat_id      = "-1001234567890"
  tg_lang             = "en-US"
  tg_bot_backup       = true
  tg_bot_login_notify = true
  tg_cpu              = 80
}
```

## Argument Reference

- `tg_bot_enable` (Optional, Boolean) - Enable Telegram bot.
- `tg_bot_token` (Optional, String, Sensitive) - Telegram bot token.
- `tg_bot_token_wo` (Optional, String, WriteOnly) - Write-only version of `tg_bot_token`. Not persisted in state. Terraform 1.11+ / OpenTofu 1.11+.
- `tg_bot_token_wo_version` (Optional, Number) - Increment to trigger re-send of `tg_bot_token_wo`. Must be set together with `tg_bot_token_wo`.
- `tg_bot_proxy` (Optional, String) - Proxy for Telegram bot.
- `tg_bot_api_server` (Optional, String) - Custom Telegram API server URL.
- `tg_bot_chat_id` (Optional, String) - Telegram chat ID for notifications.
- `tg_lang` (Optional, String) - Telegram bot language.
- `tg_run_time` (Optional, String) - Cron expression for periodic reports.
- `tg_bot_backup` (Optional, Boolean) - Enable periodic backup via Telegram.
- `tg_bot_login_notify` (Optional, Boolean, Deprecated) - Enable login notifications via Telegram. **Deprecated:** Removed from 3x-ui v3.4.0; accepted but has no effect on v3.4.0+ panels.
- `tg_cpu` (Optional, Number) - CPU usage threshold for alerts (percentage).
- `tg_enabled_events` (Optional, String) - Comma-separated event types to send via Telegram (e.g. login, backup, traffic threshold). Added in 3x-ui v3.4.0; ignored by older panels.
- `tg_memory` (Optional, Number) - Memory usage threshold (%) for Telegram alerts (0-100). Added in 3x-ui v3.4.0; ignored by older panels.

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_telegram.settings settings
```
