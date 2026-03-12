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

## Drift Detection

If the Xray version is changed outside Terraform (e.g. via the panel UI), the next `terraform plan` will detect the drift and propose an update to restore the desired version.

## Known Limitations

- **Stopped Xray**: When the Xray process is not running, the 3x-ui API reports the version as `"Unknown"`. In this case, `Read` will return an error because the actual installed version cannot be determined. Restart Xray via the panel before running Terraform.
- **Delete**: Removing this resource only clears Terraform state. The installed Xray version is not reverted.
