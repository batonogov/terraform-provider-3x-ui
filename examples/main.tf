resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
  protocol = "vless"

  settings {
    decryption = "none"
    encryption = "none"
  }

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
  value = {
    client_a = {
      id          = threexui_inbound_client.client_a.id
      client_id   = threexui_inbound_client.client_a.client_id
      email       = threexui_inbound_client.client_a.email
      enable      = threexui_inbound_client.client_a.enable
      flow        = threexui_inbound_client.client_a.flow
      limit_ip    = threexui_inbound_client.client_a.limit_ip
      total_gb    = threexui_inbound_client.client_a.total_gb
      expiry_time = threexui_inbound_client.client_a.expiry_time
      tg_id       = threexui_inbound_client.client_a.tg_id
      sub_id      = threexui_inbound_client.client_a.sub_id
      comment     = threexui_inbound_client.client_a.comment
      reset       = threexui_inbound_client.client_a.reset
      security    = threexui_inbound_client.client_a.security
    }
    client_b = {
      id          = threexui_inbound_client.client_b.id
      client_id   = threexui_inbound_client.client_b.client_id
      email       = threexui_inbound_client.client_b.email
      enable      = threexui_inbound_client.client_b.enable
      flow        = threexui_inbound_client.client_b.flow
      limit_ip    = threexui_inbound_client.client_b.limit_ip
      total_gb    = threexui_inbound_client.client_b.total_gb
      expiry_time = threexui_inbound_client.client_b.expiry_time
      tg_id       = threexui_inbound_client.client_b.tg_id
      sub_id      = threexui_inbound_client.client_b.sub_id
      comment     = threexui_inbound_client.client_b.comment
      reset       = threexui_inbound_client.client_b.reset
      security    = threexui_inbound_client.client_b.security
    }
    client_c = {
      id          = threexui_inbound_client.client_c.id
      client_id   = threexui_inbound_client.client_c.client_id
      email       = threexui_inbound_client.client_c.email
      enable      = threexui_inbound_client.client_c.enable
      flow        = threexui_inbound_client.client_c.flow
      limit_ip    = threexui_inbound_client.client_c.limit_ip
      total_gb    = threexui_inbound_client.client_c.total_gb
      expiry_time = threexui_inbound_client.client_c.expiry_time
      tg_id       = threexui_inbound_client.client_c.tg_id
      sub_id      = threexui_inbound_client.client_c.sub_id
      comment     = threexui_inbound_client.client_c.comment
      reset       = threexui_inbound_client.client_c.reset
      security    = threexui_inbound_client.client_c.security
    }
  }
}
