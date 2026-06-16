package main

import (
	"strings"
	"testing"
)

// TestFixtureFor locks down the (host, path) → fixture-file mapping that the
// proxy serves over its MITM TLS sessions. These are the only URLs 3x-ui's
// xray-version code paths hit, so a change here must be intentional.
func TestFixtureFor(t *testing.T) {
	const root = "/fixtures"

	tests := []struct {
		name     string
		host     string
		path     string
		wantAbs  string
		wantType string
		wantOK   bool
	}{
		{
			name:     "releases list (exact path)",
			host:     "api.github.com",
			path:     "/repos/XTLS/Xray-core/releases",
			wantAbs:  "/fixtures/api.github.com/repos/XTLS/Xray-core/releases.json",
			wantType: "application/json; charset=utf-8",
			wantOK:   true,
		},
		{
			// 3x-ui does not add a query, but be lenient if it ever does — the
			// path prefix is what we match, the query is ignored by ServeHTTP.
			name:     "releases list (with query path prefix still matches)",
			host:     "api.github.com",
			path:     "/repos/XTLS/Xray-core/releases", // query handled by net/http, not here
			wantAbs:  "/fixtures/api.github.com/repos/XTLS/Xray-core/releases.json",
			wantType: "application/json; charset=utf-8",
			wantOK:   true,
		},
		{
			name:     "xray zip download v26.6.1",
			host:     "github.com",
			path:     "/XTLS/Xray-core/releases/download/v26.6.1/Xray-linux-64.zip",
			wantAbs:  "/fixtures/github.com/XTLS/Xray-core/releases/download/v26.6.1/Xray-linux-64.zip",
			wantType: "application/zip",
			wantOK:   true,
		},
		{
			name:     "xray zip download v26.5.9",
			host:     "github.com",
			path:     "/XTLS/Xray-core/releases/download/v26.5.9/Xray-linux-64.zip",
			wantAbs:  "/fixtures/github.com/XTLS/Xray-core/releases/download/v26.5.9/Xray-linux-64.zip",
			wantType: "application/zip",
			wantOK:   true,
		},
		{
			name:   "unknown host → 404 (fail loud)",
			host:   "api.telegram.org",
			path:   "/bot123:abc/getMe",
			wantOK: false,
		},
		{
			name:   "github.com but unrelated path → 404",
			host:   "github.com",
			path:   "/someorg/somerepo",
			wantOK: false,
		},
		{
			name:   "api.github.com different repo → 404",
			host:   "api.github.com",
			path:   "/repos/other/core/releases",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, ct, ok := fixtureFor(root, tt.host, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("fixtureFor(%q,%q) ok = %v, want %v (abs=%q)", tt.host, tt.path, ok, tt.wantOK, abs)
			}
			if !tt.wantOK {
				return
			}
			if abs != tt.wantAbs {
				t.Errorf("abs = %q, want %q", abs, tt.wantAbs)
			}
			if ct != tt.wantType {
				t.Errorf("contentType = %q, want %q", ct, tt.wantType)
			}
		})
	}
}

// TestFixtureForPathTraversal guards against crafted GitHub download URLs that
// match the routing prefix but try to escape the fixture root via ".."
// segments. Without the prefix check in fixtureFor, filepath.Join would resolve
// such a path outside /app/fixtures — and since the MITM CA private key lives
// at /ca/ca.key (a sibling mounted volume), an escape would leak the CA the
// panel trusts. Regression test for the traversal finding in PR #286 review.
//
// The security invariant under test is: a resolved path MUST NEVER leave root,
// regardless of ".." content. Paths that stay inside root but point at a
// non-existent file are fine (they 404 at os.Open); only an actual escape is
// a vulnerability.
func TestFixtureForPathTraversal(t *testing.T) {
	const root = "/app/fixtures"
	// Each entry must either be rejected (ok=false) or resolve strictly inside root.
	cases := []string{
		// Real escapes (unencoded "..") — these would reach /etc or /ca if not guarded.
		"/XTLS/Xray-core/releases/download/../../../../../../etc/passwd",
		"/XTLS/Xray-core/releases/download/../../../../../../../ca/ca.key",
		// Traversal that collapses back INSIDE root (non-existent file → 404, but safe).
		"/XTLS/Xray-core/releases/download/v1/../../../etc/shadow",
		// Percent-encoded separators are NOT decoded by url.URL, so they stay literal
		// (a harmless, non-existent filename inside root).
		"/XTLS/Xray-core/releases/download/..%2f..%2f..%2fca/ca.key",
	}
	for _, p := range cases {
		abs, _, ok := fixtureFor(root, "github.com", p)
		escaped := abs != "" && abs != root && !strings.HasPrefix(abs, root+"/")
		t.Logf("path=%-70s ok=%-5v resolved=%s", p, ok, abs)
		if escaped {
			t.Errorf("ESCAPE: %q resolved outside root to %q", p, abs)
		}
	}

	// A legitimate path must still resolve after the hardening.
	abs, _, ok := fixtureFor(root, "github.com",
		"/XTLS/Xray-core/releases/download/v26.6.1/Xray-linux-64.zip")
	if !ok {
		t.Fatalf("legitimate path wrongly rejected")
	}
	want := "/app/fixtures/github.com/XTLS/Xray-core/releases/download/v26.6.1/Xray-linux-64.zip"
	if abs != want {
		t.Errorf("legitimate path = %q, want %q", abs, want)
	}
}
