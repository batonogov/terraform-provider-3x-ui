# Importing existing 3x-ui resources into Terraform state.
#
# If you already have inbounds and clients configured in the 3x-ui panel,
# you can import them into Terraform instead of recreating them.
#
# Prerequisites:
#   - Find inbound IDs in the 3x-ui panel (Settings column or API)
#   - Find client UUIDs in the inbound client list
#
# --------------------------------------------------------------------------
# 1. Inbound import
# --------------------------------------------------------------------------
#
# Import by inbound numeric ID (visible in the panel or via API):
#
#   terraform import threexui_inbound.existing 5
#
# After import, run `terraform plan` to see the current state.
# Copy the attributes into your configuration to avoid drift.
#
# --------------------------------------------------------------------------
# 2. Client import
# --------------------------------------------------------------------------
#
# Import format: <inbound_id>:<client_uuid>
#
#   terraform import threexui_inbound_client.existing_client 5:d4f1a2b3-c4d5-6e7f-8a9b-0c1d2e3f4a5b
#
# --------------------------------------------------------------------------
# 3. Singleton resources (panel settings)
# --------------------------------------------------------------------------
#
# Panel settings are singletons. Import with any ID (the provider ignores it):
#
#   terraform import threexui_panel_general.settings settings
#   terraform import threexui_panel_security.settings settings
#   terraform import threexui_panel_telegram.settings settings
#   terraform import threexui_panel_subscription.settings settings
#
# --------------------------------------------------------------------------
# 4. Xray settings
# --------------------------------------------------------------------------
#
# Xray settings are also singletons:
#
#   terraform import threexui_xray_basics.config settings
#   terraform import threexui_xray_dns.config settings
#   terraform import threexui_xray_routing.config settings
#   terraform import threexui_xray_outbounds.config settings
#
# --------------------------------------------------------------------------

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

# After importing, fill in the actual values from `terraform state show`.
resource "threexui_inbound" "existing" {
  port     = 443
  protocol = "vless"
  remark   = "Imported VLESS"
  enable   = true

  vless_settings {
    decryption = "none"
  }

  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "www.apple.com:443"
      server_names = ["www.apple.com"]
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

resource "threexui_inbound_client" "existing_client" {
  inbound_id = threexui_inbound.existing.id
  email      = "imported-user@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

# Panel settings: import with `terraform import threexui_panel_general.settings settings`
resource "threexui_panel_general" "settings" {
  web_port = 2053
}
