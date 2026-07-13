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

- `tag` (String, Optional) - Balancer tag name.
- `selector` (List of String, Required) - Outbound tag selectors (supports wildcards).
- `fallback_tag` (String, Optional) - Fallback balancer tag used when this balancer has no healthy outbound (xray-core balancer-to-balancer fallback). Empty means no fallback.

#### strategy (Block, Optional, Max: 1)

- `type` (String, Required) - Balancing strategy type (`random`, `roundRobin`, `leastPing`, `leastLoad`).

##### settings (Block, Optional, Max: 1)

Tuning parameters for the `leastPing` / `leastLoad` strategies.

- `expected` (Number, Optional) - Number of expected alive outbounds.
- `max_rtt` (String, Optional) - Max acceptable round-trip time (xray duration string, e.g. `"500ms"`).
- `tolerance` (Number, Optional) - Selection tolerance `0.0`–`1.0`.
- `baselines` (List of String, Optional) - Baseline outbound tags.

###### costs (Block, Optional, List)

- `regexp` (Boolean, Optional) - Treat `match` as a regular expression.
- `match` (String, Optional) - Tag match keyword or regexp.
- `value` (Number, Optional) - Cost value applied to matching outbounds.

## Attribute Reference

- `id` - The resource identifier (`xray_balancers`).

## Import

```shell
terraform import threexui_xray_balancers.config xray_balancers
```
