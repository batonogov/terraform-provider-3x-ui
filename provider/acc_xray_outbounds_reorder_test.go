package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccXrayOutboundsReorderNoFieldBleed covers issue #419 against a live
// panel. It is intentionally separate from the broad outbounds acceptance test:
// the first step seeds [direct rich, blocked sparse], the second reverses those
// semantic objects, and the checks cover Terraform state plus the generated
// Xray configuration returned by threexui_xray_config.
func TestAccXrayOutboundsReorderNoFieldBleed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayOutboundsReorderConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_outbounds.reorder", "outbound.#", "2"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.reorder", "outbound.0.tag", "direct"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.reorder", "outbound.0.send_through", "127.0.0.1"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.reorder", "outbound.0.target_strategy", "UseIPv4"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.reorder", "outbound.1.tag", "blocked"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccXrayOutboundsReorderConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayOutboundsReorderedStateExact("threexui_xray_outbounds.reorder"),
					testAccCheckXrayOutboundsReorderedRawConfig("data.threexui_xray_config.outbounds_reorder"),
				),
			},
			{
				Config:             testAccProviderConfig() + testAccXrayOutboundsReorderConfig(true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckXrayOutboundsReorderedStateExact(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		attributes := rs.Primary.Attributes
		expect := func(key, want string) error {
			if got := attributes[key]; got != want {
				return fmt.Errorf("%s.%s = %q, want %q", resourceName, key, got, want)
			}
			return nil
		}
		absent := func(key string) error {
			if got, exists := attributes[key]; exists && got != "" && got != "0" {
				return fmt.Errorf("%s.%s must be absent, got %q", resourceName, key, got)
			}
			return nil
		}

		for key, want := range map[string]string{
			"outbound.#":                                    "2",
			"outbound.0.tag":                                "blocked",
			"outbound.0.protocol":                           "blackhole",
			"outbound.0.blackhole_settings.#":               "1",
			"outbound.0.blackhole_settings.0.response_type": "none",
			"outbound.1.tag":                                "direct",
			"outbound.1.protocol":                           "freedom",
			"outbound.1.send_through":                       "127.0.0.1",
			"outbound.1.target_strategy":                    "UseIPv4",
			"outbound.1.mux.#":                              "1",
			"outbound.1.mux.0.enabled":                      "true",
			"outbound.1.mux.0.concurrency":                  "8",
			"outbound.1.freedom_settings.#":                 "1",
			"outbound.1.freedom_settings.0.domain_strategy": "AsIs",
		} {
			if err := expect(key, want); err != nil {
				return err
			}
		}

		for _, key := range []string{
			"outbound.0.send_through",
			"outbound.0.target_strategy",
			"outbound.0.mux.#",
			"outbound.0.freedom_settings.#",
			"outbound.0.dns_settings.#",
			"outbound.0.vmess_settings.#",
			"outbound.0.vless_settings.#",
			"outbound.0.trojan_settings.#",
			"outbound.0.shadowsocks_settings.#",
			"outbound.0.socks_settings.#",
			"outbound.0.http_settings.#",
			"outbound.0.wireguard_settings.#",
			"outbound.0.hysteria_settings.#",
			"outbound.1.blackhole_settings.#",
			"outbound.1.dns_settings.#",
			"outbound.1.vmess_settings.#",
			"outbound.1.vless_settings.#",
			"outbound.1.trojan_settings.#",
			"outbound.1.shadowsocks_settings.#",
			"outbound.1.socks_settings.#",
			"outbound.1.http_settings.#",
			"outbound.1.wireguard_settings.#",
			"outbound.1.hysteria_settings.#",
		} {
			if err := absent(key); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckXrayOutboundsReorderedRawConfig(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}
		var config map[string]any
		if err := json.Unmarshal([]byte(rs.Primary.Attributes["json"]), &config); err != nil {
			return fmt.Errorf("cannot parse %s.json: %w", dataSourceName, err)
		}
		outbounds, ok := config["outbounds"].([]any)
		if !ok {
			return fmt.Errorf("generated Xray config has no outbounds array")
		}
		byTag := make(map[string]map[string]any, len(outbounds))
		for _, raw := range outbounds {
			outbound, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			tag, _ := outbound["tag"].(string)
			if tag != "" {
				byTag[tag] = outbound
			}
		}
		blocked, ok := byTag["blocked"]
		if !ok {
			return fmt.Errorf("generated Xray config has no blocked outbound")
		}
		if blocked["protocol"] != "blackhole" {
			return fmt.Errorf("blocked protocol = %v, want blackhole", blocked["protocol"])
		}
		for _, key := range []string{"sendThrough", "targetStrategy", "mux"} {
			if value, exists := blocked[key]; exists {
				return fmt.Errorf("blocked outbound must not inherit %s, got %v", key, value)
			}
		}
		blockedSettings, _ := blocked["settings"].(map[string]any)
		if blockedSettings == nil {
			return fmt.Errorf("blocked outbound has no settings object")
		}
		for key := range blockedSettings {
			if key != "response" {
				return fmt.Errorf("blocked settings contain protocol-inapplicable %s: %v", key, blockedSettings[key])
			}
		}

		direct, ok := byTag["direct"]
		if !ok {
			return fmt.Errorf("generated Xray config has no direct outbound")
		}
		if direct["protocol"] != "freedom" || direct["sendThrough"] != "127.0.0.1" || direct["targetStrategy"] != "UseIPv4" {
			return fmt.Errorf("direct outbound lost configured fields: %v", direct)
		}
		if _, ok := direct["mux"].(map[string]any); !ok {
			return fmt.Errorf("direct outbound lost configured mux: %v", direct["mux"])
		}
		directSettings, _ := direct["settings"].(map[string]any)
		if directSettings == nil || directSettings["domainStrategy"] != "AsIs" {
			return fmt.Errorf("direct outbound lost freedom domainStrategy: %v", directSettings)
		}
		if _, exists := directSettings["response"]; exists {
			return fmt.Errorf("direct settings contain protocol-inapplicable response: %v", directSettings["response"])
		}
		return nil
	}
}

func testAccXrayOutboundsReorderConfig(reordered bool) string {
	direct := `
  outbound {
    tag             = "direct"
    protocol        = "freedom"
    send_through    = "127.0.0.1"
    target_strategy = "UseIPv4"

    mux {
      enabled     = true
      concurrency = 8
    }

    freedom_settings {
      domain_strategy = "AsIs"
    }
  }
`
	blocked := `
  outbound {
    tag      = "blocked"
    protocol = "blackhole"

    blackhole_settings {
      response_type = "none"
    }
  }
`
	outbounds := direct + blocked
	if reordered {
		outbounds = blocked + direct
	}
	return fmt.Sprintf(`
resource "threexui_xray_outbounds" "reorder" {
%s
}

data "threexui_xray_config" "outbounds_reorder" {
  depends_on = [threexui_xray_outbounds.reorder]
}
`, outbounds)
}
