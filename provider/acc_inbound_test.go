package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- Vmess: create + update remark ---

func TestAccInboundVmess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "vmess" {
  port     = 25001
  protocol = "vmess"
  remark   = "acc-vmess-1"
  enable   = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.vmess", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.vmess", "protocol", "vmess"),
					resource.TestCheckResourceAttr("threexui_inbound.vmess", "remark", "acc-vmess-1"),
					resource.TestCheckResourceAttr("threexui_inbound.vmess", "port", "25001"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "vmess" {
  port     = 25001
  protocol = "vmess"
  remark   = "acc-vmess-2"
  enable   = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.vmess", "remark", "acc-vmess-2"),
				),
			},
		},
	})
}

// --- Trojan ---

func TestAccInboundTrojan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "trojan" {
  port     = 25002
  protocol = "trojan"
  remark   = "acc-trojan"
  enable   = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.trojan", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.trojan", "protocol", "trojan"),
					resource.TestCheckResourceAttr("threexui_inbound.trojan", "remark", "acc-trojan"),
				),
			},
		},
	})
}

// --- Shadowsocks with method/password/network ---

func TestAccInboundShadowsocks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "ss" {
  port     = 25003
  protocol = "shadowsocks"
  remark   = "acc-ss"
  enable   = true
  shadowsocks_settings {
    method   = "aes-256-gcm"
    password = "testpassword123"
    network  = "tcp,udp"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.ss", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.ss", "protocol", "shadowsocks"),
				),
			},
		},
	})
}

// --- HTTP with auth/accounts/allow_transparent ---

func TestAccInboundHTTP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "http" {
  port     = 25004
  protocol = "http"
  remark   = "acc-http"
  enable   = true
  http_settings {
    auth              = "password"
    allow_transparent = false
    account {
      user = "testuser"
      pass = "testpass"
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.http", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.http", "protocol", "http"),
				),
			},
		},
	})
}

// --- WireGuard with peers ---

func TestAccInboundWireguard(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "wg" {
  port     = 25005
  protocol = "wireguard"
  remark   = "acc-wireguard"
  enable   = true
  wireguard_settings {
    mtu = 1420
    peer {
      public_key  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowed_ips = ["10.0.0.2/32"]
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.wg", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.wg", "protocol", "wireguard"),
				),
			},
		},
	})
}

// --- Dokodemo-door (tunnel) ---

func TestAccInboundDokodemo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "dokodemo" {
  port     = 25006
  protocol = "dokodemo-door"
  remark   = "acc-dokodemo"
  enable   = true
  dokodemo_settings {
    address          = "127.0.0.1"
    port             = 80
    network          = "tcp"
    follow_redirect  = false
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.dokodemo", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.dokodemo", "protocol", "dokodemo-door"),
				),
			},
		},
	})
}

// --- VLESS + Reality (auto keys) ---

func TestAccInboundReality(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "reality" {
  port     = 25007
  protocol = "vless"
  remark   = "acc-reality"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "google.com:443"
      server_names = ["google.com"]
    }
  }
}
`,
				// Reality auto-generates keys — UseStateForUnknown plan modifiers
				// suppress the diff since auto-generated fields are preserved.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.reality", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.reality", "protocol", "vless"),
				),
			},
		},
	})
}

// --- VLESS with fallbacks ---

func TestAccInboundFallbacks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "fallback" {
  port     = 25008
  protocol = "vless"
  remark   = "acc-fallback"
  enable   = true
  vless_settings {
    decryption = "none"
    fallback {
      name = "default"
      dest = "80"
      xver = 1
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.fallback", "id"),
				),
			},
		},
	})
}

// --- Settings: decryption preserved ---

func TestAccInboundSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "settings" {
  port     = 25009
  protocol = "vless"
  remark   = "acc-settings"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.settings", "id"),
				),
			},
			// Re-apply same config: testseed should be preserved
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "settings" {
  port     = 25009
  protocol = "vless"
  remark   = "acc-settings"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.settings", "id"),
				),
			},
		},
	})
}

// --- Stream settings + sniffing update ---

func TestAccInboundStreamSniffing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "stream" {
  port     = 25010
  protocol = "vless"
  remark   = "acc-stream"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "none"
  }
  sniffing {
    enabled       = true
    dest_override = ["http", "tls"]
    metadata_only = false
    route_only    = false
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.stream", "id"),
				),
			},
			// Update sniffing
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "stream" {
  port     = 25010
  protocol = "vless"
  remark   = "acc-stream"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "none"
  }
  sniffing {
    enabled       = true
    dest_override = ["http", "tls", "quic", "fakedns"]
    metadata_only = true
    route_only    = false
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.stream", "id"),
				),
			},
		},
	})
}

// --- Port conflict -> error ---

func TestAccInboundPortConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "conflict1" {
  port     = 25011
  protocol = "vless"
  remark   = "acc-conflict-1"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}

resource "threexui_inbound" "conflict2" {
  port     = 25011
  protocol = "vless"
  remark   = "acc-conflict-2"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  depends_on = [threexui_inbound.conflict1]
}
`,
				ExpectError: regexp.MustCompile(`(?i)(port|exist|already|error|fail)`),
			},
		},
	})
}

// --- Update port, enable, remark ---

func TestAccInboundUpdateFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundUpdateConfig("acc-update-1", 25012, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.update", "remark", "acc-update-1"),
					resource.TestCheckResourceAttr("threexui_inbound.update", "port", "25012"),
					resource.TestCheckResourceAttr("threexui_inbound.update", "enable", "true"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccInboundUpdateConfig("acc-update-2", 25013, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.update", "remark", "acc-update-2"),
					resource.TestCheckResourceAttr("threexui_inbound.update", "port", "25013"),
					resource.TestCheckResourceAttr("threexui_inbound.update", "enable", "false"),
				),
			},
		},
	})
}

// --- Idempotency: no changes on re-apply ---

func TestAccInboundIdempotency(t *testing.T) {
	config := testAccProviderConfig() + `
resource "threexui_inbound" "idem" {
  port     = 25014
  protocol = "vless"
  remark   = "acc-idempotent"
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
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.idem", "id"),
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Negative: duplicate client email ---

func TestAccInboundClientDuplicateEmail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "dup_host" {
  port     = 25021
  protocol = "vless"
  remark   = "acc-dup-email-host"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}

resource "threexui_inbound_client" "dup1" {
  inbound_id = threexui_inbound.dup_host.id
  email      = "duplicate@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "dup2" {
  inbound_id = threexui_inbound.dup_host.id
  email      = "duplicate@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
  depends_on = [threexui_inbound_client.dup1]
}
`,
				ExpectError: regexp.MustCompile(`(?i)(error|fail|duplicate|exist|already)`),
			},
		},
	})
}

// --- Stream settings: WebSocket transport ---

func TestAccInboundWebSocket(t *testing.T) {
	config := testAccProviderConfig() + `
resource "threexui_inbound" "ws" {
  port     = 25022
  protocol = "vless"
  remark   = "acc-ws"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "ws"
    security = "none"
    ws_settings {
      path = "/ws"
    }
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.ws", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.ws", "protocol", "vless"),
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Stream settings: gRPC transport ---

func TestAccInboundGRPC(t *testing.T) {
	config := testAccProviderConfig() + `
resource "threexui_inbound" "grpc" {
  port     = 25023
  protocol = "vless"
  remark   = "acc-grpc"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "grpc"
    security = "none"
    grpc_settings {
      service_name = "mygrpc"
    }
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.grpc", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.grpc", "protocol", "vless"),
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Mixed protocol + listen ---

func TestAccInboundMixed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "mixed" {
  port     = 25024
  protocol = "mixed"
  remark   = "acc-mixed"
  listen   = "127.0.0.1"
  enable   = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.mixed", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.mixed", "protocol", "mixed"),
					resource.TestCheckResourceAttr("threexui_inbound.mixed", "listen", "127.0.0.1"),
				),
			},
		},
	})
}

// --- Listen field ---

func TestAccInboundListen(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "listen" {
  port     = 25025
  protocol = "vless"
  remark   = "acc-listen"
  listen   = "0.0.0.0"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.listen", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.listen", "listen", "0.0.0.0"),
				),
			},
		},
	})
}

// --- Remark omitted (default empty string), then set, then removed ---

func TestAccInboundRemarkDefault(t *testing.T) {
	configNoRemark := testAccProviderConfig() + `
resource "threexui_inbound" "noremark" {
  port     = 25026
  protocol = "vless"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}
`
	configWithRemark := testAccProviderConfig() + `
resource "threexui_inbound" "noremark" {
  port     = 25026
  protocol = "vless"
  remark   = "now-has-remark"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create without remark — defaults to ""
			{
				Config: configNoRemark,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.noremark", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.noremark", "remark", ""),
				),
			},
			// Step 2: idempotency — no diff on re-apply
			{
				Config:             configNoRemark,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: set remark
			{
				Config: configWithRemark,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.noremark", "remark", "now-has-remark"),
				),
			},
			// Step 4: remove remark — back to default ""
			{
				Config: configNoRemark,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.noremark", "remark", ""),
				),
			},
			// Step 5: idempotency after removing remark
			{
				Config:             configNoRemark,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Config helpers ---

func testAccInboundUpdateConfig(remark string, port int, enable bool) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "update" {
  port     = %d
  protocol = "vless"
  remark   = %q
  enable   = %t
  vless_settings {
    decryption = "none"
  }
}
`, port, remark, enable)
}
