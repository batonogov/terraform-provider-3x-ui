terraform {
  required_providers {
    threexui = {
      source = "hashicorp/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

resource "threexui_inbound" "example" {
  port     = 23456
  protocol = "vless"
  remark   = "example-inbound"
  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

# resource "threexui_inbound_client" "example" {
#   inbound_id = threexui_inbound.example.id
#   email      = "client2@example.com"
#   enable     = true
#   flow       = "xtls-rprx-vision"
# }
