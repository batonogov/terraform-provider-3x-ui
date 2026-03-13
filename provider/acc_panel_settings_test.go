package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// --- Panel General: page_size, remark_model, time_location, update, idempotency ---

func TestAccPanelGeneral(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 1
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 50
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Asia/Tehran"
  traffic_diff                   = 1
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://example.com/generate_204"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "50"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Asia/Tehran"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "xray_outbound_test_url", "https://example.com/generate_204"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 25
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Local"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://www.google.com/generate_204"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "25"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Local"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "xray_outbound_test_url", "https://www.google.com/generate_204"),
				),
			},
			// ImportState
			{
				ResourceName:            "threexui_panel_general.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "settings",
				ImportStateVerifyIgnore: []string{"ldap_password"},
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 25
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Local"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://www.google.com/generate_204"
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "ldap" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = "dc=example,dc=com"
  ldap_bind_dn                   = "cn=admin,dc=example,dc=com"
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = true
  ldap_flag_field                = ""
  ldap_host                      = "ldap.example.com"
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = "ldappass"
  ldap_port                      = 636
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = true
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 25
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Local"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://www.google.com/generate_204"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_host", "ldap.example.com"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_port", "636"),
				),
			},
			// Disable LDAP (restore defaults)
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "ldap" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = "dc=example,dc=com"
  ldap_bind_dn                   = "cn=admin,dc=example,dc=com"
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = "ldappass"
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = true
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 25
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Local"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://www.google.com/generate_204"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.ldap", "ldap_enable", "false"),
				),
			},
		},
	})
}

// --- Panel Security: two_factor_enable + two_factor_token, update, idempotency ---

func TestAccPanelSecurity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_security" "test" {
  two_factor_enable = false
  two_factor_token  = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_security.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_security.test", "two_factor_enable", "false"),
				),
			},
			// Update: set a token value (but keep 2FA disabled to not block provider)
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_security" "test" {
  two_factor_enable = false
  two_factor_token  = "test-token-value"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_security.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_security.test", "two_factor_token", "test-token-value"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_panel_security.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "settings",
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_security" "test" {
  two_factor_enable = false
  two_factor_token  = "test-token-value"
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Restore defaults
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_security" "test" {
  two_factor_enable = false
  two_factor_token  = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_security.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_security.test", "two_factor_enable", "false"),
				),
			},
		},
	})
}

// --- Telegram: enable + token/chat_id/run_time/lang, update ---

func TestAccPanelTelegram(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_backup       = false
  tg_bot_chat_id      = "987654321"
  tg_bot_enable       = true
  tg_bot_login_notify = true
  tg_bot_proxy        = ""
  tg_bot_api_server   = ""
  tg_bot_token        = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
  tg_cpu              = 80
  tg_lang             = "en"
  tg_run_time         = "@daily"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_chat_id", "987654321"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_backup       = false
  tg_bot_chat_id      = ""
  tg_bot_enable       = false
  tg_bot_login_notify = true
  tg_bot_proxy        = ""
  tg_bot_api_server   = ""
  tg_bot_token        = ""
  tg_cpu              = 80
  tg_lang             = "ru"
  tg_run_time         = "@daily"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_bot_enable", "false"),
					resource.TestCheckResourceAttr("threexui_panel_telegram.test", "tg_lang", "ru"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_panel_telegram.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "settings",
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_telegram" "test" {
  tg_bot_backup       = false
  tg_bot_chat_id      = ""
  tg_bot_enable       = false
  tg_bot_login_notify = true
  tg_bot_proxy        = ""
  tg_bot_api_server   = ""
  tg_bot_token        = ""
  tg_cpu              = 80
  tg_lang             = "ru"
  tg_run_time         = "@daily"
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Subscription: enable + port/path/title, update, idempotency ---

// TestAccPanelGeneralConcurrentSettings verifies that panel_general and
// panel_subscription can be applied in the same graph without lost updates.
// Terraform applies independent resources concurrently, so both paths compete
// for the settings API. The settingsMu mutex must serialize these operations.
func TestAccPanelGeneralConcurrentSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 1
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 50
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Asia/Tehran"
  traffic_diff                   = 1
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://example.com/generate_204"
}

resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = true
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = true
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/sub/"
  sub_port           = 2096
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "concurrent-test"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					// panel_general values preserved
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "50"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Asia/Tehran"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "xray_outbound_test_url", "https://example.com/generate_204"),
					// panel_subscription values preserved
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_title", "concurrent-test"),
				),
			},
			// Idempotency — both resources stable after concurrent apply
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 1
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 50
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Asia/Tehran"
  traffic_diff                   = 1
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://example.com/generate_204"
}

resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = true
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = true
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/sub/"
  sub_port           = 2096
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "concurrent-test"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccPanelGeneralConcurrentXray verifies that panel_general
// (xray_outbound_test_url) and xray_outbounds can be applied in the
// same graph without lost updates. Both compete for the xray template
// endpoint; xrayTemplateMu must serialize them.
func TestAccPanelGeneralConcurrentXray(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 50
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Asia/Tehran"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://example.com/generate_204"
}

resource "threexui_xray_outbounds" "test" {
  outbound {
    tag      = "direct"
    protocol = "freedom"

    freedom_settings {
      domain_strategy = "AsIs"
    }
  }

  outbound {
    tag      = "blocked"
    protocol = "blackhole"

    blackhole_settings {
      response_type = "none"
    }
  }

  outbound {
    tag      = "dns-out"
    protocol = "dns"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "xray_outbound_test_url", "https://example.com/generate_204"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.0.tag", "direct"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.1.tag", "blocked"),
				),
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_general" "test" {
  date_picker                    = "gregorian"
  expire_diff                    = 0
  external_traffic_inform_enable = false
  external_traffic_inform_uri    = ""
  ldap_auto_create               = false
  ldap_auto_delete               = false
  ldap_base_dn                   = ""
  ldap_bind_dn                   = ""
  ldap_default_expiry_days       = 0
  ldap_default_limit_ip          = 0
  ldap_default_total_gb          = 0
  ldap_enable                    = false
  ldap_flag_field                = ""
  ldap_host                      = ""
  ldap_inbound_tags              = ""
  ldap_invert_flag               = false
  ldap_password                  = ""
  ldap_port                      = 389
  ldap_sync_cron                 = "@every 1m"
  ldap_truthy_values             = "true,1,yes,on"
  ldap_use_tls                   = false
  ldap_user_attr                 = "mail"
  ldap_user_filter               = "(objectClass=person)"
  ldap_vless_field               = "vless_enabled"
  page_size                      = 50
  remark_model                   = "-ieo"
  session_max_age                = 360
  time_location                  = "Asia/Tehran"
  traffic_diff                   = 0
  web_base_path                  = "/"
  web_cert_file                  = ""
  web_domain                     = ""
  web_key_file                   = ""
  web_listen                     = ""
  web_port                       = 2053
  xray_outbound_test_url         = "https://example.com/generate_204"
}

resource "threexui_xray_outbounds" "test" {
  outbound {
    tag      = "direct"
    protocol = "freedom"

    freedom_settings {
      domain_strategy = "AsIs"
    }
  }

  outbound {
    tag      = "blocked"
    protocol = "blackhole"

    blackhole_settings {
      response_type = "none"
    }
  }

  outbound {
    tag      = "dns-out"
    protocol = "dns"
  }
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccPanelSubscription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = true
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = true
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/sub/"
  sub_port           = 2096
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "acc-test-sub"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_enable", "true"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_title", "acc-test-sub"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_port", "2096"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = true
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = false
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/newsub/"
  sub_port           = 2097
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "acc-test-sub-updated"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_title", "acc-test-sub-updated"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_port", "2097"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_path", "/newsub/"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_panel_subscription.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "settings",
			},
			// Idempotency
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = true
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = false
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/newsub/"
  sub_port           = 2097
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "acc-test-sub-updated"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Disable subscription (restore)
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_subscription" "test" {
  sub_announce       = ""
  sub_cert_file      = ""
  sub_domain         = ""
  sub_enable         = false
  sub_enable_routing = true
  sub_encrypt        = true
  sub_json_enable    = false
  sub_json_fragment  = ""
  sub_json_mux       = ""
  sub_json_noises    = ""
  sub_json_path      = "/json/"
  sub_json_rules     = ""
  sub_json_uri       = ""
  sub_key_file       = ""
  sub_listen         = ""
  sub_path           = "/sub/"
  sub_port           = 2096
  sub_profile_url    = ""
  sub_routing_rules  = ""
  sub_show_info      = true
  sub_support_url    = ""
  sub_title          = "acc-test-sub-updated"
  sub_updates        = 12
  sub_uri            = ""
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_subscription.test", "sub_enable", "false"),
				),
			},
		},
	})
}

// TestAccPanelGeneralBasePathChange verifies that changing web_base_path and
// xray_outbound_test_url in the same operation succeeds. Before the fix, the
// Xray update after restart would fail because client.basePath was stale.
//
// This test drives the client directly (not through terraform-plugin-testing)
// because the framework re-configures the provider between apply and
// post-apply plan, which would create a new client with the old base_path.
func TestAccPanelGeneralBasePathChange(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}

	client, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("client init: %v", err)
	}
	ctx := context.Background()

	// Read current settings so we can restore them later.
	original, err := client.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	originalTestURL, err := client.GetXrayOutboundTestURL(ctx)
	if err != nil {
		t.Fatalf("get test url: %v", err)
	}

	// Ensure we restore the original base path regardless of outcome.
	t.Cleanup(func() {
		// Panel is on /testbp/ — keep basePath as-is to reach it.
		client.SetBasePath("/testbp/")
		_ = client.UpdateSettings(ctx, original) // restores webBasePath to "/"
		_ = client.SendRestart(ctx)              // restart on /testbp/ (current)
		client.SetBasePath("/")                  // panel will restart on /
		_ = client.WaitForReady(ctx)
		_ = client.SetXrayOutboundTestURL(ctx, originalTestURL)
	})

	// --- Simulate what applyPanelGeneral does ---

	// 1. Update settings: change webBasePath.
	desired := map[string]any{"webBasePath": "/testbp/"}
	merged := mergeSettings(original, desired)
	if err := client.UpdateSettings(ctx, merged); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	// 2. Send restart on the OLD path, then update basePath, then wait.
	if err := client.SendRestart(ctx); err != nil {
		t.Fatalf("send restart: %v", err)
	}
	client.SetBasePath("/testbp/")
	if err := client.WaitForReady(ctx); err != nil {
		t.Fatalf("wait for ready: %v", err)
	}

	// 3. Set xray outbound test URL on the NEW path.
	newTestURL := "https://example.com/generate_204"
	if err := client.SetXrayOutboundTestURL(ctx, newTestURL); err != nil {
		t.Fatalf("set xray outbound test url after base path change: %v", err)
	}

	// Verify the test URL was applied.
	got, err := client.GetXrayOutboundTestURL(ctx)
	if err != nil {
		t.Fatalf("get test url: %v", err)
	}
	if got != newTestURL {
		t.Fatalf("xray outbound test url = %q, want %q", got, newTestURL)
	}
}
