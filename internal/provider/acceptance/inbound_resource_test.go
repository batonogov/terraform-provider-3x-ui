package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/batonogov/terraform-provider-3x-ui/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"3xui": providerserver.NewProtocol6WithError(provider.New("acc-test")()),
}

func TestAccInbound_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set")
	}
	if os.Getenv("THREEXUI_BASE_URL") == "" {
		t.Skip("THREEXUI_BASE_URL must be set for acceptance tests")
	}
	if os.Getenv("THREEXUI_USERNAME") == "" || os.Getenv("THREEXUI_PASSWORD") == "" {
		t.Skip("THREEXUI_USERNAME and THREEXUI_PASSWORD must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInboundConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("3xui_inbound.test", "protocol", "vless"),
				),
			},
			{
				ResourceName:      "3xui_inbound.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInboundConfig() string {
	return `
provider "3xui" {
  base_url        = "${env.THREEXUI_BASE_URL}"
  username        = "${env.THREEXUI_USERNAME}"
  password        = "${env.THREEXUI_PASSWORD}"
  tls_skip_verify = true
}

resource "3xui_inbound" "test" {
  remark   = "tf-acc"
  protocol = "vless"
  port     = 28000
  settings_json = jsonencode({
    clients = []
  })
}
`
}
