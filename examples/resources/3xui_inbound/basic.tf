terraform {
  required_providers {
    threexui = {
      source  = "registry.terraform.io/batonogov/3x-ui"
      version = ">= 0.1.0"
    }
  }
}

provider "3xui" {
  base_url        = var.threexui_base_url
  username        = var.threexui_username
  password        = var.threexui_password
  tls_skip_verify = var.threexui_tls_skip_verify
}

resource "3xui_inbound" "basic" {
  protocol = "vless"
  listen   = "0.0.0.0"
  port     = 443
  remark   = "basic inbound"
  enable   = true

  stream {
    network  = "tcp"
    security = "tls"
    tls {
      sni = "example.com"
    }
  }

  settings_vless {
    clients = [{
      id    = "00000000-0000-4000-8000-000000000000"
      email = "user@example.com"
    }]
  }
}
