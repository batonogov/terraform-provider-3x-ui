package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Protocol round-trip acceptance matrix (#70)
//
// Each protocol entry defines HCL configs for create and update, plus
// protocol-specific checks. Every entry exercises the full lifecycle:
//   Create -> Read -> ImportStateVerify -> Update -> PlanOnly (idempotency)
//
// Client-bearing protocols (vless, vmess, trojan, shadowsocks, hysteria)
// additionally create an inbound_client resource and verify it survives
// the inbound update.
// ---------------------------------------------------------------------------

type protocolMatrixEntry struct {
	// protocol is the Terraform protocol value (e.g. "vless", "vmess").
	protocol string
	// tfName is the Terraform resource name suffix (e.g. "vless" -> threexui_inbound.vless).
	tfName string
	// port is a unique port for this protocol's test.
	port int
	// createHCL returns the HCL for the initial create step.
	createHCL func(port int) string
	// updateHCL returns the HCL for the update step (must change at least one
	// protocol-specific field beyond remark).
	updateHCL func(port int) string
	// createChecks returns TestCheckFuncs evaluated after create.
	createChecks func(tfAddr string) []resource.TestCheckFunc
	// updateChecks returns TestCheckFuncs evaluated after update.
	updateChecks func(tfAddr string) []resource.TestCheckFunc
	// hasClient indicates this protocol supports clients.
	hasClient bool
	// clientHCL returns an inbound_client resource referencing the inbound.
	// Only used when hasClient is true.
	clientHCL func(inboundTFAddr string) string
	// clientChecks returns TestCheckFuncs for the client resource.
	clientChecks func(clientTFAddr string) []resource.TestCheckFunc
	// importStateVerifyIgnore lists attributes to ignore on import verification.
	importStateVerifyIgnore []string
}

// protocolMatrix returns the full set of protocol test entries.
func protocolMatrix() []protocolMatrixEntry {
	return []protocolMatrixEntry{
		matrixVless(),
		matrixVmess(),
		matrixTrojan(),
		matrixShadowsocks(),
		matrixHTTP(),
		matrixSocks(),
		matrixWireguard(),
		matrixDokodemo(),
		matrixHysteria(),
	}
}

// ---------------------------------------------------------------------------
// TestAccInboundProtocolMatrix — table-driven round-trip test
// ---------------------------------------------------------------------------

func TestAccInboundProtocolMatrix(t *testing.T) {
	for _, entry := range protocolMatrix() {
		entry := entry // capture range variable
		t.Run(entry.protocol, func(t *testing.T) {
			t.Parallel()

			inboundAddr := fmt.Sprintf("threexui_inbound.%s", entry.tfName)
			createConfig := testAccProviderConfig() + entry.createHCL(entry.port)
			updateConfig := testAccProviderConfig() + entry.updateHCL(entry.port)

			// If protocol supports clients, append client resources.
			clientAddr := ""
			if entry.hasClient && entry.clientHCL != nil {
				clientAddr = fmt.Sprintf("threexui_inbound_client.%s_client", entry.tfName)
				clientHCL := entry.clientHCL(inboundAddr)
				createConfig += clientHCL
				updateConfig += clientHCL
			}

			steps := []resource.TestStep{
				// Step 1: Create
				{
					Config: createConfig,
					Check:  resource.ComposeTestCheckFunc(matrixCreateChecks(inboundAddr, entry, clientAddr)...),
				},
				// Step 2: Import
				{
					ResourceName:            inboundAddr,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: entry.importStateVerifyIgnore,
				},
				// Step 3: Update (protocol-specific change)
				{
					Config: updateConfig,
					Check:  resource.ComposeTestCheckFunc(matrixUpdateChecks(inboundAddr, entry, clientAddr)...),
				},
				// Step 4: Idempotency — no diff on re-apply
				{
					Config:             updateConfig,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				CheckDestroy:             testAccCheckInboundDestroyed,
				Steps:                    steps,
			})
		})
	}
}

// ---------------------------------------------------------------------------
// Check helpers
// ---------------------------------------------------------------------------

func matrixCreateChecks(inboundAddr string, e protocolMatrixEntry, clientAddr string) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(inboundAddr, "id"),
		resource.TestCheckResourceAttr(inboundAddr, "protocol", e.protocol),
		resource.TestCheckResourceAttr(inboundAddr, "port", fmt.Sprintf("%d", e.port)),
	}
	if e.createChecks != nil {
		checks = append(checks, e.createChecks(inboundAddr)...)
	}
	if clientAddr != "" && e.clientChecks != nil {
		checks = append(checks, e.clientChecks(clientAddr)...)
	}
	return checks
}

func matrixUpdateChecks(inboundAddr string, e protocolMatrixEntry, clientAddr string) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(inboundAddr, "id"),
	}
	if e.updateChecks != nil {
		checks = append(checks, e.updateChecks(inboundAddr)...)
	}
	// After update, client must still exist.
	if clientAddr != "" && e.clientChecks != nil {
		checks = append(checks, e.clientChecks(clientAddr)...)
	}
	return checks
}

// ---------------------------------------------------------------------------
// Per-protocol definitions
// ---------------------------------------------------------------------------

func matrixVless() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "vless",
		tfName:   "mx_vless",
		port:     26001,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_vless" {
  port     = %d
  protocol = "vless"
  remark   = "matrix-vless-create"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "none"
    tcp_settings {
      accept_proxy_protocol = false
      header_type           = "none"
    }
  }
  sniffing {
    enabled       = true
    dest_override = ["http", "tls"]
    metadata_only = false
    route_only    = false
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_vless" {
  port     = %d
  protocol = "vless"
  remark   = "matrix-vless-updated"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "ws"
    security = "none"
    ws_settings {
      path = "/vless-ws"
    }
  }
  sniffing {
    enabled       = true
    dest_override = ["http", "tls"]
    metadata_only = false
    route_only    = false
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-vless-create"),
				resource.TestCheckResourceAttr(addr, "stream_settings.network", "tcp"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-vless-updated"),
				resource.TestCheckResourceAttr(addr, "stream_settings.network", "ws"),
			}
		},
		hasClient: true,
		clientHCL: func(inboundAddr string) string {
			return fmt.Sprintf(`
resource "threexui_inbound_client" "mx_vless_client" {
  inbound_id = %s.id
  email      = "matrix-vless@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
`, inboundAddr)
		},
		clientChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(addr, "id"),
				resource.TestCheckResourceAttr(addr, "email", "matrix-vless@test.com"),
			}
		},
	}
}

func matrixVmess() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "vmess",
		tfName:   "mx_vmess",
		port:     26002,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_vmess" {
  port     = %d
  protocol = "vmess"
  remark   = "matrix-vmess-create"
  enable   = true
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_vmess" {
  port     = %d
  protocol = "vmess"
  remark   = "matrix-vmess-updated"
  enable   = false
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-vmess-create"),
				resource.TestCheckResourceAttr(addr, "enable", "true"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-vmess-updated"),
				resource.TestCheckResourceAttr(addr, "enable", "false"),
			}
		},
		hasClient: true,
		clientHCL: func(inboundAddr string) string {
			return fmt.Sprintf(`
resource "threexui_inbound_client" "mx_vmess_client" {
  inbound_id = %s.id
  email      = "matrix-vmess@test.com"
  enable     = true
  security   = "auto"
}
`, inboundAddr)
		},
		clientChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(addr, "id"),
				resource.TestCheckResourceAttr(addr, "email", "matrix-vmess@test.com"),
				resource.TestCheckResourceAttr(addr, "security", "auto"),
			}
		},
	}
}

func matrixTrojan() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "trojan",
		tfName:   "mx_trojan",
		port:     26003,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_trojan" {
  port     = %d
  protocol = "trojan"
  remark   = "matrix-trojan-create"
  enable   = true
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_trojan" {
  port     = %d
  protocol = "trojan"
  remark   = "matrix-trojan-updated"
  enable   = true
  sniffing {
    enabled       = true
    dest_override = ["http", "tls"]
    metadata_only = false
    route_only    = false
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-trojan-create"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-trojan-updated"),
				resource.TestCheckResourceAttr(addr, "sniffing.enabled", "true"),
			}
		},
		hasClient: true,
		clientHCL: func(inboundAddr string) string {
			return fmt.Sprintf(`
resource "threexui_inbound_client" "mx_trojan_client" {
  inbound_id = %s.id
  email      = "matrix-trojan@test.com"
  enable     = true
  password   = "trojan-matrix-pass"
}
`, inboundAddr)
		},
		clientChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(addr, "id"),
				resource.TestCheckResourceAttr(addr, "email", "matrix-trojan@test.com"),
			}
		},
	}
}

func matrixShadowsocks() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "shadowsocks",
		tfName:   "mx_ss",
		port:     26004,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_ss" {
  port     = %d
  protocol = "shadowsocks"
  remark   = "matrix-ss-create"
  enable   = true
  shadowsocks_settings {
    method   = "aes-256-gcm"
    password = "matrix-ss-pass"
    network  = "tcp,udp"
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_ss" {
  port     = %d
  protocol = "shadowsocks"
  remark   = "matrix-ss-updated"
  enable   = true
  shadowsocks_settings {
    method   = "chacha20-ietf-poly1305"
    password = "matrix-ss-pass"
    network  = "tcp,udp"
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-ss-create"),
				resource.TestCheckResourceAttr(addr, "shadowsocks_settings.method", "aes-256-gcm"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-ss-updated"),
				resource.TestCheckResourceAttr(addr, "shadowsocks_settings.method", "chacha20-ietf-poly1305"),
			}
		},
		hasClient: true,
		clientHCL: func(inboundAddr string) string {
			return fmt.Sprintf(`
resource "threexui_inbound_client" "mx_ss_client" {
  inbound_id = %s.id
  email      = "matrix-ss@test.com"
  enable     = true
  password   = "ss-client-pass"
}
`, inboundAddr)
		},
		clientChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(addr, "id"),
				resource.TestCheckResourceAttr(addr, "email", "matrix-ss@test.com"),
			}
		},
		importStateVerifyIgnore: []string{
			"shadowsocks_settings.password",
		},
	}
}

func matrixHTTP() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "http",
		tfName:   "mx_http",
		port:     26005,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_http" {
  port     = %d
  protocol = "http"
  remark   = "matrix-http-create"
  enable   = true
  http_settings {
    auth              = "password"
    allow_transparent = false
    account {
      user = "httpuser"
      pass = "httppass"
    }
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_http" {
  port     = %d
  protocol = "http"
  remark   = "matrix-http-updated"
  enable   = true
  http_settings {
    auth              = "password"
    allow_transparent = true
    account {
      user = "httpuser"
      pass = "httppass"
    }
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-http-create"),
				resource.TestCheckResourceAttr(addr, "http_settings.allow_transparent", "false"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-http-updated"),
				resource.TestCheckResourceAttr(addr, "http_settings.allow_transparent", "true"),
			}
		},
	}
}

func matrixSocks() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "socks",
		tfName:   "mx_socks",
		port:     26006,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_socks" {
  port     = %d
  protocol = "socks"
  remark   = "matrix-socks-create"
  enable   = true
  socks_settings {
    auth = "password"
    udp  = true
    ip   = "127.0.0.1"
    account {
      user = "socksuser"
      pass = "sockspass"
    }
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_socks" {
  port     = %d
  protocol = "socks"
  remark   = "matrix-socks-updated"
  enable   = true
  socks_settings {
    auth = "password"
    udp  = false
    ip   = "127.0.0.1"
    account {
      user = "socksuser"
      pass = "sockspass"
    }
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-socks-create"),
				resource.TestCheckResourceAttr(addr, "socks_settings.udp", "true"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-socks-updated"),
				resource.TestCheckResourceAttr(addr, "socks_settings.udp", "false"),
			}
		},
	}
}

func matrixWireguard() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "wireguard",
		tfName:   "mx_wg",
		port:     26007,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_wg" {
  port     = %d
  protocol = "wireguard"
  remark   = "matrix-wg-create"
  enable   = true
  wireguard_settings {
    mtu = [1420, 1280]
    peer {
      public_key  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowed_ips = ["10.0.0.2/32"]
    }
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_wg" {
  port     = %d
  protocol = "wireguard"
  remark   = "matrix-wg-updated"
  enable   = true
  wireguard_settings {
    mtu = [1400, 1280]
    peer {
      public_key  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowed_ips = ["10.0.0.2/32"]
    }
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-wg-create"),
				resource.TestCheckResourceAttr(addr, "wireguard_settings.mtu.0", "1420"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-wg-updated"),
				resource.TestCheckResourceAttr(addr, "wireguard_settings.mtu.0", "1400"),
			}
		},
		importStateVerifyIgnore: []string{
			"wireguard_settings.secret_key",
		},
	}
}

func matrixDokodemo() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "dokodemo-door",
		tfName:   "mx_dokodemo",
		port:     26008,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_dokodemo" {
  port     = %d
  protocol = "dokodemo-door"
  remark   = "matrix-dokodemo-create"
  enable   = true
  dokodemo_settings {
    address         = "127.0.0.1"
    port            = 80
    network         = "tcp"
    follow_redirect = false
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_dokodemo" {
  port     = %d
  protocol = "dokodemo-door"
  remark   = "matrix-dokodemo-updated"
  enable   = true
  dokodemo_settings {
    address         = "127.0.0.1"
    port            = 443
    network         = "tcp,udp"
    follow_redirect = false
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-dokodemo-create"),
				resource.TestCheckResourceAttr(addr, "dokodemo_settings.network", "tcp"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-dokodemo-updated"),
				resource.TestCheckResourceAttr(addr, "dokodemo_settings.network", "tcp,udp"),
				resource.TestCheckResourceAttr(addr, "dokodemo_settings.port", "443"),
			}
		},
	}
}

func matrixHysteria() protocolMatrixEntry {
	return protocolMatrixEntry{
		protocol: "hysteria",
		tfName:   "mx_hysteria",
		port:     26009,
		createHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_hysteria" {
  port     = %d
  protocol = "hysteria"
  remark   = "matrix-hysteria-create"
  enable   = true
  hysteria_settings {
    version = 2
  }
  stream_settings {
    network  = "hysteria"
    security = "tls"
  }
}
`, port)
		},
		updateHCL: func(port int) string {
			return fmt.Sprintf(`
resource "threexui_inbound" "mx_hysteria" {
  port     = %d
  protocol = "hysteria"
  remark   = "matrix-hysteria-updated"
  enable   = true
  hysteria_settings {
    version = 2
  }
  stream_settings {
    network  = "hysteria"
    security = "tls"
  }
}
`, port)
		},
		createChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-hysteria-create"),
				resource.TestCheckResourceAttr(addr, "hysteria_settings.version", "2"),
			}
		},
		updateChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(addr, "remark", "matrix-hysteria-updated"),
			}
		},
		hasClient: true,
		clientHCL: func(inboundAddr string) string {
			return fmt.Sprintf(`
resource "threexui_inbound_client" "mx_hysteria_client" {
  inbound_id = %s.id
  email      = "matrix-hysteria@test.com"
  auth       = "hysteria-secret"
  enable     = true
}
`, inboundAddr)
		},
		clientChecks: func(addr string) []resource.TestCheckFunc {
			return []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(addr, "id"),
				resource.TestCheckResourceAttr(addr, "email", "matrix-hysteria@test.com"),
				resource.TestCheckResourceAttr(addr, "auth", "hysteria-secret"),
			}
		},
	}
}
