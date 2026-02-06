# Terraform/OpenTofu провайдер для 3x-ui

Провайдер для управления инбаундами и клиентами 3x-ui через HTTP API панели.

## Конфигурация провайдера

```hcl
provider "threexui" {
  endpoint            = "http://localhost:2053"
  username            = "admin"
  password            = "admin"
  # base_path           = "/"           # опционально
  # insecure_skip_verify = true          # для self-signed HTTPS
  # request_timeout      = "30s"
}
```

## Ресурсы

### `threexui_inbound`

```hcl
resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
  protocol = "vless"

  # Опциональные настройки (пример для VLESS)
  # settings {
  #   decryption = "none"
  #   encryption = "none"
  # }
}
```

Основные поля:
- `remark` — описание инбаунда.
- `port` — порт прослушивания.
- `protocol` — протокол (`vless`, `vmess`, `trojan`, `shadowsocks`, ...).
- `settings` — JSON‑настройки инбаунда (без клиентов).

Полезно знать:
- `settings` инбаунда больше не управляет клиентами.
- Клиенты создаются только через `threexui_inbound_client`.

### `threexui_inbound_client`

```hcl
resource "threexui_inbound_client" "client_a" {
  inbound_id = threexui_inbound.example.id
  email      = "client-a@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
```

Основные поля:
- `inbound_id` — ID инбаунда.
- `email` — идентификатор клиента.
- `enable` — включён ли клиент.
- `flow` — flow для VLESS (`xtls-rprx-vision` и т.д.).
- `expiry_time` — время истечения в миллисекундах Unix epoch (число).
- `limit_ip` — лимит IP.
- `total_gb` — лимит трафика.
- `security` / `password` — используются для некоторых протоколов (чувствительные).

## Outputs (пример)

```hcl
output "inbound_clients" {
  value = {
    client_a = {
      id          = threexui_inbound_client.client_a.id
      client_id   = threexui_inbound_client.client_a.client_id
      email       = threexui_inbound_client.client_a.email
      enable      = threexui_inbound_client.client_a.enable
      flow        = threexui_inbound_client.client_a.flow
      limit_ip    = threexui_inbound_client.client_a.limit_ip
      total_gb    = threexui_inbound_client.client_a.total_gb
      expiry_time = threexui_inbound_client.client_a.expiry_time
      tg_id       = threexui_inbound_client.client_a.tg_id
      sub_id      = threexui_inbound_client.client_a.sub_id
      comment     = threexui_inbound_client.client_a.comment
      reset       = threexui_inbound_client.client_a.reset
      security    = threexui_inbound_client.client_a.security
    }
  }
}
```

## Импорт

```bash
# inbound
terraform import threexui_inbound.example 123

# inbound client: <inbound_id>:<client_id>
terraform import threexui_inbound_client.client_a 123:client-id
```
