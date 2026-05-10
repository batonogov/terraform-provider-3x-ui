---
page_title: "threexui_xray_routing Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages Xray routing configuration in the 3x-ui panel.
---

# threexui_xray_routing (Resource)

Manages the routing section of the Xray template configuration. Uses a **set path** strategy -- the provided configuration completely replaces the `routing` key in the Xray template.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the routing configuration.

## Example Usage

```hcl
resource "threexui_xray_routing" "config" {
  domain_strategy = "AsIs"
  domain_matcher  = "hybrid"

  rule {
    type         = "field"
    ip           = ["geoip:private"]
    outbound_tag = "blocked"
  }

  rule {
    type         = "field"
    domain       = ["geosite:category-ads"]
    outbound_tag = "blocked"
  }
}
```

> **Note:** 3x-ui v2.9.4+ manages the internal `api` inbound to `api` outbound routing rule automatically. The provider omits that internal rule from Terraform state to avoid drift.

## Argument Reference

### Top-level attributes

- `domain_strategy` (String, Optional) - Domain resolution strategy (e.g. `AsIs`, `IPIfNonMatch`, `IPOnDemand`).
- `domain_matcher` (String, Optional) - Domain matcher type (e.g. `hybrid`, `linear`).

### rule (Block, Optional, List)

- `type` (String, Optional) - Rule type (e.g. `field`).
- `domain` (List of String, Optional) - Domain matching patterns.
- `ip` (List of String, Optional) - IP/CIDR matching patterns (e.g. `geoip:private`).
- `port` (String, Optional) - Port range (e.g. `"80"`, `"1000-2000"`).
- `source_port` (String, Optional) - Source port range.
- `network` (String, Optional) - Network type (`tcp`, `udp`, `tcp,udp`).
- `source` (List of String, Optional) - Source IP/CIDR patterns.
- `user` (List of String, Optional) - User email patterns.
- `inbound_tag` (List of String, Optional) - Inbound tags to match.
- `protocol` (List of String, Optional) - Protocols to match (e.g. `http`, `tls`, `bittorrent`).
- `attrs` (String, Optional) - Advanced attribute matching (JSON string).
- `outbound_tag` (String, Optional) - Target outbound tag.
- `balancer_tag` (String, Optional) - Target balancer tag.

## Attribute Reference

- `id` - The resource identifier (`xray_routing`).

## Import

```shell
terraform import threexui_xray_routing.config xray_routing
```
