# План тестирования `threexui_inbound`

Дата: 2026-02-05

Репозиторий 3x-ui: `3x-ui-2.8.9/`

## 1. Подготовка окружения

- [ ] Чистый запуск панели 3x-ui (docker compose up)
- [ ] Проверить доступность UI и API (логин)
- [ ] Очистить состояние OpenTofu/TF в `examples/`
- [ ] Зафиксировать версию панели (в логах)

## 2. Базовое создание inbound

- [ ] Создать inbound без `settings`
- [ ] Создать inbound с пустым `settings {}`
- [ ] Создать inbound с `settings { decryption = "none", encryption = "none" }`
- [ ] Проверить, что inbound появляется в UI
- [ ] Проверить поля: `port`, `protocol`, `remark`, `tag`, `enable`
- [ ] Проверить `listen` с нестандартным значением (конкретный IP)
- [ ] Проверить автогенерацию `tag`
- [ ] Проверить `up`, `down`, `total`, `all_time` (трафик)
- [ ] Проверить `expiry_time`, `traffic_reset`

Примечание: проверка полей выполняется через API (без ручного вмешательства).

## 3. Проверка `settings`

- [ ] `settings {}` не приводит к drift при повторном apply
- [ ] `settings` сохраняет `testseed` (не сбрасывается при update — `preserveSettingsKey`)
- [ ] `settings` сохраняет `decryption`/`encryption` (если заданы)
- [ ] `settings` без клиентов не ломает addClient
- [ ] `settings` с `fallbacks` (для vless/trojan) — name, alpn, path, dest, xver
- [ ] `settings` с `method` + `password` для shadowsocks
- [ ] `settings` с `selected_auth` → автозаполнение decryption/encryption через API
- [ ] `settings` с `accounts` (для http-протокола) — user, pass
- [ ] `settings` с `peers` (для wireguard) — private_key, public_key, allowed_ips, keep_alive
- [ ] `settings` с `port_map` (для tunnel)
- [ ] Дефолтные settings для vless: `decryption=none, encryption=none, testseed=[900,500,900,256]`
- [ ] Пустые settings → `applyDefaultInboundSettings` подставляет дефолты
- [ ] `preserveInboundSettings` при update — clients и testseed из existing

Примечание: проверка `settings` подтверждена через API.

## 4. Проверка `stream_settings`

- [ ] Задать `stream_settings` c `network = "tcp"`
- [ ] Проверить `tcp_settings.header.type = "none"`
- [ ] Проверить `accept_proxy_protocol`
- [ ] Проверить сохранение/чтение `stream_settings` после apply
- [ ] `external_proxy` — dest, port, remark, force_tls

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

- [ ] Задать `stream_settings.security = "reality"`
- [ ] Задать `reality_settings.target` и `server_names`
- [ ] Проверить автогенерацию `private_key`/`public_key` (через `GetNewX25519Cert`)
- [ ] Проверить автогенерацию `short_ids` (8 значений разной длины)
- [ ] Без target/server_names → дефолт `www.apple.com:443`
- [ ] `reality_settings.settings` — public_key, fingerprint, server_name, spider_x
- [ ] `mldsa65_seed` / `mldsa65_verify`
- [ ] `min_client_ver`, `max_client_ver`, `max_timediff`
- [ ] При update — ключи сохраняются из existing (`mergeRealityFromExisting`)
- [ ] Проверить, что `reality_settings` читаются обратно

## 6. Sniffing

- [ ] Включить `sniffing.enabled = true`
- [ ] `sniffing.dest_override` (http/tls/quic)
- [ ] Проверить сохранение `metadata_only`, `route_only`

## 7. Обновления inbound

- [ ] Изменить `remark`
- [ ] Изменить `port` (проверить поведение API)
- [ ] Изменить `enable` true/false
- [ ] Добавить/убрать `settings {}` и проверить стабильность
- [ ] Обновить `stream_settings` (tcp -> ws, при необходимости)
- [ ] Обновить `sniffing` значения

## 8. Идемпотентность

- [ ] `apply` без изменений → `No changes`
- [ ] `apply` после UI‑изменений → корректный drift
- [ ] `apply` после перезапуска панели → state совпадает

## 9. Клиенты (не управляются inbound)

- [ ] Создать inbound и отдельно 3 клиента
- [ ] `apply` → `No changes` при наличии клиентов
- [ ] Изменение inbound не удаляет клиентов
- [ ] Удаление клиента не ломает inbound
- [ ] Проверить `expiryTime`, `totalGB`, `limitIp` на клиентах
- [ ] Проверить `flow` для vless-клиентов
- [ ] Обновление клиента не сбрасывает `created_at`

## 10. Удаление

- [ ] `destroy` inbound без клиентов
- [ ] `destroy` inbound с клиентами (через отдельный ресурс)
- [ ] Отсутствие ошибки "no client remained in Inbound"

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

- [ ] `vless` (базовый) — decryption, encryption, testseed, fallbacks, selected_auth
- [ ] `vmess` (создание/обновление/удаление) — дефолтные settings = `{clients:[]}`
- [ ] `trojan` (создание/обновление/удаление) — fallbacks
- [ ] `shadowsocks` (создание/обновление/удаление) — method, password, network, iv_check
- [ ] `http` — auth, accounts (user/pass), allow_transparent
- [ ] `mixed` — auth, accounts
- [ ] `wireguard` — secret_key, mtu, peers, no_kernel_tun
- [ ] `tunnel` — address, port, port_map
- [ ] Проверить совместимость `settings` под каждый протокол

## 13. Import

- [ ] Import inbound по ID
- [ ] Проверить корректность state после import

## 14. Результаты

- [ ] Все пункты выполнены
- [ ] Зафиксированы отклонения/баги
