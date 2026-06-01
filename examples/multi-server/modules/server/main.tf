# Per-server module: configures the provider and manages resources
# for a single 3x-ui instance.

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 3.0"
    }
  }
}

variable "endpoint" {
  description = "Base URL of the 3x-ui panel."
  type        = string
}

variable "base_path" {
  description = "Base path configured in 3x-ui (webBasePath)."
  type        = string
  default     = "/"
}

variable "username" {
  description = "3x-ui username."
  type        = string
}

variable "password" {
  description = "3x-ui password."
  type        = string
  sensitive   = true
}

variable "insecure_skip_verify" {
  description = "Skip TLS certificate verification."
  type        = bool
  default     = false
}

provider "threexui" {
  endpoint             = var.endpoint
  base_path            = var.base_path
  username             = var.username
  password             = var.password
  insecure_skip_verify = var.insecure_skip_verify
}

# --- Resources for this server ---

resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

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
    tcp_settings {
      accept_proxy_protocol = false
      header_type           = "none"
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}

resource "threexui_inbound_client" "user1" {
  inbound_id = threexui_inbound.vless.id
  email      = "user1@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

output "inbound_id" {
  description = "ID of the created inbound."
  value       = threexui_inbound.vless.id
}
