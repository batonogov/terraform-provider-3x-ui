---
page_title: "3xui_server_status Data Source"
subcategory: ""
description: |-
  Возвращает текущее состояние сервера и сервиса Xray с панели 3x-ui.
---

# Дата-сорс `3xui_server_status`

Получает данные `/panel/api/server/status`: загрузку CPU, память, swap, диски, версию ядра, uptime и состояние сервиса Xray.

## Пример

```hcl
data "3xui_server_status" "current" {}

output "cpu_usage" {
  value = data.3xui_server_status.current.cpu.current
}
```

## Атрибуты

| Имя | Тип | Описание |
|-----|-----|----------|
| `cpu` | number | Текущая загрузка CPU. |
| `cpu_cores` | number | Количество физических ядер. |
| `logical_processors` | number | Количество логических потоков. |
| `cpu_speed_mhz` | number | Частота CPU в MHz. |
| `mem` | object | Структура `{ current, total }` по оперативной памяти. |
| `swap` | object | `{ current, total }` по swap. |
| `disk` | object | `{ current, total }` по дискам. |
| `xray` | object | Параметры службы Xray (`state`, `error_msg`, `version`). |
| `uptime` | number | Аптайм сервера (секунды). |
| `loads` | list(number) | Список средних нагрузок. |
| `tcp_count` / `udp_count` | number | Текущие TCP/UDP соединения. |
| `net_io` | object | Байты `up` / `down` с момента старта панели. |
| `net_traffic` | object | Байты `sent` / `recv`. |
| `public_ip` | object | Публичные IPv4/IPv6 адреса. |
| `app_stats` | object | Параметры процесса панели (`threads`, `mem`, `uptime`). |

> Точная структура будет синхронизирована с JSON, возвращаемым API. После реализации ресурса документация обновится автоматически командой `tfplugindocs generate`.
