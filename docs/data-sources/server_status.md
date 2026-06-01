---
page_title: "threexui_server_status Data Source - 3x-ui"
subcategory: "Server"
description: |-
  Retrieves the server status from the 3x-ui panel.
---

# threexui_server_status (Data Source)

Retrieves the current server status from the 3x-ui panel as a JSON string. Includes system information such as CPU, memory, disk usage, uptime, and network statistics.

## Example Usage

```hcl
data "threexui_server_status" "current" {}

output "server_status" {
  value = data.threexui_server_status.current.json
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `id` (String) - Length of the JSON payload in bytes.
- `json` (String) - Server status as a JSON string.
