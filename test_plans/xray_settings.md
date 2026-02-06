# План тестирования Xray-настроек

Дата: 2026-02-06

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

- [ ] Поле `json` — принимает валидный JSON
- [ ] `DiffSuppressFunc` — эквивалентный JSON не вызывает diff
- [ ] `StateFunc` — JSON нормализуется
- [ ] Невалидный JSON → ошибка
- [ ] Пустая строка → ошибка

---

## 2. `threexui_xray_basics` (mode: merge root)

- [ ] Задать `log`, `policy`
- [ ] Проверить, что мержится с существующим конфигом (не затирает `dns`, `routing`)
- [ ] Обновить `log.loglevel` — проверить идемпотентность

---

## 3. `threexui_xray_dns` (path: `dns`)

- [ ] Задать DNS-серверы
- [ ] Проверить чтение обратно
- [ ] Обновить → проверить идемпотентность

---

## 4. `threexui_xray_routing` (path: `routing`)

- [ ] Задать правила маршрутизации
- [ ] Проверить чтение/обновление

---

## 5. `threexui_xray_balancers` (path: `routing.balancers`)

- [ ] Задать балансировщики
- [ ] Проверить вложенный путь

---

## 6. `threexui_xray_reverse` (path: `reverse`)

- [ ] Задать reverse-прокси
- [ ] Проверить чтение

---

## 7. `threexui_xray_outbounds` (path: `outbounds`)

- [ ] Задать outbound'ы (freedom, blackhole)
- [ ] Проверить чтение массива

---

## 8. `threexui_xray_advanced` (mode: replace all)

- [ ] Полная замена конфига
- [ ] Проверить, что старый конфиг полностью заменяется

---

## 9. Негативные сценарии

- [ ] Невалидный JSON в xray-ресурсах → ошибка
- [ ] Пустая строка в `json` → ошибка

---

## 10. Результаты

- [ ] Все пункты выполнены
- [ ] Зафиксированы отклонения/баги
