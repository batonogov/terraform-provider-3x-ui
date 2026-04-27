---
page_title: "3x-ui Provider"
subcategory: ""
description: |-
  Terraform provider for managing 3x-ui panel resources.
---

# 3x-ui Provider

The 3x-ui provider allows you to manage [3x-ui](https://github.com/MHSanaei/3x-ui) panel resources using Terraform. It supports managing inbounds, clients, panel settings, and Xray configuration.

## Example Usage

```hcl
provider "threexui" {
  endpoint = "http://localhost:2053"
  username = "admin"
  password = "admin"
}

resource "threexui_inbound" "vless" {
  port     = 443
  protocol = "vless"
  enable   = true
  remark   = "VLESS Reality"

  vless_settings {
    decryption = "none"
  }

  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "www.apple.com:443"
      server_names = ["www.apple.com"]
    }
  }

  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
  }
}
```

## Authentication

The provider authenticates to the 3x-ui panel using username and password. These can be provided directly in the provider configuration or via environment variables.

-> **Note:** The provider has **partial** 2FA support. You can supply a TOTP code via the `two_factor_code` attribute, and it will be sent with the initial login request. However, TOTP codes expire every 30 seconds. Because the provider performs automatic re-login when the session expires (on HTTP 401/404), subsequent logins will fail once the original code is no longer valid. For short-lived operations (a single `terraform apply`) this may work, but long-running or repeated runs will require a fresh code each time.

## Argument Reference

- `endpoint` (Required, String) - Base URL of the 3x-ui panel, e.g. `http://localhost:2053`.
- `base_path` (Optional, String) - Base path configured in 3x-ui (`webBasePath`). Default is `/`.
- `username` (Optional, String) - 3x-ui username. Default is `admin`.
- `password` (Optional, String, Sensitive) - 3x-ui password. Default is `admin`.
- `two_factor_code` (Optional, String, Sensitive) - TOTP code for 2FA login. Used for the initial authentication request. See the note above about re-login limitations.
- `insecure_skip_verify` (Optional, Boolean) - Skip TLS certificate verification (useful for self-signed certs). Default is `false`.
- `request_timeout` (Optional, String) - HTTP request timeout (e.g. `30s`, `1m`). Default is `30s`.
- `max_retries` (Optional, Number) - Maximum number of additional attempts on transient HTTP 5xx responses for **idempotent write endpoints only** (`UpdateInbound`, `UpdateInboundClient`, `UpdateSettings`, `UpdateXrayTemplate`, `SetXrayOutboundTestURL`). Each retry waits 500ms and emits a `Warn`-level log entry, so upstream flakiness is observable rather than silently absorbed. Set to `0` to disable retries entirely. Default is `1`.

-> **Why this exists:** 3x-ui's pre-v2.9.0 inbound update path is not transactional, and a panic in the panel's handler goroutine surfaces as a 5xx via `gin.Recovery`. v2.9.0 reworked the path to run inside a SQLite transaction (`buildRuntimeInboundForAPI(tx, ...)`) and is more robust, but older versions still in our compatibility matrix occasionally surface this. The retry is scoped to write endpoints whose service-level handlers replace the entire row by id (truly idempotent at the SQL level), so repeating the same request after a transient failure is safe.
