---
page_title: "Provider: 3x-ui"
subcategory: ""
description: |
  Провайдер 3x-ui позволяет управлять inbound'ами, клиентами и состоянием панели MHSanaei/3x-ui через Terraform/OpenTofu.
---

# Провайдер 3x-ui

Провайдер подключается к панели [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) и автоматизирует управление инбаундами, клиентами, подписками и серверными настройками.

## Пример

```hcl
terraform {
  required_providers {
    threexui = {
      source  = "registry.terraform.io/batonogov/3x-ui"
      version = ">= 0.1.0"
    }
  }
}

provider "3xui" {
  base_url        = "https://localhost:2053"
  username        = var.threexui_username
  password        = var.threexui_password
  tls_skip_verify = true
}
```

## Настройка

См. `docs/provider-config.md` для полного описания параметров. Кратко:
- `base_url` — адрес панели.
- `username`/`password` — учётные данные (или используйте `session_cookie`).
- `tls_skip_verify`, `request_timeout`, `max_retries` — дополнительные настройки клиента.

## Ресурсы и дата-сорсы

Ресурсы:
- [`3xui_inbound`](resources/inbound.md)
- [`3xui_user`](resources/user.md) *(планируется)*
- `3xui_subscription`, `3xui_panel_setting`, `3xui_cron_job` *(будут добавлены позже)*

Дата-сорсы:
- [`3xui_server_status`](data-sources/server_status.md)
- `3xui_inbounds`, `3xui_usage_stats`, `3xui_raw` *(в разработке)*

## Генерация документации

Документация поддерживает структуру `tfplugindocs`. Команду `tfplugindocs generate` можно запускать после реализации провайдера, чтобы синхронизировать файлы в `docs/` и примеры в `examples/`.
