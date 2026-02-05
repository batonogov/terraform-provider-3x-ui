resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
  protocol = "vless"

  # settings {
  #   decryption = "none"
  #   encryption = "none"
  # }

  # stream_settings {
  #   network  = "tcp"
  #   security = "reality"

  #   reality_settings {
  #     show   = false
  #     xver   = 0
  #     target = "google.com:443"
  #     server_names = [
  #       "google.com",
  #       "www.google.com"
  #     ]
  #     min_client_ver = ""
  #     max_client_ver = ""
  #     max_timediff   = 0

  #     settings {
  #       fingerprint    = "chrome"
  #       server_name    = ""
  #       spider_x       = "/"
  #       mldsa65_verify = ""
  #     }
  #   }

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

output "reality_settings" {
  value = {
    target       = try(threexui_inbound.example.stream_settings[0].reality_settings[0].target, null)
    server_names = try(threexui_inbound.example.stream_settings[0].reality_settings[0].server_names, null)
  }
}

resource "threexui_inbound_client" "client_a" {
  inbound_id = threexui_inbound.example.id
  email      = "client-a@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "client_b" {
  inbound_id = threexui_inbound.example.id
  email      = "client-b@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "client_c" {
  inbound_id = threexui_inbound.example.id
  email      = "client-c@example.com"
  enable     = false
  flow       = "xtls-rprx-vision"
}

output "inbound_clients" {
  sensitive = true
  value = {
    client_a = threexui_inbound_client.client_a
    client_b = threexui_inbound_client.client_b
    client_c = threexui_inbound_client.client_c
  }
}
