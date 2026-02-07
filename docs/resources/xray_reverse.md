---
page_title: "threexui_xray_reverse Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray reverse proxy configuration in the 3x-ui panel.
---

# threexui_xray_reverse (Resource)

Manages the reverse proxy section of the Xray template configuration. Uses a **set path** strategy -- the provided configuration completely replaces the `reverse` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the reverse proxy configuration.

## Example Usage

```hcl
resource "threexui_xray_reverse" "config" {
  bridge {
    tag    = "bridge"
    domain = "test.example.com"
  }

  portal {
    tag    = "portal"
    domain = "test.example.com"
  }
}
```

## Argument Reference

### bridge (Block, Optional, List)

- `tag` (String, Required) - Bridge tag name.
- `domain` (String, Required) - Bridge domain.

### portal (Block, Optional, List)

- `tag` (String, Required) - Portal tag name.
- `domain` (String, Required) - Portal domain.

## Attribute Reference

- `id` - The resource identifier (`xray_reverse`).
