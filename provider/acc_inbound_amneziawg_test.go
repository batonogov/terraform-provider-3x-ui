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

// TestAccInboundAmneziawgPartialServerKeepsObfuscation guards the reason
// AmneziaWG exists at all.
//
// 3x-ui generates its randomised obfuscation set only when the settings it
// receives carry NO `server` object; a partial block is taken literally and
// every omitted field is stored as its zero value. An inbound configured with
// just a subnet therefore ends up with jc=0, blank h1-h4 and no header
// protection — plain WireGuard, trivially fingerprinted, with nothing in the
// panel or in Terraform reporting it. Measured directly against v3.7.0.
//
// The provider works around it by creating the inbound without the block and
// applying the configured fields afterwards (splitAmneziawgServer /
// applyAmneziawgServerOverrides). This test pins both halves: the configured
// values must win, and the generated ones must survive.
func TestAccInboundAmneziawgPartialServerKeepsObfuscation(t *testing.T) {
	requireMinVersion(t, "v3.7.0")

	const port = 26013
	config := testAccProviderConfig() + fmt.Sprintf(`
resource "threexui_inbound" "awg_partial" {
  port     = %d
  protocol = "amneziawg"
  remark   = "acc-awg-partial"
  enable   = true

  amneziawg_settings {
    server {
      subnet_ip   = "10.9.2.0"
      subnet_cidr = 24
      primary_dns = "1.1.1.1"
    }
  }
}
`, port)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Configured fields win.
					resource.TestCheckResourceAttr("threexui_inbound.awg_partial", "amneziawg_settings.server.subnet_ip", "10.9.2.0"),
					resource.TestCheckResourceAttr("threexui_inbound.awg_partial", "amneziawg_settings.server.primary_dns", "1.1.1.1"),
					// Generated fields survive.
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.jc", nonZeroAttr("jc")),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.jmin", nonZeroAttr("jmin")),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.jmax", nonZeroAttr("jmax")),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.s1", nonZeroAttr("s1")),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.s2", nonZeroAttr("s2")),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.h1", func(v string) error {
						if v == "" {
							return fmt.Errorf("h1 is blank: magic-header obfuscation was not generated")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("threexui_inbound.awg_partial", "amneziawg_settings.server.header_protection_key", func(v string) error {
						if v == "" {
							return fmt.Errorf("header_protection_key is blank: header protection was not generated")
						}
						return nil
					}),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// TestAccInboundAmneziawgRemoveLastPeer covers peer removal, which is only
// possible because preserveInboundSettings skips protocols whose clients the
// inbound owns.
//
// For vmess/vless/… the provider re-injects the existing clients[] on update,
// since those peers belong to threexui_inbound_client and must not be clobbered
// by an inbound-level write. AmneziaWG peers are owned by the inbound, so the
// same behaviour would put a deleted peer straight back: the apply fails with
// "block count changed from 0 to 1" and the peer keeps connecting — a removal
// that reports failure while leaving access intact.
func TestAccInboundAmneziawgRemoveLastPeer(t *testing.T) {
	requireMinVersion(t, "v3.7.0")

	const port = 26014
	base := `
resource "threexui_inbound" "awg_peers" {
  port     = %d
  protocol = "amneziawg"
  remark   = "acc-awg-peers"
  enable   = true

  amneziawg_settings {
    server {}
%s
  }
}
`
	peer := `
    clients {
      email       = "awg-last-peer@test.com"
      enable      = true
      public_key  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowed_ips = ["10.8.1.5/32"]
    }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(base, port, peer),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.awg_peers", "amneziawg_settings.clients.#", "1"),
					resource.TestCheckResourceAttr("threexui_inbound.awg_peers", "amneziawg_settings.clients.0.email", "awg-last-peer@test.com"),
				),
			},
			{
				// The only peer is removed. It must actually go away.
				Config: testAccProviderConfig() + fmt.Sprintf(base, port, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.awg_peers", "amneziawg_settings.clients.#", "0"),
				),
			},
			{
				Config:   testAccProviderConfig() + fmt.Sprintf(base, port, ""),
				PlanOnly: true,
			},
		},
	})
}

// TestAccInboundAmneziawgRecreateWithSamePeerEmail is the regression guard for
// #452: destroy followed by apply, which is what a `-replace`, a moved resource
// or a rebuilt workspace does.
//
// 3x-ui's DelInbound drops the inbound-to-client links but keeps the `clients`
// rows, which carry the unique index on `email`. Peers owned by the inbound
// (AmneziaWG, and WireGuard `clients[]` since v3.4.2) are therefore left
// occupying their address, and the next create fails with
// "Duplicate email: <address>" naming a client that no longer appears under any
// inbound. The provider now deletes each peer through the per-client endpoint
// before removing the inbound.
//
// The framework's own destroy runs between steps, so a second step with an
// identical configuration is enough: the first create+destroy has to leave the
// email free for the second create. Both steps deliberately use the SAME email.
func TestAccInboundAmneziawgRecreateWithSamePeerEmail(t *testing.T) {
	requireMinVersion(t, "v3.7.0")

	const port = 26015
	config := func(remark string) string {
		return testAccProviderConfig() + fmt.Sprintf(`
resource "threexui_inbound" "awg_recreate" {
  port     = %d
  protocol = "amneziawg"
  remark   = %q
  enable   = true

  amneziawg_settings {
    server {}
    clients {
      email       = "awg-recreated@test.com"
      enable      = true
      public_key  = "dGVzdHB1YmxpY2tleXRlc3RwdWJsaWNrZXkxMjM0NQ=="
      allowed_ips = ["10.8.1.7/32"]
    }
  }
}
`, port, remark)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config("acc-awg-recreate-first"),
				Check: resource.TestCheckResourceAttr("threexui_inbound.awg_recreate",
					"amneziawg_settings.clients.0.email", "awg-recreated@test.com"),
			},
			{
				// Forces destroy + create of the inbound with the same peer email.
				Taint:  []string{"threexui_inbound.awg_recreate"},
				Config: config("acc-awg-recreate-second"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.awg_recreate", "remark", "acc-awg-recreate-second"),
					resource.TestCheckResourceAttr("threexui_inbound.awg_recreate",
						"amneziawg_settings.clients.0.email", "awg-recreated@test.com"),
				),
			},
		},
	})
}
