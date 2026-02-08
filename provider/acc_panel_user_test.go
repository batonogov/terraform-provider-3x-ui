package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPanelUser verifies the resource lifecycle (create, update, idempotency)
// by setting the same credentials that the provider is already using.
// This avoids the issue where the testing framework creates a new provider
// instance for the post-apply plan check and can't login with stale creds.
func TestAccPanelUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create: set admin/admin (same as default).
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_user" "test" {
  username = "admin"
  password = "admin"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("threexui_panel_user.test", "id", "user"),
					resource.TestCheckResourceAttr("threexui_panel_user.test", "username", "admin"),
				),
			},
			// Idempotency.
			{
				Config: testAccProviderConfig() + `
resource "threexui_panel_user" "test" {
  username = "admin"
  password = "admin"
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccPanelUserCredentialChange is a Go-level integration test that verifies
// actual credential changes via the UpdateUser API. It changes the admin
// credentials and then restores them.
func TestAccPanelUserCredentialChange(t *testing.T) {
	testAccPreCheck(t)

	client, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("client init: %s", err)
	}

	ctx := context.Background()

	// Change admin/admin → testuser/testpass.
	if err := client.UpdateUser(ctx, "admin", "admin", "testuser", "testpass"); err != nil {
		t.Fatalf("UpdateUser (change): %s", err)
	}

	// Verify new creds work by logging in.
	if err := client.Login(ctx); err != nil {
		// Try to restore before failing.
		_ = restoreCredentials(ctx, client)
		t.Fatalf("Login with new creds failed: %s", err)
	}

	// Verify old creds no longer work.
	oldClient, err := testAccClientFromEnvNoLogin()
	if err != nil {
		_ = restoreCredentials(ctx, client)
		t.Fatalf("client init (old): %s", err)
	}
	if err := oldClient.Login(ctx); err == nil {
		// Old creds should fail. Restore and fail.
		_ = restoreCredentials(ctx, client)
		t.Fatal("Login with old creds should have failed but succeeded")
	}

	// Restore: testuser/testpass → admin/admin.
	if err := client.UpdateUser(ctx, "testuser", "testpass", "admin", "admin"); err != nil {
		t.Fatalf("UpdateUser (restore): %s", err)
	}

	// Verify restored creds work.
	if err := client.Login(ctx); err != nil {
		t.Fatalf("Login after restore failed: %s", err)
	}
}

func restoreCredentials(ctx context.Context, client *Client) error {
	return client.UpdateUser(ctx, "testuser", "testpass", "admin", "admin")
}
