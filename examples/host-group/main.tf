terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "https://panel.example.com:2053"
  username = "admin"
  password = "admin"
}

# Create a VLESS inbound first
resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"
}

# Host group routing traffic to multiple hosts through the inbound
resource "threexui_host_group" "cdn" {
  remark      = "CDN Rotation"
  hosts       = ["cdn1.example.com", "cdn2.example.com"]
  inbound_ids = [threexui_inbound.vless.id]
  sort_order  = 1
  tags        = ["production", "cdn"]
}
