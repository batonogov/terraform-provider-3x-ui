# План тестирования настроек панели и Xray

Дата: 2026-02-06

## Ресурсы

- `threexui_panel_general` — общие настройки панели
- `threexui_panel_security` — 2FA
- `threexui_panel_telegram` — Telegram-бот
- `threexui_panel_subscription` — подписки
- `threexui_xray_basics` — базовые настройки Xray
- `threexui_xray_dns` — DNS Xray
- `threexui_xray_routing` — маршрутизация Xray
- `threexui_xray_balancers` — балансировщики Xray
- `threexui_xray_reverse` — reverse proxy Xray
- `threexui_xray_outbounds` — outbound'ы Xray
- `threexui_xray_advanced` — полная замена конфига Xray

---

## 1. `threexui_panel_general`

### 1.1 Базовое создание
- [ ] Создать ресурс с дефолтными значениями
- [ ] Проверить чтение всех полей из API

### 1.2 Основные поля
- [ ] `web_port` — изменить (вызывает restart панели)
- [ ] `web_base_path` — изменить (вызывает restart)
- [ ] `web_listen` — задать конкретный IP
- [ ] `web_domain` — задать домен
- [ ] `session_max_age` — изменить
- [ ] `page_size` — изменить
- [ ] `remark_model` — изменить
- [ ] `time_location` — изменить
- [ ] `date_picker` — gregorian/jalali

### 1.3 Restart панели
- [ ] Изменение `web_port` → панель перезапускается
- [ ] Изменение `web_base_path` → панель перезапускается
- [ ] Изменение `web_cert_file`/`web_key_file` → restart
- [ ] Изменение `session_max_age` → restart
- [ ] Изменение `page_size` → **не** вызывает restart

### 1.4 LDAP-настройки
- [ ] `ldap_enable` — включить/выключить
- [ ] `ldap_host`, `ldap_port` — задать
- [ ] `ldap_bind_dn`, `ldap_password`, `ldap_base_dn`
- [ ] `ldap_auto_create`, `ldap_auto_delete`
- [ ] `ldap_default_total_gb`, `ldap_default_expiry_days`, `ldap_default_limit_ip`

### 1.5 Идемпотентность
- [ ] `apply` без изменений → `No changes`
- [ ] Повторный apply после ручных изменений в UI → drift

---

## 2. `threexui_panel_security`

- [ ] Создать ресурс
- [ ] `two_factor_enable` — включить/выключить
- [ ] `two_factor_token` — задать/считать (sensitive)
- [ ] Идемпотентность

---

## 3. `threexui_panel_telegram`

- [ ] Создать ресурс
- [ ] `tg_bot_enable` — включить
- [ ] `tg_bot_token` — задать (sensitive)
- [ ] `tg_bot_chat_id` — задать
- [ ] `tg_run_time` — изменить (@daily, @weekly)
- [ ] `tg_bot_backup` — true/false
- [ ] `tg_bot_login_notify` — true/false
- [ ] `tg_cpu` — порог CPU
- [ ] `tg_lang` — язык
- [ ] Идемпотентность

---

## 4. `threexui_panel_subscription`

- [ ] Создать ресурс
- [ ] `sub_enable` — включить
- [ ] `sub_listen`, `sub_port`, `sub_path`, `sub_domain`
- [ ] `sub_cert_file`, `sub_key_file`
- [ ] `sub_json_enable` — JSON-эндпоинт
- [ ] `sub_encrypt` — шифрование
- [ ] `sub_show_info` — показывать инфо
- [ ] `sub_title`, `sub_support_url`, `sub_announce`
- [ ] `sub_json_fragment`, `sub_json_noises`, `sub_json_mux`, `sub_json_rules`
- [ ] Идемпотентность

---

## 5. Xray-секции

### 5.1 Общее для всех xray-ресурсов
- [ ] Поле `json` — принимает валидный JSON
- [ ] `DiffSuppressFunc` — эквивалентный JSON не вызывает diff
- [ ] `StateFunc` — JSON нормализуется
- [ ] Невалидный JSON → ошибка
- [ ] Пустая строка → ошибка

### 5.2 `threexui_xray_basics` (mode: merge root)
- [ ] Задать `log`, `policy`
- [ ] Проверить, что мержится с существующим конфигом (не затирает `dns`, `routing`)
- [ ] Обновить `log.loglevel` — проверить идемпотентность

### 5.3 `threexui_xray_dns` (path: `dns`)
- [ ] Задать DNS-серверы
- [ ] Проверить чтение обратно
- [ ] Обновить → проверить идемпотентность

### 5.4 `threexui_xray_routing` (path: `routing`)
- [ ] Задать правила маршрутизации
- [ ] Проверить чтение/обновление

### 5.5 `threexui_xray_balancers` (path: `routing.balancers`)
- [ ] Задать балансировщики
- [ ] Проверить вложенный путь

### 5.6 `threexui_xray_reverse` (path: `reverse`)
- [ ] Задать reverse-прокси
- [ ] Проверить чтение

### 5.7 `threexui_xray_outbounds` (path: `outbounds`)
- [ ] Задать outbound'ы (freedom, blackhole)
- [ ] Проверить чтение массива

### 5.8 `threexui_xray_advanced` (mode: replace all)
- [ ] Полная замена конфига
- [ ] Проверить, что старый конфиг полностью заменяется

---

## 6. Delete-поведение

- [ ] Все settings-ресурсы используют `resourceSettingsDelete`
- [ ] Delete не сбрасывает настройки на дефолт (ожидаемое поведение?)
- [ ] State очищается после destroy

---

## 7. Негативные сценарии

- [ ] Невалидный `web_port` (0, -1, 99999) → ошибка
- [ ] Невалидный JSON в xray-ресурсах → ошибка
- [ ] Конфликт `sub_port` = `web_port` → поведение?
- [ ] Пустой `tg_bot_token` при `tg_bot_enable = true` → поведение?

---

## 8. Результаты

- [ ] Все пункты выполнены
- [ ] Зафиксированы отклонения/баги
