---
page_title: "threexui_node Resource - 3x-ui"
subcategory: "Cluster"
description: |-
  Manages a remote 3x-ui panel registered as a cluster node (3x-ui multi-node surface).
---

# threexui_node (Resource)

Manages a remote 3x-ui panel registered as a cluster node on the central panel. Wraps the 3x-ui multi-node API (`/panel/api/nodes`), available since 3x-ui v3.0.2.

> **Reachability constraint:** The central panel probes the node for reachability (`ensureReachable`) during create/update before persisting it. The node's web API **must be reachable from the central panel at apply time**, otherwise the create/update fails. There is no way to bypass this from the provider — it is inherent to the 3x-ui server flow.
>
> **Scope (M2):** This resource implements **Create + Read + Import**. **In-place Update and Delete are placeholders** that emit a warning and do not change the node on the panel; they are implemented in M3 (#318). To change a node today, change `name`/`address` (forces replacement) or recreate the resource. Write-only secrets (`api_token_wo`) are M4 (#319).

## Example Usage

```hcl
resource "threexui_node" "fra1" {
  name        = "de-fra-1"
  remark      = "Frankfurt edge"
  scheme      = "https"
  address     = "node1.example.com"
  port        = 2053
  base_path   = "/"
  api_token   = var.fra1_api_token
  enable      = true

  tls_verify_mode = "verify"
}
```

Import an existing node by its numeric id:

```bash
terraform import threexui_node.fra1 7
```

## Argument Reference

- `name` (Required, String) - Unique node name (upstream `uniqueIndex`). Changing this forces a new resource.
- `address` (Required, String) - Node host. Changing this forces a new resource.
- `port` (Required, Number) - Node web API port, 1-65535.
- `remark` (Optional, String) - Free-form note.
- `scheme` (Optional, String) - `http` or `https`. Defaults to `https`.
- `base_path` (Optional, String) - Node web API base path. Defaults to `/`.
- `api_token` (Optional, String, Sensitive) - Bearer API token the central panel uses to authenticate to the node. Required unless `tls_verify_mode` is `mtls`. The panel returns this raw without redaction.
- `enable` (Optional, Bool) - Whether the node is enabled. Defaults to `true`.
- `allow_private_address` (Optional, Bool) - Allow the node address to resolve to a private IP. Defaults to `false`.
- `tls_verify_mode` (Optional, String) - `verify`, `skip`, `pin`, or `mtls`. Defaults to `verify`.
- `pinned_cert_sha256` (Optional, String, Sensitive) - Pinned certificate fingerprint, required when `tls_verify_mode` is `pin`.
- `inbound_sync_mode` (Optional, String) - `all` or `selected`. Defaults to `all`.
- `inbound_tags` (Optional, List of String) - Inbound tags to sync when `inbound_sync_mode` is `selected`.
- `outbound_tag` (Optional, String) - Xray outbound/balancer tag bridging this node.

## Attribute Reference

In addition to the arguments above, the following observed-state attributes are read-only (`Computed`):

- `id` (String) - Numeric node id (as a string). Import key.
- `guid` (String) - Remote panel stable GUID.
- `status` (String) - `online`, `offline`, or `unknown`.
- `last_heartbeat` (Number) - Unix seconds of the last successful heartbeat (0 = never).
- `latency_ms` (Number) - Last heartbeat latency in ms.
- `xray_version` (String) - Xray version reported by the node.
- `panel_version` (String) - 3x-ui panel version reported by the node.
- `cpu_pct` / `mem_pct` (Number) - Node CPU / memory usage percent.
- `uptime_secs` / `net_up` / `net_down` (Number) - Uptime (s) and network byte counters.
- `last_error` (String) - Last heartbeat error message.
- `xray_state` / `xray_error` (String) - Xray core state / error from the node.
- `config_dirty` (Bool) / `config_dirty_at` (Number) - Whether the node config differs from the central panel.
- `inbound_count` / `client_count` / `online_count` / `active_count` / `disabled_count` / `depleted_count` (Number) - Client/inbound counts.
- `parent_guid` (String) / `transitive` (Bool) - Node-tree attribution (read-only transitive sub-nodes have `transitive = true`).
- `created_at` / `updated_at` (Number) - Unix millis.

## Import

Import is supported by numeric id:

```bash
terraform import threexui_node.NAME <id>
```
