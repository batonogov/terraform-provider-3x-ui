# Multi-server example

Terraform requires provider configurations to be declared statically. It does
not support `for_each` on provider blocks, and a module cannot dynamically pick
an aliased provider from a map.

This example therefore declares one `threexui` provider alias and one module
call per panel. Each module call passes its provider explicitly:

```hcl
module "finland" {
  source = "./modules/server"

  providers = {
    threexui.target = threexui.finland
  }
}
```

The child module declares `threexui.target` in `configuration_aliases` and does
not contain credentials or a provider configuration of its own. To add another
panel, add its connection values, a root provider alias, and a static module
call.

Copy `terraform.tfvars.example` to a gitignored variable file or supply the
sensitive `servers` map from a secret store. Never commit real panel passwords.

See HashiCorp's [Providers Within Modules](https://developer.hashicorp.com/terraform/language/modules/develop/providers)
documentation for the provider inheritance and alias rules behind this pattern.
