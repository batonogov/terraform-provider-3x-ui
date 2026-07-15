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

# Configure SMTP email notifications (3x-ui v3.4.0+)
resource "threexui_panel_email" "notifications" {
  smtp_host     = "smtp.gmail.com"
  smtp_port     = 587
  smtp_username = "alerts@example.com"
  smtp_password = "app-specific-password"
  from_address  = "alerts@example.com"
  notify_on     = ["traffic_limit", "expiry", "login"]
}
