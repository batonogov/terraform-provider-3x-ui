package provider

import (
	"fmt"
	"io"
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

// TestAccPanelSubscription_JsonBodySettingsTakeEffect is the end-to-end guard
// for #443.
//
// The subscription server reads its JSON-body settings — sub_json_mux,
// sub_json_rules, sub_json_final_mask and (v3.7.0+) sub_json_observatory — once,
// inside (*sub.Server).initRouter(), which freezes them into the SUBController it
// builds (3x-ui-3.7.0/internal/sub/sub.go:143-158, :251) and runs only from
// Start(). Before #443 those keys were not in restartKeys, so `terraform apply`
// wrote them to the panel database, state matched the panel, and every served
// JSON subscription kept the OLD body indefinitely — the same silent no-op class
// as #291.
//
// The test asserts the user-visible contract rather than the restart itself: it
// applies a non-default sub_json_path together with a mux blob carrying a
// recognisable concurrency value, then GETs the JSON subscription. Two things
// have to be true for that to pass, and both require the router to have been
// rebuilt: the new path must be registered (otherwise 404), and the served
// outbound must carry the configured mux (json_service.go:519-534 copies it into
// outbound.Mux verbatim, and the per-inbound xmux override that would suppress it
// only applies to xhttp streams — this inbound is plain tcp).
func TestAccPanelSubscription_JsonBodySettingsTakeEffect(t *testing.T) {
	const (
		subPort     = 2096
		subPath     = "/sub3/"
		subJSONPath = "/json3/"
		// Distinctive value so a stale body cannot accidentally match.
		muxConcurrency = 37
	)
	muxBlob := fmt.Sprintf(`{\"enabled\":true,\"concurrency\":%d}`, muxConcurrency)

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
  sub_json_path   = %q
  sub_json_mux    = "%s"
}

resource "threexui_inbound" "host" {
  port     = 25211
  protocol = "vless"
  remark   = "acc-sub-json-body-host"
  enable   = true
  vless_settings {
    decryption = "none"
  }
  stream_settings {
    network  = "tcp"
    security = "none"
  }
}

resource "threexui_inbound_client" "cli" {
  inbound_id = threexui_inbound.host.id
  email      = "sub-json-body@test.com"
  enable     = true
}
`, subPort, subPath, subJSONPath, muxBlob),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_subscription.sub", "sub_json_path", subJSONPath),
					resource.TestCheckResourceAttrWith("threexui_inbound_client.cli", "sub_id", func(subID string) error {
						if subID == "" {
							return fmt.Errorf("client sub_id is empty; cannot probe the JSON subscription endpoint")
						}
						jsonURL := subscriptionTestURL(t, subPort, subJSONPath, subID)
						body, err := waitForSubscriptionBody(t, jsonURL)
						if err != nil {
							return err
						}
						// The panel re-marshals the blob, so match on the value, not on the
						// exact spacing of the JSON we sent.
						if !strings.Contains(strings.ReplaceAll(body, " ", ""), fmt.Sprintf(`"concurrency":%d`, muxConcurrency)) {
							return fmt.Errorf("JSON subscription does not carry the configured mux (want concurrency %d); "+
								"the sub server is still serving a pre-restart body. Got: %s",
								muxConcurrency, truncateForError(body))
						}
						return nil
					}),
				),
			},
		},
	})
}

// waitForSubscriptionBody polls a subscription endpoint until it answers HTTP 200
// and returns the response body, covering the panel-restart window.
func waitForSubscriptionBody(t *testing.T, subURL string) (string, error) {
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
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && readErr == nil {
				t.Logf("subscription endpoint ready at %s after %d attempt(s)", subURL, i+1)
				return string(body), nil
			}
			if readErr != nil {
				lastErr = fmt.Errorf("reading subscription body: %w", readErr)
			} else {
				lastErr = fmt.Errorf("subscription endpoint returned HTTP %d (want 200)", resp.StatusCode)
			}
		} else {
			lastErr = fmt.Errorf("subscription endpoint not reachable: %w", err)
		}
		time.Sleep(wait)
	}
	return "", fmt.Errorf("subscription endpoint not ready after %d attempts: %w", attempts, lastErr)
}

func truncateForError(s string) string {
	const max = 400
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}
