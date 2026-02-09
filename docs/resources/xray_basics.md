---
page_title: "threexui_xray_basics Resource - 3x-ui"
subcategory: "Xray Settings"
description: |-
  Manages basic Xray configuration in the 3x-ui panel.
---

# threexui_xray_basics (Resource)

Manages the basic Xray template configuration using a **merge root** strategy. The provided configuration is deep-merged into the existing Xray template. Only the keys you specify are updated; other keys are preserved.

This is a singleton resource. Deleting this resource only removes it from Terraform state; it does not reset the Xray configuration.

## Example Usage

```hcl
resource "threexui_xray_basics" "config" {
  log {
    loglevel = "warning"
    access   = "/var/log/xray/access.log"
    error    = "/var/log/xray/error.log"
    dns_log  = false
  }

  policy {
    system {
      stats_inbound_downlink  = true
      stats_inbound_uplink    = true
      stats_outbound_downlink = true
      stats_outbound_uplink   = true
    }

    level {
      id              = 0
      handshake       = 4
      conn_idle       = 300
      uplink_only     = 2
      downlink_only   = 5
      stats_user_uplink   = false
      stats_user_downlink = false
      buffer_size     = 4
    }
  }

  api {
    tag      = "api"
    services = ["HandlerService", "LoggerService", "StatsService"]
  }

  stats {}
}
```

## Argument Reference

### log (Block, Optional, Max: 1)

- `loglevel` (String, Optional) - Log level: `debug`, `info`, `warning`, `error`, `none`.
- `access` (String, Optional) - Path to access log file.
- `error` (String, Optional) - Path to error log file.
- `dns_log` (Bool, Optional) - Enable DNS query logging.

### policy (Block, Optional, Max: 1)

#### system (Block, Optional, Max: 1)

- `stats_inbound_downlink` (Bool, Optional) - Enable inbound downlink stats.
- `stats_inbound_uplink` (Bool, Optional) - Enable inbound uplink stats.
- `stats_outbound_downlink` (Bool, Optional) - Enable outbound downlink stats.
- `stats_outbound_uplink` (Bool, Optional) - Enable outbound uplink stats.

#### level (Block, Optional, List)

- `id` (Int, Required) - Policy level ID (used as map key in Xray JSON, e.g. `"0"`, `"1"`).
- `handshake` (Int, Optional) - Handshake timeout in seconds.
- `conn_idle` (Int, Optional) - Connection idle timeout in seconds.
- `uplink_only` (Int, Optional) - Uplink-only timeout in seconds.
- `downlink_only` (Int, Optional) - Downlink-only timeout in seconds.
- `stats_user_uplink` (Bool, Optional) - Enable per-user uplink stats.
- `stats_user_downlink` (Bool, Optional) - Enable per-user downlink stats.
- `buffer_size` (Int, Optional) - Buffer size per connection in KB.

### api (Block, Optional, Max: 1)

- `tag` (String, Optional) - API inbound tag.
- `services` (List of String, Optional) - API services to enable (e.g. `HandlerService`, `LoggerService`, `StatsService`).

### stats (Block, Optional, Max: 1)

Empty block. Presence enables Xray stats collection.

## Attribute Reference

- `id` - The resource identifier (`xray_basics`).

## Import

```shell
terraform import threexui_xray_basics.config xray_basics
```
