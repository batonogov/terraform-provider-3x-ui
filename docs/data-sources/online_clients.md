---
page_title: "threexui_online_clients Data Source - 3x-ui"
subcategory: "Clients"
description: |-
  Retrieves the list of currently online clients from the 3x-ui panel.
---

# threexui_online_clients (Data Source)

Retrieves the list of currently online client emails from the 3x-ui panel. Useful for monitoring active connections, dashboards, and automation based on client presence.

## Example Usage

```hcl
data "threexui_online_clients" "current" {}

output "online_clients" {
  value = data.threexui_online_clients.current.clients
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `clients` (List of String) - List of online client emails.
