# CLAUDE.md

Правила для агентов, работающих в этом репозитории.

## Цель

Terraform-провайдер для панели [3x-ui](https://github.com/MHSanaei/3x-ui) (Go, terraform-plugin-framework).

## Структура проекта

```
provider/              — весь код провайдера
  provider.go          — ThreeXUIProvider (framework): Metadata, Schema, Configure, Resources, DataSources
  client.go            — HTTP-клиент к 3x-ui API (cookie auth, auto re-login)
  types.go             — Inbound, ClientTraffic, APIResponse, ParseJSONField
  resource_inbound.go  — ресурс threexui_inbound (CRUD, Reality, settings defaults)
  resource_inbound_client.go — ресурс threexui_inbound_client (мьютекс, UUID)
  resource_settings_tabs.go  — panel_general/security/telegram/subscription (typed атрибуты)
  resource_panel_user.go     — ресурс threexui_panel_user (смена логина/пароля админа)
  resource_xray_settings.go  — CRUD для xray_basics/dns/routing/balancers/reverse/outbounds (typed атрибуты)
  xray_basics_schema.go      — модель, схема, expand/flatten для xray_basics (log, policy, api, stats)
  xray_dns_schema.go         — модель, схема, expand/flatten для xray_dns (servers, hosts)
  xray_routing_schema.go     — модель, схема, expand/flatten для xray_routing (rules)
  xray_balancers_schema.go   — модель, схема, expand/flatten для xray_balancers
  xray_reverse_schema.go     — модель, схема, expand/flatten для xray_reverse (bridges, portals)
  xray_outbounds_schema.go   — модель, схема, expand/flatten для xray_outbounds (per-protocol settings)
  inbound_settings_schema.go      — модель, схема, expand/flatten для per-protocol settings (vless, trojan, ss, http, socks, wg, dokodemo)
  inbound_stream_settings_schema.go — модель, схема, expand/flatten для stream_settings (tcp, ws, grpc, httpupgrade, xhttp, kcp, reality, sockopt)
  inbound_sniffing_schema.go      — модель, схема, expand/flatten для sniffing
  settings.go          — buildSettingsJSON(map[string]any), flattenSettings(string), expand/flatten clients/fallbacks/peers
  stream_settings.go   — buildStreamSettingsJSON(map[string]any), flattenStreamSettings(string), expand/flatten per-transport
  sniffing.go          — buildSniffingJSON(map[string]any), flattenSniffing(string)
  settings_helpers.go  — mergeSettings
  default_settings.go  — дефолтные settings по протоколу, applyDefaultInboundSettings
  data_source_*.go     — data sources (inbounds, server_status, settings, xray_config, xray_versions)
examples/              — примеры TF-конфигов для ручного тестирования
3x-ui-2.8.9/          — исходники 3x-ui v2.8.9 (в .gitignore, для справки)
docker-compose.yaml    — 3x-ui v2.8.9 на порту 2053
Taskfile.yml           — task build / test / fmt
```

## Ресурсы провайдера

| Terraform-ресурс | Файл | Описание |
|---|---|---|
| `threexui_inbound` | resource_inbound.go + inbound_*_schema.go | Inbound (vless/vmess/trojan/ss/http/mixed/wg/tunnel). Typed блоки для settings/stream_settings/sniffing |
| `threexui_inbound_client` | resource_inbound_client.go | Клиент внутри inbound. Typed атрибуты |
| `threexui_panel_general` | resource_settings_tabs.go | Настройки панели (web, LDAP). Typed атрибуты |
| `threexui_panel_security` | resource_settings_tabs.go | 2FA. Typed атрибуты |
| `threexui_panel_user` | resource_panel_user.go | Смена логина/пароля админа. Write-only (нет read API) |
| `threexui_panel_telegram` | resource_settings_tabs.go | Telegram-бот. Typed атрибуты |
| `threexui_panel_subscription` | resource_settings_tabs.go | Подписки. Typed атрибуты |
| `threexui_xray_basics` | resource_xray_settings.go + xray_basics_schema.go | Базовый Xray-конфиг (merge root). Typed блоки |
| `threexui_xray_dns` | resource_xray_settings.go + xray_dns_schema.go | DNS (set path). Typed блоки |
| `threexui_xray_routing` | resource_xray_settings.go + xray_routing_schema.go | Маршрутизация (set path). Typed блоки |
| `threexui_xray_balancers` | resource_xray_settings.go + xray_balancers_schema.go | Балансировщики (set path). Typed блоки |
| `threexui_xray_reverse` | resource_xray_settings.go + xray_reverse_schema.go | Reverse proxy (set path). Typed блоки |
| `threexui_xray_outbounds` | resource_xray_settings.go + xray_outbounds_schema.go | Outbound'ы (set path). Typed блоки |

## Data Sources

| Terraform data source | Описание |
|---|---|
| `threexui_inbounds` | Список всех inbound'ов (JSON-строка) |
| `threexui_server_status` | Статус сервера (JSON) |
| `threexui_xray_versions` | Доступные версии Xray (list of strings) |
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
- `POST /panel/setting/updateUser` — сменить логин/пароль админа (JSON: oldUsername, oldPassword, newUsername, newPassword)
- `POST /panel/xray` — Xray template (xraySetting)
- `POST /panel/xray/update` — обновить Xray template

Неавторизованные запросы возвращают 404 (не 401). Клиент делает auto re-login при 401/404.

## Важные особенности кода

### Framework (terraform-plugin-framework)
- Провайдер: `ThreeXUIProvider` реализует `provider.Provider` (Metadata, Schema, Configure, Resources, DataSources)
- Фабрика: `New() provider.Provider`
- Ресурсы реализуют `resource.Resource` + `resource.ResourceWithImportState`
- Data sources реализуют `datasource.DataSource`
- Модели используют `types.String`, `types.Int64`, `types.Bool` с тегами `tfsdk:"..."`
- Plan modifiers: `stringplanmodifier.RequiresReplace()`, `int64planmodifier.RequiresReplace()`
- Defaults: `booldefault.StaticBool()`, `stringdefault.StaticString()`
- Import: `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`

### Inbound / Client
- `settings`, `stream_settings`, `sniffing` — JSON-строки в API, typed блоки в TF schema
- Трёхслойная конвертация: Typed Model ↔ Untyped Map (expand/flatten*FromModel/*ToModel) ↔ JSON String (build*/flatten*)
- Per-protocol settings блоки: `vless_settings`, `trojan_settings`, `shadowsocks_settings`, `http_settings`, `socks_settings`, `wireguard_settings`, `dokodemo_settings`
- stream_settings поддерживает транспорты: tcp, ws, grpc, httpupgrade, xhttp, kcp + reality, sockopt, external_proxy
- `alignBlocksWithPlan` — предотвращает ошибки "was absent, but now present" для Optional блоков (Create/Read/Update); пропускается при Import (detect: `state.Protocol.IsNull()`)
- `preserveInboundSettings` — при update сохраняет clients и testseed из existing inbound
- `ensureRealityKeys` — автогенерация private/public key и short_ids
- `ensureInboundClientIDs` — автогенерация UUID для клиентов без id
- `applyDefaultInboundSettings` — дефолтные settings по протоколу (vless: decryption=none, testseed)
- `inboundClientMu` — мьютекс для конкурентных операций с клиентами
- `email` в `threexui_inbound_client` — **Required** (без email 3x-ui падает с SQL error при добавлении следующего клиента)
- `isSubset` — standalone утилита для проверки подмножества JSON

### Panel Settings
- Settings-ресурсы — синглтоны (ID = `"settings"`), один экземпляр на тип
- Typed атрибуты (Optional + Computed + UseStateForUnknown) — каждое поле отдельный атрибут в schema
- Per-resource модели: `PanelGeneralModel`, `PanelSecurityModel`, `PanelTelegramModel`, `PanelSubscriptionModel`
- `settingsApplyTyped` / `settingsReadTyped` — shared CRUD логика (expand model → API → flatten → model)
- Delete только очищает TF state, **не** сбрасывает настройки в API
- Subscription resource делает двойной apply (обходит баг 3x-ui: sub_json_enable не сохраняется при первом apply совместно с sub_enable)
- Включение 2FA блокирует провайдер (login не поддерживает 2FA-код) — добавлен Warning
- Изменение `web_base_path` требует обновления `base_path` в provider config — добавлен Warning
- `panelSettingsNeedRestart` — ключи: webListen, webDomain, webPort, webBasePath, webCertFile, webKeyFile, sessionMaxAge

### Panel User
- `threexui_panel_user` — синглтон (ID = `"user"`), управляет admin credentials
- Write-only: нет API для чтения username/password, Read — no-op (state preserved)
- Create использует `r.client.username/password` как old credentials
- Update использует предыдущий state как old credentials
- После успешного UpdateUser клиент обновляет свои хранимые credentials для последующих запросов
- Delete только очищает TF state, credentials на панели не откатываются
- Warning напоминает обновить provider config после смены credentials

### Xray Settings
- Typed блоки (ListNestedBlock) — каждый ресурс имеет свою модель и schema в `*_schema.go`
- Per-resource модели: `XrayBasicsModel`, `XrayDNSModel`, `XrayRoutingModel`, `XrayBalancersModel`, `XrayReverseModel`, `XrayOutboundsModel`
- Двухслойная конвертация: typed model ↔ untyped map (expand/flatten) ↔ Xray JSON (build/flattenToMap)
- Xray-ресурсы работают в 2 режимах: merge root (`xray_basics`), set path (остальные)
- `xrayTemplateMu` — мьютекс для сериализации read-modify-write на xray template (предотвращает race condition)
- `xrayApplyTyped` / `xrayReadSection` — shared CRUD логика
- CRUD: plan.Get → expand → build → xrayApplyTyped → xrayReadSection → flattenToMap → flatten → state.Set
- DNS servers: address-only → сериализуется как строка в JSON, с доп. полями → как объект
- Outbound settings: per-protocol блоки (`freedom_settings`, `blackhole_settings`, ...) определяются значением `protocol`
- Policy levels: в Xray JSON map `{"0": {...}}`, в TF list `[{id=0, ...}]`
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

**golangci-lint** включён в pre-commit и CI.

Конфигурации: `.pre-commit-config.yaml`, `.golangci.yml`

## Тестовое окружение

```bash
task test              # Полный цикл: docker up, acc-тесты (OpenTofu), docker down
docker compose up -d   # Запуск 3x-ui v2.8.9 на localhost:2053
# Логин: admin / admin
# Docker image v2.8.9 по умолчанию webBasePath = / (НЕ /panel/)
# Не задавать THREEXUI_BASE_PATH
```

Acc-тесты используют `terraform-plugin-testing`:
- `testAccProtoV6ProviderFactories()` — возвращает `map[string]func() (tfprotov6.ProviderServer, error)`
- `ProtoV6ProviderFactories` в TestCase (не `ProviderFactories`)
- HCL-конфиги используют typed блоки и атрибуты (не `jsonencode()`)

Acc-тесты требуют OpenTofu и переменные окружения для корректного provider namespace:
- `TF_ACC_TERRAFORM_PATH` — абсолютный путь к `tofu`
- `TF_ACC_PROVIDER_NAMESPACE=batonogov`
- `TF_ACC_PROVIDER_HOST=registry.opentofu.org`

Всё это уже настроено в `Taskfile.yml` → `task test`.

## Основные принципы

- Действуй прагматично: сначала понять задачу, затем изменить минимально необходимое.
- Не ломать обратную совместимость без явного запроса.
- Сохранять стиль кода и структуру проекта.
- Изменения делай точечно, избегай массовых переформатирований.
- После изменения кода запускай `task build`.
- Пиши кратко и по делу. Указывай, какие файлы изменены.
