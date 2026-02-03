# Terraform/OpenTofu провайдер для 3x-ui

Провайдер для управления inbounds и клиентами 3x-ui через HTTP API панели.

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

### threexui_inbound

```hcl
resource "threexui_inbound" "example" {
  port     = 23456
  protocol = "vless"
  remark   = "example-inbound"
  enable   = true

  settings = jsonencode({
    clients = [{
      id    = "11111111-1111-1111-1111-111111111111"
      email = "client@example.com"
      flow  = "xtls-rprx-vision"
    }]
    decryption = "none"
  })

  stream_settings = jsonencode({})
  sniffing        = jsonencode({})
}
```

### threexui_inbound_client

```hcl
resource "threexui_inbound_client" "example" {
  inbound_id = threexui_inbound.example.id
  client_id  = "22222222-2222-2222-2222-222222222222"
  email      = "client2@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
```

## Data sources

```hcl
data "threexui_inbounds" "all" {}

data "threexui_server_status" "status" {}

data "threexui_xray_versions" "versions" {}

data "threexui_xray_config" "config" {}

data "threexui_settings" "settings" {}
```

## Импорт

```bash
# inbound
terraform import threexui_inbound.example 123

# inbound client: <inbound_id>:<client_id>
terraform import threexui_inbound_client.example 123:client-id
```

## Acceptance-тесты

```bash
TF_ACC=1 \
THREEXUI_ENDPOINT=http://localhost:2053 \
THREEXUI_USERNAME=admin \
THREEXUI_PASSWORD=admin \
go test ./provider -run TestAcc
```

## Примеры

Смотри `examples/`.
