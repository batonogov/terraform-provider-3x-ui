# Ресурсы и дата-сорсы провайдера 3x-ui

## Ресурсы

### `3xui_inbound`
- **Описание**: управляет inbound-записями (порт, протокол, stream settings, clients).
- **API**: `/panel/api/inbounds/add`, `/update/:id`, `/del/:id`, `/get/:id`.
- **Ключевые поля**:
  - `protocol` (enum: vless, vmess, trojan, shadowsocks, dokodemo и т.д.).
  - `listen`, `port`, `tag`, `remark`, `enable`.
  - `settings` (jsonobject → отдельные вложенные блоки по протоколу).
  - `stream` (transport settings: `network`, `security`, `tls`, `reality`, `ws`, `grpc`, ...).
  - `clients` (вложенный блок для предсоздания пользователей).
  - `sniffing`, `allocate`, `fallbacks`, `ruleList`.
- **Возможности**: CRUD, импорт по ID (внутренний `db` ID). Дополнительно методы для ресета трафика будут вынесены в отдельные действия (через `CustomizeDiff` или отдельные ресурсы).

### `3xui_user`
- **Описание**: аккаунты панели (`/panel/api/inbounds/addClient`/`updateClient`). В 3x-ui пользователи связаны с inbound-клиентами; ресурс позволит управлять клиентами внутри конкретного inbound.
- **API**: `/panel/api/inbounds/addClient`, `/updateClient/:clientId`, `/:inboundId/delClient/:clientId`, `/:inboundId/delClientByEmail/:email`.
- **Схема**:
  - `inbound_id` (int, required).
  - `email` (string, уникальный для клиента).
  - `enable`, `total_up`, `total_down`, `expiry_time`.
  - `settings`/`flow`/`limit_ip` и др.
- **Импорт**: комбинация `<inbound_id>/<client_id>`.

### `3xui_subscription`
- **Описание**: управление ссылками подписки (в 3x-ui реализовано через `sub` пакет). Планируется ресурс, создающий публичные subscription entries для клиентов.
- **API**: REST эндпоинты находятся в `web/controller/subscription.go` (надо уточнить после анализа). Ресурс будет включать `title`, `prefix`, `expire_days`, `limit_ips` и т.д.

### `3xui_panel_setting`
- **Описание**: глобальные настройки панели (SMTP, Telegram bot, LDAP, brand). Пока нет публичного API, но через `web/html/settings.html` видно, что используется `/panel/api/setting/save`. Потребуется изучить контроллер (вероятно `web/controller/setting.go`).
- **Схема**: набор параметров из `entity.Setting` (панель, Xray, уведомления).

### `3xui_cron_job`
- **Описание**: задания для сброса трафика/отправки отчётов (см. `web/job/*`). Возможно придётся задействовать методы панели (например, `resetAllTraffics`). Альтернативно, этот ресурс может представлять планировщик (если API нет — отложим).

### Дополняющие ресурсы
- `3xui_server_geofile` — управление обновлением geo файлов (`POST /panel/api/server/updateGeofile`).
- `3xui_server_logs` — ресурс/действие для экспорта логов (скорее data source).

## Дата-сорсы

### `3xui_inbounds`
- Список inbound'ов с фильтрами (`protocol`, `port`, `tag`). Использует `GET /panel/api/inbounds/list` и фильтрует на стороне провайдера.

### `3xui_inbound`
- Получение одного inbound по ID или тегу.

### `3xui_user`
- Поиск клиента по email внутри inbound.

### `3xui_usage_stats`
- Данные о трафике (использует `/panel/api/inbounds/getClientTraffics` и `/onlines`). Возвращает up/down, последний онлайн, текущие подключения.

### `3xui_server_status`
- Обёртка над `/panel/api/server/status`, выдаёт CPU, память, диски, uptime.

### `3xui_raw`
- Универсальный дата-сорс для обращения к неподдерживаемым эндпоинтам: пользователь задаёт `path`, `method`, `body`, получает JSON. Позволит пользователям покрывать edge-кейсы без правок провайдера.

## План реализации
1. Начать с `3xui_inbound` (наибольший эффект).
2. Добавить `3xui_user` как зависимый ресурс (операции с клиентами).
3. Реализовать `3xui_server_status`/`3xui_usage_stats` как дата-сорсы (для мониторинга).
4. Остальные сущности — по мере необходимости (подписки, панельные настройки).
