# Shadowsocks inbound with AEAD cipher.

terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

resource "threexui_inbound" "shadowsocks" {
  port     = 9001
  protocol = "shadowsocks"
  remark   = "Shadowsocks"
  enable   = true

  shadowsocks_settings {
    method   = "aes-256-gcm"
    password = "change-me-to-a-strong-password"
    network  = "tcp,udp"
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}
