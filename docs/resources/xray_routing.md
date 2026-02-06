---
page_title: "threexui_xray_routing Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray routing configuration in the 3x-ui panel.
---

# threexui_xray_routing (Resource)

Manages the routing section of the Xray template configuration. Uses a **set path** strategy -- the provided JSON completely replaces the `routing` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the routing configuration.

## Example Usage

```hcl
resource "threexui_xray_routing" "config" {
  json = jsonencode({
    domainStrategy = "AsIs"
    rules = [
      {
        type        = "field"
        ip          = ["geoip:private"]
        outboundTag = "blocked"
      }
    ]
  })
}
```

## Argument Reference

- `json` (Optional, String) - Routing configuration as a JSON string.

## Attribute Reference

- `id` - The resource identifier (`xray_routing`).
- `json` - The current routing configuration from the Xray template.
