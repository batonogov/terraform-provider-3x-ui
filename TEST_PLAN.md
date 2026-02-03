# План тестирования: Terraform/OpenTofu провайдер для 3x-ui

## Область
План покрывает Terraform/OpenTofu провайдер, который управляет inbounds и клиентами 3x-ui через HTTP API. Включены unit-тесты для API клиента и преобразований данных, а также acceptance-тесты на реальном 3x-ui.

## Цели
- Проверить поведение API клиента, обработку ошибок и повторную аутентификацию.
- Проверить корректность преобразований schema <-> API payload.
- Проверить CRUD и import ресурсов на реальном 3x-ui.
- Проверить data sources.
- Зафиксировать регрессии вокруг генерации ID клиентов и дефолтных настроек.

## Входит в объем
- Конфигурация провайдера и инициализация клиента.
- Ресурс: `threexui_inbound`.
- Ресурс: `threexui_inbound_client`.
- Data sources: `threexui_inbounds`, `threexui_server_status`, `threexui_xray_versions`, `threexui_xray_config`, `threexui_settings`.

## Не входит в объем
- Внутренние детали 3x-ui вне публичного API.
- Нагрузочное/производительное тестирование.
- UI 3x-ui.

## Уровни тестирования

### 1) Unit-тесты: чистые функции и преобразования
Целевые файлы в `provider/`.

- [x] Unit: `normalizeBasePath`, `resolvePath` (валидные/невалидные base path и relative path).
- [x] Unit: `parseID`, `splitInboundClientID`, `makeInboundClientID`.
- [x] Unit: `newUUID` (формат и непустое значение).
- [x] Unit: `ParseJSONField`, `isSubset`, `jsonSubsetDiffSuppress`.
- [x] Unit: `buildSettingsJSON` <-> `flattenSettings`.
- [x] Unit: `buildStreamSettingsJSON` <-> `flattenStreamSettings`.
- [x] Unit: `buildSniffingJSON` <-> `flattenSniffing`.
- [x] Unit: `ensureInboundClientIDs` при отсутствии `id` в `settings.clients`.
- [x] Unit: `preserveInboundClientIDs` при обновлении inbound.

### 2) Unit-тесты: API клиент с `httptest`
Целевые файлы: `provider/client.go` и связанные тесты.

- [x] API: `Login` успешный кейс.
- [x] API: `Login` неуспешный кейс.
- [x] API: обработка 2FA (`twoFactorCode`).
- [x] API: авто-логин при `401` и `404` с негативными кейсами.
- [x] API: `decodeAPIResponse` для пустого тела при `status >= 400`.
- [x] API: `decodeAPIResponse` для невалидного JSON.
- [x] API: `decodeAPIResponse` для `success=false` с/без `msg`.
- [x] API: `AddInbound`, `UpdateInbound`, `GetInbound`, `DeleteInbound`.
- [x] API: `AddInboundClient`, `UpdateInboundClient`, `DeleteInboundClient`.
- [x] API: `GetInbounds`, `GetServerStatus`, `GetXrayVersions`, `GetXrayConfig`, `GetSettings`.

### 3) Unit-тесты: конфигурация провайдера
Целевой файл: `provider/provider.go`.

- [x] Provider: невалидный `request_timeout` возвращает diagnostics.
- [x] Provider: `endpoint` без схемы приводит к ошибке.
- [x] Provider: `base_path` нормализуется корректно.

### 4) Acceptance-тесты (реальный 3x-ui)
Целевые файлы: `provider/acc_test.go` и новые acceptance тесты.

Примечание: acceptance-тесты требуют запуска вне песочницы, т.к. go-plugin использует unix-сокеты.

- [x] Acc: поднять 3x-ui через `docker-compose.yaml`.
- [x] Acc: установить переменные окружения `TF_ACC=1`, `THREEXUI_ENDPOINT`, `THREEXUI_USERNAME`, `THREEXUI_PASSWORD`.
- [x] Acc inbound: Create с минимальной конфигурацией.
- [x] Acc inbound: Update `remark`, `settings`, `stream_settings`, `sniffing`.
- [x] Acc inbound: Delete и проверка удаления.
- [x] Acc inbound: Import по ID.
- [x] Acc inbound_client: Create под inbound.
- [x] Acc inbound_client: Update `email`, `enable`, `flow`.
- [x] Acc inbound_client: Delete и проверка удаления.
- [x] Acc inbound_client: Import по `<inbound_id>:<client_id>`.
- [x] Acc data source: `threexui_inbounds` возвращает список.
- [x] Acc data source: `threexui_server_status` возвращает ожидаемые ключи.
- [x] Acc data source: `threexui_xray_versions` возвращает список.
- [x] Acc data source: `threexui_xray_config` возвращает JSON.
- [x] Acc data source: `threexui_settings` возвращает JSON.

### 5) Регрессии и edge cases
- [x] Reg: обновление inbound с частично заданными `client_id` не пересоздает ID.
- [x] Reg: пустые/отсутствующие `settings`/`stream_settings`/`sniffing` не падают, defaults применяются.
- [x] Reg: невалидный JSON в settings дает понятную ошибку.
- [x] Reg: `base_path` с/без слэшей ведет себя одинаково.

## Тестовые данные
- Фиксированные порты (например, `23456`, `23457`) для inbounds.
- Уникальные `remark` в тестах для удобной проверки.

## Запуск
- Unit-тесты:
  - `task test` или `go test ./provider`
- Все тесты:
  - `task test-all` или `go test ./...`
- Acceptance-тесты:
  - `task acc`

## Риски и допущения
- Acceptance-тесты требуют запущенный 3x-ui и стабильный API.
- Некоторые поля зависят от версии 3x-ui, поэтому образ в `docker-compose.yaml` закреплен.

## Критерии готовности
- Все unit-тесты проходят.
- Acceptance-тесты проходят на локальном 3x-ui.
- Нет регрессий относительно текущих тестов.
