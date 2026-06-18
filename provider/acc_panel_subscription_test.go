package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPanelSubscription_RestartsPanel is the end-to-end regression guard for #291.
//
// #292 claimed to fix #291 by adding the subscription server-binding keys to
// panelSettingsNeedRestart, but it never wired the restart into applySubscription,
// so a change to a server-binding field (sub_port/sub_path/sub_listen/…) was
// written to the database without the 3x-ui subscription server actually
// rebinding — the sub server kept serving the OLD path/port until a manual panel
// restart.
//
// Why sub_path (not sub_enable/sub_port): 3x-ui v3.x ships with subEnable=true,
// subPort=2096 and subPath="/sub/" by DEFAULT (see 3x-ui/internal/web/service/
// setting.go defaults), so the sub server is already listening on :2096 from a
// fresh start and merely re-asserting those defaults would not exercise the
// restart path. Changing sub_path to a NON-default value ("/sub2/") forces a
// router re-registration that only happens at sub-server startup (3x-ui/internal/
// sub/sub.go initRouter(), called from Start()), i.e. only after a panel restart.
//
// The test applies a non-default sub_path together with a real inbound client so
// we have a valid subscription id, then GETs the subscription endpoint under the
// NEW path. After the provider-triggered panel restart the sub server serves the
// client's page there (200); without the restart the router still has the old
// path and the request 404s. A bounded poll covers the panel-restart window.
//
// Why 200 rather than "not 404": the 3x-ui subscription handler returns 404 for
// an UNKNOWN subId (see 3x-ui/internal/sub/controller.go), so "non-404" would be
// an unreliable success signal. With a real client's subId and the correct path
// the handler serves the subscription page (200) — the actual user-visible
// contract (the subscription URL works).
//
// Requires the subscription port to be published from the acc-test container —
// see the "2096:2096" port mapping in docker-compose.yaml.
func TestAccPanelSubscription_RestartsPanel(t *testing.T) {
	// The subscription server + its link-generation page exist in all supported
	// 3x-ui lines (v3.1.x–v3.3.x); no version guard needed.

	const subPort = 2096
	const subPath = "/sub2/"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "threexui_panel_subscription" "sub" {
  sub_enable      = true
  sub_listen      = ""
  sub_domain      = ""
  sub_port        = %d
  sub_path        = %q
  sub_cert_file   = ""
  sub_key_file    = ""
  sub_json_enable = true
}

resource "threexui_inbound" "host" {
  port     = 25210
  protocol = "vless"
  remark   = "acc-sub-restart-host"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "reality"
    reality_settings {
      target       = "google.com:443"
      server_names = ["google.com"]
    }
  }
}

resource "threexui_inbound_client" "cli" {
  inbound_id = threexui_inbound.host.id
  email      = "sub-restart@test.com"
  enable     = true
}
`, subPort, subPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.sub", "sub_path", subPath),
					resource.TestCheckResourceAttrWith("threexui_inbound_client.cli", "sub_id", func(subID string) error {
						if subID == "" {
							return fmt.Errorf("client sub_id is empty; cannot probe subscription endpoint")
						}
						subURL := subscriptionTestURL(t, subPort, subPath, subID)
						return waitForSubscriptionReady(t, subURL)
					}),
				),
			},
		},
	})
}

// subscriptionTestURL builds the subscription endpoint URL for the acc-test.
// Host is taken from THREEXUI_ENDPOINT (defaults to localhost when unset/unparseable);
// the subscription server is published on its own port. subPath is "/"-surrounded
// (e.g. "/sub2/"); the route is registered as <subPath>:subid (see sub.go initRouter),
// so the final URL is host:port/<subPath without trailing slash>/<subId>.
func subscriptionTestURL(t *testing.T, subPort int, subPath, subID string) string {
	t.Helper()
	host := "localhost"
	if ep := os.Getenv("THREEXUI_ENDPOINT"); ep != "" {
		if u, err := url.Parse(ep); err == nil && u.Hostname() != "" {
			host = u.Hostname()
		}
	}
	return fmt.Sprintf("http://%s:%d%s/%s", host, subPort, strings.TrimRight(subPath, "/"), subID)
}

// waitForSubscriptionReady polls the subscription endpoint until it answers
// with HTTP 200 (the sub server rebound and serves the client's subscription
// page under the new path), covering the panel-restart window. Bounded to ~60s.
func waitForSubscriptionReady(t *testing.T, subURL string) error {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	const (
		attempts = 60
		wait     = 1 * time.Second
	)
	var lastErr error
	for i := 0; i < attempts; i++ {
		resp, err := client.Get(subURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("subscription endpoint ready at %s after %d attempt(s)", subURL, i+1)
				return nil
			}
			lastErr = fmt.Errorf("subscription endpoint returned HTTP %d (want 200)", resp.StatusCode)
		} else {
			lastErr = fmt.Errorf("subscription endpoint not reachable: %w", err)
		}
		time.Sleep(wait)
	}
	return fmt.Errorf("subscription endpoint not ready after %d attempts: %w", attempts, lastErr)
}
