package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// --- Panel General: page_size, remark_model, time_location, update, idempotency ---

func TestAccPanelGeneral(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  page_size     = 50
  remark_model  = "-ieo"
  time_location = "Asia/Tehran"
  expire_diff   = 1
  traffic_diff  = 1
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "50"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "remark_model", "-ieo"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Asia/Tehran"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "expire_diff", "1"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "traffic_diff", "1"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  page_size     = 25
  remark_model  = "-ieo"
  time_location = "Local"
  expire_diff   = 0
  traffic_diff  = 0
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "25"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Local"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "expire_diff", "0"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "traffic_diff", "0"),
				),
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  page_size     = 25
  remark_model  = "-ieo"
  time_location = "Local"
  expire_diff   = 0
  traffic_diff  = 0
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Panel General: LDAP fields ---

func TestAccPanelGeneralLDAP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "ldap" {
  ldap_enable   = true
  ldap_host     = "ldap.example.com"
  ldap_port     = 636
  ldap_use_tls  = true
  ldap_bind_dn  = "cn=admin,dc=example,dc=com"
  ldap_password = "ldappass"
  ldap_base_dn  = "dc=example,dc=com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_host", "ldap.example.com"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_port", "636"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_use_tls", "true"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_bind_dn", "cn=admin,dc=example,dc=com"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_base_dn", "dc=example,dc=com"),
				),
			},
			// Disable LDAP (restore defaults)
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "ldap" {
  ldap_enable = false
  ldap_host   = ""
  ldap_port   = 389
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_enable", "false"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_port", "389"),
				),
			},
		},
	})
}

// --- Telegram: enable + token/chat_id/run_time/lang, update ---

func TestAccPanelTelegram(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_enable = true
  tg_bot_token  = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
  tg_bot_chat_id = "987654321"
  tg_run_time    = "@daily"
  tg_lang        = "en"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_token", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_chat_id", "987654321"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_run_time", "@daily"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_lang", "en"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_enable = false
  tg_bot_token  = ""
  tg_bot_chat_id = ""
  tg_run_time    = "@daily"
  tg_lang        = "ru"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_enable", "false"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_lang", "ru"),
				),
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_enable = false
  tg_bot_token  = ""
  tg_bot_chat_id = ""
  tg_run_time    = "@daily"
  tg_lang        = "ru"
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Subscription: enable + port/path/title, update, idempotency ---

func TestAccPanelSubscription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_enable      = true
  sub_json_enable = true
  sub_port        = 2096
  sub_path        = "/sub/"
  sub_title       = "acc-test-sub"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_json_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_port", "2096"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_path", "/sub/"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_title", "acc-test-sub"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_enable      = true
  sub_json_enable = false
  sub_port        = 2097
  sub_path        = "/newsub/"
  sub_title       = "acc-test-sub-updated"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_json_enable", "false"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_port", "2097"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_path", "/newsub/"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_title", "acc-test-sub-updated"),
				),
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_enable      = true
  sub_json_enable = false
  sub_port        = 2097
  sub_path        = "/newsub/"
  sub_title       = "acc-test-sub-updated"
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Disable subscription (restore)
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_enable      = false
  sub_json_enable = false
  sub_port        = 2096
  sub_path        = "/sub/"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_enable", "false"),
				),
			},
		},
	})
}
