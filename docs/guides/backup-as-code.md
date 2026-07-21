---
page_title: "Backup-as-code with the 3x-ui provider"
subcategory: "Guides"
description: |-
  Keep Terraform-managed 3x-ui configuration reviewable and pair it with a panel database backup for disaster recovery.
---

# Backup-as-code

A Terraform configuration is a useful recovery specification for the 3x-ui
objects it manages. It is not, by itself, a complete panel backup: unmanaged
objects, write-only values, panel database metadata, certificates, and other
server-side files are outside that configuration.

For disaster recovery, keep both of these:

- the Terraform configuration and its encrypted remote state;
- a tested backup of the 3x-ui database and any server-side files it references.

## What you get

- **Point-in-time configuration history** — every code change is a commit.
- **Review before merge** — pull requests expose accidental client deletions.
- **Drift detection** — `terraform plan` flags changes made through the UI.
- **Repeatable rebuilds** — explicitly declared IDs and secrets can be applied
  into a fresh Terraform state when preserving the old database is not required.

## Layout

```text
panel-backup/
  main.tf            # provider + locals
  inbounds.tf        # all inbounds
  clients.tf         # all clients
  panel.tf           # general/security/telegram/subscription settings
  xray.tf            # routing, DNS, outbounds
  terraform.tfvars   # endpoint + credentials (gitignored)
```

## Provider configuration

```hcl
terraform {
  required_providers {
    threexui = {
      source = "batonogov/threexui"
    }
  }
  backend "s3" {
    bucket         = "panel-state"
    key            = "production/terraform.tfstate"
    region         = "eu-central-1"
    encrypt        = true
    dynamodb_table = "tf-locks"
  }
}

provider "threexui" {
  endpoint = var.endpoint
  username = var.username
  password = var.password
}
```

## Importing what you already have

If the panel is already populated, import it once instead of recreating it:

```bash
# inbounds (look up IDs in the panel or via the API)
terraform import threexui_inbound.vless 5
terraform import threexui_inbound.trojan 6

# clients use composite IDs: <inbound_id>:<client_uuid>
terraform import 'threexui_inbound_client.alice' '5:d4f1a2b3-c4d5-6e7f-8a9b-0c1d2e3f4a5b'

# panel settings are singletons
terraform import threexui_panel_general.this settings
terraform import threexui_panel_telegram.this settings
```

After import, run `terraform plan` and copy non-default attributes into your
`.tf` files until the plan is clean. Import only covers the resources you name;
it does not discover every object in the panel automatically.

## Daily flow

1. Someone proposes a change as a PR (`feat: add three new clients for team-x`).
2. CI runs `terraform plan` against staging and posts the diff as a PR comment.
3. A reviewer approves; merging triggers `terraform apply`.
4. The panel and the configuration history stay in lock-step.

## Recovery path A — preserve the existing panel identity

Use a database restore when existing client links, generated keys, object IDs,
and panel-only data must remain unchanged.

1. Quiesce writes to the source panel.
2. Back up the complete 3x-ui database with the procedure appropriate for its
   database engine. Also copy referenced TLS certificates and other mounted
   files.
3. Restore those assets on the destination and start a compatible 3x-ui
   version.
4. Point the provider at the restored panel and run `terraform plan`.

Because the database contains the same resource IDs, Terraform can refresh the
existing state against the destination. Investigate any unexpected plan before
applying it.

## Recovery path B — rebuild into a fresh Terraform state

Use a fresh backend key or workspace when you deliberately want Terraform to
create new panel objects:

```bash
# Configure a new backend key, or create a new workspace first.
terraform workspace new replacement
terraform apply
```

Do not point the old state at an empty panel and expect refresh to convert every
missing remote ID into a create operation. Reads for missing objects can fail,
and the old state still represents the source panel.

For a reproducible rebuild, declare every identity that must stay stable, such
as client UUIDs, subscription IDs, Reality keys/short IDs, and WireGuard keys.
Keep write-only passwords and tokens in a secret manager and supply them again;
write-only values cannot be recovered from Terraform state. Any value left for
the new panel to generate may differ from the old deployment.

## State and secret handling

`terraform state pull` exports only the objects tracked by that Terraform
state. It does not include unmanaged panel objects, the panel database,
certificates, or write-only secrets, so it is not a full 3x-ui export.

Terraform state can still contain sensitive values in plaintext. Encrypt the
remote backend, restrict access, and handle any downloaded state file as a
secret. See [SECURITY.md](../../SECURITY.md).

## Related

- [Server migration guide](server-migration.md) — move a live panel while preserving identities.
- [Multi-server example](https://github.com/batonogov/terraform-provider-threexui/tree/main/examples/multi-server) — manage a fixed set of panels with provider aliases.
