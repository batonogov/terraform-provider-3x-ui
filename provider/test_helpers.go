package provider

import (
	"os"
	"testing"

	"golang.org/x/mod/semver"
)

// requireMinVersion skips the test if THREEXUI_VERSION is set and is older than min.
// Both values must use "v" prefix (e.g. "v2.9.0").
func requireMinVersion(t skipFatalHelper, min string) {
	t.Helper()

	v := os.Getenv("THREEXUI_VERSION")
	if v == "" {
		return // no version constraint, run the test
	}

	if !semver.IsValid(v) {
		return // can't parse, don't skip
	}

	if !semver.IsValid(min) {
		t.Fatal("invalid min version " + min)
	}

	if semver.Compare(v, min) < 0 {
		t.Skipf("requires 3x-ui >= %s, running %s", min, v)
	}
}

// requireBelowVersion skips the test if THREEXUI_VERSION is set and is >= max.
// Use for protocol features removed in a specific 3x-ui version.
// Both values must use "v" prefix (e.g. "v3.2.0").
func requireBelowVersion(t skipFatalHelper, max string) {
	t.Helper()

	v := os.Getenv("THREEXUI_VERSION")
	if v == "" {
		return // no version constraint, run the test
	}

	if !semver.IsValid(v) {
		return // can't parse, don't skip
	}

	if !semver.IsValid(max) {
		t.Fatal("invalid max version " + max)
	}

	if semver.Compare(v, max) >= 0 {
		t.Skipf("removed in 3x-ui >= %s, running %s", max, v)
	}
}

// skipFatalHelper is the narrow subset of *testing.T this package's
// version-gating helpers depend on. Defined as an interface so the
// helpers can be unit-tested with a fake — testing.TB is sealed and
// cannot be implemented outside the standard library.
type skipFatalHelper interface {
	Helper()
	Skipf(format string, args ...any)
	Fatal(args ...any)
}

// skipOnFlakyVersions skips the test if THREEXUI_VERSION matches any of
// the listed versions. Use for tests that fail deterministically on a
// specific upstream panel version due to a known upstream bug — not for
// transient flakes (use skipIfFlaky for those). `reason` should reference
// a tracking issue so the gate is removed once upstream is fixed.
func skipOnFlakyVersions(t skipFatalHelper, reason string, versions ...string) {
	t.Helper()
	if len(versions) == 0 {
		t.Fatal("skipOnFlakyVersions requires at least one version")
	}
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

// needs sub-day quarantine (see CLAUDE.md -> Testing). Removing
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
