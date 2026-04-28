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

// skipOnFlakyVersions skips the test if THREEXUI_VERSION matches any of
// the listed versions. Use for tests that fail deterministically on a
// specific upstream panel version due to a known upstream bug — not for
// transient flakes (use skipIfFlaky for those). The last argument is the
// reason and should reference a tracking issue so the gate is removed
// once upstream is fixed.
func skipOnFlakyVersions(t *testing.T, versionsAndReason ...string) {
	t.Helper()
	if len(versionsAndReason) < 2 {
		t.Fatal("skipOnFlakyVersions requires at least one version plus a reason")
	}
	reason := versionsAndReason[len(versionsAndReason)-1]
	versions := versionsAndReason[:len(versionsAndReason)-1]
	v := os.Getenv("THREEXUI_VERSION")
	if v == "" {
		return
	}
	for _, bad := range versions {
		if v == bad {
			t.Skipf("known-broken on %s: %s", v, reason)
		}
	}
}

// needs sub-day quarantine (see CLAUDE.md → CI Flake Mitigation). Removing
// it forces every quarantine to start with a Taskfile/CI plumbing change.
//
// skipIfFlaky quarantines a known-flaky test when THREEXUI_SKIP_FLAKY is
// set (any non-empty value). The intent is to give us a sub-day mitigation
// path when a test starts firing falsely — flip the env var in CI, file an
// issue with the failure log, and unblock contributors. Tests gated this
// way must be tracked in #161 (or a follow-up issue) and either fixed or
// removed; the gate is not a permanent home. `reason` is included in the
// skip message so reviewers can see *why* a test is quarantined without
// chasing git blame.
//
//nolint:unused // intentional pre-deployed gate; called only when a test
func skipIfFlaky(t *testing.T, reason string) {
	t.Helper()
	if os.Getenv("THREEXUI_SKIP_FLAKY") == "" {
		return
	}
	t.Skipf("quarantined as flaky (THREEXUI_SKIP_FLAKY set): %s", reason)
}
