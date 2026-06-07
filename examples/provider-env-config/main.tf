# Provider configuration using environment variables.
#
# Set via environment variables:
#   export THREEXUI_ENDPOINT="https://panel.example.com:2053"
#   export THREEXUI_USERNAME="admin"
#   export THREEXUI_PASSWORD="s3cret"
#   export THREEXUI_BASE_PATH="/"
#
# Then run:
#   terraform plan
#   terraform apply

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 3.0"
    }
  }
}

# All attributes are read from THREEXUI_* environment variables.
provider "threexui" {}
