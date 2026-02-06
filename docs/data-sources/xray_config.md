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
  value = data.threexui_xray_config.current.json
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `json` (String) - The current Xray template configuration as a JSON string.
