# План тестирования Data Sources

Дата: 2026-02-06

## Data Sources

- `threexui_inbounds` — список inbound'ов
- `threexui_server_status` — статус сервера
- `threexui_xray_versions` — версии Xray
- `threexui_xray_config` — конфигурация Xray
- `threexui_settings` — настройки панели

---

## 1. `threexui_inbounds`

### 1.1 Базовое чтение
- [ ] Прочитать при наличии inbound'ов
- [ ] Прочитать при отсутствии inbound'ов (пустой список)
- [ ] Проверить `id` ресурса (= id первого inbound или "0")

### 1.2 Поля inbound
- [ ] `id`, `port`, `protocol`, `remark`, `tag`, `enable`
- [ ] `up`, `down`, `total`, `all_time` (трафик)
- [ ] `expiry_time`, `traffic_reset`, `last_traffic_reset_time`
- [ ] `listen`

### 1.3 Вложенные блоки
- [ ] `settings` — корректно разбирается из JSON
- [ ] `stream_settings` — корректно разбирается
- [ ] `sniffing` — `enabled`, `dest_override`, `metadata_only`, `route_only`

### 1.4 Несколько inbound'ов
- [ ] 2+ inbound'а с разными протоколами
- [ ] Проверить, что все возвращаются

---

## 2. `threexui_server_status`

- [ ] Прочитать статус сервера
- [ ] Поле `json` содержит валидный JSON
- [ ] Проверить наличие данных о CPU, памяти, uptime
- [ ] Повторный read → актуальные данные

---

## 3. `threexui_xray_versions`

- [ ] Прочитать список версий
- [ ] Поле `versions` — список строк
- [ ] Не пустой список (хотя бы текущая версия)

---

## 4. `threexui_xray_config`

- [ ] Прочитать конфигурацию Xray
- [ ] Поле `json` содержит валидный JSON
- [ ] Содержит ожидаемые секции (`log`, `routing`, `outbounds`)
- [ ] После изменения через `threexui_xray_*` ресурсы — данные обновляются

---

## 5. `threexui_settings`

- [ ] Прочитать все настройки
- [ ] Поле `json` содержит валидный JSON
- [ ] Содержит ключи: `webPort`, `webBasePath`, `sessionMaxAge` и др.
- [ ] После изменения через `threexui_panel_*` ресурсы — данные обновляются

---

## 6. Негативные сценарии

- [ ] Недоступный API → ошибка (для всех data sources)
- [ ] Невалидная сессия → ошибка авторизации

---

## 7. Результаты

- [ ] Все пункты выполнены
- [ ] Зафиксированы отклонения/баги
