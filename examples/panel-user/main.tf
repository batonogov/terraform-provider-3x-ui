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

# Rotate the panel admin credentials.
# The provider uses the current credentials from provider config to authenticate,
# then updates the panel user to the new credentials.
resource "threexui_panel_user" "admin" {
  username = "admin"
  password = "new-secure-password-2024"
}
