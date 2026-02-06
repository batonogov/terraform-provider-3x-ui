---
page_title: "threexui_xray_reverse Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray reverse proxy configuration in the 3x-ui panel.
---

# threexui_xray_reverse (Resource)

Manages the reverse proxy section of the Xray template configuration. Uses a **set path** strategy -- the provided JSON completely replaces the `reverse` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the reverse proxy configuration.

## Example Usage

```hcl
resource "threexui_xray_reverse" "config" {
  json = jsonencode({
    bridges = [
      {
        tag    = "bridge"
        domain = "test.example.com"
      }
    ]
  })
}
```

## Argument Reference

- `json` (Optional, String) - Reverse proxy configuration as a JSON string.

## Attribute Reference

- `id` - The resource identifier (`xray_reverse`).
- `json` - The current reverse proxy configuration from the Xray template.
