---
page_title: "threexui_xray_balancers Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray balancers configuration in the 3x-ui panel.
---

# threexui_xray_balancers (Resource)

Manages the balancers section within routing of the Xray template configuration. Uses a **set path** strategy -- the provided configuration completely replaces the `routing.balancers` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the balancers configuration.

## Example Usage

```hcl
resource "threexui_xray_balancers" "config" {
  balancer {
    tag      = "balancer-1"
    selector = ["proxy-*"]

    strategy {
      type = "leastPing"
    }
  }

  balancer {
    tag      = "balancer-2"
    selector = ["direct-*"]
  }
}
```

## Argument Reference

### balancer (Block, Optional, List)

- `tag` (String, Required) - Balancer tag name.
- `selector` (List of String, Required) - Outbound tag selectors (supports wildcards).

#### strategy (Block, Optional, Max: 1)

- `type` (String, Optional) - Balancing strategy type (e.g. `random`, `leastPing`, `leastLoad`).

## Attribute Reference

- `id` - The resource identifier (`xray_balancers`).
