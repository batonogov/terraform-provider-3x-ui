---
page_title: "threexui_xray_config Data Source - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Retrieves the current Xray configuration from the 3x-ui panel.
---

# threexui_xray_config (Data Source)

Retrieves the current Xray template configuration from the 3x-ui panel as a JSON string.

## Example Usage

```hcl
data "threexui_xray_config" "current" {}

output "xray_config" {
  value     = data.threexui_xray_config.current.json
  sensitive = true
}
```

> **Upgrade note:** The `json` attribute is marked sensitive because the Xray template includes outbound credentials (Shadowsocks/Trojan/SOCKS/HTTP passwords, VLESS/VMess UUIDs, WireGuard `secretKey`, Reality `privateKey`) and inbound client identifiers. Any `output` referencing it (or values derived from it via `jsondecode(...)`) must declare `sensitive = true`, or be wrapped in `nonsensitive(...)` for fields that are safe to expose. Without it, `terraform plan` fails with `Output refers to sensitive values`.

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `id` (String) - Length of the JSON payload in bytes.
- `json` (String, Sensitive) - The current Xray template configuration as a JSON string. Marked sensitive because the payload includes outbound credentials, client UUIDs, WireGuard `secretKey`, and Reality `privateKey`.
