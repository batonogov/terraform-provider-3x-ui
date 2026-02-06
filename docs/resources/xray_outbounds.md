---
page_title: "threexui_xray_outbounds Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray outbounds configuration in the 3x-ui panel.
---

# threexui_xray_outbounds (Resource)

Manages the outbounds section of the Xray template configuration. Uses a **set path** strategy -- the provided JSON completely replaces the `outbounds` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the outbounds configuration.

## Example Usage

```hcl
resource "threexui_xray_outbounds" "config" {
  json = jsonencode([
    {
      tag      = "direct"
      protocol = "freedom"
      settings = {}
    },
    {
      tag      = "blocked"
      protocol = "blackhole"
      settings = {}
    }
  ])
}
```

## Argument Reference

- `json` (Optional, String) - Outbounds configuration as a JSON string (typically a JSON array).

## Attribute Reference

- `id` - The resource identifier (`xray_outbounds`).
- `json` - The current outbounds configuration from the Xray template.
