# Обзор API 3x-ui

Документ фиксирует основные HTTP-эндпоинты панели 3x-ui (https://github.com/MHSanaei/3x-ui) для дальнейшей реализации Terraform/OpenTofu-провайдера.

## Базовая информация
- **Базовый путь**: `https://<host>:<port>/` (по умолчанию 2053 для HTTPS или 9999 для HTTP в docker-compose).
- **Панель**: UI живёт под `/panel`, API — под `/panel/api`.
- **Формат**: REST JSON, но поля и структуры совпадают с моделями `database/model` в исходниках.

## Аутентификация и сессии
- Логин выполняется POST `/login` (form-urlencoded): поля `username`, `password`, опционально `twoFactorCode`.
- При успешной аутентификации сервер устанавливает cookie с сериализованным `model.User` (Gin sessions). Cookie имеет `HttpOnly`, `SameSite=Lax`, `Path=/` и TTL `sessionMaxAge` минут (настраивается в панели).
- Все маршруты `/panel/api/**` проходят middleware `checkAPIAuth`, который возвращает `404` при отсутствии валидной сессии. API-ключей или токенов нет.
- Для автоматизации потребуется хранить cookie сессии или выполнять логин программно; иначе API недоступно. HTTPS обязателен, чтобы защитить пароль/куки.

## Разделы API
Главные группы определены в `web/controller/api.go`:

### `/panel/api/inbounds`
Работа с inbound'ами и клиентами. Основные методы:
- `GET /list` — список инбаундов текущего пользователя.
- `GET /get/:id` — данные одного inbound (ID в БД).
- `GET /getClientTraffics/:email` и `/getClientTrafficsById/:id` — статистика клиентов.
- `POST /add`, `/update/:id`, `/del/:id` — CRUD инбаундов; тело соответствует `model.Inbound`.
- `POST /addClient`, `/updateClient/:clientId`, `/:id/delClient/:clientId`, `/:id/delClientByEmail/:email` — управление клиентами внутри inbound.
- `POST /clientIps/:email` и `/clearClientIps/:email` — просмотр/очистка IP.
- `POST /resetAllTraffics`, `/resetAllClientTraffics/:id`, `/resetAllTraffics`, `/:id/resetClientTraffic/:email` — сброс статистики.
- `POST /delDepletedClients/:id` — удалить клиентов с исчерпанным трафиком.
- `POST /onlines` и `/lastOnline` — онлайн статусы клиентов.
- `POST /updateClientTraffic/:email` — ручное изменение трафика.
- `POST /import` — импорт инбаундов (ожидает JSON с полем `data`).

Дополнительно в шаблонах HTML используются эндпоинты:
- `POST /addClient`, `/updateClient` с payload `clientStats`, `settings`, `stream`, и т. д.
- `POST /inbounds/resetAllClientTraffics/:id`, `/inbounds/delDepletedClients/:id` и др. (см. `web/html/inbounds.html`).

### `/panel/api/server`
Мониторинг и операции над Xray/сервером. Методы:
- `GET /status` — агрегированный статус (CPU, RAM, диски, аптайм и т. п.).
- `GET /cpuHistory/:bucket` — исторический CPU (bucket: 2, 30, 60, 120, 180, 300).
- `GET /getXrayVersion` — список доступных версий Xray (кеш 60 с).
- `GET /getConfigJson` — текущий конфиг Xray.
- `GET /getDb` — дамп SQLite базы (используется UI для бэкапов).
- `GET /getNewUUID`, `/getNewX25519Cert`, `/getNewmldsa65`, `/getNewmlkem768`, `/getNewVlessEnc` — генерация служебных значений.
- `POST /stopXrayService`, `/restartXrayService` — управление сервисом.
- `POST /installXray/:version` — обновление Xray.
- `POST /updateGeofile` и `/updateGeofile/:fileName` — обновить geo файлы (валидируется имя файла).
- `POST /logs/:count`, `/xraylogs/:count` — получение логов (body включает фильтры `level`, `syslog`, `filter`, `showDirect` и т. д.).
- `POST /importDB` — импорт БД из загруженного файла.
- `POST /getNewEchCert` — генерация ECH сертификата (ожидает `{ "sni": "example.com" }`).

### Прочие маршруты
- `GET /panel/api/backuptotgbot` — инициирует отправку бэкапа в Telegram (использует `service.Tgbot`).
- Через HTML интерфейс видны и другие POST-запросы (например, `/panel/api/server/updateGeofile` с формой, `/panel/api/server/importDB` с `multipart/form-data`).

## Типовые ответы
Почти все обработчики используют `jsonMsg`/`jsonMsgObj`/`jsonObj` (`web/controller/base.go`). Структура `entity.Msg`:
```json
{
  "success": true,
  "msg": "localized message",
  "obj": { ...optional payload... }
}
```
При ошибке `success=false`, `msg` содержит локализованную подпись, `obj` может содержать данные или быть `null`.

## Ограничения и особенности
- API не предназначено для внешней интеграции: нет rate limit, версионирования или токенов; все изменения могут ломать совместимость.
- Ответы зависят от локализации (переводы). Для провайдера лучше ориентироваться на поле `success` и код HTTP.
- Некоторые операции (логин, импорт БД) требуют `multipart/form-data` или `application/x-www-form-urlencoded`, остальные — JSON.
- При использовании автоматизации учитывать, что панель фиксирует все логины и отправляет уведомления в Telegram (если настроено).
- Для сценариев без сессий можно рассмотреть реверс-прокси, который подставляет cookie, но это выходит за рамки API.

## Следующие шаги
- Детализировать схему `model.Inbound`, `model.Client` для маппинга в Terraform.
- Проверить, какие эндпоинты нужны для cron, подписок, панельных настроек (`web/controller/setting.go` не экспортируется в API — возможно, придётся использовать внутренние страницы).
