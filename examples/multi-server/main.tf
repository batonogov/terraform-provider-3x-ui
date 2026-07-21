# Multi-server management with static provider aliases.
#
# Terraform cannot create provider configurations with for_each or dynamically
# select a provider alias for a module instance. Declare one provider and one
# module call per panel, then pass that provider through the module providers map.

terraform {
  required_providers {
    threexui = {
      source  = "batonogov/threexui"
      version = "~> 3.0"
    }
  }
}

variable "servers" {
  description = "Connection settings for the statically declared 3x-ui panels."
  type = map(object({
    endpoint             = string
    base_path            = optional(string, "/")
    username             = string
    password             = string
    insecure_skip_verify = optional(bool, false)
  }))
  sensitive = true
}

provider "threexui" {
  alias = "finland"

  endpoint             = var.servers["finland"].endpoint
  base_path            = var.servers["finland"].base_path
  username             = var.servers["finland"].username
  password             = var.servers["finland"].password
  insecure_skip_verify = var.servers["finland"].insecure_skip_verify
}

provider "threexui" {
  alias = "netherlands"

  endpoint             = var.servers["netherlands"].endpoint
  base_path            = var.servers["netherlands"].base_path
  username             = var.servers["netherlands"].username
  password             = var.servers["netherlands"].password
  insecure_skip_verify = var.servers["netherlands"].insecure_skip_verify
}

provider "threexui" {
  alias = "germany"

  endpoint             = var.servers["germany"].endpoint
  base_path            = var.servers["germany"].base_path
  username             = var.servers["germany"].username
  password             = var.servers["germany"].password
  insecure_skip_verify = var.servers["germany"].insecure_skip_verify
}

module "finland" {
  source = "./modules/server"

  providers = {
    threexui.target = threexui.finland
  }
}

module "netherlands" {
  source = "./modules/server"

  providers = {
    threexui.target = threexui.netherlands
  }
}

module "germany" {
  source = "./modules/server"

  providers = {
    threexui.target = threexui.germany
  }
}
