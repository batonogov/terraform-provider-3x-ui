# Complete workflow: VLESS Reality inbound with multiple clients.
#
# This example shows how to:
#   1. Create an inbound with protocol-specific settings
#   2. Add multiple clients to the same inbound
#   3. Reference the inbound ID in client resources

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 2.0"
    }
  }
}

provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

# Step 1: Create an inbound.
resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  remark   = "VLESS Reality"
  enable   = true

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
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

# Step 2: Add clients. Each client references the inbound by ID.
resource "threexui_inbound_client" "alice" {
  inbound_id = threexui_inbound.vless.id
  email      = "alice@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "bob" {
  inbound_id = threexui_inbound.vless.id
  email      = "bob@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"

  # Limit total traffic to 50 GB.
  total_gb = 50
}

# Use the subscription URL from the client output.
output "alice_sub_id" {
  description = "Alice subscription ID (use in subscription URL)"
  value       = threexui_inbound_client.alice.sub_id
}
