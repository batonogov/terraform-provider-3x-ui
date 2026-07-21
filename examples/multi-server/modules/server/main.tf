# Reusable per-server resources. Provider credentials stay in the root module;
# callers pass one statically declared provider alias as threexui.target.

terraform {
  required_providers {
    threexui = {
      source                = "batonogov/threexui"
      version               = "~> 3.0"
      configuration_aliases = [threexui.target]
    }
  }
}

# --- Resources for this server ---

resource "threexui_inbound" "vless" {
  provider = threexui.target

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
      target       = "www.amazon.com:443"
      server_names = ["www.amazon.com"]
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
  provider = threexui.target

  inbound_id = threexui_inbound.vless.id
  email      = "user1@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

output "inbound_id" {
  description = "ID of the created inbound."
  value       = threexui_inbound.vless.id
}
