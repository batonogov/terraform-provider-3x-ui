package main

import "testing"

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
