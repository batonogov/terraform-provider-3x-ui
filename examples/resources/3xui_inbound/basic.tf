resource "3xui_inbound" "basic" {
  protocol = "vless"
  listen   = "0.0.0.0"
  port     = 443
  remark   = "basic inbound"
  enable   = true

  stream {
    network  = "tcp"
    security = "tls"
    tls {
      sni = "example.com"
    }
  }

  settings_vless {
    clients = [{
      id    = "00000000-0000-4000-8000-000000000000"
      email = "user@example.com"
    }]
  }
}
