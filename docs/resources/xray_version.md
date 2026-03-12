---
page_title: "threexui_xray_version Resource - 3x-ui"
subcategory: "Xray"
description: |-
  Manages the installed Xray core version on the 3x-ui panel.
---

# threexui_xray_version (Resource)

Manages the installed Xray core version on the 3x-ui panel.

This is a singleton resource -- only one instance should exist per provider. Deleting this resource only removes it from Terraform state; the installed Xray version is not reverted.

## Example Usage

```hcl
data "threexui_xray_versions" "available" {}

resource "threexui_xray_version" "this" {
  version = data.threexui_xray_versions.available.versions[0]
}
```

## Argument Reference

- `version` (Required, String) - The desired Xray version to install (e.g. `"v25.1.1"`). Must include the `v` prefix. Available versions can be retrieved via the `threexui_xray_versions` data source.

## Attribute Reference

- `id` (String) - Always `"xray_version"`.
- `current_version` (String) - The currently installed Xray version (with `v` prefix).
