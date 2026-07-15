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

# Configure Xray Observatory to monitor outbound health.
# The panel stores this as part of the Xray template config.
resource "threexui_xray_observatory" "main" {
  observer {
    subject_selector = ["outbound"]
    probe_url        = "https://www.google.com/generate_204"
    probe_interval   = "30s"
  }
}
