# План тестирования Xray-настроек

Дата: 2026-02-06
Обновлено: 2026-02-06 (structured blocks вместо jsonencode)

## Ресурсы

- `threexui_xray_basics` — базовые настройки Xray (mode: merge root), блоки: log, policy, api, stats
- `threexui_xray_dns` — DNS Xray (path: `dns`), блоки: server, атрибуты: hosts, query_strategy и др.
- `threexui_xray_routing` — маршрутизация Xray (path: `routing`), блоки: rule, атрибуты: domain_strategy, domain_matcher
- `threexui_xray_balancers` — балансировщики Xray (path: `routing.balancers`), блоки: balancer
- `threexui_xray_reverse` — reverse proxy Xray (path: `reverse`), блоки: bridge, portal
- `threexui_xray_outbounds` — outbound'ы Xray (path: `outbounds`), блоки: outbound с per-protocol settings

Удалён: `threexui_xray_advanced` (полная замена конфига).

---

## 1. Общее для всех xray-ресурсов

- [ ] Все ресурсы используют structured blocks (нативные HCL-блоки) вместо `json` атрибута
- [ ] Terraform SDK корректно сравнивает блоки (нет ложных diff)
- [ ] Delete = только очищает TF state, не сбрасывает xray-конфиг
- [ ] Мьютекс `xrayTemplateMu` предотвращает race condition при параллельном создании
- [ ] Каждый ресурс — синглтон (ID = section.id)

---

## 2. `threexui_xray_basics` (mode: merge root)

- [ ] Блок `log` — задать loglevel, access, error, dns_log
- [ ] Блок `policy.system` — stats_inbound_downlink/uplink, stats_outbound_downlink/uplink
- [ ] Блок `policy.level` — id, handshake, conn_idle, buffer_size и др. (levels = map по id в JSON)
- [ ] Блок `api` — tag, services
- [ ] Блок `stats` — пустой блок (наличие = включение stats)
- [ ] Merge root — мержится с существующим конфигом, не затирает другие ключи
- [ ] Идемпотентность — повторный apply без изменений не вызывает diff

---

## 3. `threexui_xray_dns` (path: `dns`)

- [ ] Блок `server` с одним address → сериализуется как строка в JSON
- [ ] Блок `server` с address + port/domains/expect_ips → сериализуется как объект
- [ ] Flatten: строковый сервер из API → блок `server { address = "..." }`
- [ ] Атрибуты: hosts (map), query_strategy, tag, disable_cache, disable_fallback, client_ip
- [ ] Идемпотентность подтверждена

---

## 4. `threexui_xray_routing` (path: `routing`)

- [ ] Атрибуты: domain_strategy, domain_matcher
- [ ] Блок `rule` — type (default "field"), domain, ip, port, source_port, network, source, user, inbound_tag, protocol, attrs, outbound_tag, balancer_tag
- [ ] Множество правил — порядок сохраняется
- [ ] Идемпотентность подтверждена

---

## 5. `threexui_xray_balancers` (path: `routing.balancers`)

- [ ] Блок `balancer` — tag (Required), selector (Required), strategy.type
- [ ] Массив балансировщиков корректно сериализуется
- [ ] Идемпотентность подтверждена

---

## 6. `threexui_xray_reverse` (path: `reverse`)

- [ ] Блок `bridge` — tag (Required), domain (Required)
- [ ] Блок `portal` — tag (Required), domain (Required)
- [ ] Сериализация: bridge → bridges, portal → portals в JSON
- [ ] Идемпотентность подтверждена

---

## 7. `threexui_xray_outbounds` (path: `outbounds`)

### Базовые атрибуты outbound
- [ ] tag, protocol (Required), send_through
- [ ] Блок `mux` — enabled, concurrency, xudp_concurrency, xudp_proxy_udp443

### Per-protocol settings (взаимоисключающие блоки)
- [ ] `freedom_settings` — domain_strategy, redirect, fragment (packets/length/interval), noises (type/packet/delay)
- [ ] `blackhole_settings` — response_type → сериализуется как `response.type`
- [ ] `dns_settings` — network, address, port, non_ip_query, block_types
- [ ] `vmess_settings` — address, port, id, security → сериализуется через vnext[].users[]
- [ ] `vless_settings` — address, port, id, flow, encryption → сериализуется через vnext[].users[]
- [ ] `trojan_settings` — address, port, password (Sensitive) → сериализуется через servers[]
- [ ] `shadowsocks_settings` — address, port, password (Sensitive), method, uot, uot_version
- [ ] `socks_settings` — address, port, user, pass (Sensitive) → servers[].users[]
- [ ] `http_settings` — address, port, user, pass (Sensitive) → servers[].users[]
- [ ] `wireguard_settings` — mtu, secret_key (Sensitive), address (list), workers, domain_strategy, reserved (list int), no_kernel_tun, peer (public_key, pre_shared_key, allowed_ips, endpoint, keep_alive)
- [ ] `hysteria_settings` — address, port, version

### Flatten
- [ ] API → structured blocks: protocol определяет какой блок заполнять
- [ ] vnext/servers структура корректно маппится на плоские атрибуты
- [ ] Идемпотентность подтверждена

---

## 8. Unit-тесты

- [ ] `TestFlattenXrayReverseToMap` — bridges/portals → bridge/portal блоки
- [ ] `TestFlattenXrayBalancersToMap` — массив балансировщиков
- [ ] `TestFlattenXrayDNSToMap` — строковые и объектные серверы
- [ ] `TestExpandDNSServers_StringOnly` — address-only → строка
- [ ] `TestExpandDNSServers_WithPort` — address+port → объект
- [ ] `TestFlattenXrayRoutingToMap` — domainStrategy, rules
- [ ] `TestFlattenXrayBasicsToMap` — log, api, stats
- [ ] `TestFlattenXrayOutboundsToMap` — freedom + blackhole
- [ ] `TestExpandReverseEntries` — tag/domain
- [ ] `TestExpandRoutingRules` — type/ip/outboundTag
- [ ] `TestFlattenWireguardOutSettings` — secretKey, peers, reserved
- [ ] `TestFlattenBasicsPolicyLevels` — map "0" → list с id=0

---

## 9. Результаты

- [ ] Все пункты выполнены
- [ ] Зафиксированы отклонения/баги

### Особенности

1. **Delete = только state** — destroy только очищает TF state, не сбрасывает xray-конфиг.
2. **Мьютекс** — `xrayTemplateMu` сериализует read-modify-write операции.
3. **DNS smart serialization** — сервер без доп. полей → строка, с полями → объект.
4. **Outbound settings** — блок определяется значением `protocol`.
5. **Policy levels** — в Xray JSON map `{"0": {...}}`, в TF list `[{id=0, ...}]`.
