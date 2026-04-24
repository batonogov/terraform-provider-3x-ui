# Provider configuration using Terraform variables.
#
# Set via environment variables:
#   export TF_VAR_threexui_endpoint="https://panel.example.com:2053"
#   export TF_VAR_threexui_username="admin"
#   export TF_VAR_threexui_password="s3cret"
#   export TF_VAR_threexui_base_path="/"
#
# Then run:
#   terraform plan
#   terraform apply

variable "threexui_endpoint" {
  description = "Base URL of the 3x-ui panel"
  type        = string
}

variable "threexui_username" {
  description = "3x-ui admin username"
  type        = string
  default     = "admin"
}

variable "threexui_password" {
  description = "3x-ui admin password"
  type        = string
  sensitive   = true
  default     = "admin"
}

variable "threexui_base_path" {
  description = "Base path configured in 3x-ui (webBasePath)"
  type        = string
  default     = "/"
}

terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
}

provider "threexui" {
  endpoint  = var.threexui_endpoint
  username  = var.threexui_username
  password  = var.threexui_password
  base_path = var.threexui_base_path

  # Skip TLS verification for self-signed certificates
  insecure_skip_verify = true
}
