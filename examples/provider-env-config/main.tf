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

# The supported attributes listed above are read from THREEXUI_* environment
# variables. bootstrap_username, bootstrap_password, and two_factor_code have
# no environment-variable fallback and must be configured explicitly when used.
provider "threexui" {}
