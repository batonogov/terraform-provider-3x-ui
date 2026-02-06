# План тестирования `threexui_inbound`

Дата: 2026-02-05
Тестирование: 2026-02-06

Репозиторий 3x-ui: `3x-ui-2.8.9/`

## 1. Подготовка окружения

- [x] Чистый запуск панели 3x-ui (docker compose up)
- [x] Проверить доступность UI и API (логин)
- [x] Очистить состояние OpenTofu/TF в `examples/`
- [x] Зафиксировать версию панели (v2.8.9)

## 2. Базовое создание inbound

- [x] Создать inbound без `settings`
- [x] Создать inbound с пустым `settings {}`
- [x] Создать inbound с `settings { decryption = "none", encryption = "none" }`
- [x] Проверить, что inbound появляется в UI
- [x] Проверить поля: `port`, `protocol`, `remark`, `tag`, `enable`
- [x] Проверить `listen` с нестандартным значением (конкретный IP) — `tag = "inbound-127.0.0.1:10004"`
- [x] Проверить автогенерацию `tag` — `inbound-{port}` или `inbound-{ip}:{port}`
- [x] Проверить `up`, `down`, `total`, `all_time` (трафик) — все 0 при создании
- [x] Проверить `expiry_time`, `traffic_reset` — 0 и "never" по умолчанию

Примечание: проверка полей выполняется через API (без ручного вмешательства).

## 3. Проверка `settings`

- [x] `settings {}` не приводит к drift при повторном apply
- [x] `settings` сохраняет `testseed` (не сбрасывается при update — `preserveSettingsKey`)
  - Был баг: `flattenSettings` не читал testseed из API. Исправлено в `settings.go` (добавлен `flattenIntList`).
- [x] `settings` сохраняет `decryption`/`encryption` (если заданы)
- [x] `settings` без клиентов не ломает addClient
- [x] `settings` с `fallbacks` (для vless/trojan) — name, alpn, path, dest, xver — идемпотентно
- [x] `settings` с `method` + `password` для shadowsocks — работает, идемпотентно
- [ ] `settings` с `selected_auth` → автозаполнение decryption/encryption через API — не тестировали (нужен специальный API endpoint)
- [ ] `settings` с `accounts` (для http-протокола) — не тестировали
- [ ] `settings` с `peers` (для wireguard) — не тестировали
- [ ] `settings` с `port_map` (для tunnel) — не тестировали
- [x] Дефолтные settings для vless: `decryption=none, encryption=none, testseed=[900,500,900,256]` — в API верно, в state testseed=[]
- [x] Пустые settings → `applyDefaultInboundSettings` подставляет дефолты
- [x] `preserveInboundSettings` при update — clients OK, testseed OK (после fix)

Примечание: проверка `settings` подтверждена через API.

## 4. Проверка `stream_settings`

- [x] Задать `stream_settings` c `network = "tcp"`
- [x] Проверить `tcp_settings.header.type = "none"`
- [x] Проверить `accept_proxy_protocol`
- [x] Проверить сохранение/чтение `stream_settings` после apply
- [ ] `external_proxy` — dest, port, remark, force_tls — не тестировали

Примечание: `tls_settings` не реализован в провайдере.

## 4.1 Ограничения `stream_settings` в провайдере

**Важно:** провайдер сейчас поддерживает только:
- `tcp_settings` (header, accept_proxy_protocol)
- `reality_settings` (полный блок)
- `external_proxy` (dest, port, remark, force_tls)

**НЕ реализованы** в провайдере (только в 3x-ui API):
- `ws_settings`, `grpc_settings`, `http_settings` (h2)
- `quic_settings`, `kcp_settings`, `httpupgrade_settings`
- `tls_settings` (сертификаты)

Тесты для нереализованных транспортов не применимы.

## 5. Reality (VLESS)

- [x] Задать `stream_settings.security = "reality"`
- [x] Задать `reality_settings.target` и `server_names` — google.com:443
- [x] Проверить автогенерацию `private_key`/`public_key` (через `GetNewX25519Cert`)
- [x] Проверить автогенерацию `short_ids` (8 значений разной длины) — ✓
- [ ] Без target/server_names → дефолт `www.apple.com:443` — не тестировали
- [x] `reality_settings.settings` — fingerprint, spider_x — работает
- [ ] `mldsa65_seed` / `mldsa65_verify` — не тестировали
- [ ] `min_client_ver`, `max_client_ver`, `max_timediff` — не тестировали
- [x] При update — ключи сохраняются из existing (`mergeRealityFromExisting`)
- [x] Проверить, что `reality_settings` читаются обратно — идемпотентно

## 6. Sniffing

- [x] Включить `sniffing.enabled = true`
- [x] `sniffing.dest_override` (http/tls/quic/fakedns)
- [x] Проверить сохранение `metadata_only`, `route_only`

## 7. Обновления inbound

- [x] Изменить `remark`
- [ ] Изменить `port` (проверить поведение API) — не тестировали
- [x] Изменить `enable` true/false
- [ ] Добавить/убрать `settings {}` и проверить стабильность — не тестировали
- [ ] Обновить `stream_settings` (tcp -> ws, при необходимости) — не применимо (ws не реализован)
- [x] Обновить `sniffing` значения — dest_override, metadata_only, route_only

## 8. Идемпотентность

- [x] `apply` без изменений → `No changes`
- [ ] `apply` после UI‑изменений → корректный drift — не тестировали
- [ ] `apply` после перезапуска панели → state совпадает — не тестировали

## 9. Клиенты (не управляются inbound)

- [x] Создать inbound и отдельно 3 клиента
- [x] `apply` → `No changes` при наличии клиентов
- [x] Изменение inbound не удаляет клиентов (подтверждено через API)
- [ ] Удаление клиента не ломает inbound — не тестировали отдельно
- [ ] Проверить `expiryTime`, `totalGB`, `limitIp` на клиентах — не тестировали
- [ ] Проверить `flow` для vless-клиентов — client-a имеет flow, работает
- [ ] Обновление клиента не сбрасывает `created_at` — не тестировали

## 10. Удаление

- [x] `destroy` inbound без клиентов
- [x] `destroy` inbound с клиентами (через отдельный ресурс)
- [x] Отсутствие ошибки "no client remained in Inbound"

## 11. Негативные сценарии

- [ ] Неверный `protocol` → ошибка API
- [ ] Занятый `port` → ошибка API
- [ ] Пустой `port` → ошибка валидации
- [ ] Неверный JSON в `settings` → ошибка валидации
- [ ] Некорректный `stream_settings` JSON → ошибка валидации
- [ ] Некорректный `sniffing` JSON → ошибка валидации
- [ ] Дублирующийся `email` клиента между inbound'ами → ошибка
- [ ] Дублирующийся `port` на том же `listen` → ошибка
- [ ] Пустой `client.id` → автогенерация UUID (не ошибка!)
- [ ] API retry при 401/404 → автоматический re-login

## 12. Протоколы

- [x] `vless` (базовый) — decryption, encryption, testseed, fallbacks
- [ ] `vmess` (создание/обновление/удаление) — дефолтные settings = `{clients:[]}`
- [ ] `trojan` (создание/обновление/удаление) — fallbacks
- [x] `shadowsocks` (создание/обновление/удаление) — method, password, network
- [ ] `http` — auth, accounts (user/pass), allow_transparent
- [ ] `mixed` — auth, accounts
- [ ] `wireguard` — secret_key, mtu, peers, no_kernel_tun
- [ ] `tunnel` — address, port, port_map
- [ ] Проверить совместимость `settings` под каждый протокол

## 13. Import

- [x] Import inbound по ID — работает
- [x] Проверить корректность state после import — `No changes` после import

## 14. Результаты

- [ ] Все пункты выполнены
- [x] Зафиксированы отклонения/баги

### Найденные и исправленные баги

1. **testseed теряется при update** — исправлено. Добавлен `flattenIntList` + чтение testseed в `flattenSettings` (`settings.go`).
