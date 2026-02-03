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

  # settings/stream_settings/sniffing теперь опциональны.
  # Если settings не задан, провайдер подставит дефолт как в UI 3x-ui:
  # vless/vmess/trojan/shadowsocks — один случайный клиент + нужные поля.
  # settings {
  #   decryption = "none"
  #   encryption = "none"
  #   clients {
  #     email  = "client@example.com"
  #     enable = true
  #   }
  # }
  # stream_settings {
  #   network  = "tcp"
  #   security = "reality"
  #
  #   external_proxy = []
  #
  #   reality_settings {
  #     show          = false
  #     xver          = 0
  #     target        = "caddy:443"
  #     server_names  = ["ns-k1.lifelink.space", "www.ns-k1.lifelink.space"]
  #     private_key   = "..."
  #     min_client_ver = ""
  #     max_client_ver = ""
  #     max_timediff  = 0
  #     short_ids     = ["af9094bc", "01"]
  #     mldsa65_seed  = ""
  #     settings {
  #       public_key    = "..."
  #       fingerprint   = "chrome"
  #       server_name   = ""
  #       spider_x      = "/"
  #       mldsa65_verify = ""
  #     }
  #   }
  #
  #   tcp_settings {
  #     accept_proxy_protocol = false
  #     header {
  #       type = "none"
  #     }
  #   }
  # }
  # sniffing {
  #   enabled       = true
  #   dest_override = ["http", "tls", "quic", "fakedns"]
  #   metadata_only = false
  #   route_only    = false
  # }
}
```

### threexui_inbound_client

```hcl
resource "threexui_inbound_client" "example" {
  inbound_id = threexui_inbound.example.id
  email      = "client2@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
  # client_id можно не задавать — он генерируется автоматически
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
