resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

  vless_settings {
    decryption = "none"
  }

  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "www.apple.com:443"
      server_names = ["www.apple.com"]
    }
    tcp_settings {
      accept_proxy_protocol = false
      header_type           = "none"
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

resource "threexui_inbound_client" "user1" {
  inbound_id = threexui_inbound.vless.id
  email      = "user1@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
