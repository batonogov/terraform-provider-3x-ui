# План работ для Terraform/OpenTofu провайдера 3x-ui

Актуальная версия исходников: `v2.8.9` (detached HEAD).
План основан на исходниках веб‑части 3x-ui в `3x-ui/` и реализованных там API‑роутах.

## Поверхность API (по коду)

Базовый путь: настройка `webBasePath` (по умолчанию `/`). Все маршруты идут от базового пути.
Аутентификация: cookie‑сессия `3x-ui`, получаемая через `POST {basePath}login`.
Ответы: JSON `{"success":bool,"msg":string,"obj":any}`.

### Auth
- `POST {basePath}login` (form/json: `username`, `password`, опц. `twoFactorCode`)
- `GET {basePath}logout`

### Inbounds API (нужен логин; при отсутствии — 404)
`{basePath}panel/api/inbounds`
- `GET /list`
- `GET /get/:id`
- `GET /getClientTraffics/:email`
- `GET /getClientTrafficsById/:id`
- `POST /add` (form: поля inbound; `settings`, `streamSettings`, `sniffing` — JSON‑строки)
- `POST /del/:id`
- `POST /update/:id` (form: поля inbound)
- `POST /clientIps/:email`
- `POST /clearClientIps/:email`
- `POST /addClient` (form: `id`, `settings` с `{"clients":[...]}`)
- `POST /:id/delClient/:clientId`
- `POST /updateClient/:clientId` (form: `id`, `settings` с одним клиентом)
- `POST /:id/resetClientTraffic/:email`
- `POST /resetAllTraffics`
- `POST /resetAllClientTraffics/:id`
- `POST /delDepletedClients/:id`
- `POST /import` (form: `data` = inbound JSON)
- `POST /onlines`
- `POST /lastOnline`
- `POST /updateClientTraffic/:email` (JSON body: `upload`, `download`)
- `POST /:id/delClientByEmail/:email`

### Server API (нужен логин)
`{basePath}panel/api/server`
- `GET /status`
- `GET /cpuHistory/:bucket`
- `GET /getXrayVersion`
- `GET /getConfigJson`
- `GET /getDb` (скачать БД)
- `GET /getNewUUID`
- `GET /getNewX25519Cert`
- `GET /getNewmldsa65`
- `GET /getNewmlkem768`
- `GET /getNewVlessEnc`
- `POST /stopXrayService`
- `POST /restartXrayService`
- `POST /installXray/:version`
- `POST /updateGeofile` или `/updateGeofile/:fileName`
- `POST /logs/:count` (form: `level`, `syslog`)
- `POST /xraylogs/:count` (form: `filter`, `showDirect`, `showBlocked`, `showProxy`)
- `POST /importDB` (multipart поле `db`)
- `POST /getNewEchCert` (form: `sni`)

### Настройки панели (нужен логин)
`{basePath}panel/setting`
- `POST /all`
- `POST /defaultSettings`
- `POST /update` (form: AllSetting)
- `POST /updateUser`
- `POST /restartPanel`
- `GET /getDefaultJsonConfig`

### Настройки Xray (нужен логин)
`{basePath}panel/xray`
- `GET /getDefaultJsonConfig`
- `GET /getOutboundsTraffic`
- `GET /getXrayResult`
- `POST /` (шаблон xray config + inbound tags)
- `POST /warp/:action` (actions: `data`, `del`, `config`, `reg`, `license`)
- `POST /update` (form: `xraySetting`)
- `POST /resetOutboundsTraffic` (form: `tag`)

## Дизайн провайдера

### Конфигурация провайдера
- `endpoint` (base URL, напр. `http://localhost:2053`)
- `base_path` (опционально; по умолчанию `/`)
- `username`, `password`, `two_factor_code` (опционально)
- `insecure_skip_verify` (опционально для HTTPS)
- `request_timeout`

### Планируемые ресурсы
1. `3xui_inbound`
   - Управление inbound через `/panel/api/inbounds/add`, `/update/:id`, `/del/:id`.
   - Схема повторяет `model.Inbound` + `settings`, `stream_settings`, `sniffing` как JSON‑строки.
   - Read через `/list` или `/get/:id`.

2. `3xui_inbound_client`
   - Управление клиентами в inbound через `/addClient`, `/updateClient/:clientId`, `/delClient`.
   - Входные поля: `inbound_id`, `protocol`, client‑идентификаторы (id/password/email), клиентские поля.
   - Read через `/get/:id` и разбор `settings.clients` или `/getClientTrafficsById/:id`.

### Планируемые data sources
- `3xui_inbounds` (список)
- `3xui_server_status`
- `3xui_xray_versions`
- `3xui_xray_config` (из `/getConfigJson`)
- `3xui_settings` (из `/panel/setting/all`)

## Шаги реализации (с чек‑листом)

Правило: тесты пишем всегда вместе с новой функциональностью.

- [x] 1. Клиент + авторизация
- [x] HTTP‑клиент с cookie‑jar.
- [x] Логин (POST form) и авто‑relogin при 404/401 от API.
- [x] Поддержка base path и TLS‑настроек.

- [x] 2. API‑биндинги
- [x] Go‑структуры для `entity.Msg`, `model.Inbound`, payload‑ов клиентов.
- [x] Хелперы для form vs JSON запросов.
- [x] Парсинг JSON‑строковых полей (`settings`, `streamSettings`, `sniffing`).

- [x] 3. Ресурсы
- [x] `3xui_inbound` CRUD через API.
- [x] `3xui_inbound_client` CRUD через client‑эндпоинты.
- [x] Import поддержка (read‑only или import ID) для существующих объектов.

- [x] 4. Data sources
- [x] Реализация перечисленных data sources.

- [x] 5. Тесты
- [x] Локальные acceptance‑тесты против `docker-compose.yaml` (3x‑ui на 2053).
- [x] Кейс‑тесты: create/update/delete inbound и client.
- [x] Env‑переменные для endpoint/creds (без секретов в конфиге).

- [x] 6. Документация + примеры
- [x] Пример конфигурации провайдера.
- [x] Примеры ресурсов inbound + client.

## Открытые вопросы
- Какой минимальный набор ресурсов нужен для v1 (только inbound? inbound + client? ещё настройки/сервер?)
- Как обрабатывать поля `settings/stream/sniffing`: сырые JSON‑строки или типизированные схемы?
- Какие дефолтные креды использовать для docker‑compose тестов?
