# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 2.x     | Yes       |
| < 2.0   | No        |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Email:** [fekinos@me.com](mailto:fekinos@me.com)

Please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact

**Do not** open a public GitHub issue for security vulnerabilities.

## Response

- You can expect an initial response within **48 hours**.
- We will work with you to understand and address the issue before any public disclosure.

## Sensitive Data Handled by the Provider

The 3x-ui panel issues and stores secrets that this provider reads and writes. The following fields are marked `Sensitive` in the schema and are never logged in plaintext:

| Surface | Sensitive fields |
| --- | --- |
| Provider config | `password`, `bootstrap_password`, `two_factor_code` |
| `threexui_inbound` (`stream_settings.reality_settings`) | `private_key`, auto-generated short IDs |
| `threexui_inbound` (`wireguard_settings`) | `secret_key`, peer `private_key` |
| `threexui_inbound_client` | client `id` (UUID), `password` (trojan/ss), `auth` (hysteria) |
| `threexui_panel_general` | LDAP `bind_password` |
| `threexui_panel_telegram` | `bot_token`, `chat_id` |
| `threexui_panel_user` | `old_password`, `new_password` |
| `threexui_xray_outbounds` | per-protocol credentials (e.g. `password`, `users[].password`) |
| Data sources | `threexui_inbounds.inbounds`, `threexui_settings.json`, `threexui_xray_config.json` (full payloads) |

## Protecting Terraform State

Terraform state stores **all** sensitive values in plaintext, regardless of the `Sensitive` flag in the schema. Always:

- **Use a remote backend** with encryption at rest (S3 with SSE, GCS with CMEK, Terraform Cloud, etc.).
- **Restrict state-file access** to the smallest set of operators who already need panel access.
- **Avoid committing `terraform.tfstate`** or `*.tfstate.backup` to git — they are in the default `.gitignore`, but worth verifying.
- **Rotate the panel password** if a state file is exposed; the provider stores the credentials it was given.

## Network and Authentication Notes

- The provider authenticates to 3x-ui over HTTP/HTTPS with form-based session cookies.
- Use HTTPS in production. The `insecure_skip_verify` flag exists for self-signed lab environments only.
- The provider performs automatic re-login on HTTP 401/404. 2FA support is partial — see the provider docs for the trade-offs.
- Auto-generated Reality keys, short IDs, and WireGuard keys are returned by 3x-ui on create. The provider preserves them across `Read` operations via `UseStateForUnknown` plan modifiers, so a `terraform plan` after `apply` will not surface them as drift.

## Responsible Disclosure Window

We aim to publish a fix and advisory within **30 days** of a confirmed report. If a vulnerability is being actively exploited, the timeline tightens — please flag this in your initial email.
