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
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "log.0.loglevel", "warning"),
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "api.0.tag", "api"),
				),
			},
			// Update loglevel
			{
				Config: testAccProviderConfig() + testAccXrayBasicsConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "id", "xray_basics"),
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "log.0.loglevel", "info"),
					resource.TestCheckResourceAttr("threexui_xray_basics.test", "policy.0.level.0.handshake", "4"),
				),
			},
			// ImportState
			{
				ResourceName:      "threexui_xray_basics.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "xray_basics",
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
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "query_strategy", "UseIP"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "enable_parallel_query", "true"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "use_system_hosts", "false"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "server.0.address", "8.8.8.8"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "server.1.address", "localhost"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "server.1.port", "53"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayDNSConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "id", "xray_dns"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "query_strategy", "UseIPv4"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "enable_parallel_query", "false"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "use_system_hosts", "true"),
					resource.TestCheckResourceAttr("threexui_xray_dns.test", "server.0.address", "1.1.1.1"),
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
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "domain_strategy", "AsIs"),
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "rule.0.outbound_tag", "direct"),
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "rule.1.outbound_tag", "blocked"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayRoutingConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "id", "xray_routing"),
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "domain_strategy", "IPIfNonMatch"),
					resource.TestCheckResourceAttr("threexui_xray_routing.test", "rule.0.outbound_tag", "direct"),
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
					resource.TestCheckResourceAttr("threexui_xray_balancers.test", "balancer.0.tag", "bal1"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayBalancersConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_balancers.test", "id", "xray_balancers"),
					resource.TestCheckResourceAttr("threexui_xray_balancers.test", "balancer.0.tag", "bal-updated"),
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
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "bridge.0.tag", "bridge1"),
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "portal.0.tag", "portal1"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayReverseConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "id", "xray_reverse"),
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "bridge.0.tag", "bridge-updated"),
					resource.TestCheckResourceAttr("threexui_xray_reverse.test", "portal.0.tag", "portal-updated"),
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
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.0.tag", "direct"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.0.protocol", "freedom"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.1.tag", "blocked"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.2.tag", "dns-out"),
				),
			},
			// Update
			{
				Config: testAccProviderConfig() + testAccXrayOutboundsConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "id", "xray_outbounds"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.0.tag", "direct"),
					resource.TestCheckResourceAttr("threexui_xray_outbounds.test", "outbound.1.tag", "blocked"),
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
  log {
    loglevel = "warning"
    access   = "none"
    dns_log  = false
  }

  policy {
    system {
      stats_inbound_downlink  = true
      stats_inbound_uplink    = true
      stats_outbound_downlink = false
      stats_outbound_uplink   = false
    }

    level {
      id                  = 0
      stats_user_uplink   = true
      stats_user_downlink = true
    }
  }

  api {
    tag      = "api"
    services = ["HandlerService", "LoggerService", "StatsService"]
  }

  stats {}
}
`
}

func testAccXrayBasicsConfigUpdated() string {
	return `
resource "threexui_xray_basics" "test" {
  log {
    loglevel = "info"
    access   = "none"
    dns_log  = false
  }

  policy {
    system {
      stats_inbound_downlink  = true
      stats_inbound_uplink    = true
      stats_outbound_downlink = false
      stats_outbound_uplink   = false
    }

    level {
      id                  = 0
      handshake           = 4
      conn_idle           = 300
      stats_user_uplink   = true
      stats_user_downlink = true
    }
  }

  api {
    tag      = "api"
    services = ["HandlerService", "LoggerService", "StatsService"]
  }

  stats {}
}
`
}

func testAccXrayDNSConfig() string {
	return `
resource "threexui_xray_dns" "test" {
  query_strategy       = "UseIP"
  enable_parallel_query = true
  use_system_hosts      = false

  server {
    address = "8.8.8.8"
  }

  server {
    address = "localhost"
    port    = 53
    domains = ["geosite:cn"]
  }
}
`
}

func testAccXrayDNSConfigUpdated() string {
	return `
resource "threexui_xray_dns" "test" {
  query_strategy       = "UseIPv4"
  enable_parallel_query = false
  use_system_hosts      = true

  server {
    address = "1.1.1.1"
  }
}
`
}

func testAccXrayRoutingConfig() string {
	return `
resource "threexui_xray_routing" "test" {
  domain_strategy = "AsIs"

  rule {
    type         = "field"
    ip           = ["geoip:private"]
    outbound_tag = "direct"
  }

  rule {
    type         = "field"
    domain       = ["geosite:category-ads"]
    outbound_tag = "blocked"
  }
}
`
}

func testAccXrayRoutingConfigUpdated() string {
	return `
resource "threexui_xray_routing" "test" {
  domain_strategy = "IPIfNonMatch"

  rule {
    type         = "field"
    ip           = ["geoip:private"]
    outbound_tag = "direct"
  }
}
`
}

func testAccXrayBalancersConfig() string {
	return `
resource "threexui_xray_balancers" "test" {
  balancer {
    tag      = "bal1"
    selector = ["proxy-*"]

    strategy {
      type = "leastPing"
    }
  }
}
`
}

func testAccXrayBalancersConfigUpdated() string {
	return `
resource "threexui_xray_balancers" "test" {
  balancer {
    tag      = "bal-updated"
    selector = ["proxy-*", "direct-*"]

    strategy {
      type = "random"
    }
  }
}
`
}

func testAccXrayReverseConfig() string {
	return `
resource "threexui_xray_reverse" "test" {
  bridge {
    tag    = "bridge1"
    domain = "test.example.com"
  }

  portal {
    tag    = "portal1"
    domain = "test.example.com"
  }
}
`
}

func testAccXrayReverseConfigUpdated() string {
	return `
resource "threexui_xray_reverse" "test" {
  bridge {
    tag    = "bridge-updated"
    domain = "updated.example.com"
  }

  portal {
    tag    = "portal-updated"
    domain = "updated.example.com"
  }
}
`
}

func testAccXrayOutboundsConfig() string {
	return `
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
`
}

func testAccXrayOutboundsConfigUpdated() string {
	return `
resource "threexui_xray_outbounds" "test" {
  outbound {
    tag      = "direct"
    protocol = "freedom"

    freedom_settings {
      domain_strategy = "UseIP"
    }
  }

  outbound {
    tag      = "blocked"
    protocol = "blackhole"

    blackhole_settings {
      response_type = "http"
    }
  }

  outbound {
    tag      = "dns-out"
    protocol = "dns"
  }
}
`
}
