package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceInbounds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccDataSourceInboundsConfig(),
				Check: resource.ComposeTestCheckFunc(
					// Verify inbound fields
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds.0.id"),
					resource.TestCheckResourceAttr("data.threexui_inbounds.all", "inbounds.0.protocol", "vless"),
					resource.TestCheckResourceAttr("data.threexui_inbounds.all", "inbounds.0.remark", "acc-ds-inbound-1"),
					resource.TestCheckResourceAttr("data.threexui_inbounds.all", "inbounds.0.enable", "true"),
					resource.TestCheckResourceAttr("data.threexui_inbounds.all", "inbounds.0.port", "24001"),
					// Verify nested block counts
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds.0.settings.#"),
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds.0.stream_settings.#"),
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds.0.sniffing.#"),
				),
			},
		},
	})
}

func TestAccDataSourceInboundsMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccDataSourceInboundsMultipleConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceInboundsCount("data.threexui_inbounds.all", 2),
				),
			},
		},
	})
}

func TestAccDataSourceServerStatus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "threexui_server_status" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_server_status.test", "json"),
					testAccCheckJSONValid("data.threexui_server_status.test", "json"),
				),
			},
		},
	})
}

func TestAccDataSourceXrayVersions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "threexui_xray_versions" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_xray_versions.test", "versions.0"),
				),
			},
		},
	})
}

func TestAccDataSourceXrayConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "threexui_xray_config" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_xray_config.test", "json"),
					testAccCheckJSONValid("data.threexui_xray_config.test", "json"),
					testAccCheckJSONContainsKey("data.threexui_xray_config.test", "json", "log"),
					testAccCheckJSONContainsKey("data.threexui_xray_config.test", "json", "outbounds"),
				),
			},
		},
	})
}

func TestAccDataSourceSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "threexui_settings" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_settings.test", "json"),
					testAccCheckJSONValid("data.threexui_settings.test", "json"),
					testAccCheckJSONContainsKey("data.threexui_settings.test", "json", "webPort"),
					testAccCheckJSONContainsKey("data.threexui_settings.test", "json", "webBasePath"),
				),
			},
		},
	})
}

// --- Check helpers ---

func testAccCheckJSONValid(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		val := rs.Primary.Attributes[attrName]
		if val == "" {
			return fmt.Errorf("%s.%s is empty", resourceName, attrName)
		}
		var raw any
		if err := json.Unmarshal([]byte(val), &raw); err != nil {
			return fmt.Errorf("%s.%s is not valid JSON: %w", resourceName, attrName, err)
		}
		return nil
	}
}

func testAccCheckJSONContainsKey(resourceName, attrName, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		val := rs.Primary.Attributes[attrName]
		var m map[string]any
		if err := json.Unmarshal([]byte(val), &m); err != nil {
			return fmt.Errorf("%s.%s is not valid JSON object: %w", resourceName, attrName, err)
		}
		if _, ok := m[key]; !ok {
			return fmt.Errorf("%s.%s does not contain key %q", resourceName, attrName, key)
		}
		return nil
	}
}

func testAccCheckDataSourceInboundsCount(resourceName string, minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		countStr, ok := rs.Primary.Attributes["inbounds.#"]
		if !ok {
			return fmt.Errorf("inbounds.# not found in %s", resourceName)
		}
		count := 0
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("cannot parse inbounds.#: %s", countStr)
		}
		if count < minCount {
			return fmt.Errorf("expected at least %d inbounds, got %d", minCount, count)
		}
		return nil
	}
}

// --- Config helpers ---

func testAccDataSourceInboundsConfig() string {
	return `
resource "threexui_inbound" "ds_test" {
  port     = 24001
  protocol = "vless"
  remark   = "acc-ds-inbound-1"
  enable   = true
}

data "threexui_inbounds" "all" {
  depends_on = [threexui_inbound.ds_test]
}
`
}

func testAccDataSourceInboundsMultipleConfig() string {
	return `
resource "threexui_inbound" "ds_test1" {
  port     = 24002
  protocol = "vless"
  remark   = "acc-ds-multi-1"
  enable   = true
}

resource "threexui_inbound" "ds_test2" {
  port     = 24003
  protocol = "vmess"
  remark   = "acc-ds-multi-2"
  enable   = true
}

data "threexui_inbounds" "all" {
  depends_on = [threexui_inbound.ds_test1, threexui_inbound.ds_test2]
}
`
}
