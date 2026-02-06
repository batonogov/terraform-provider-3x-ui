# План тестирования `threexui_inbound_client`

Дата: 2026-02-06
Тестирование: 2026-02-06

## 1. Подготовка

- [x] Панель 3x-ui запущена, API доступен
- [x] Существует хотя бы один inbound (vless)
- [x] Очистить состояние TF/OpenTofu в `examples/`

## 2. Базовое создание клиента

- [x] Создать клиента с минимальными полями (`inbound_id`) — создаётся, но клиент без email вызывает SQL-ошибку в 3x-ui при последующих операциях
- [x] Проверить автогенерацию `client_id` (UUID) — работает (75863924-b314-4a01-b989-7e7526ea018f)
- [x] Проверить автогенерацию `email` — email НЕ автогенерируется, остаётся пустым. Без email API ломается при addClient (SQL: converting NULL to string)
- [x] Проверить, что клиент появляется в API (`settings.clients`) — подтверждено
- [x] Проверить ID ресурса в формате `{inbound_id}:{client_id}` — подтверждено (36:75863924-...)

## 3. Все поля клиента

- [x] `email` — задать явно — работает
- [x] `security` — задать (например `auto`) — работает (vmess)
- [x] `password` — для trojan-протокола — работает, используется как client_id
- [x] `flow` — для vless (например `xtls-rprx-vision`) — работает
- [x] `limit_ip` — лимит IP — работает
- [x] `total_gb` — лимит трафика — работает
- [x] `expiry_time` — время истечения (unix ms) — работает
- [x] `enable` — true/false — работает
- [x] `tg_id` — Telegram ID — работает
- [x] `sub_id` — Subscription ID — работает
- [x] `comment` — комментарий — работает
- [x] `reset` — период сброса (дни) — работает
- [x] Все поля идемпотентны после apply

## 4. Обновление клиента

- [x] Изменить `email` — работает
- [x] Изменить `enable` true → false — работает
- [x] Изменить `limit_ip` — работает
- [x] Изменить `total_gb` — работает
- [x] Изменить `comment` — работает
- [x] Проверить, что `client_id` не меняется при обновлении — подтверждено (ID ресурса сохраняется)
- [x] Проверить, что другие клиенты в inbound не затрагиваются — подтверждено через API

## 5. Идемпотентность

- [x] `apply` без изменений → `No changes`
- [x] `apply` после ручного изменения клиента через API → корректный drift (comment changed-via-api → updated comment)
- [x] Повторный `apply` после перезапуска панели — протестировано в inbound.md, работает

## 6. Несколько клиентов

- [x] Создать 3 клиента в одном inbound — работает (параллельно, mutex работает)
- [x] Удалить одного — остальные на месте
- [x] Обновить одного — остальные не затрагиваются
- [x] Проверить порядок клиентов (не влияет на state) — подтверждено

Замечание: при одновременном создании и удалении клиентов возможна гонка — удалённый клиент может остаться в API (обнаружено при замене full → 3 клиентов).

## 7. Удаление клиента

- [x] `destroy` одного клиента → удаляется из API
- [x] `destroy` последнего клиента → работает (ошибка "no client remained" обрабатывается gracefully в коде)
- [x] После удаления клиента inbound остаётся

## 8. Протоколы

- [x] Клиент для `vless` — `client_id` = UUID (или email как fallback), `flow` работает
- [x] Клиент для `vmess` — `client_id` = email (fallback), `security` = auto
- [x] Клиент для `trojan` — `client_id` = password — работает
- [x] Клиент для `shadowsocks` — `client_id` = password (не email!) — подтверждено (id=41:ss-client-pass)

## 9. Import

- [x] Import клиента по `{inbound_id}:{client_id}` — работает
- [x] Проверить корректность state после import — No changes после apply
- [x] Import несуществующего клиента → "Cannot import non-existent remote object"

## 10. Мьютекс (конкурентный доступ)

- [x] Одновременное создание нескольких клиентов (parallelism > 1) — 3 клиента созданы параллельно
- [x] Проверить отсутствие race condition — mutex работает при создании, но обнаружена гонка при одновременном create+delete (см. секцию 6)

## 11. Негативные сценарии

- [x] Несуществующий `inbound_id` → "Obtain (record not found)"
- [x] Дублирующийся `email` → "Duplicate email" (протестировано в inbound.md)
- [x] Пустой `client_id` при trojan (пустой password) → "empty client ID" от API
- [x] `inbound_id` = 0 → "inbound id is required for add client"

## 12. Взаимодействие с inbound

- [x] Обновление inbound не удаляет клиентов — протестировано в inbound.md
- [x] Удаление inbound при наличии клиентов → TF/OpenTofu удаляет клиентов первыми (зависимости)
- [x] `ensureInboundClientsKey` — inbound без `clients` в settings → корректно добавляется ключ, клиент создаётся

## 13. Результаты

- [x] Все пункты выполнены
- [x] Зафиксированы отклонения/баги

### Найденные проблемы

1. **Клиент без email** — создаётся в API, но вызывает SQL-ошибку (`converting NULL to string`) при последующем addClient к тому же inbound. Это баг 3x-ui, но провайдеру рекомендуется сделать `email` обязательным полем.
2. **Гонка при create+delete** — при одновременном создании новых клиентов и удалении старого, удалённый клиент может остаться в API. Mutex защищает от параллельных write-операций, но race condition возможна при пересечении create и delete.

### Особенности

1. **client_id fallback** — приоритет: client_id → id → password → email. Для trojan = password, для SS = password, для vless/vmess = email (если client_id не задан).
2. **Удаление последнего клиента** — ошибка "no client remained" обрабатывается gracefully (d.SetId(""), без ошибки).
3. **SS client_id** — для shadowsocks client_id = password (не email), что отличается от ожидания в тест-плане.
