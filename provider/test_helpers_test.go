package provider

import (
	"fmt"
	"testing"
)

// fakeT captures the side-effects of skipFatalHelper assertions so the
// helpers themselves can be unit-tested.
type fakeT struct {
	skipped  bool
	skipMsg  string
	failed   bool
	failMsg  string
	helperOK bool
}

func (f *fakeT) Helper() { f.helperOK = true }
func (f *fakeT) Skipf(format string, args ...any) {
	f.skipped = true
	f.skipMsg = fmt.Sprintf(format, args...)
}
func (f *fakeT) Fatal(args ...any) { f.failed = true; f.failMsg = fmt.Sprint(args...) }

func TestSkipOnFlakyVersions(t *testing.T) {
	t.Run("no env runs the test", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "")
		f := &fakeT{}
		skipOnFlakyVersions(f, "reason", "v2.8.9")
		if f.skipped {
			t.Fatalf("should not skip when THREEXUI_VERSION is empty (msg=%q)", f.skipMsg)
		}
		if !f.helperOK {
			t.Fatal("Helper() should be called")
		}
	})

	t.Run("env matches a listed version skips with reason", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v2.9.1")
		f := &fakeT{}
		skipOnFlakyVersions(f, "tracked in #163", "v2.8.9", "v2.9.1")
		if !f.skipped {
			t.Fatal("should skip when THREEXUI_VERSION matches a listed version")
		}
		if !contains(f.skipMsg, "v2.9.1") || !contains(f.skipMsg, "tracked in #163") {
			t.Fatalf("skip message should include version and reason, got %q", f.skipMsg)
		}
	})

	t.Run("env does not match runs the test", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v2.9.3")
		f := &fakeT{}
		skipOnFlakyVersions(f, "reason", "v2.8.9", "v2.9.1")
		if f.skipped {
			t.Fatalf("should not skip when version does not match (msg=%q)", f.skipMsg)
		}
	})

	t.Run("zero versions is a usage error", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v2.9.0")
		f := &fakeT{}
		skipOnFlakyVersions(f, "reason")
		if !f.failed {
			t.Fatal("should Fatal when no versions are passed")
		}
	})
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRequireBelowVersion(t *testing.T) {
	t.Run("no env runs the test", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "")
		ft := &fakeT{}
		requireBelowVersion(ft, "v3.2.0")
		if ft.skipped {
			t.Fatalf("should not skip when THREEXUI_VERSION is empty (msg=%q)", ft.skipMsg)
		}
	})

	t.Run("version at max skips", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v3.2.0")
		ft := &fakeT{}
		requireBelowVersion(ft, "v3.2.0")
		if !ft.skipped {
			t.Fatal("should skip when version equals max")
		}
	})

	t.Run("version above max skips", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v3.3.0")
		ft := &fakeT{}
		requireBelowVersion(ft, "v3.2.0")
		if !ft.skipped {
			t.Fatal("should skip when version above max")
		}
	})

	t.Run("version below max runs the test", func(t *testing.T) {
		t.Setenv("THREEXUI_VERSION", "v3.1.0")
		ft := &fakeT{}
		requireBelowVersion(ft, "v3.2.0")
		if ft.skipped {
			t.Fatalf("should not skip when version below max (msg=%q)", ft.skipMsg)
		}
	})
}
