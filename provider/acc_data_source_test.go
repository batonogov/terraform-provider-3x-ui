package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceInbounds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccDataSourceInboundsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds"),
				),
			},
		},
	})
}

func TestAccDataSourceInboundsMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
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

func TestAccDataSourceOnlineClients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "threexui_online_clients" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_online_clients.test", "id"),
					resource.TestCheckResourceAttrSet("data.threexui_online_clients.test", "clients.#"),
				),
			},
		},
	})
}

func TestAccDataSourceClientTraffics(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccDataSourceClientTrafficsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_client_traffics.test", "id"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "email", "acc-ds-traffic-client"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "up", "0"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "down", "0"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "total", "0"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "expiry_time", "0"),
					resource.TestCheckResourceAttr("data.threexui_client_traffics.test", "enable", "true"),
					resource.TestCheckResourceAttrPair(
						"data.threexui_client_traffics.test", "inbound_id",
						"threexui_inbound.ds_traffic_test", "id",
					),
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

func testAccCheckJSONContainsKey(resourceName, attrName, key string) resource.TestCheckFunc { //nolint:unparam // attrName may vary in future tests
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
		inboundsJSON := rs.Primary.Attributes["inbounds"]
		if inboundsJSON == "" {
			return fmt.Errorf("inbounds attribute is empty in %s", resourceName)
		}
		var arr []any
		if err := json.Unmarshal([]byte(inboundsJSON), &arr); err != nil {
			return fmt.Errorf("cannot parse inbounds JSON: %w", err)
		}
		if len(arr) < minCount {
			return fmt.Errorf("expected at least %d inbounds, got %d", minCount, len(arr))
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
  vless_settings {
    decryption = "none"
  }
}

data "threexui_inbounds" "all" {
  depends_on = [threexui_inbound.ds_test]
}
`
}

func testAccDataSourceClientTrafficsConfig() string {
	return `
resource "threexui_inbound" "ds_traffic_test" {
  port     = 24010
  protocol = "vless"
  remark   = "acc-ds-traffic"
  enable   = true
  vless_settings {
    decryption = "none"
  }
}

resource "threexui_inbound_client" "ds_traffic_client" {
  inbound_id = threexui_inbound.ds_traffic_test.id
  email      = "acc-ds-traffic-client"
  enable     = true
}

data "threexui_client_traffics" "test" {
  email      = threexui_inbound_client.ds_traffic_client.email
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
  vless_settings {
    decryption = "none"
  }
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
