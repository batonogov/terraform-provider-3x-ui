# Дизайн схем Terraform ↔ 3x-ui JSON

## Общие принципы
- Структуры Terraform максимально повторяют `database/model` из 3x-ui, чтобы избежать несовместимостей.
- Поля с типом `map[string]any` в API (например, `settings`, `stream`) раскладываются на вложенные блоки и типизированные атрибуты.
- Все числовые значения трафика (`totalUp`, `totalDown`, `up`/`down`) хранятся в `int64`.
- Даты — Unix millis (`expiry_time`) → в Terraform представляем как `int64` или используем `time` (позже можно добавить `TypeTimestamp`).
- Для bool-полей используем `types.Bool`, по возможности указываем значения по умолчанию.

## Пример: `3xui_inbound`
Исходная модель (`database/model/inbound.go`):
```go
type Inbound struct {
    Id          int    `json:"id"`
    UserId      int    `json:"userId"`
    Remark      string `json:"remark"`
    Listen      string `json:"listen"`
    Port        int    `json:"port"`
    Protocol    string `json:"protocol"`
    Settings    string `json:"settings"`
    StreamSettings string `json:"stream"`
    Sniffing    string `json:"sniffing"`
    Allocate    string `json:"allocate"`
    Tag         string `json:"tag"`
    Up          int64  `json:"up"`
    Down        int64  `json:"down"`
    Total       int64  `json:"total"`
    Enable      bool   `json:"enable"`
    ExpiryTime  int64  `json:"expiryTime"`
    Clients     string `json:"clients"`
    /* ... */
}
```
В API `settings`, `stream`, `sniffing`, `allocate`, `clients` передаются как JSON-строки. В провайдере:
- Добавляем вложенный блок `settings { ... }`, который мапится на разные типы в зависимости от `protocol` (например, `vmess_settings`, `vless_settings`).
- Блок `stream` → `stream { network, security, tls { server_name, ... }, reality { ... }, ws { path, headers }, grpc { service_name, multi_mode } }`.
- `clients` → отдельные вложенные блоки, либо вынесены в ресурс `3xui_user`.
- `sniffing` → `sniffing { enabled, dest_override = list(string) }`.
- `allocate` → `allocate { strategy, concurrency, refresh }`.

### HCL-структура
```hcl
resource "3xui_inbound" "example" {
  protocol = "vless"
  listen   = "0.0.0.0"
  port     = 443
  remark   = "prod inbound"
  enable   = true

  stream {
    network  = "tcp"
    security = "tls"
    tls {
      sni         = "example.com"
      cert_path   = var.cert_path
      key_path    = var.key_path
      reality     = false
    }
  }

  settings_vless {
    clients = [
      {
        id           = "uuid"
        flow         = "xtls-rprx-vision"
        email        = "user@example.com"
        limit_ip     = 2
        total_ip     = 0
        enable       = true
        expiry_time  = 0
      }
    ]
  }
}
```
Сериализация: при `Create`/`Update` блоки конвертируются в структуры Go, которые затем маршалятся в JSON-строки (`settings`, `stream`).

## Валидаторы
- `protocol` — enum.
- `port` — 1..65535.
- `listen` — IPv4/IPv6 или пустая строка.
- `clients.email` — valid email (regexp).
- `stream.network` — enum (`tcp`, `ws`, `grpc`, `httpupgrade`, `splithttp`...).
- `stream.security` — enum (`none`, `tls`, `reality`...).

## Nested-to-string handling
Поскольку API ожидает строки с JSON, вводим helper функции:
```go
type JSONString[T any] struct {
    Raw string
    Obj *T
}
```
В ресурсе заполняем `Obj` из схемы, а клиент маршалит `Obj` в `Raw`. При чтении из API — наоборот (unmarshal строки в структуру и записываем в state).

## Data source schema
- Дата-сорсы просто вычисляемые поля, поэтому многие строковые JSONы можно отдавать как `map[string]any` или `string`. Если нужно давать пользователям сложные структуры (например, `server_status`), используем вложенные объекты.

## Дополнительные поля
- Для некоторых полей UI хранит `null`/пустые строки. В провайдере используем `Optional` + `Computed`, чтобы поддерживать «проброс» значений от сервера.
- Импорт: `ImportStatePassthroughID` + кастомный `Decode` для разделённых ID (`inboundID/clientID`).

## Тестирование схем
- Табличные тесты, которые проверяют round-trip: Terraform config → Go struct → JSON → обратно.
- Используем fixtures из `testdata/*.json` (пример реального inbound).
