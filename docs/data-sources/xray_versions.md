---
page_title: "threexui_xray_versions Data Source - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Retrieves the available Xray versions from the 3x-ui panel.
---

# threexui_xray_versions (Data Source)

Retrieves the list of available Xray versions that can be installed on the 3x-ui panel.

## Example Usage

```hcl
data "threexui_xray_versions" "available" {}

output "xray_versions" {
  value = data.threexui_xray_versions.available.versions
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `versions` (List of String) - Available Xray versions.
