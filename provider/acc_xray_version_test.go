package provider

import (
	"context"
	"strings"
	"testing"

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
