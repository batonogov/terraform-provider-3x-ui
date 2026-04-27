---
page_title: "Migrating a 3x-ui panel to a new server"
subcategory: "Guides"
description: |-
  Move an entire 3x-ui panel — inbounds, clients, settings, Xray config — to a new VPS without re-typing anything.
---

# Migrating 3x-ui between servers

Migrating a populated 3x-ui panel by hand is painful: every inbound, every client UUID, every Reality key, every routing rule has to be re-entered, and the QR codes you already shipped to users must keep working.

With Terraform, the migration is a two-line change.

## Prerequisites

- Source panel (the one being retired) is reachable.
- Destination panel (the new VPS) is reachable and has 3x-ui installed but **empty**.
- Your panel is already managed by this provider — see [Backup-as-code](backup-as-code.md) if it isn't yet.

## Step 1 — confirm clean state on the source

```bash
terraform plan
# expected: "No changes. Your infrastructure matches the configuration."
```

If `plan` shows drift, resolve it before migrating. You do not want to carry surprises forward.

## Step 2 — point the provider at the new host

In `terraform.tfvars`:

```hcl
endpoint = "https://new-host.example.com:2053"
username = "admin"
password = "newpanel-temporary"
```

Or, equivalently, swap the values via environment variables:

```bash
export TF_VAR_endpoint="https://new-host.example.com:2053"
export TF_VAR_password="newpanel-temporary"
```

## Step 3 — apply against the new host

```bash
terraform apply
```

Terraform sees the new panel as empty, the state as full, and creates every resource from scratch.

What is preserved verbatim:

- Inbound `port`, `remark`, `protocol`
- Client UUIDs and emails (links and QR codes still work)
- Reality `private_key` / `public_key` / `short_ids`
- WireGuard `secret_key` / `public_key`
- Routing rules, DNS hosts, outbound chains
- Telegram bot token, subscription path, panel base path

What you may need to redo manually:

- TLS certificates if `web_cert_file` / `web_key_file` reference filesystem paths on the old host. Either copy the files or switch to a different cert source on the new host.
- LDAP password if you stored it outside Terraform.

## Step 4 — flip DNS

Once `terraform apply` finishes successfully, point your DNS records at the new VPS. Sessions on the old host will keep working until you take it down.

## Step 5 — retire the old host

After a grace period:

```bash
ssh old-host "docker compose down -v"
```

The Terraform state still describes only one panel — the live one — so no further changes are needed.

## Variations

### Active/active staging ↔ production

Use the [multi-server example](https://github.com/batonogov/terraform-provider-threexui/tree/main/examples/multi-server) to manage both panels from a single config and copy state selectively with `terraform state mv`.

### Cutover with zero downtime

1. Apply against the new panel (Step 3).
2. Lower DNS TTL.
3. Flip DNS.
4. Drain old panel after TTL expires.

The provider has no opinions about the network layer — anything you can do with two endpoints, you can do with two `provider` blocks (via aliases) or two workspaces.

## Related

- [Backup-as-code](backup-as-code.md) — the pattern this guide assumes you already use.
- [Bulk client onboarding](bulk-clients.md) — once migrated, scale the panel without manual work.
