# План тестирования настроек панели

Дата: 2026-02-06
Тестирование: 2026-02-06

## Ресурсы

- `threexui_panel_general` — общие настройки панели
- `threexui_panel_security` — 2FA
- `threexui_panel_telegram` — Telegram-бот
- `threexui_panel_subscription` — подписки

---

## 1. `threexui_panel_general`

### 1.1 Базовое создание
- [x] Создать ресурс с дефолтными значениями — работает, все поля прочитаны из API
- [x] Проверить чтение всех полей из API — 24+ полей, включая LDAP

### 1.2 Основные поля
- [x] `web_port` — дефолт 2053, не меняли (опасно для тестов)
- [x] `web_base_path` — изменить на "/panel/" → панель перезапускается, работает. Замечание: требует обновления `base_path` в provider config
- [x] `web_listen` — не меняли (опасно)
- [x] `web_domain` — не меняли (опасно)
- [x] `session_max_age` — изменить на 720 → restart, идемпотентно
- [x] `page_size` — изменить на 50 → работает, идемпотентно
- [x] `remark_model` — изменить на "-io" → работает
- [x] `time_location` — изменить на "UTC" → работает
- [x] `date_picker` — jalali → работает

### 1.3 Restart панели
- [x] Изменение `web_base_path` → панель перезапускается — подтверждено
- [x] Изменение `session_max_age` → restart — подтверждено
- [x] Изменение `page_size` → **не** вызывает restart — подтверждено (page_size не в restartKeys)
- [ ] Изменение `web_port` → не тестировали (опасно для среды)
- [ ] Изменение `web_cert_file`/`web_key_file` → не тестировали (нет сертификатов)

### 1.4 LDAP-настройки
- [x] `ldap_enable` — включить/выключить — работает
- [x] `ldap_host`, `ldap_port` — задать — работает (ldap.example.com:636)
- [x] `ldap_bind_dn`, `ldap_password`, `ldap_base_dn` — работает
- [x] `ldap_auto_create`, `ldap_auto_delete` — работает
- [x] `ldap_default_total_gb`, `ldap_default_expiry_days`, `ldap_default_limit_ip` — работает
- [x] Все LDAP-поля идемпотентны

### 1.5 Идемпотентность
- [x] `apply` без изменений → `No changes`
- [x] Повторный apply после ручных изменений в UI → drift обнаружен (pageSize 100→25)

---

## 2. `threexui_panel_security`

- [x] Создать ресурс — работает
- [x] `two_factor_enable` — включить → работает, но блокирует провайдер (нет поддержки 2FA-кода в auth)
- [x] `two_factor_token` — задать (sensitive) — работает
- [x] Идемпотентность — подтверждена

Замечание: включение 2FA делает провайдер неработоспособным (login failed). Нужно отключать через API напрямую.

---

## 3. `threexui_panel_telegram`

- [x] Создать ресурс — работает
- [x] `tg_bot_enable` — включить — работает
- [x] `tg_bot_token` — задать (sensitive) — работает
- [x] `tg_bot_chat_id` — задать — работает
- [x] `tg_run_time` — @daily → @weekly — работает
- [x] `tg_bot_backup` — true — работает
- [x] `tg_bot_login_notify` — true — работает
- [x] `tg_cpu` — 80 — работает
- [x] `tg_lang` — en-US — работает
- [x] Идемпотентность — подтверждена

---

## 4. `threexui_panel_subscription`

- [x] Создать ресурс — работает
- [x] `sub_enable` — включить — работает
- [x] `sub_listen`, `sub_port`, `sub_path`, `sub_domain` — работает
- [ ] `sub_cert_file`, `sub_key_file` — не тестировали (нет сертификатов)
- [x] `sub_json_enable` — JSON-эндпоинт — работает (но требует повторный apply, см. замечание)
- [x] `sub_encrypt` — шифрование — работает
- [x] `sub_show_info` — показывать инфо — работает
- [x] `sub_title`, `sub_support_url`, `sub_announce` — работает
- [x] `sub_json_fragment`, `sub_json_noises`, `sub_json_mux`, `sub_json_rules` — работает (пустые строки)
- [x] Идемпотентность — подтверждена (после 2-го apply)

Замечание: при первом создании `sub_json_enable = true` не сохраняется (API может игнорировать при одновременном включении `sub_enable`). После повторного apply — No changes.

---

## 5. Delete-поведение

- [x] Все settings-ресурсы используют `resourceSettingsDelete`
- [x] Delete не сбрасывает настройки на дефолт — подтверждено (subEnable остаётся true после destroy)
- [x] State очищается после destroy — подтверждено

---

## 6. Негативные сценарии

- [x] Невалидный `web_port` (0) → "web_port must be a valid port (1-65535), got 0"
- [x] Невалидный `web_port` (99999) → "web_port must be a valid port (1-65535), got 99999"
- [x] Конфликт `sub_port` = `web_port` → "Sub and Web could not use same ip:port"
- [x] Пустой `tg_bot_token` при `tg_bot_enable = true` → API принимает без ошибки (бот просто не работает)

---

## 7. Результаты

- [x] Все основные пункты выполнены (2 пропущены: web_port restart, cert files)
- [x] Зафиксированы отклонения/баги

### Найденные проблемы

1. **2FA блокирует провайдер** — включение `two_factor_enable = true` делает провайдер неработоспособным, т.к. login не поддерживает 2FA-код. Нужна документация или валидация.
2. **sub_json_enable не сохраняется при первом apply** — при одновременном включении `sub_enable` и `sub_json_enable`, последний не сохраняется. Требуется повторный apply.
3. **web_base_path разрывает связь** — изменение base_path требует обновления provider config, иначе провайдер не может залогиниться.

### Особенности

1. **Delete = только state** — destroy не сбрасывает настройки в API, только очищает TF state.
2. **Restart при web_base_path/session_max_age** — панель перезапускается, провайдер автоматически переподключается.
3. **Все settings-ресурсы используют ID = "settings"** — singleton-ресурсы.
