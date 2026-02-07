package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- All fields: email, flow, limit_ip, total_gb, expiry_time, enable, tg_id, sub_id, comment, reset ---

func TestAccInboundClientAllFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "client_host" {
  port     = 25101
  protocol = "vless"
  remark   = "acc-client-host-allfields"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound_client" "allfields" {
  inbound_id  = threexui_inbound.client_host.id
  email       = "allfields@test.com"
  enable      = true
  flow        = "xtls-rprx-vision"
  limit_ip    = 3
  total_gb    = 10737418240
  expiry_time = 0
  tg_id       = 123456
  sub_id      = "test-sub-id"
  comment     = "test comment"
  reset       = 0
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.allfields", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "email", "allfields@test.com"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "enable", "true"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "flow", "xtls-rprx-vision"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "limit_ip", "3"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "total_gb", "10737418240"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "tg_id", "123456"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "sub_id", "test-sub-id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "comment", "test comment"),
					resource.TestCheckResourceAttr("threexui_inbound_client.allfields", "reset", "0"),
				),
			},
		},
	})
}

// --- Update: email, enable, limit_ip, comment ---

func TestAccInboundClientUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundClientUpdateConfig("client-upd@test.com", true, 2, "initial"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "email", "client-upd@test.com"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "enable", "true"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "limit_ip", "2"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "comment", "initial"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccInboundClientUpdateConfig("client-upd-new@test.com", false, 5, "updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "email", "client-upd-new@test.com"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "enable", "false"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "limit_ip", "5"),
					resource.TestCheckResourceAttr("threexui_inbound_client.upd", "comment", "updated"),
				),
			},
			{
				ResourceName:            "threexui_inbound_client.upd",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// --- Multiple clients in one inbound ---

func TestAccInboundClientMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "multi_host" {
  port     = 25103
  protocol = "vless"
  remark   = "acc-multi-client-host"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound_client" "multi1" {
  inbound_id = threexui_inbound.multi_host.id
  email      = "multi1@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "multi2" {
  inbound_id = threexui_inbound.multi_host.id
  email      = "multi2@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
  depends_on = [threexui_inbound_client.multi1]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.multi1", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.multi1", "email", "multi1@test.com"),
					resource.TestCheckResourceAttrSet("threexui_inbound_client.multi2", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.multi2", "email", "multi2@test.com"),
				),
			},
		},
	})
}

// --- Vmess client with security ---

func TestAccInboundClientVmess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "vmess_host" {
  port     = 25104
  protocol = "vmess"
  remark   = "acc-vmess-client-host"
  enable   = true
}

resource "threexui_inbound_client" "vmess_cl" {
  inbound_id = threexui_inbound.vmess_host.id
  email      = "vmess-cl@test.com"
  enable     = true
  security   = "auto"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.vmess_cl", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.vmess_cl", "email", "vmess-cl@test.com"),
					resource.TestCheckResourceAttr("threexui_inbound_client.vmess_cl", "security", "auto"),
				),
			},
		},
	})
}

// --- Trojan client with password ---

func TestAccInboundClientTrojan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "trojan_host" {
  port     = 25105
  protocol = "trojan"
  remark   = "acc-trojan-client-host"
  enable   = true
}

resource "threexui_inbound_client" "trojan_cl" {
  inbound_id = threexui_inbound.trojan_host.id
  email      = "trojan-cl@test.com"
  enable     = true
  password   = "mytrojanpass"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.trojan_cl", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.trojan_cl", "email", "trojan-cl@test.com"),
				),
			},
		},
	})
}

// --- Client removal: 2 clients -> 1 client ---

func TestAccInboundClientRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "removal_host" {
  port     = 25106
  protocol = "vless"
  remark   = "acc-removal-host"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound_client" "remove1" {
  inbound_id = threexui_inbound.removal_host.id
  email      = "remove1@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}

resource "threexui_inbound_client" "remove2" {
  inbound_id = threexui_inbound.removal_host.id
  email      = "remove2@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
  depends_on = [threexui_inbound_client.remove1]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.remove1", "id"),
					resource.TestCheckResourceAttrSet("threexui_inbound_client.remove2", "id"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "removal_host" {
  port     = 25106
  protocol = "vless"
  remark   = "acc-removal-host"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound_client" "remove1" {
  inbound_id = threexui_inbound.removal_host.id
  email      = "remove1@test.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.remove1", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.remove1", "email", "remove1@test.com"),
				),
			},
		},
	})
}

// --- Config helpers ---

func testAccInboundClientUpdateConfig(email string, enable bool, limitIP int, comment string) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "upd_host" {
  port     = 25102
  protocol = "vless"
  remark   = "acc-client-upd-host"
  enable   = true
  settings = jsonencode({
    decryption = "none"
  })
}

resource "threexui_inbound_client" "upd" {
  inbound_id = threexui_inbound.upd_host.id
  email      = %q
  enable     = %t
  flow       = "xtls-rprx-vision"
  limit_ip   = %d
  comment    = %q
}
`, email, enable, limitIP, comment)
}
