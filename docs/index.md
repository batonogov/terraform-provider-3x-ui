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

  stream_settings {
    network  = "tcp"
    security = "reality"

    reality_settings {
      target = "www.apple.com:443"
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

-> **Note:** The provider does not currently support two-factor authentication codes during login. Enabling 2FA on the panel will prevent the provider from authenticating.

## Argument Reference

- `endpoint` (Required, String) - Base URL of the 3x-ui panel, e.g. `http://localhost:2053`.
- `base_path` (Optional, String) - Base path configured in 3x-ui (`webBasePath`). Default is `/`.
- `username` (Optional, String) - 3x-ui username. Default is `admin`.
- `password` (Optional, String, Sensitive) - 3x-ui password. Default is `admin`.
- `two_factor_code` (Optional, String, Sensitive) - Optional 2FA code for login.
- `insecure_skip_verify` (Optional, Boolean) - Skip TLS certificate verification (useful for self-signed certs). Default is `false`.
- `request_timeout` (Optional, String) - HTTP request timeout (e.g. `30s`, `1m`). Default is `30s`.
