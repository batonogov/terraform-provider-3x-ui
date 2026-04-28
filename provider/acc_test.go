package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	envEndpoint           = "THREEXUI_ENDPOINT"
	envBasePath           = "THREEXUI_BASE_PATH"
	envUsername           = "THREEXUI_USERNAME"
	envPassword           = "THREEXUI_PASSWORD"
	envInsecureSkipVerify = "THREEXUI_INSECURE_SKIP_VERIFY"
)

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"threexui": providerserver.NewProtocol6WithError(New("test")()),
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

	namespace := os.Getenv("TF_ACC_PROVIDER_NAMESPACE")
	host := os.Getenv("TF_ACC_PROVIDER_HOST")
	if namespace == "" {
		namespace = "hashicorp"
	}
	if host == "" {
		host = "registry.terraform.io"
	}

	config := fmt.Sprintf(`terraform {
  required_providers {
    threexui = {
      source = "%s/%s/threexui"
    }
  }
}

provider "threexui" {
  endpoint            = %q
  username            = %q
  password            = %q
`, host, namespace, endpoint, username, password)
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

func testAccClientFromEnvNoLogin() (*Client, error) {
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

	return NewClient(ClientConfig{
		Endpoint:           endpoint,
		BasePath:           basePath,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecureBool,
		Timeout:            30 * time.Second,
	})
}

// destroyVisibilityAttempts × destroyVisibilityBackoff is how long the
// CheckDestroy helpers wait for a successful DELETE to become visible to a
// follow-up GET. 3x-ui's DELETE endpoint can return success while a
// concurrent GET still observes the row under SQLite contention — see
// issues #157 and #161. CI runners are slower than local hardware and the
// matrix test hammers SQLite for tens of seconds before late subtests run.
// 15 × 500ms = 7.5s was insufficient on v2.9.1 in CI; 30 × 500ms = 15s
// matches the worst-case lag observed.
const (
	destroyVisibilityAttempts = 30
	destroyVisibilityBackoff  = 500 * time.Millisecond
)

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
		var lastErr error
		gone := false
		for attempt := 0; attempt < destroyVisibilityAttempts; attempt++ {
			if _, getErr := client.GetInbound(context.Background(), id); getErr != nil {
				gone = true
				break
			}
			lastErr = nil
			time.Sleep(destroyVisibilityBackoff)
		}
		if !gone {
			if lastErr != nil {
				return fmt.Errorf("inbound %d still exists after %d attempts: %w", id, destroyVisibilityAttempts, lastErr)
			}
			return fmt.Errorf("inbound %d still exists after %d attempts", id, destroyVisibilityAttempts)
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
		gone := false
		for attempt := 0; attempt < destroyVisibilityAttempts; attempt++ {
			inbound, getErr := client.GetInbound(context.Background(), inboundID)
			if getErr != nil {
				gone = true
				break
			}
			settings, parseErr := parseInboundSettings(inbound.Settings)
			if parseErr != nil {
				return parseErr
			}
			if findClientByID(settings.Clients, clientID) == nil {
				gone = true
				break
			}
			time.Sleep(destroyVisibilityBackoff)
		}
		if !gone {
			return fmt.Errorf("inbound client %s still exists after %d attempts", rs.Primary.ID, destroyVisibilityAttempts)
		}
	}
	return nil
}

func TestAccInboundBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundDestroyed,
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
					resource.TestCheckResourceAttr("threexui_inbound.test", "stream_settings.network", "tcp"),
					resource.TestCheckResourceAttr("threexui_inbound.test", "sniffing.enabled", "true"),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInboundClientDestroyed,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundConfig("acc-inbound-ds") + testAccDataSourcesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.threexui_inbounds.all", "inbounds"),
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
  vless_settings {
    decryption = "none"
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
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "none"
    tcp_settings {
      accept_proxy_protocol = false
      header_type           = "none"
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
  vless_settings {
    decryption = "none"
  }
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
