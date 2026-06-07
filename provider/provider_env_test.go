package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvString_TFValue(t *testing.T) {
	got := envString(types.StringValue("from-hcl"), "THREEXUI_UNUSED_TEST", "fallback")
	if got != "from-hcl" {
		t.Fatalf("expected from-hcl, got %q", got)
	}
}

func TestEnvString_EnvFallback(t *testing.T) {
	t.Setenv("THREEXUI_TEST_STRING", "from-env")
	got := envString(types.StringNull(), "THREEXUI_TEST_STRING", "fallback")
	if got != "from-env" {
		t.Fatalf("expected from-env, got %q", got)
	}
}

func TestEnvString_Default(t *testing.T) {
	got := envString(types.StringNull(), "THREEXUI_UNUSED_TEST", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestEnvString_TFOverrulesEnv(t *testing.T) {
	t.Setenv("THREEXUI_TEST_STRING", "from-env")
	got := envString(types.StringValue("from-hcl"), "THREEXUI_TEST_STRING", "fallback")
	if got != "from-hcl" {
		t.Fatalf("expected from-hcl (HCL takes precedence), got %q", got)
	}
}

func TestEnvString_UnknownUsesEnv(t *testing.T) {
	t.Setenv("THREEXUI_TEST_STRING", "from-env")
	got := envString(types.StringUnknown(), "THREEXUI_TEST_STRING", "fallback")
	if got != "from-env" {
		t.Fatalf("expected from-env, got %q", got)
	}
}

func TestEnvConstants(t *testing.T) {
	want := map[string]string{
		"THREEXUI_ENDPOINT":             envEndpoint,
		"THREEXUI_BASE_PATH":            envBasePath,
		"THREEXUI_USERNAME":             envUsername,
		"THREEXUI_PASSWORD":             envPassword,
		"THREEXUI_INSECURE_SKIP_VERIFY": envInsecureSkipVerify,
		"THREEXUI_REQUEST_TIMEOUT":      envRequestTimeout,
		"THREEXUI_MAX_RETRIES":          envMaxRetries,
	}
	for expected, constant := range want {
		if constant != expected {
			t.Errorf("expected %q, got %q", expected, constant)
		}
	}
}

func TestEnvString_EmptyEnvUsesDefault(t *testing.T) {
	t.Setenv("THREEXUI_TEST_STRING", "")
	got := envString(types.StringNull(), "THREEXUI_TEST_STRING", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback (empty env ignored), got %q", got)
	}
}

func TestEndpointRequiredWithoutEnvOrConfig(t *testing.T) {
	// Ensure endpoint error is produced when neither HCL nor env provides it.
	// This is tested via Configure — we just verify envString returns empty.
	got := envString(types.StringNull(), "THREEXUI_ENDPOINT_NOT_SET_XYZ", "")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
