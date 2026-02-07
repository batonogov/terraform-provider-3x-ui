# PLAN.md

План развития Terraform-провайдера для 3x-ui.

## Покрытие тестами

### Acceptance-тесты ресурсов

- [x] `threexui_inbound` — 13 тестов (vmess, trojan, shadowsocks, http, wireguard, dokodemo, reality, fallbacks, settings, stream/sniffing, port conflict, update, idempotency)
- [x] `threexui_inbound_client` — 5 тестов (all fields, update, multiple, vmess, trojan)
- [x] `threexui_panel_general` — 2 теста (basic + LDAP)
- [x] `threexui_panel_security` — 1 тест (create, update, idempotency, restore)
- [x] `threexui_panel_telegram` — 1 тест (enable/disable, update, idempotency)
- [x] `threexui_panel_subscription` — 1 тест (enable/disable, update, idempotency)
- [x] `threexui_xray_basics` — 1 тест
- [x] `threexui_xray_dns` — 1 тест
- [x] `threexui_xray_routing` — 1 тест
- [x] `threexui_xray_balancers` — 1 тест
- [x] `threexui_xray_reverse` — 1 тест
- [x] `threexui_xray_outbounds` — 1 тест

### Acceptance-тесты data sources

- [x] `threexui_inbounds` — 2 теста (basic + multiple)
- [x] `threexui_server_status` — 1 тест
- [x] `threexui_xray_versions` — 1 тест
- [x] `threexui_xray_config` — 1 тест
- [x] `threexui_settings` — 1 тест

## Новые data sources

| Приоритет | Data source | API-эндпоинт | Описание |
|---|---|---|---|
| Высокий | `threexui_client_traffic` | `GET /panel/api/inbounds/getClientTraffics/:email` | Трафик клиента по email |
| Средний | `threexui_online_clients` | `POST /panel/api/inbounds/onlines` | Список онлайн-клиентов |
| Средний | `threexui_client_ips` | `POST /panel/api/inbounds/clientIps/:email` | IP-адреса клиента |
| Средний | `threexui_outbounds_traffic` | `GET /panel/xray/getOutboundsTraffic` | Трафик по outbound'ам |
| Низкий | `threexui_cpu_history` | `GET /panel/api/server/cpuHistory/:bucket` | История CPU |

- [ ] `threexui_client_traffic`
- [ ] `threexui_online_clients`
- [ ] `threexui_client_ips`
- [ ] `threexui_outbounds_traffic`
- [ ] `threexui_cpu_history`

## Новые ресурсы

| Приоритет | Ресурс | API-эндпоинт | Описание |
|---|---|---|---|
| Высокий | `threexui_xray_install` | `POST /panel/api/server/installXray/:version` | Управление версией Xray |
| Средний | `threexui_panel_user` | `POST /panel/setting/updateUser` | Смена логина/пароля админа |
| Средний | `threexui_warp` | `POST /panel/xray/warp/*` | Cloudflare WARP конфигурация |

- [ ] `threexui_xray_install`
- [ ] `threexui_panel_user`
- [ ] `threexui_warp`

## Непокрытые операционные эндпоинты

Эти эндпоинты не подходят для декларативных Terraform-ресурсов, но могут быть полезны как provisioner или отдельные утилиты:

- `POST /panel/api/inbounds/:id/resetClientTraffic/:email` — сброс трафика клиента
- `POST /panel/api/inbounds/resetAllClientTraffics/:id` — сброс трафика всех клиентов inbound'а
- `POST /panel/api/inbounds/resetAllTraffics` — сброс всего трафика
- `POST /panel/api/inbounds/clearClientIps/:email` — очистка IP клиента
- `POST /panel/api/inbounds/delDepletedClients/:id` — удаление клиентов с исчерпанным трафиком
- `POST /panel/api/server/stopXrayService` — остановка Xray
- `POST /panel/api/server/restartXrayService` — перезапуск Xray
- `POST /panel/api/server/updateGeofile` — обновление geo-файлов
- `POST /panel/api/server/importDB` — импорт базы данных
- `GET /panel/api/server/getDb` — экспорт базы данных
- `POST /panel/api/server/logs/:count` — логи панели
- `POST /panel/api/server/xraylogs/:count` — логи Xray
