package provider

import (
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
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "50"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Asia/Tehran"),
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
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_general.test", "id", "settings"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "page_size", "25"),
					resource.TestCheckResourceAttr("threexui_panel_general.test", "time_location", "Local"),
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
