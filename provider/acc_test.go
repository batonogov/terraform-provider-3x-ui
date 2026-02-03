package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		"threexui": func() (*schema.Provider, error) { return Provider(), nil },
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

func TestAccInboundBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
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
		},
	})
}

func TestAccInboundClientBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccInboundWithClientConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("threexui_inbound_client.test", "id"),
					resource.TestCheckResourceAttr("threexui_inbound_client.test", "email", "acc-client@example.com"),
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
  settings = jsonencode({
    clients = [{
      id    = "11111111-1111-1111-1111-111111111111"
      email = "acc-client@example.com"
      flow  = "xtls-rprx-vision"
    }]
    decryption = "none"
  })
  stream_settings = jsonencode({})
  sniffing        = jsonencode({})
}
`, remark)
}

func testAccInboundWithClientConfig() string {
	return `
resource "threexui_inbound" "test" {
  port     = 23457
  protocol = "vless"
  remark   = "acc-inbound-client"
  enable   = true
  settings = jsonencode({
    clients = [{
      id    = "22222222-2222-2222-2222-222222222222"
      email = "acc-bootstrap@example.com"
      flow  = "xtls-rprx-vision"
    }]
    decryption = "none"
  })
  stream_settings = jsonencode({})
  sniffing        = jsonencode({})
}

resource "threexui_inbound_client" "test" {
  inbound_id = threexui_inbound.test.id
  client_id  = "33333333-3333-3333-3333-333333333333"
  email      = "acc-client@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
`
}
