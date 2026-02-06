---
page_title: "threexui_xray_basics Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages basic Xray configuration in the 3x-ui panel.
---

# threexui_xray_basics (Resource)

Manages the basic Xray template configuration using a **merge root** strategy. The provided JSON is deep-merged into the existing Xray template. Only the keys you specify are updated; other keys are preserved.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the Xray configuration.

-> **Note:** This resource uses subset diff suppression. No diff is shown if your configuration is a subset of the current state.

## Example Usage

```hcl
resource "threexui_xray_basics" "config" {
  json = jsonencode({
    log = {
      loglevel = "warning"
    }
    policy = {
      system = {
        statsInboundDownlink  = true
        statsInboundUplink    = true
        statsOutboundDownlink = true
        statsOutboundUplink   = true
      }
    }
  })
}
```

## Argument Reference

- `json` (Optional, String) - Xray configuration as a JSON string. Must be a JSON object. Keys are deep-merged into the existing template.

## Attribute Reference

- `id` - The resource identifier (`xray_basics`).
- `json` - The current Xray template content for the managed sections (log, policy, routing, outbounds).
