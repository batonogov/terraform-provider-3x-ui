package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccXrayBasics(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayBasicsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "id", "xray_basics"),
					resource.TestCheckResourceAttrSet("threexui_xray_basics.test", "json"),
				),
			},
			// Update loglevel
			{
				Config: testAccProviderConfig() + testAccXrayBasicsConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "id", "xray_basics"),
					resource.TestCheckResourceAttrSet("threexui_xray_basics.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:            "threexui_xray_basics.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "xray_basics",
				ImportStateVerifyIgnore: []string{"json"},
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayBasicsConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccXrayDNS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayDNSConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "id", "xray_dns"),
					resource.TestCheckResourceAttrSet("threexui_xray_dns.test", "json"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayDNSConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "id", "xray_dns"),
					resource.TestCheckResourceAttrSet("threexui_xray_dns.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_dns",
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayDNSConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccXrayRouting(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayRoutingConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "id", "xray_routing"),
					resource.TestCheckResourceAttrSet("threexui_xray_routing.test", "json"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayRoutingConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "id", "xray_routing"),
					resource.TestCheckResourceAttrSet("threexui_xray_routing.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_routing.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_routing",
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayRoutingConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccXrayBalancers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayBalancersConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_balancers.test", "id", "xray_balancers"),
					resource.TestCheckResourceAttrSet("threexui_xray_balancers.test", "json"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayBalancersConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_balancers.test", "id", "xray_balancers"),
					resource.TestCheckResourceAttrSet("threexui_xray_balancers.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_balancers.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_balancers",
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayBalancersConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccXrayReverse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayReverseConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "id", "xray_reverse"),
					resource.TestCheckResourceAttrSet("threexui_xray_reverse.test", "json"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayReverseConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "id", "xray_reverse"),
					resource.TestCheckResourceAttrSet("threexui_xray_reverse.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_reverse.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_reverse",
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayReverseConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccXrayOutbounds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + testAccXrayOutboundsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "id", "xray_outbounds"),
					resource.TestCheckResourceAttrSet("threexui_xray_outbounds.test", "json"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayOutboundsConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "id", "xray_outbounds"),
					resource.TestCheckResourceAttrSet("threexui_xray_outbounds.test", "json"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_outbounds.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_outbounds",
			},
			// Idempotency
			{
				Config:             testAccProviderConfig() + testAccXrayOutboundsConfigUpdated(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- Config helpers ---

func testAccXrayBasicsConfig() string {
	return `
resource "threexui_xray_basics" "test" {
  json = jsonencode({
    log = {
      loglevel = "warning"
      access   = "none"
      dns_log  = false
    }
    policy = {
      system = {
        stats_inbound_downlink  = true
        stats_inbound_uplink    = true
        stats_outbound_downlink = false
        stats_outbound_uplink   = false
      }
      level = [
        {
          id                  = 0
          stats_user_uplink   = true
          stats_user_downlink = true
        }
      ]
    }
    api = {
      tag      = "api"
      services = ["HandlerService", "LoggerService", "StatsService"]
    }
    stats = {}
  })
}
`
}

func testAccXrayBasicsConfigUpdated() string {
	return `
resource "threexui_xray_basics" "test" {
  json = jsonencode({
    log = {
      loglevel = "info"
      access   = "none"
      dns_log  = false
    }
    policy = {
      system = {
        stats_inbound_downlink  = true
        stats_inbound_uplink    = true
        stats_outbound_downlink = false
        stats_outbound_uplink   = false
      }
      level = [
        {
          id                  = 0
          handshake           = 4
          conn_idle           = 300
          stats_user_uplink   = true
          stats_user_downlink = true
        }
      ]
    }
    api = {
      tag      = "api"
      services = ["HandlerService", "LoggerService", "StatsService"]
    }
    stats = {}
  })
}
`
}

func testAccXrayDNSConfig() string {
	return `
resource "threexui_xray_dns" "test" {
  json = jsonencode({
    server = [
      {
        address = "8.8.8.8"
      },
      {
        address = "localhost"
        port    = 53
        domains = ["geosite:cn"]
      }
    ]
    query_strategy = "UseIP"
  })
}
`
}

func testAccXrayDNSConfigUpdated() string {
	return `
resource "threexui_xray_dns" "test" {
  json = jsonencode({
    server = [
      {
        address = "1.1.1.1"
      }
    ]
    query_strategy = "UseIPv4"
  })
}
`
}

func testAccXrayRoutingConfig() string {
	return `
resource "threexui_xray_routing" "test" {
  json = jsonencode({
    domain_strategy = "AsIs"
    rule = [
      {
        type         = "field"
        ip           = ["geoip:private"]
        outbound_tag = "direct"
      },
      {
        type         = "field"
        domain       = ["geosite:category-ads"]
        outbound_tag = "blocked"
      }
    ]
  })
}
`
}

func testAccXrayRoutingConfigUpdated() string {
	return `
resource "threexui_xray_routing" "test" {
  json = jsonencode({
    domain_strategy = "IPIfNonMatch"
    rule = [
      {
        type         = "field"
        ip           = ["geoip:private"]
        outbound_tag = "direct"
      }
    ]
  })
}
`
}

func testAccXrayBalancersConfig() string {
	return `
resource "threexui_xray_balancers" "test" {
  json = jsonencode({
    balancer = [
      {
        tag      = "bal1"
        selector = ["proxy-*"]
        strategy = [
          {
            type = "leastPing"
          }
        ]
      }
    ]
  })
}
`
}

func testAccXrayBalancersConfigUpdated() string {
	return `
resource "threexui_xray_balancers" "test" {
  json = jsonencode({
    balancer = [
      {
        tag      = "bal-updated"
        selector = ["proxy-*", "direct-*"]
        strategy = [
          {
            type = "random"
          }
        ]
      }
    ]
  })
}
`
}

func testAccXrayReverseConfig() string {
	return `
resource "threexui_xray_reverse" "test" {
  json = jsonencode({
    bridge = [
      {
        tag    = "bridge1"
        domain = "test.example.com"
      }
    ]
    portal = [
      {
        tag    = "portal1"
        domain = "test.example.com"
      }
    ]
  })
}
`
}

func testAccXrayReverseConfigUpdated() string {
	return `
resource "threexui_xray_reverse" "test" {
  json = jsonencode({
    bridge = [
      {
        tag    = "bridge-updated"
        domain = "updated.example.com"
      }
    ]
    portal = [
      {
        tag    = "portal-updated"
        domain = "updated.example.com"
      }
    ]
  })
}
`
}

func testAccXrayOutboundsConfig() string {
	return `
resource "threexui_xray_outbounds" "test" {
  json = jsonencode({
    outbound = [
      {
        tag      = "direct"
        protocol = "freedom"
        freedom_settings = [
          {
            domain_strategy = "AsIs"
          }
        ]
      },
      {
        tag      = "blocked"
        protocol = "blackhole"
        blackhole_settings = [
          {
            response_type = "none"
          }
        ]
      },
      {
        tag      = "dns-out"
        protocol = "dns"
      }
    ]
  })
}
`
}

func testAccXrayOutboundsConfigUpdated() string {
	return `
resource "threexui_xray_outbounds" "test" {
  json = jsonencode({
    outbound = [
      {
        tag      = "direct"
        protocol = "freedom"
        freedom_settings = [
          {
            domain_strategy = "UseIP"
          }
        ]
      },
      {
        tag      = "blocked"
        protocol = "blackhole"
        blackhole_settings = [
          {
            response_type = "http"
          }
        ]
      },
      {
        tag      = "dns-out"
        protocol = "dns"
      }
    ]
  })
}
`
}
