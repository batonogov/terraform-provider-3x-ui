---
page_title: "threexui_client_traffics Data Source - 3x-ui"
subcategory: "Clients"
description: |-
  Retrieves traffic statistics for a client by email.
---

# threexui_client_traffics (Data Source)

Retrieves traffic statistics for a specific client by email from the 3x-ui panel. The 3x-ui panel enforces email uniqueness per client in the `client_traffics` table, so email is the canonical lookup key for this API endpoint.

Useful for monitoring client traffic usage (e.g., alerting on quota thresholds, generating reports).

## Example Usage

```hcl
data "threexui_client_traffics" "example" {
  email = "my-client"
}

output "upload_bytes" {
  value = data.threexui_client_traffics.example.up
}

output "download_bytes" {
  value = data.threexui_client_traffics.example.down
}
```

## Argument Reference

- `email` (Required, String) - Client email to look up. Must match an existing client email in the panel.

## Attribute Reference

- `id` (String) - Internal traffic record ID.
- `inbound_id` (Number) - Associated inbound ID.
- `enable` (Boolean) - Whether the client is enabled.
- `up` (Number) - Upload bytes.
- `down` (Number) - Download bytes.
- `total` (Number) - Traffic limit in bytes (0 = unlimited).
- `expiry_time` (Number) - Expiration timestamp in milliseconds since epoch (0 = no expiry).
