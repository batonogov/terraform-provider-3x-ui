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
  settings = jsonencode({
    method   = "aes-256-gcm"
    password = "testpassword123"
    network  = "tcp,udp"
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.ss", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.ss", "protocol", "shadowsocks"),
					resource.TestCheckResourceAttrSet("threexui_inbound.ss", "settings"),
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
  settings = jsonencode({
    auth             = "password"
    allowTransparent = false
    accounts = [{
      user = "testuser"
      pass = "testpass"
    }]
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.http", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.http", "protocol", "http"),
					resource.TestCheckResourceAttrSet("threexui_inbound.http", "settings"),
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
  settings = jsonencode({
    mtu = 1420
    peers = [{
      publicKey  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowedIPs = ["10.0.0.2/32"]
    }]
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.wg", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.wg", "protocol", "wireguard"),
					resource.TestCheckResourceAttrSet("threexui_inbound.wg", "settings"),
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
  settings = jsonencode({
    address        = "127.0.0.1"
    port           = 80
    network        = "tcp"
    followRedirect = false
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.dokodemo", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.dokodemo", "protocol", "dokodemo-door"),
					resource.TestCheckResourceAttrSet("threexui_inbound.dokodemo", "settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "reality"
    realitySettings = {
      target      = "google.com:443"
      serverNames = ["google.com"]
    }
  })
}
`,
				// Reality auto-generates keys — the jsonSubsetPlanModifier
				// suppresses the diff since config is a subset of state.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.reality", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.reality", "protocol", "vless"),
					resource.TestCheckResourceAttrSet("threexui_inbound.reality", "stream_settings"),
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
  settings = jsonencode({
    decryption = "none"
    fallbacks = [{
      name = "default"
      dest = "80"
      xver = 1
    }]
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.fallback", "id"),
					resource.TestCheckResourceAttrSet("threexui_inbound.fallback", "settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.settings", "id"),
					resource.TestCheckResourceAttrSet("threexui_inbound.settings", "settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.settings", "settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
  })
  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls"]
    metadataOnly = false
    routeOnly    = false
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.stream", "stream_settings"),
					resource.TestCheckResourceAttrSet("threexui_inbound.stream", "sniffing"),
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
  })
  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls", "quic", "fakedns"]
    metadataOnly = true
    routeOnly    = false
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.stream", "sniffing"),
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
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound" "conflict2" {
  port     = 25011
  protocol = "vless"
  remark   = "acc-conflict-2"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
    tcpSettings = {
      acceptProxyProtocol = false
      header = {
        type = "none"
      }
    }
  })
  sniffing = jsonencode({
    enabled      = true
    destOverride = ["http", "tls"]
    metadataOnly = false
    routeOnly    = false
  })
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
  settings = jsonencode({
    decryption = "none"
  })
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "ws"
    security = "none"
    wsSettings = {
      path = "/ws"
    }
  })
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
					resource.TestCheckResourceAttrSet("threexui_inbound.ws", "stream_settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "grpc"
    security = "none"
    grpcSettings = {
      serviceName = "mygrpc"
    }
  })
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
					resource.TestCheckResourceAttrSet("threexui_inbound.grpc", "stream_settings"),
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
  settings = jsonencode({
    decryption = "none"
  })
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

// --- Config helpers ---

func testAccInboundUpdateConfig(remark string, port int, enable bool) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "update" {
  port     = %d
  protocol = "vless"
  remark   = %q
  enable   = %t
  settings = jsonencode({
    decryption = "none"
  })
}
`, port, remark, enable)
}
