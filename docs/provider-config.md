# Конфигурация провайдера 3x-ui

Провайдер на Terraform/OpenTofu будет называться `3xui`. Основные параметры конфигурации, задаваемые в `provider "3xui" { ... }`:

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `base_url` | string | да | Базовый URL панели 3x-ui, например `https://localhost:2053`. Должен включать схему (http/https) и порт. Используется для всех API-запросов. |
| `username` | string | да* | Имя пользователя панели. Обязательно, если не используется `session_cookie`. |
| `password` | string | да* | Пароль пользователя панели. Обязателен вместе с `username`, если не указан `session_cookie`. |
| `session_cookie` | string | нет | Готовое значение cookie (например, `session=...`) для случаев, когда логин выполняется вне провайдера. Если задано, провайдер не выполняет `/login`. |
| `tls_skip_verify` | bool | нет (по умолчанию `false`) | Принудительно отключает проверку TLS-сертификата (полезно для self-signed dev-стендов). |
| `request_timeout` | string | нет (по умолчанию `30s`) | Таймаут HTTP-запросов (формат Go duration). |
| `poll_interval` | string | нет (по умолчанию `5s`) | Интервал ожидания при операциях, требующих повторных запросов (например, ожидание применения изменений, если потребуется). |
| `max_retries` | number | нет (по умолчанию `3`) | Количество повторов при временных ошибках (timeouts, 5xx). |

`username`/`password` и `session_cookie` взаимоисключающие: провайдер либо сам логинится и хранит cookie в памяти, либо получает cookie от пользователя. Это важно, т.к. API 3x-ui не предоставляет отдельного токена.

## Поведение Configure
1. Провайдер валидирует `base_url` (парсинг URL и проверка схемы).
2. Если указаны `username`/`password`, выполняет POST `/login` (формат `application/x-www-form-urlencoded`), обрабатывает 2FA (на первом этапе просто ошибка, если сервер её требует).
3. Сохраняет `session` cookie в клиенте (использует cookie jar) и выполняет тестовый запрос `GET /panel/api/inbounds/list`, чтобы проверить доступ.
4. Если задан `session_cookie`, она добавляется в cookie jar без логина.
5. При ошибках отдаёт диагностические сообщения Terraform (`tfsdk.Diagnostics`).

## Пример конфигурации
```hcl
provider "3xui" {
  base_url        = "https://localhost:2053"
  username        = var.threexui_username
  password        = var.threexui_password
  tls_skip_verify = true
}
```

Для окружений с внешним логином:
```hcl
provider "3xui" {
  base_url       = "https://panel.example.com"
  session_cookie = var.threexui_session
}
```

## Переменные окружения
Параметры можно задавать через env vars (Terraform Plugin Framework поддерживает `TF_VAR` и собственные env). Планируем сопоставление:
- `THREEXUI_BASE_URL`
- `THREEXUI_USERNAME`
- `THREEXUI_PASSWORD`
- `THREEXUI_SESSION_COOKIE`
- `THREEXUI_TLS_SKIP_VERIFY`
- `THREEXUI_REQUEST_TIMEOUT`
- `THREEXUI_POLL_INTERVAL`
- `THREEXUI_MAX_RETRIES`

## Будущие расширения
- Поддержка 2FA (возможно, через TOTP секрет, если панель позволит автоматизацию).
- Возможность прокси (`http_proxy`, `https_proxy`).
- Поддержка client certificates.
