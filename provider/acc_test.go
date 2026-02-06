package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	envEndpoint           = "THREEXUI_ENDPOINT"
	envBasePath           = "THREEXUI_BASE_PATH"
	envUsername           = "THREEXUI_USERNAME"
	envPassword           = "THREEXUI_PASSWORD"
	envInsecureSkipVerify = "THREEXUI_INSECURE_SKIP_VERIFY"
)

func testAccProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"threexui": func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam // signature required by Terraform SDK
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	if os.Getenv(envEndpoint) == "" {
		t.Skipf("%s not set", envEndpoint)
	}
}

func testAccProviderConfig() string {
	endpoint := os.Getenv(envEndpoint)
	basePath := os.Getenv(envBasePath)
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	insecure := os.Getenv(envInsecureSkipVerify)
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}

	config := fmt.Sprintf("provider \"threexui\" {\n  endpoint            = %q\n  username            = %q\n  password            = %q\n", endpoint, username, password)
	if basePath != "" {
		config += fmt.Sprintf("  base_path           = %q\n", basePath)
	}
	if insecure != "" {
		config += fmt.Sprintf("  insecure_skip_verify = %s\n", insecure)
	}
	config += "}\n"
	return config + "\n"
}

func testAccClientFromEnv() (*Client, error) {
	endpoint := os.Getenv(envEndpoint)
	basePath := os.Getenv(envBasePath)
	username := os.Getenv(envUsername)
	password := os.Getenv(envPassword)
	insecure := os.Getenv(envInsecureSkipVerify)
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}
	insecureBool := false
	if insecure != "" {
		if v, err := strconv.ParseBool(insecure); err == nil {
			insecureBool = v
		}
	}

	client, err := NewClient(ClientConfig{
		Endpoint:           endpoint,
		BasePath:           basePath,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecureBool,
		Timeout:            30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if err := client.Login(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}

func testAccCheckInboundDestroyed(state *terraform.State) error {
	client, err := testAccClientFromEnv()
	if err != nil {
		return fmt.Errorf("client init failed: %w", err)
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "threexui_inbound" {
			continue
		}
		id, err := parseID(rs.Primary.ID)
		if err != nil {
			continue
		}
		if _, err := client.GetInbound(context.Background(), id); err == nil {
			return fmt.Errorf("inbound %d still exists", id)
		}
	}
	return nil
}

func testAccCheckInboundClientDestroyed(state *terraform.State) error {
	client, err := testAccClientFromEnv()
	if err != nil {
		return fmt.Errorf("client init failed: %w", err)
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "threexui_inbound_client" {
			continue
		}
		inboundID, clientID, err := splitInboundClientID(rs.Primary.ID)
		if err != nil {
			continue
		}
		inbound, err := client.GetInbound(context.Background(), inboundID)
		if err != nil {
			continue
		}
		settings, err := parseInboundSettings(inbound.Settings)
		if err != nil {
			return err
		}
		if found := findClientByID(settings.Clients, clientID); found != nil {
			return fmt.Errorf("inbound client %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccInboundBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundConfig("acc-inbound-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound.test", "id"),
					resource.TestCheckResourceAttr("threexui_inbound.test", "remark", "acc-inbound-1"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccInboundConfig("acc-inbound-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.test", "remark", "acc-inbound-2"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccInboundConfigWithExtras("acc-inbound-3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound.test", "remark", "acc-inbound-3"),
					resource.TestCheckResourceAttr("threexui_inbound.test", "stream_settings.0.network", "tcp"),
					resource.TestCheckResourceAttr("threexui_inbound.test", "sniffing.0.enabled", "true"),
				),
			},
			{
				ResourceName:      "threexui_inbound.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInboundClientBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckInboundClientDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundWithClientConfig("acc-client@example.com", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.test", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.test", "email", "acc-client@example.com"),
				),
			},
			{
				Config: testAccProviderConfig() + testAccInboundWithClientConfig("acc-client-updated@example.com", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_inbound_client.test", "email", "acc-client-updated@example.com"),
					resource.TestCheckResourceAttr("threexui_inbound_client.test", "enable", "false"),
				),
			},
			{
				ResourceName:            "threexui_inbound_client.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundConfig("acc-inbound-ds") + testAccDataSourcesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds.0.id"),
					resource.TestCheckResourceAttrSet("data.threexui_server_status.status", "json"),
					resource.TestCheckResourceAttrSet("data.threexui_xray_versions.versions", "versions.0"),
					resource.TestCheckResourceAttrSet("data.threexui_xray_config.config", "json"),
					resource.TestCheckResourceAttrSet("data.threexui_settings.settings", "json"),
				),
			},
		},
	})
}

func testAccInboundConfig(remark string) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "test" {
  port     = 23456
  protocol = "vless"
  remark   = %q
  enable   = true
  settings {
    decryption = "none"
    clients {
      email = "acc-client@example.com"
      flow  = "xtls-rprx-vision"
    }
  }
}
`, remark)
}

func testAccInboundConfigWithExtras(remark string) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "test" {
  port     = 23456
  protocol = "vless"
  remark   = %q
  enable   = true
  settings {
    decryption = "none"
    clients {
      email = "acc-client@example.com"
      flow  = "xtls-rprx-vision"
    }
  }
  stream_settings {
    network  = "tcp"
    security = "none"
    tcp_settings {
      accept_proxy_protocol = false
      header {
        type = "none"
      }
    }
  }
  sniffing {
    enabled       = true
    dest_override = ["http", "tls"]
    metadata_only = false
    route_only    = false
  }
}
`, remark)
}

func testAccInboundWithClientConfig(clientEmail string, clientEnable bool) string {
	return fmt.Sprintf(`
resource "threexui_inbound" "test" {
  port     = 23457
  protocol = "vless"
  remark   = "acc-inbound-client"
  enable   = true
}

resource "threexui_inbound_client" "test" {
  inbound_id = threexui_inbound.test.id
  email      = %q
  enable     = %t
  flow       = "xtls-rprx-vision"
}
`, clientEmail, clientEnable)
}

func testAccDataSourcesConfig() string {
	return `
data "threexui_inbounds" "all" {
  depends_on = [threexui_inbound.test]
}

data "threexui_server_status" "status" {
  depends_on = [threexui_inbound.test]
}

data "threexui_xray_versions" "versions" {
  depends_on = [threexui_inbound.test]
}

data "threexui_xray_config" "config" {
  depends_on = [threexui_inbound.test]
}

data "threexui_settings" "settings" {
  depends_on = [threexui_inbound.test]
}
`
}
