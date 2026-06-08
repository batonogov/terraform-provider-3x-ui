package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccXrayVersion(t *testing.T) {
	testAccPreCheck(t)
	client, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("client init: %s", err)
	}
	ctx := context.Background()

	// Verify GetCurrentXrayVersion returns a v-prefixed string.
	currentVersion, err := client.GetCurrentXrayVersion(ctx)
	if err != nil {
		t.Fatalf("GetCurrentXrayVersion: %s", err)
	}
	if currentVersion == "" {
		t.Fatal("GetCurrentXrayVersion returned empty string")
	}
	if !strings.HasPrefix(currentVersion, "v") {
		t.Fatalf("GetCurrentXrayVersion should return v-prefixed version, got %q", currentVersion)
	}
	t.Logf("current xray version: %s", currentVersion)

	// Verify GetXrayVersions returns a non-empty list with v-prefixed versions.
	versions, err := client.GetXrayVersions(ctx)
	if err != nil {
		t.Fatalf("GetXrayVersions: %s", err)
	}
	if len(versions) == 0 {
		t.Fatal("GetXrayVersions returned empty list")
	}
	if !strings.HasPrefix(versions[0], "v") {
		t.Fatalf("GetXrayVersions should return v-prefixed versions, got %q", versions[0])
	}
	t.Logf("available versions: %d (first: %s)", len(versions), versions[0])
}

func TestAccXrayVersionResource(t *testing.T) {
	// NOTE: this test requires network access from the 3x-ui container
	// to download Xray releases from GitHub. It is skipped in Docker
	// environments where outbound access is restricted.
	testAccPreCheck(t)
	client, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("client init: %s", err)
	}
	ctx := context.Background()

	currentVersion, err := client.GetCurrentXrayVersion(ctx)
	if err != nil {
		t.Fatalf("GetCurrentXrayVersion: %s", err)
	}

	// Pre-flight: try installing the current version. If it fails
	// (e.g. Docker can't reach GitHub), skip the test.
	if err := client.InstallXray(ctx, currentVersion); err != nil {
		t.Skipf("InstallXray not available in this environment: %s", err)
	}

	config := testAccProviderConfig() + `
resource "threexui_xray_version" "test" {
  version = "` + currentVersion + `"
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_version.test", "id", "xray_version"),
					resource.TestCheckResourceAttr("threexui_xray_version.test", "version", currentVersion),
					resource.TestCheckResourceAttr("threexui_xray_version.test", "current_version", currentVersion),
				),
			},
			// Idempotency.
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccXrayVersionDrift verifies that the provider detects external drift.
// It installs version A via Terraform, changes the version to B outside
// Terraform via the API, then asserts that the next plan is non-empty
// (drift detected) and that applying brings the version back to A.
func TestAccXrayVersionDrift(t *testing.T) {
	testAccPreCheck(t)
	// InstallXray accepts the request but the panel never picks up the new
	// binary regardless of how many retries or how long we wait. This is a
	// deterministic upstream defect that cannot be fixed at the provider level.
	// Remove the gate once the upstream pickup is fixed.
	//
	// v2.9.1: verified across two full CI retries on PR #162. See #163.
	// v2.9.2: same symptom — version unchanged after 3×20 polls (PR #229 CI).
	// v2.9.3: intermittent — drift simulation succeeds but version reverts
	// before Terraform's Read, causing "empty refresh plan" (PR #229 CI).
	// v3.0.0, v3.0.1: panel intermittently fails to pick up the new binary
	// within the poll budget. See #224.
	skipOnFlakyVersions(t,
		"InstallXray pickup is broken or unreliable on this panel version",
		"v2.9.1", "v2.9.2", "v2.9.3", "v3.0.0", "v3.0.1",
		"v3.2.5", "v3.2.6", "v3.2.7")
	client, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("client init: %s", err)
	}
	ctx := context.Background()

	// Get available versions; we need at least two distinct ones.
	versions, err := client.GetXrayVersions(ctx)
	if err != nil {
		t.Fatalf("GetXrayVersions: %s", err)
	}

	// GetCurrentXrayVersion may return ErrXrayVersionUnknown if xray is
	// restarting after a previous test's InstallXray. Retry with the same
	// budget as waitForXrayVersion (90s) to accommodate slow CI runners.
	var currentVersion string
	for i := 0; i < 90; i++ {
		currentVersion, err = client.GetCurrentXrayVersion(ctx)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrXrayVersionUnknown) {
			t.Fatalf("GetCurrentXrayVersion: %s", err)
		}
		select {
		case <-ctx.Done():
			t.Fatalf("GetCurrentXrayVersion: context cancelled: %s", ctx.Err())
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		t.Fatalf("GetCurrentXrayVersion: xray not running after 90s: %s", err)
	}

	// Find an alternative version different from the current one.
	var altVersion string
	for _, v := range versions {
		if v != currentVersion {
			altVersion = v
			break
		}
	}
	if altVersion == "" {
		t.Skip("only one Xray version available, cannot test drift")
	}

	// Pre-flight: verify InstallXray works.
	if err := client.InstallXray(ctx, currentVersion); err != nil {
		t.Skipf("InstallXray not available in this environment: %s", err)
	}

	config := testAccProviderConfig() + fmt.Sprintf(`
resource "threexui_xray_version" "test" {
  version = %q
}
`, currentVersion)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: create the resource at currentVersion.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_version.test", "version", currentVersion),
					resource.TestCheckResourceAttr("threexui_xray_version.test", "current_version", currentVersion),
				),
			},
			// Step 2: simulate external drift by installing altVersion via API,
			// then run a plan-only step expecting drift to be detected.
			{
				PreConfig: func() {
					const maxInstallAttempts = 3
					const maxPollAttempts = 20
					const maxBackoff = 8 * time.Second

					for installAttempt := 0; installAttempt < maxInstallAttempts; installAttempt++ {
						if err := client.InstallXray(ctx, altVersion); err != nil {
							t.Fatalf("failed to simulate drift by installing %s: %s", altVersion, err)
						}

						pollBackoff := 2 * time.Second
						for pollAttempt := 0; pollAttempt < maxPollAttempts; pollAttempt++ {
							cur, err := client.GetCurrentXrayVersion(ctx)
							if err == nil && cur == altVersion {
								t.Logf("drift simulated: version changed to %s (install %d, poll %d)",
									altVersion, installAttempt+1, pollAttempt+1)
								return
							}
							select {
							case <-ctx.Done():
								t.Fatalf("context cancelled during drift poll: %s", ctx.Err())
							case <-time.After(pollBackoff):
							}
							pollBackoff *= 2
							if pollBackoff > maxBackoff {
								pollBackoff = maxBackoff
							}
						}

						if installAttempt < maxInstallAttempts-1 {
							t.Logf("InstallXray(%s) accepted but version unchanged after %d polls — re-issuing install (attempt %d/%d)",
								altVersion, maxPollAttempts, installAttempt+2, maxInstallAttempts)
						}
					}
					t.Fatalf("InstallXray(%s) returned success but version did not change after %d install attempts x %d polls",
						altVersion, maxInstallAttempts, maxPollAttempts)
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Step 3: apply should restore the desired version.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_xray_version.test", "version", currentVersion),
					resource.TestCheckResourceAttr("threexui_xray_version.test", "current_version", currentVersion),
				),
			},
		},
	})
}
