terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint = "https://central.example.com:2053"
  username = "admin"
  password = "admin"
}

# Register a remote 3x-ui node in the central panel's cluster
resource "threexui_node" "de_fra_1" {
  name    = "de-fra-1"
  address = "node1.example.com"
  port    = 2053
  scheme  = "https"

  remark = "Frankfurt edge node"

  # API token for inter-node communication (sensitive)
  api_token = "node-api-token-secret"

  enable = true

  # Sync only selected inbounds to this node
  inbound_sync_mode = "selected"
  inbound_tags      = ["eu", "vless"]

  outbound_tag = "proxy-out"
}
