resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

  stream_settings = jsonencode({
    network  = "tcp"
    security = "reality"
    realitySettings = {
      dest        = "www.apple.com:443"
      serverNames = ["www.apple.com"]
    }
    tcpSettings = {
      acceptProxyProtocol = false
      header = { type = "none" }
    }
  })

  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls", "quic", "fakedns"]
  })
}

resource "threexui_inbound_client" "user1" {
  inbound_id = threexui_inbound.vless.id
  email      = "user1@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
