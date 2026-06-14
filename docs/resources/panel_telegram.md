---
page_title: "threexui_panel_telegram Resource - 3x-ui"
subcategory: "Panel Settings"
description: |-
  Manages Telegram bot settings in the 3x-ui panel.
---

# threexui_panel_telegram (Resource)

Manages the Telegram bot integration settings of the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; it does not reset the settings.

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
- `tg_bot_login_notify` (Optional, Boolean) - Enable login notifications via Telegram.
- `tg_cpu` (Optional, Number) - CPU usage threshold for alerts (percentage).

## Attribute Reference

All arguments are also exported as attributes.

## Import

```shell
terraform import threexui_panel_telegram.settings settings
```
