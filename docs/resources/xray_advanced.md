---
page_title: "threexui_xray_advanced Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages the complete Xray template in the 3x-ui panel.
---

# threexui_xray_advanced (Resource)

Manages the entire Xray template configuration using a **replace all** strategy. The provided JSON completely replaces the entire Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the Xray configuration.

~> **Warning:** This resource replaces the **entire** Xray template. Use with caution, as it will overwrite any changes made by other `threexui_xray_*` resources or manual edits.

## Example Usage

```hcl
resource "threexui_xray_advanced" "config" {
  json = jsonencode({
    log = {
      loglevel = "warning"
    }
    api = {
      tag      = "api"
      services = ["HandlerService", "LoggerService", "StatsService"]
    }
    policy = {
      system = {
        statsInboundDownlink  = true
        statsInboundUplink    = true
        statsOutboundDownlink = true
        statsOutboundUplink   = true
      }
    }
    routing = {
      domainStrategy = "AsIs"
      rules          = []
    }
    outbounds = [
      {
        tag      = "direct"
        protocol = "freedom"
        settings = {}
      }
    ]
  })
}
```

## Argument Reference

- `json` (Optional, String) - Complete Xray template as a JSON string. Must be a JSON object.

## Attribute Reference

- `id` - The resource identifier (`xray_advanced`).
- `json` - The current complete Xray template.
