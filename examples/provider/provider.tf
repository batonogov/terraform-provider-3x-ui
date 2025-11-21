terraform {
  required_providers {
    threexui = {
      source  = "registry.terraform.io/batonogov/3x-ui"
      version = ">= 0.1.0"
    }
  }
}

provider "3xui" {
  base_url        = var.threexui_base_url
  username        = var.threexui_username
  password        = var.threexui_password
  tls_skip_verify = true
}
