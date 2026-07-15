package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccInboundMtproto verifies the mtproto_settings typed block round-trips
// through the 3x-ui panel: create with a fake_tls_domain, then update it and
// confirm no drift.
func TestAccInboundMtproto(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "mtproto" {
  port     = 26001
  protocol = "mtproto"
  remark   = "acc-mtproto-1"
  enable   = true
  mtproto_settings {
    fake_tls_domain = "www.cloudflare.com"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.mtproto", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "protocol", "mtproto"),
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "remark", "acc-mtproto-1"),
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "port", "26001"),
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "mtproto_settings.0.fake_tls_domain", "www.cloudflare.com"),
				),
			},
			{
				Config: testAccProviderConfig() + `
resource "threexui_inbound" "mtproto" {
  port     = 26001
  protocol = "mtproto"
  remark   = "acc-mtproto-2"
  enable   = true
  mtproto_settings {
    fake_tls_domain = "bing.com"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "remark", "acc-mtproto-2"),
					resource.TestCheckResourceAttr("threexui_inbound.mtproto", "mtproto_settings.0.fake_tls_domain", "bing.com"),
				),
			},
		},
	})
}
