# Shadowsocks inbound.
#
# Cipher choice: 3x-ui v2.9.3 removed the legacy AEAD ciphers
# (aes-128-gcm, aes-256-gcm) from the xray user-registration switch.
# Configs that still pass them silently route through Shadowsocks-2022
# using the password as the 2022 key, so AEAD clients stop connecting.
# chacha20-ietf-poly1305 works on every supported 3x-ui version; for
# 2022-edition use a 2022-blake3-* method with a base64 32-byte key.

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 2.0"
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
    method   = "chacha20-ietf-poly1305"
    password = "change-me-to-a-strong-password"
    network  = "tcp,udp"
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}
