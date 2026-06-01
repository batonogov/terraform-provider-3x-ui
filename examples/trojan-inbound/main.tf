# Trojan inbound with WebSocket transport.

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 3.0"
    }
  }
}

provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

resource "threexui_inbound" "trojan" {
  port     = 8443
  protocol = "trojan"
  remark   = "Trojan WS"
  enable   = true

  stream_settings {
    network = "ws"
    # security = "none" is for local/testing only.
    # In production, use "tls" or put the inbound behind a reverse proxy with TLS.
    security = "none"
    ws_settings {
      path = "/trojan-ws"
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

# Add a client to the Trojan inbound.
resource "threexui_inbound_client" "trojan_user" {
  inbound_id = threexui_inbound.trojan.id
  email      = "trojan-user@example.com"
  enable     = true
  password   = "my-trojan-password"
}
