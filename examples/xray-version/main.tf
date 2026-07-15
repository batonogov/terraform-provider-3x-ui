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

# Pin a specific Xray core version.
# The provider calls InstallXray and polls until the panel reports
# the requested version. Delete is a no-op (does not revert the version).
resource "threexui_xray_version" "pinned" {
  version = "v25.6.6"
}
