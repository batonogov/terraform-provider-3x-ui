---
page_title: "threexui_xray_balancers Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray balancers configuration in the 3x-ui panel.
---

# threexui_xray_balancers (Resource)

Manages the balancers section within routing of the Xray template configuration. Uses a **set path** strategy -- the provided JSON completely replaces the `routing.balancers` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the balancers configuration.

## Example Usage

```hcl
resource "threexui_xray_balancers" "config" {
  json = jsonencode([
    {
      tag      = "balancer-1"
      selector = ["proxy-*"]
      strategy = {
        type = "leastPing"
      }
    }
  ])
}
```

## Argument Reference

- `json` (Optional, String) - Balancers configuration as a JSON string (typically a JSON array).

## Attribute Reference

- `id` - The resource identifier (`xray_balancers`).
- `json` - The current balancers configuration from the Xray template.
