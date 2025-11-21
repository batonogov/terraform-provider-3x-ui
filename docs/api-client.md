# Архитектура API-клиента

## Цели
- Инкапсулировать HTTP-взаимодействие с 3x-ui (логин, cookie, REST-запросы).
- Обеспечить повторное использование между ресурсами/дата-сорсами.
- Упростить тестирование за счёт интерфейсов и подменяемого транспорта.

## Пакет `internal/client`
Структура:
```
internal/
  client/
    client.go       // публичный конструктор и интерфейс
    http.go         // низкоуровневые HTTP вызовы, retry/backoff
    auth.go         // логин/управление cookie
    types.go        // структуры запросов/ответов (Inbound, ServerStatus и т.д.)
    errors.go       // нормализация ошибок API
```

### Интерфейс
```go
type Client interface {
    ListInbounds(ctx context.Context) ([]Inbound, error)
    GetInbound(ctx context.Context, id int) (*Inbound, error)
    CreateInbound(ctx context.Context, req InboundRequest) (*Inbound, error)
    UpdateInbound(ctx context.Context, id int, req InboundRequest) (*Inbound, error)
    DeleteInbound(ctx context.Context, id int) error

    AddClient(ctx context.Context, inboundID int, payload ClientRequest) (*Client, error)
    // ... остальные методы (reset traffic, list clients, server status, etc.)

    ServerStatus(ctx context.Context) (*ServerStatus, error)
    RestartXray(ctx context.Context) error
}
```
Интерфейс будет расширяться по мере добавления ресурсов. Для тестов можно реализовывать мок, чтобы проверять логику `Resource` без настоящей панели.

### Конфигурация клиента
```go
type Config struct {
    BaseURL        *url.URL
    Username       string
    Password       string
    SessionCookie  string
    TLSSkipVerify  bool
    RequestTimeout time.Duration
    MaxRetries     int
    Logger         hclog.Logger
}
```
Конструктор `New(Config)` валидирует параметры, создаёт `http.Client` с:
- кастомным `http.Transport` (TLS настройки, `DisableCompression=false`, `MaxIdleConns` >= 10).
- cookie jar (`http.CookieJar`) для хранения сессии.
- middleware retry/backoff (экспоненциальный с jitter).

### Аутентификация
- Если `SessionCookie` пуст, клиент вызывает `login(ctx)` один раз при первом запросе (ленивая инициализация).
- `login` отправляет `POST /login` с `application/x-www-form-urlencoded` телом, парсит ответ (ожидает `success=true`).
- Cookie сохраняется в jar автоматически; дополнительно можно кешировать expiration (по MaxAge cookie).
- В диагностике Terraform выводим подсказки: неверный пароль, включена 2FA, истёк сертификат.

### Выполнение запросов
`do(ctx, method, path, body, opts...)`:
1. Собирает URL: `baseURL.ResolveReference(path)`.
2. Добавляет нужные заголовки (`Content-Type`, `Accept: application/json`).
3. Сериализует тело через `json.Marshal` или `url.Values` (в зависимости от эндпоинта).
4. Выполняет запрос с контекстом и таймаутом.
5. Если ответ `401/404`, повторяет логин (в случае истёкшей сессии) и ретраит один раз.
6. Декодирует `entity.Msg` (поле `success`, `msg`, `obj`). При `success=false` возвращает `*APIError{Message, Code, RawObj}`.

### Ошибки и диагностика
`errors.go` содержит тип:
```go
type APIError struct {
    StatusCode int
    Message    string
    Raw        json.RawMessage
}
```
Клиент оборачивает сетевые ошибки (`context.DeadlineExceeded`, `net.Error`) в `RetryableError`, чтобы retry-механизм мог обрабатывать их автоматически.

### Логирование
Используем `terraform-plugin-log/tflog`. Клиент принимает `tflog.Logger` или `context.Context` и записывает отладочные сообщения при `TF_LOG=DEBUG` (например, URL, длительность, ответ).

### Тесты
- Юнит-тесты с `httptest.Server`, проверяющие логин, retry, десериализацию.
- Табличные тесты на ошибки (404, 500, `success=false`).
- Файловые фикстуры JSON для типовых ответов (`testdata/*.json`).

## Взаимодействие с ресурсами
Каждый `Resource`/`DataSource` получает `client.Client` в `Configure`. В CRUD методах вызываются соответствующие методы клиента. Это снижает дублирование и упрощает мокинг в юнит-тестах.
