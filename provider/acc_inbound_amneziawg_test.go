package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccInboundAmneziawgKeysSurviveUnrelatedUpdate is the regression guard for
// the sharpest failure mode of the AmneziaWG surface (#441).
//
// 3x-ui regenerates the entire server block — a fresh keypair included —
// whenever an inbound is saved with settings that carry no `server` object
// (normalizeAmneziaWGSettings, internal/web/service/inbound_amneziawg.go:171-200).
// That runs on UpdateInbound too, so anything that lets the provider send an
// empty blob on a later apply silently rotates the server keys and invalidates
// every peer config already distributed. Confirmed against a live v3.7.0 panel:
// posting `settings = {}` on update returns a different publicKey than create
// did.
//
// The schema-level defence is amneziawgServerRequiredValidator, which refuses a
// configuration without the block. This test covers the other half: that when
// the block IS declared but every attribute is left to the panel, the values it
// generated on create survive an unrelated edit. That works only because the
// server attributes are Optional+Computed with UseStateForUnknown, so the plan
// replays what Read recorded — a regression in any of those three would rotate
// the keys here.
func TestAccInboundAmneziawgKeysSurviveUnrelatedUpdate(t *testing.T) {
	requireMinVersion(t, "v3.7.0")

	const port = 26012
	config := func(remark string) string {
		return testAccProviderConfig() + fmt.Sprintf(`
resource "threexui_inbound" "awg_keys" {
  port     = %d
  protocol = "amneziawg"
  remark   = %q
  enable   = true

  # Everything is left to the panel: this is the shape that regenerates the
  # server block if the provider ever sends an empty settings blob.
  amneziawg_settings {
    server {}
  }
}
`, port, remark)
	}

	var publicKeyAfterCreate, privateKeyAfterCreate string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config("acc-awg-keys"),
				Check: resource.ComposeTestCheckFunc(
					// The panel must have filled the block in.
					resource.TestCheckResourceAttrSet("threexui_inbound.awg_keys", "amneziawg_settings.server.public_key"),
					resource.TestCheckResourceAttrSet("threexui_inbound.awg_keys", "amneziawg_settings.server.private_key"),
					resource.TestCheckResourceAttrSet("threexui_inbound.awg_keys", "amneziawg_settings.server.jc"),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_keys", "amneziawg_settings.server.public_key", func(v string) error {
						publicKeyAfterCreate = v
						return nil
					}),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_keys", "amneziawg_settings.server.private_key", func(v string) error {
						privateKeyAfterCreate = v
						return nil
					}),
				),
			},
			{
				// Only the remark changes. Nothing about the tunnel should move.
				Config: config("acc-awg-keys-renamed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.awg_keys", "remark", "acc-awg-keys-renamed"),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_keys", "amneziawg_settings.server.public_key", func(v string) error {
						if v != publicKeyAfterCreate {
							return fmt.Errorf("server public key rotated on an unrelated update: %q -> %q; "+
								"every peer configuration handed out before this apply is now invalid",
								publicKeyAfterCreate, v)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_keys", "amneziawg_settings.server.private_key", func(v string) error {
						if v != privateKeyAfterCreate {
							return fmt.Errorf("server private key rotated on an unrelated update")
						}
						return nil
					}),
				),
			},
			{
				// And the whole thing must be driftless afterwards.
				Config:   config("acc-awg-keys-renamed"),
				PlanOnly: true,
			},
		},
	})
}
