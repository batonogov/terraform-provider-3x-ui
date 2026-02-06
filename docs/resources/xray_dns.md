---
page_title: "threexui_xray_dns Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray DNS configuration in the 3x-ui panel.
---

# threexui_xray_dns (Resource)

Manages the DNS section of the Xray template configuration. Uses a **set path** strategy -- the provided JSON completely replaces the `dns` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the DNS configuration.

## Example Usage

```hcl
resource "threexui_xray_dns" "config" {
  json = jsonencode({
    servers = [
      "https+local://1.1.1.1/dns-query",
      {
        address = "localhost"
        domains = ["geosite:cn"]
      }
    ]
    queryStrategy = "UseIP"
  })
}
```

## Argument Reference

- `json` (Optional, String) - DNS configuration as a JSON string.

## Attribute Reference

- `id` - The resource identifier (`xray_dns`).
- `json` - The current DNS configuration from the Xray template.
