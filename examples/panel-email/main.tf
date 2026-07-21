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

variable "smtp_password" {
  description = "SMTP password supplied by a secret store or TF_VAR_smtp_password."
  type        = string
  sensitive   = true
}

# Configure SMTP email notifications (3x-ui v3.4.0+)
resource "threexui_panel_email" "notifications" {
  smtp_enable              = true
  smtp_host                = "smtp.gmail.com"
  smtp_port                = 587
  smtp_username            = "alerts@example.com"
  smtp_password_wo         = var.smtp_password
  smtp_password_wo_version = 1
  smtp_to                  = "alerts@example.com"
  smtp_encryption_type     = "starttls"
  smtp_enabled_events      = "login,backup"
}
