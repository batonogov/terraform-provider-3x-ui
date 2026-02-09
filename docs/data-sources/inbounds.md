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

output "inbound_ports" {
  value = [for i in local.inbounds : i.port]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `id` (String) - ID derived from the first inbound.
- `inbounds` (String) - JSON-encoded array of all inbound objects. Use `jsondecode()` to work with the data. Each object contains fields like `id`, `port`, `protocol`, `remark`, `enable`, `settings`, `streamSettings`, `sniffing`, etc.
