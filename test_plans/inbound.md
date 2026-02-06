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
- [x] `settings` с `accounts` (для http-протокола) — работает (http с auth, accounts, allow_transparent)
- [x] `settings` с `peers` (для wireguard) — работает (private_key, public_key, allowed_ips, keep_alive)
- [x] `settings` с `port_map` (для dokodemo-door) — работает (address, port)
- [x] Дефолтные settings для vless: `decryption=none, encryption=none, testseed=[900,500,900,256]` — в API верно, в state testseed=[]
- [x] Пустые settings → `applyDefaultInboundSettings` подставляет дефолты
- [x] `preserveInboundSettings` при update — clients OK, testseed OK (после fix)

Примечание: проверка `settings` подтверждена через API.

## 4. Проверка `stream_settings`

- [x] Задать `stream_settings` c `network = "tcp"`
- [x] Проверить `tcp_settings.header.type = "none"`
- [x] Проверить `accept_proxy_protocol`
- [x] Проверить сохранение/чтение `stream_settings` после apply
- [x] `external_proxy` — dest, port, remark, force_tls — работает, идемпотентно

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
- [x] Без target/server_names → дефолт `www.apple.com:443` — подтверждено
- [x] `reality_settings.settings` — fingerprint, spider_x — работает
- [ ] `mldsa65_seed` / `mldsa65_verify` — не тестировали (экспериментальные поля)
- [x] `min_client_ver`, `max_client_ver`, `max_timediff` — работает, идемпотентно
- [x] При update — ключи сохраняются из existing (`mergeRealityFromExisting`)
- [x] Проверить, что `reality_settings` читаются обратно — идемпотентно

## 6. Sniffing

- [x] Включить `sniffing.enabled = true`
- [x] `sniffing.dest_override` (http/tls/quic/fakedns)
- [x] Проверить сохранение `metadata_only`, `route_only`

## 7. Обновления inbound

- [x] Изменить `remark`
- [x] Изменить `port` (проверить поведение API) — работает
- [x] Изменить `enable` true/false
- [x] Добавить/убрать `settings {}` и проверить стабильность — работает (добавление вызывает drift по testseed, после apply идемпотентно; удаление — No changes)
- [x] Обновить `stream_settings` (tcp -> ws, при необходимости) — не применимо (ws не реализован)
- [x] Обновить `sniffing` значения — dest_override, metadata_only, route_only

## 8. Идемпотентность

- [x] `apply` без изменений → `No changes`
- [x] `apply` после UI‑изменений → корректный drift — подтверждено (remark изменён через API, plan обнаружил)
- [x] `apply` после перезапуска панели → state совпадает — подтверждено (docker compose restart → No changes)

## 9. Клиенты (не управляются inbound)

- [x] Создать inbound и отдельно 3 клиента
- [x] `apply` → `No changes` при наличии клиентов
- [x] Изменение inbound не удаляет клиентов (подтверждено через API)
- [x] Удаление клиента не ломает inbound — подтверждено (удаление 1 из 2 клиентов, No changes после)
- [x] Проверить `expiryTime`, `totalGB`, `limitIp` на клиентах — работает, идемпотентно
- [x] Проверить `flow` для vless-клиентов — flow=xtls-rprx-vision, идемпотентно
- [x] Обновление клиента не сбрасывает `created_at` — подтверждено (created_at неизменен, updated_at обновился)

## 10. Удаление

- [x] `destroy` inbound без клиентов
- [x] `destroy` inbound с клиентами (через отдельный ресурс)
- [x] Отсутствие ошибки "no client remained in Inbound"

## 11. Негативные сценарии

- [x] Неверный `protocol` → API принимает без ошибки (нет валидации на стороне 3x-ui)
- [x] Занятый `port` → ошибка API "Port already exists"
- [x] Пустой `port` → ошибка валидации ("The argument \"port\" is required")
- [x] Неверный JSON в `settings` → не применимо (settings — блок, не JSON строка)
- [x] Некорректный `stream_settings` JSON → не применимо (stream_settings — блок)
- [x] Некорректный `sniffing` JSON → не применимо (sniffing — блок)
- [x] Дублирующийся `email` клиента между inbound'ами → "Duplicate email: same-email"
- [x] Дублирующийся `port` на том же `listen` → "UNIQUE constraint failed: inbounds.tag"
- [x] Пустой `client.id` → email используется как fallback ID (UUID генерируется только если email тоже пуст)
- [x] API retry при 401/404 → автоматический re-login — подтверждено (restart панели сбрасывает сессии, провайдер перелогинивается автоматически)

## 12. Протоколы

- [x] `vless` (базовый) — decryption, encryption, testseed, fallbacks
- [x] `vmess` (создание/обновление/удаление) — дефолтные settings, идемпотентно
- [x] `trojan` (создание/обновление/удаление) — идемпотентно
- [x] `shadowsocks` (создание/обновление/удаление) — method, password, network
- [x] `http` — auth, accounts (user/pass), allow_transparent — работает, идемпотентно
- [x] `mixed` — auth, accounts — работает, идемпотентно
- [x] `wireguard` — mtu, peers (private_key, public_key, allowed_ips, keep_alive), no_kernel_tun — идемпотентно
- [x] `dokodemo-door` (tunnel) — address, port — работает, идемпотентно
- [x] Проверить совместимость `settings` под каждый протокол — все протоколы протестированы

## 13. Import

- [x] Import inbound по ID — работает
- [x] Проверить корректность state после import — `No changes` после import

## 14. Результаты

- [x] Все основные пункты выполнены (2 пункта пропущены: selected_auth, mldsa65)
- [x] Зафиксированы отклонения/баги

### Найденные и исправленные баги

1. **testseed теряется при update** — исправлено. Добавлен `flattenIntList` + чтение testseed в `flattenSettings` (`settings.go`).

### Особенности

1. **Неверный protocol** — 3x-ui API не валидирует протокол, принимает любую строку.
2. **Дублирующийся порт** — ошибка идёт по UNIQUE constraint на tag, а не по port.
3. **Добавление `settings {}`** к существующему inbound — вызывает drift по testseed (пустой блок сбрасывает), после apply идемпотентно.
4. **client.id fallback** — если client_id не задан, используется email; UUID генерируется только если все поля пусты.
