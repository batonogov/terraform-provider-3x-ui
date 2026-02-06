# CLAUDE.md

Правила для агентов, работающих в этом репозитории.

## Цель

Terraform-провайдер для панели [3x-ui](https://github.com/MHSanaei/3x-ui) (Go, terraform-plugin-sdk/v2).

## Структура проекта

```
provider/              — весь код провайдера
  provider.go          — регистрация ресурсов и data sources
  client.go            — HTTP-клиент к 3x-ui API (cookie auth, auto re-login)
  types.go             — Inbound, ClientTraffic, APIResponse, ParseJSONField
  resource_inbound.go  — ресурс threexui_inbound (CRUD, Reality, settings defaults)
  resource_inbound_client.go — ресурс threexui_inbound_client (мьютекс, UUID)
  resource_settings_tabs.go  — panel_general/security/telegram/subscription
  resource_xray_settings.go  — xray_basics/dns/routing/balancers/reverse/outbounds/advanced
  settings.go          — settings schema, build/flatten JSON, expand/flatten clients/fallbacks/peers
  stream_settings.go   — stream_settings schema (tcp_settings, reality_settings, external_proxy)
  sniffing.go          — sniffing schema, build/flatten
  settings_helpers.go  — merge, getField helpers, port validation
  default_settings.go  — дефолтные settings по протоколу, applyDefaultInboundSettings
  data_source_*.go     — data sources (inbounds, server_status, settings, xray_config, xray_versions)
examples/              — примеры TF-конфигов для ручного тестирования
test_plans/            — тест-планы для каждого ресурса
3x-ui-2.8.9/          — исходники 3x-ui v2.8.9 (в .gitignore, для справки)
docker-compose.yaml    — 3x-ui v2.8.9 на порту 2053
Taskfile.yml           — task build / test / fmt
```

## Ресурсы провайдера

| Terraform-ресурс | Файл | Описание |
|---|---|---|
| `threexui_inbound` | resource_inbound.go | Inbound (vless/vmess/trojan/ss/http/mixed/wg/tunnel) |
| `threexui_inbound_client` | resource_inbound_client.go | Клиент внутри inbound |
| `threexui_panel_general` | resource_settings_tabs.go | Настройки панели (web, LDAP) |
| `threexui_panel_security` | resource_settings_tabs.go | 2FA |
| `threexui_panel_telegram` | resource_settings_tabs.go | Telegram-бот |
| `threexui_panel_subscription` | resource_settings_tabs.go | Подписки |
| `threexui_xray_basics` | resource_xray_settings.go | Базовый Xray-конфиг (merge root) |
| `threexui_xray_dns` | resource_xray_settings.go | DNS (set path) |
| `threexui_xray_routing` | resource_xray_settings.go | Маршрутизация (set path) |
| `threexui_xray_balancers` | resource_xray_settings.go | Балансировщики (set path) |
| `threexui_xray_reverse` | resource_xray_settings.go | Reverse proxy (set path) |
| `threexui_xray_outbounds` | resource_xray_settings.go | Outbound'ы (set path) |
| `threexui_xray_advanced` | resource_xray_settings.go | Полная замена конфига (replace all) |

## Data Sources

| Terraform data source | Описание |
|---|---|
| `threexui_inbounds` | Список всех inbound'ов |
| `threexui_server_status` | Статус сервера (JSON) |
| `threexui_xray_versions` | Доступные версии Xray |
| `threexui_xray_config` | Текущий Xray-конфиг (JSON) |
| `threexui_settings` | Все настройки панели (JSON) |

## API 3x-ui (ключевые эндпоинты)

- `POST /login` — авторизация (form: username, password, twoFactorCode)
- `GET /panel/api/inbounds/list` — все inbound'ы
- `GET /panel/api/inbounds/get/:id` — один inbound
- `POST /panel/api/inbounds/add` — создать (form-encoded)
- `POST /panel/api/inbounds/update/:id` — обновить
- `POST /panel/api/inbounds/del/:id` — удалить
- `POST /panel/api/inbounds/addClient` — добавить клиента
- `POST /panel/api/inbounds/updateClient/:clientId` — обновить клиента
- `POST /panel/api/inbounds/:id/delClient/:clientId` — удалить клиента
- `POST /panel/setting/all` — все настройки
- `POST /panel/setting/update` — обновить настройки (JSON body)
- `POST /panel/xray` — Xray template (xraySetting)
- `POST /panel/xray/update` — обновить Xray template

Неавторизованные запросы возвращают 404 (не 401). Клиент делает auto re-login при 401/404.

## Важные особенности кода

### Inbound / Client
- `settings`, `stream_settings`, `sniffing` — JSON-строки в API, но structured blocks в TF schema
- `stream_settings` поддерживает **только**: tcp_settings, reality_settings, external_proxy (ws/grpc/h2/quic/kcp НЕ реализованы)
- `preserveInboundSettings` — при update сохраняет clients и testseed из existing inbound
- `ensureRealityKeys` — автогенерация private/public key и short_ids
- `ensureInboundClientIDs` — автогенерация UUID для клиентов без id
- `applyDefaultInboundSettings` — дефолтные settings по протоколу (vless: decryption=none, testseed)
- `inboundClientMu` — мьютекс для конкурентных операций с клиентами
- `email` в `threexui_inbound_client` — **Required** (без email 3x-ui падает с SQL error при добавлении следующего клиента)
- `jsonSubsetDiffSuppress` / `isSubset` — используется для settings и merge root, подавляет diff если config подмножество state

### Panel Settings
- Settings-ресурсы — синглтоны (ID = `"settings"`), один экземпляр на тип
- `resourceSettingsDelete` — только очищает TF state, **не** сбрасывает настройки в API
- `resourceSubscriptionSettingsApply` — делает двойной apply (обходит баг 3x-ui: sub_json_enable не сохраняется при первом apply совместно с sub_enable)
- Включение 2FA (`two_factor_enable`) блокирует провайдер (login не поддерживает 2FA-код) — добавлен Warning
- Изменение `web_base_path` требует обновления `base_path` в provider config — добавлен Warning
- `panelSettingsNeedRestart` — ключи: webListen, webDomain, webPort, webBasePath, webCertFile, webKeyFile, sessionMaxAge

### Xray Settings
- Xray-ресурсы работают в 3 режимах: merge root, set path, replace all
- `xrayTemplateMu` — мьютекс для сериализации read-modify-write на xray template (предотвращает race condition)
- Merge root (`xray_basics`) использует `jsonSubsetDiffSuppress` — state содержит полный конфиг, но diff подавляется если config пользователя является подмножеством
- SetPath / ReplaceAll используют `jsonEqualDiffSuppress` — точное сравнение JSON
- Delete для xray-ресурсов — только очищает TF state, не сбрасывает xray-конфиг

## Команды

```bash
task build       # Собрать бинарь
task test        # Запустить acceptance-тесты (нужен docker)
task fmt         # gofmt
task vet         # go vet
task lint        # golangci-lint (не запускается автоматически в pre-commit)
task pre-commit  # Запустить все pre-commit проверки вручную
```

## Pre-commit hooks

В проекте настроены автоматические проверки перед коммитом:
- **go-fmt** — форматирование кода
- **go-vet** — статический анализ
- **go-build** — проверка компиляции
- Проверки YAML/JSON, trailing whitespace, EOF

**golangci-lint** намеренно НЕ включён в pre-commit (медленный), но рекомендуется запускать вручную перед PR: `task lint`.

Конфигурации: `.pre-commit-config.yaml`, `.golangci.yml`

## Тестовое окружение

```bash
docker compose up -d   # Запуск 3x-ui v2.8.9 на localhost:2053
# Логин: admin / admin
# Docker image v2.8.9 по умолчанию webBasePath = /panel/
# provider.tf: base_path должен совпадать с webBasePath панели
```

## Основные принципы

- Действуй прагматично: сначала понять задачу, затем изменить минимально необходимое.
- Не ломать обратную совместимость без явного запроса.
- Сохранять стиль кода и структуру проекта.
- Изменения делай точечно, избегай массовых переформатирований.
- После изменения кода запускай `task build`.
- Пиши кратко и по делу. Указывай, какие файлы изменены.
