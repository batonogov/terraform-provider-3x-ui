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
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "example-inbound"
}
