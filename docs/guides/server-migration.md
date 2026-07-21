---
page_title: "Migrating a 3x-ui panel to a new server"
subcategory: "Guides"
description: |-
  Move a 3x-ui panel to a new server by restoring its database, then verify the restored objects with Terraform.
---

# Migrating 3x-ui between servers

To preserve existing client links, generated keys, resource IDs, and panel-only
data, migrate the 3x-ui database and server-side assets. Terraform then verifies
that the restored panel still matches the managed configuration.

Changing only the provider endpoint while retaining state is not a migration
mechanism: the state contains remote IDs from the source, and refresh can fail
when those IDs do not exist on an empty destination.

## Prerequisites

- The source panel is reachable and changes can be paused during the backup.
- The destination can run a compatible 3x-ui version and database engine.
- You can back up and restore the panel database and referenced server files.
- Terraform reports a clean plan against the source.

## Step 1 — freeze changes and confirm state

Pause UI and Terraform changes, then confirm that the source matches the
configuration:

```bash
terraform plan
# Expected: "No changes. Your infrastructure matches the configuration."
```

Resolve unexplained drift before taking the migration backup.

## Step 2 — back up the source

Create all of the following before changing the destination:

- an engine-appropriate backup of the complete 3x-ui database;
- copies of TLS certificates and other files referenced by panel settings;
- a backup of the encrypted Terraform backend/state;
- the exact 3x-ui version and database configuration used by the source.

Test that the database backup can be restored. `terraform state pull` is useful
for backing up Terraform state, but it is not a substitute for the panel
database: it includes only Terraform-managed objects and omits write-only
secrets and server-side data.

## Step 3 — restore the destination

1. Keep the destination private while restoring it.
2. Restore the database with the procedure for its database engine.
3. Restore certificates and mounted files at the paths expected by the panel.
4. Start a compatible 3x-ui version and verify panel health.

The restored database retains inbound IDs, client UUIDs, subscription IDs,
generated keys, settings, and other panel records.

## Step 4 — verify with Terraform

Update the provider variables to use the destination:

```hcl
endpoint = "https://new-host.example.com:2053"
username = "admin"
password = "restored-panel-password"
```

Then refresh and review:

```bash
terraform plan
```

Because the destination was restored from the source database, the IDs in
Terraform state should resolve there. A clean plan is the migration gate. Paths
such as `web_cert_file` may require an intentional update if the destination
uses a different filesystem layout.

## Step 5 — validate clients and cut over

Before changing DNS, test representative client links, subscriptions, Reality
and WireGuard connections, panel login, and any external integrations. Lower
DNS TTL as appropriate, switch traffic to the destination, and monitor it.

Keep the source stopped but recoverable, together with the tested backup, until
the rollback window has elapsed. Retire it only after the destination has been
verified.

## Alternative — rebuild instead of preserving the database

If new IDs and regenerated secrets are acceptable, use a separate backend key
or a fresh Terraform workspace and apply the configuration as a new deployment.
Do not reuse the source state against an empty destination.

To preserve selected identities during a rebuild, explicitly configure them:
client UUIDs, subscription IDs, protocol passwords, Reality keys/short IDs, and
WireGuard keys. Supply write-only values again from a secret manager. Computed
or panel-generated values that are not in configuration are not guaranteed to
survive this path.

The [multi-server example](https://github.com/batonogov/terraform-provider-threexui/tree/main/examples/multi-server)
shows the static provider-alias pattern for managing source and destination as
separate deployments during a staged rebuild.

## Related

- [Backup-as-code](backup-as-code.md) — understand the separate roles of configuration, state, and database backups.
- [Bulk client onboarding](bulk-clients.md) — manage larger client sets declaratively.
