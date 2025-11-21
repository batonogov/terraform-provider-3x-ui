---
page_title: "3xui_inbound Resource"
subcategory: ""
description: |-
  Создаёт/управляет inbound-конфигурациями в панели 3x-ui.
---

# Ресурс `3xui_inbound`

> ⚠️ Эта страница описывает планируемую схему. Реализация появится после завершения шага 4.

Позволяет управлять inbound-записями панели. Каждая запись определяет порт, протокол и параметры транспорта (stream settings).

## Пример

```hcl
resource "3xui_inbound" "example" {
  protocol = "vless"
  listen   = "0.0.0.0"
  port     = 443
  remark   = "prod"
  enable   = true

  stream {
    network  = "tcp"
    security = "tls"
    tls {
      sni       = "example.com"
      cert_path = "/etc/ssl/example.crt"
      key_path  = "/etc/ssl/example.key"
    }
  }

  settings_vless {
    clients = [{
      id    = "UUID"
      email = "user@example.com"
    }]
  }
}
```

## Аргументы

| Имя | Тип | Обязательный | Описание |
|-----|-----|--------------|----------|
| `protocol` | string | да | Протокол inbound (`vless`, `vmess`, `trojan`, `shadowsocks`, ...). |
| `listen` | string | нет | Адрес прослушивания (по умолчанию `0.0.0.0`). |
| `port` | number | да | Порт 1-65535. |
| `remark` | string | нет | Комментарий. |
| `enable` | bool | нет | Включить/отключить inbound. |
| `stream` | block | да | Сетевые настройки (см. ниже). |
| `settings_*` | блоки | зависят от `protocol` | Параметры конкретного протокола (например, `settings_vless`). |

### Блок `stream`

| Имя | Тип | Обязательный | Описание |
|-----|-----|--------------|----------|
| `network` | string | да | Тип транспорта (`tcp`, `ws`, `grpc`, ...). |
| `security` | string | нет | `none`, `tls`, `reality`. |
| `tls` | block | нет | TLS-параметры. |
| `reality` | block | нет | Параметры Reality. |
| `ws`/`grpc` | block | нет | Настройки конкретного транспорта. |

## Атрибуты

- `id` — числовой идентификатор inbound в БД 3x-ui.
- `tag` — автоматически вычисляемый тег (например, `inbound-443`).
- `up`, `down`, `total` — статистика трафика.

## Импорт

```bash
terraform import 3xui_inbound.example 123
```
где `123` — ID inbound в панели.
