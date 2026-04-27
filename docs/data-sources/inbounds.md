---
page_title: "threexui_inbounds Data Source - 3x-ui"
subcategory: "Inbound"
description: |-
  Retrieves the list of all inbounds from the 3x-ui panel.
---

# threexui_inbounds (Data Source)

Retrieves the list of all inbounds configured in the 3x-ui panel as a JSON string.

## Example Usage

```hcl
data "threexui_inbounds" "all" {}

locals {
  inbounds = jsondecode(data.threexui_inbounds.all.inbounds)
}

# Ports themselves are not secrets, so wrap with `nonsensitive()` to expose
# them in plan output. Sensitivity propagates from the source attribute, so
# any expression derived from `inbounds` is sensitive by default.
output "inbound_ports" {
  value = nonsensitive([for i in local.inbounds : i.port])
}
```

> **Upgrade note:** The `inbounds` attribute is marked sensitive because each inbound object contains client UUIDs/passwords, Reality `privateKey`, WireGuard `secretKey`, and similar credentials. Outputs that reference it (or values derived from it) must declare `sensitive = true` or be wrapped in `nonsensitive(...)` for fields that are safe to expose.

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `id` (String) - ID derived from the first inbound.
- `inbounds` (String, Sensitive) - JSON-encoded array of all inbound objects. Use `jsondecode()` to work with the data. Each object contains fields like `id`, `port`, `protocol`, `remark`, `enable`, `settings`, `streamSettings`, `sniffing`, etc. Marked sensitive because the payload includes client UUIDs/passwords, Reality `privateKey`, WireGuard `secretKey`, and other credentials.
