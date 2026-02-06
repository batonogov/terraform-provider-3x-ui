# План тестирования Xray-настроек

Дата: 2026-02-06
Тестирование: 2026-02-06

## Ресурсы

- `threexui_xray_basics` — базовые настройки Xray (mode: merge root)
- `threexui_xray_dns` — DNS Xray (path: `dns`)
- `threexui_xray_routing` — маршрутизация Xray (path: `routing`)
- `threexui_xray_balancers` — балансировщики Xray (path: `routing.balancers`)
- `threexui_xray_reverse` — reverse proxy Xray (path: `reverse`)
- `threexui_xray_outbounds` — outbound'ы Xray (path: `outbounds`)
- `threexui_xray_advanced` — полная замена конфига Xray (mode: replace all)

---

## 1. Общее для всех xray-ресурсов

- [x] Поле `json` — принимает валидный JSON
- [x] `DiffSuppressFunc` — эквивалентный JSON не вызывает diff (formatted vs compact — подтверждено)
- [x] `StateFunc` — JSON нормализуется (compact form)
- [x] Невалидный JSON → ошибка: "json must be valid JSON: invalid character..."
- [x] Пустая строка → не ошибка: `GetOkExists` считает "" как "не задано", ресурс просто читает текущий конфиг

---

## 2. `threexui_xray_basics` (mode: merge root)

- [x] Задать `log` — создание работает
- [x] Проверить, что мержится с существующим конфигом — работает (мьютекс предотвращает гонку)
- [x] Обновить `log.loglevel` — идемпотентность подтверждена (jsonSubsetDiffSuppress)

---

## 3. `threexui_xray_dns` (path: `dns`)

- [x] Задать DNS-серверы — работает
- [x] Проверить чтение обратно — JSON совпадает
- [x] Обновить (добавить серверы) → идемпотентность подтверждена

---

## 4. `threexui_xray_routing` (path: `routing`)

- [x] Задать правила маршрутизации (domainStrategy, rules) — работает
- [x] Проверить чтение/обновление — идемпотентность подтверждена

---

## 5. `threexui_xray_balancers` (path: `routing.balancers`)

- [x] Задать балансировщики (массив) — работает
- [x] Проверить вложенный путь — идемпотентность подтверждена

---

## 6. `threexui_xray_reverse` (path: `reverse`)

- [x] Задать reverse-прокси (bridges) — работает
- [x] Проверить чтение — идемпотентность подтверждена

---

## 7. `threexui_xray_outbounds` (path: `outbounds`)

- [x] Задать outbound'ы (freedom, blackhole) — работает
- [x] Проверить чтение массива — JSON-массив корректно сохраняется и читается

---

## 8. `threexui_xray_advanced` (mode: replace all)

- [x] Полная замена конфига — работает
- [x] Проверить, что старый конфиг полностью заменяется — подтверждено (policy, routing, dns исчезли)
- [x] Идемпотентность — подтверждена

---

## 9. Негативные сценарии

- [x] Невалидный JSON в xray-ресурсах → ошибка "json must be valid JSON: ..."
- [x] Пустая строка в `json` → не ошибка (GetOkExists zero value), ресурс читает текущий конфиг

---

## 10. Результаты

- [x] Все пункты выполнены
- [x] Зафиксированы отклонения/баги

### Найденные и исправленные проблемы

1. **Merge root mode не идемпотентен** — `extractXraySection` возвращает 4 фиксированных ключа из полного конфига, а пользователь задал только часть → вечный diff. **Исправлено**: используется `jsonSubsetDiffSuppress` вместо `jsonEqualDiffSuppress` для merge root — diff подавляется если config является подмножеством state.

2. **Race condition при параллельном создании xray-ресурсов** — оба читают конфиг, модифицируют и пишут, второй затирает изменения первого. **Исправлено**: добавлен `xrayTemplateMu sync.Mutex` в `resourceXraySectionApply` — операции сериализуются.

### Особенности

1. **Пустая строка `json = ""` не вызывает ошибку** — SDK `GetOkExists` считает пустую строку как "не задано", ресурс просто читает текущий конфиг.
2. **Delete = только state** — destroy только очищает TF state, не сбрасывает xray-конфиг.
3. **Все ресурсы идемпотентны** после исправлений.
