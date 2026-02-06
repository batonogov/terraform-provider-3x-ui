---
page_title: "threexui_inbounds Data Source - 3x-ui"
subcategory: "Inbound"
description: |-
  Retrieves the list of all inbounds from the 3x-ui panel.
---

# threexui_inbounds (Data Source)

Retrieves the list of all inbounds configured in the 3x-ui panel.

## Example Usage

```hcl
data "threexui_inbounds" "all" {}

output "inbound_ports" {
  value = [for i in data.threexui_inbounds.all.inbounds : i.port]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `inbounds` - A list of inbound objects. Each inbound has the following attributes:
  - `id` (Number) - Inbound ID.
  - `up` (Number) - Upload traffic in bytes.
  - `down` (Number) - Download traffic in bytes.
  - `total` (Number) - Total traffic limit in bytes.
  - `all_time` (Number) - All-time traffic in bytes.
  - `remark` (String) - Inbound label/name.
  - `enable` (Boolean) - Whether the inbound is enabled.
  - `expiry_time` (Number) - Expiry time as Unix timestamp in milliseconds.
  - `traffic_reset` (String) - Traffic reset period.
  - `last_traffic_reset_time` (Number) - Last traffic reset timestamp.
  - `listen` (String) - Listen address.
  - `port` (Number) - Port number.
  - `protocol` (String) - Protocol type.
  - `settings` (List) - Protocol-specific settings (see `threexui_inbound` resource for schema).
  - `stream_settings` (List) - Transport/security settings (see `threexui_inbound` resource for schema).
  - `sniffing` (List) - Sniffing settings.
  - `tag` (String) - Inbound tag.
