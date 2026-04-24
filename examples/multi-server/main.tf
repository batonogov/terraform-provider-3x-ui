# Multi-server management with map(object) pattern
#
# Terraform does not support for_each on providers, so each server
# is managed via a reusable module that receives connection parameters
# as input variables.

variable "servers" {
  description = "Map of 3x-ui servers to manage. Key = server name."
  type = map(object({
    endpoint             = string
    base_path            = optional(string, "/")
    username             = string
    password             = string
    insecure_skip_verify = optional(bool, false)
  }))
  sensitive = true
}

module "server" {
  source   = "./modules/server"
  for_each = var.servers

  endpoint             = each.value.endpoint
  base_path            = each.value.base_path
  username             = each.value.username
  password             = each.value.password
  insecure_skip_verify = each.value.insecure_skip_verify
}
