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

### Разделенные настройки (по вкладкам)

```hcl
resource "threexui_panel_general" "panel" {
  web_port      = 2053
  web_base_path = "/"
}

Mapping (panel/general.html -> API -> resource fields):

| UI control (v-model) | API key | Resource field |
| --- | --- | --- |
| allSetting.webListen | webListen | web_listen |
| allSetting.webDomain | webDomain | web_domain |
| allSetting.webPort | webPort | web_port |
| allSetting.webBasePath | webBasePath | web_base_path |
| allSetting.sessionMaxAge | sessionMaxAge | session_max_age |
| allSetting.pageSize | pageSize | page_size |
| remarkModel | remarkModel | remark_model |
| datepicker | datepicker | date_picker |
| allSetting.timeLocation | timeLocation | time_location |
| allSetting.expireDiff | expireDiff | expire_diff |
| allSetting.trafficDiff | trafficDiff | traffic_diff |
| allSetting.webCertFile | webCertFile | web_cert_file |
| allSetting.webKeyFile | webKeyFile | web_key_file |
| allSetting.externalTrafficInformEnable | externalTrafficInformEnable | external_traffic_inform_enable |
| allSetting.externalTrafficInformURI | externalTrafficInformURI | external_traffic_inform_uri |
| allSetting.ldapEnable | ldapEnable | ldap_enable |
| allSetting.ldapHost | ldapHost | ldap_host |
| allSetting.ldapPort | ldapPort | ldap_port |
| allSetting.ldapUseTLS | ldapUseTLS | ldap_use_tls |
| allSetting.ldapBindDN | ldapBindDN | ldap_bind_dn |
| allSetting.ldapPassword | ldapPassword | ldap_password |
| allSetting.ldapBaseDN | ldapBaseDN | ldap_base_dn |
| allSetting.ldapUserFilter | ldapUserFilter | ldap_user_filter |
| allSetting.ldapUserAttr | ldapUserAttr | ldap_user_attr |
| allSetting.ldapVlessField | ldapVlessField | ldap_vless_field |
| allSetting.ldapSyncCron | ldapSyncCron | ldap_sync_cron |
| allSetting.ldapFlagField | ldapFlagField | ldap_flag_field |
| allSetting.ldapTruthyValues | ldapTruthyValues | ldap_truthy_values |
| allSetting.ldapInvertFlag | ldapInvertFlag | ldap_invert_flag |
| ldapInboundTagList | ldapInboundTags | ldap_inbound_tags |
| allSetting.ldapAutoCreate | ldapAutoCreate | ldap_auto_create |
| allSetting.ldapAutoDelete | ldapAutoDelete | ldap_auto_delete |
| allSetting.ldapDefaultTotalGB | ldapDefaultTotalGB | ldap_default_total_gb |
| allSetting.ldapDefaultExpiryDays | ldapDefaultExpiryDays | ldap_default_expiry_days |
| allSetting.ldapDefaultLimitIP | ldapDefaultLimitIP | ldap_default_limit_ip |

Note: `lang` (UI language selector) is stored in browser cookies and is not part of `/panel/setting/update`, so it cannot be managed via `threexui_panel_general`.

resource "threexui_panel_security" "account" {
  two_factor_enable = true
}

resource "threexui_panel_telegram" "telegram" {
  tg_bot_enable = true
  tg_bot_token  = "..."
}

resource "threexui_panel_subscription" "subscription" {
  sub_enable = true
  sub_title  = "My Sub"
}
```

### Xray настройки (по вкладкам)

```hcl
resource "threexui_xray_basics" "basics" {
  json = jsonencode({
    log = {
      loglevel = "warning"
    }
  })
}

resource "threexui_xray_dns" "dns" {
  json = jsonencode({
    servers = [
      "1.1.1.1",
      "8.8.8.8"
    ]
  })
}

resource "threexui_xray_routing" "routing" {
  json = jsonencode({
    domainStrategy = "AsIs"
    rules          = []
  })
}

resource "threexui_xray_balancers" "balancers" {
  json = jsonencode([])
}

resource "threexui_xray_reverse" "reverse" {
  json = jsonencode({
    portals = []
    bridges = []
  })
}

resource "threexui_xray_outbounds" "outbounds" {
  json = jsonencode([])
}

resource "threexui_xray_advanced" "advanced" {
  json = jsonencode({
    log = {
      loglevel = "warning"
    }
  })
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
