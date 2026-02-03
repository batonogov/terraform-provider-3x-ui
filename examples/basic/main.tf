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
  enable   = true

  settings = jsonencode({
    clients = [{
      id    = "11111111-1111-1111-1111-111111111111"
      email = "client@example.com"
      flow  = "xtls-rprx-vision"
    }]
    decryption = "none"
  })

  stream_settings = jsonencode({})
  sniffing        = jsonencode({})
}

resource "threexui_inbound_client" "example" {
  inbound_id = threexui_inbound.example.id
  client_id  = "22222222-2222-2222-2222-222222222222"
  email      = "client2@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

data "threexui_inbounds" "all" {}

data "threexui_server_status" "status" {}
