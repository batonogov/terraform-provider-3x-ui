package provider

import (
	"os"
	"testing"

	"golang.org/x/mod/semver"
)

// requireMinVersion skips the test if THREEXUI_VERSION is set and is older than min.
// Both values must use "v" prefix (e.g. "v2.9.0").
func requireMinVersion(t *testing.T, min string) {
	t.Helper()

	v := os.Getenv("THREEXUI_VERSION")
	if v == "" {
		return // no version constraint, run the test
	}

	if !semver.IsValid(v) {
		return // can't parse, don't skip
	}

	if !semver.IsValid(min) {
		t.Fatalf("invalid min version %q", min)
	}

	if semver.Compare(v, min) < 0 {
		t.Skipf("requires 3x-ui >= %s, running %s", min, v)
	}
}
