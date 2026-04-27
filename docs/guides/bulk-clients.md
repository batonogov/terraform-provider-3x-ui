---
page_title: "Onboarding many 3x-ui clients at once"
subcategory: "Guides"
description: |-
  Patterns for adding tens or hundreds of clients in a single terraform apply — for_each, CSV-driven onboarding, and per-team modules.
---

# Onboarding many clients at once

Adding 100 clients through the panel UI is 100 forms × 30 seconds each. With Terraform, it is one `for_each` and one `apply`.

This guide covers three common patterns.

## Pattern 1 — inline map

Useful for small, hand-curated client lists.

```hcl
locals {
  clients = {
    "alice@example.com"   = { flow = "xtls-rprx-vision" }
    "bob@example.com"     = { flow = "xtls-rprx-vision" }
    "carol@example.com"   = { flow = "" }
  }
}

resource "threexui_inbound_client" "by_email" {
  for_each = local.clients

  inbound_id = threexui_inbound.vless.id
  email      = each.key
  enable     = true
  flow       = each.value.flow
}
```

Adding a person is one new line. Removing them is one deleted line. Review happens in the PR.

## Pattern 2 — CSV-driven

Useful when client lists come from another system (HR, ticketing, billing).

`clients.csv`:

```csv
email,team,flow,total_gb,expiry_days
alice@example.com,engineering,xtls-rprx-vision,500,30
bob@example.com,engineering,xtls-rprx-vision,500,30
carol@example.com,sales,,200,7
```

`main.tf`:

```hcl
locals {
  raw     = csvdecode(file("${path.module}/clients.csv"))
  clients = { for c in local.raw : c.email => c }
}

resource "threexui_inbound_client" "from_csv" {
  for_each = local.clients

  inbound_id  = threexui_inbound.vless.id
  email       = each.key
  enable      = true
  flow        = each.value.flow
  total_gb    = tonumber(each.value.total_gb) * 1024 * 1024 * 1024
  expiry_time = each.value.expiry_days != "" ? (
    timestamp_plus_days(timestamp(), tonumber(each.value.expiry_days))
  ) : 0
}
```

Edit the CSV, commit, `apply`. The diff in the PR shows exactly who is being added or removed.

## Pattern 3 — per-team modules

Useful when each team owns its own client list.

```hcl
module "engineering" {
  source     = "./modules/team"
  inbound_id = threexui_inbound.vless.id
  team_name  = "engineering"
  members    = ["alice@example.com", "bob@example.com"]
  quota_gb   = 500
}

module "sales" {
  source     = "./modules/team"
  inbound_id = threexui_inbound.vless.id
  team_name  = "sales"
  members    = ["carol@example.com"]
  quota_gb   = 200
}
```

Inside `./modules/team/main.tf`:

```hcl
variable "inbound_id"  { type = number }
variable "team_name"   { type = string }
variable "members"     { type = list(string) }
variable "quota_gb"    { type = number }

resource "threexui_inbound_client" "member" {
  for_each = toset(var.members)

  inbound_id = var.inbound_id
  email      = each.value
  enable     = true
  total_gb   = var.quota_gb * 1024 * 1024 * 1024
}
```

Each team's section is small, focused, and reviewable independently.

## Performance notes

- Client operations within a single inbound are serialized via an internal mutex (the panel itself races otherwise). 100 clients on one inbound take roughly 100 × the round-trip time.
- Spreading clients across multiple inbounds parallelizes safely.
- The provider retries transient HTTP 5xx on idempotent writes (default `max_retries = 1`). Bumping it to `2`–`3` helps on flaky networks.

## What not to do

- **Do not** omit `email` on a client — 3x-ui will crash with a SQL error when the next client is added. The provider marks `email` as `Required` for this reason.
- **Do not** reuse client UUIDs across inbounds. Let the provider auto-generate them.
- **Do not** put 1000+ clients in a single `for_each` if you can avoid it. Group them by inbound or team module.

## Related

- [Backup-as-code](backup-as-code.md) — the foundation that makes bulk changes safe.
- [Server migration](server-migration.md) — once you have many clients, you really do not want to re-enter them by hand.
