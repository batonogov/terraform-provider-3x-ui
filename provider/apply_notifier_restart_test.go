package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// TestSettingsApplyTyped_RestartsForNotifierCronKeys is the end-to-end guard for
// #449.
//
// panel_telegram, panel_email and panel_security share settingsApplyTyped, which
// had no restart path at all — unlike panel_general and panel_subscription,
// which have had one since #291. But the panel decides ONCE, in
// web.Server.Start(), whether to register the periodic stats report and the CPU
// and memory alarm jobs, and on what schedule (internal/web/web.go:387-412,
// :439-448, :469-478). Changing any of the settings those decisions read
// therefore applied to the database and to Terraform state while the running
// panel kept the old schedule — or never started the job at all.
//
// The unit tests on panelSettingsNeedRestart protect the key LIST; this protects
// the "a restart is actually issued" contract, and that panel_security — which
// shares the same code path but owns no startup-read key — stays quiet.
func TestSettingsApplyTyped_RestartsForNotifierCronKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		existing    map[string]any
		desired     map[string]any
		wantRestart int32
	}{
		{
			name:        "stats-notify schedule (tgRunTime)",
			existing:    map[string]any{"tgRunTime": "@daily"},
			desired:     map[string]any{"tgRunTime": "@every 6h"},
			wantRestart: 1,
		},
		{
			name:        "enabling the bot registers its cron jobs",
			existing:    map[string]any{"tgBotEnable": false},
			desired:     map[string]any{"tgBotEnable": true},
			wantRestart: 1,
		},
		{
			name:        "alarm event subscription (tgEnabledEvents)",
			existing:    map[string]any{"tgEnabledEvents": "client.expiry"},
			desired:     map[string]any{"tgEnabledEvents": "client.expiry,cpu.high"},
			wantRestart: 1,
		},
		{
			name:        "CPU alarm threshold decides whether the job exists",
			existing:    map[string]any{"tgCpu": float64(0)},
			desired:     map[string]any{"tgCpu": 80},
			wantRestart: 1,
		},
		{
			name:        "memory alarm threshold",
			existing:    map[string]any{"tgMemory": float64(0)},
			desired:     map[string]any{"tgMemory": 90},
			wantRestart: 1,
		},
		{
			name:        "SMTP notifier can want the same alarms",
			existing:    map[string]any{"smtpEnable": false},
			desired:     map[string]any{"smtpEnable": true},
			wantRestart: 1,
		},
		{
			name:        "SMTP alarm thresholds",
			existing:    map[string]any{"smtpCpu": float64(0), "smtpMemory": float64(0)},
			desired:     map[string]any{"smtpCpu": 85, "smtpMemory": 85},
			wantRestart: 1,
		},
		{
			name:        "SMTP event subscription",
			existing:    map[string]any{"smtpEnabledEvents": ""},
			desired:     map[string]any{"smtpEnabledEvents": "memory.high"},
			wantRestart: 1,
		},
		{
			// The bot process itself is hot-reloaded by the panel, so credentials
			// must not bounce it (controller/setting.go:165-172 → web.go:650).
			name:        "bot credentials are hot-reloaded, no restart",
			existing:    map[string]any{"tgBotToken": "old", "tgBotChatId": "1", "tgBotAPIServer": ""},
			desired:     map[string]any{"tgBotToken": "new", "tgBotChatId": "2", "tgBotAPIServer": "https://api.example"},
			wantRestart: 0,
		},
		{
			// panel_security shares this code path and owns no startup-read key.
			name:        "panel_security keys do not restart",
			existing:    map[string]any{"twoFactorEnable": false},
			desired:     map[string]any{"twoFactorEnable": true},
			wantRestart: 0,
		},
		{
			// SMTP transport settings are read per send, not at startup.
			name:        "SMTP transport settings do not restart",
			existing:    map[string]any{"smtpHost": "old.example", "smtpPort": float64(587)},
			desired:     map[string]any{"smtpHost": "new.example", "smtpPort": 465},
			wantRestart: 0,
		},
		{
			// Re-tuning a threshold that is already on changes nothing about cron
			// registration — the job publishes a raw metric and the notifier does the
			// comparison per event, so the new value is live immediately.
			name:        "re-tuning an active threshold does not restart",
			existing:    map[string]any{"tgCpu": float64(80), "smtpMemory": float64(70)},
			desired:     map[string]any{"tgCpu": 90, "smtpMemory": 75},
			wantRestart: 0,
		},
		{
			name:        "switching a threshold off deregisters the job",
			existing:    map[string]any{"tgCpu": float64(80)},
			desired:     map[string]any{"tgCpu": 0},
			wantRestart: 1,
		},
		{
			// Only cpu.high / memory.high membership gates a job; the rest is
			// filtered at delivery.
			name:        "adding an unrelated event does not restart",
			existing:    map[string]any{"tgEnabledEvents": "cpu.high,login.attempt"},
			desired:     map[string]any{"tgEnabledEvents": "cpu.high,login.attempt,backup"},
			wantRestart: 0,
		},
		{
			name:        "dropping cpu.high deregisters the sampler",
			existing:    map[string]any{"smtpEnabledEvents": "memory.high,cpu.high"},
			desired:     map[string]any{"smtpEnabledEvents": "memory.high"},
			wantRestart: 1,
		},
		{
			// A panel predating the key drops it on write, so restarting would
			// bounce it on every apply for a value that is never stored.
			name:        "key unknown to the panel does not restart",
			existing:    map[string]any{"tgBotEnable": true},
			desired:     map[string]any{"smtpCpu": 80, "smtpEnabledEvents": "cpu.high"},
			wantRestart: 0,
		},
		{
			name:        "re-applying identical values stays quiet",
			existing:    map[string]any{"tgRunTime": "@daily", "tgCpu": float64(80)},
			desired:     map[string]any{"tgRunTime": "@daily", "tgCpu": 80},
			wantRestart: 0,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var restarts int32
			mu := sync.Mutex{}
			server := make(map[string]any, len(c.existing))
			for k, v := range c.existing {
				server[k] = v
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/login":
					_, _ = w.Write(okResponse(nil))
				case "/panel/api/setting/all":
					mu.Lock()
					snapshot := make(map[string]any, len(server))
					for k, v := range server {
						snapshot[k] = v
					}
					mu.Unlock()
					_, _ = w.Write(okResponse(snapshot))
				case "/panel/api/setting/update":
					var in map[string]any
					_ = json.NewDecoder(r.Body).Decode(&in)
					mu.Lock()
					for k, v := range in {
						server[k] = v
					}
					mu.Unlock()
					_, _ = w.Write(okResponse(nil))
				case "/panel/api/setting/restartPanel":
					atomic.AddInt32(&restarts, 1)
					_, _ = w.Write(okResponse(nil))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			client := newTestClient(t, srv.URL)
			client.settingsAPIMu.Lock()
			v := true
			client.settingsUnderAPI = &v
			client.settingsAPIMu.Unlock()

			var d diag.Diagnostics
			settingsApplyTyped(context.Background(), c.desired, &d, client)

			if d.HasError() {
				t.Fatalf("settingsApplyTyped returned errors: %v", d.Errors())
			}
			if got := atomic.LoadInt32(&restarts); got != c.wantRestart {
				t.Fatalf("restartPanel calls = %d, want %d", got, c.wantRestart)
			}
		})
	}
}

// A panel that refuses the restart must surface as an error, not as a silent
// success: the settings were written, so the practitioner has to know the panel
// is still running the old wiring.
func TestSettingsApplyTyped_RestartFailureSurfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			_, _ = w.Write(okResponse(nil))
		case "/panel/api/setting/all":
			_, _ = w.Write(okResponse(map[string]any{"tgRunTime": "@daily"}))
		case "/panel/api/setting/update":
			_, _ = w.Write(okResponse(nil))
		case "/panel/api/setting/restartPanel":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	client.settingsAPIMu.Lock()
	v := true
	client.settingsUnderAPI = &v
	client.settingsAPIMu.Unlock()

	var d diag.Diagnostics
	settingsApplyTyped(context.Background(), map[string]any{"tgRunTime": "@every 6h"}, &d, client)

	if !d.HasError() {
		t.Fatal("a refused restart must be reported")
	}
}

// The event-list comparison reads whatever the panel returned, which is not
// guaranteed to be a string on a malformed response.
func TestEventListContains_NonString(t *testing.T) {
	if eventListContains(42, "cpu.high") {
		t.Error("a non-string value cannot contain an event")
	}
	if eventListContains(nil, "cpu.high") {
		t.Error("a nil value cannot contain an event")
	}
	if !eventListContains("memory.high,cpu.high", "cpu.high") {
		t.Error("a plain list must still match")
	}
}

// The remaining error paths through settingsApplyTyped. Each has to report
// rather than fail silently, and — since settingsMu is now unlocked by hand
// rather than by defer — each has to leave the lock released, which the
// subsequent call in this test would otherwise hang on.
func TestSettingsApplyTyped_ErrorPaths(t *testing.T) {
	newClient := func(t *testing.T, handler http.HandlerFunc) *Client {
		t.Helper()
		srv := httptest.NewServer(handler)
		t.Cleanup(srv.Close)
		client := newTestClient(t, srv.URL)
		client.settingsAPIMu.Lock()
		v := true
		client.settingsUnderAPI = &v
		client.settingsAPIMu.Unlock()
		return client
	}

	t.Run("unreadable settings", func(t *testing.T) {
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/login" {
				_, _ = w.Write(okResponse(nil))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		})

		var d diag.Diagnostics
		settingsApplyTyped(context.Background(), map[string]any{"tgRunTime": "@daily"}, &d, client)
		if !d.HasError() {
			t.Fatal("a failed settings read must be reported")
		}
	})

	t.Run("rejected update", func(t *testing.T) {
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/login":
				_, _ = w.Write(okResponse(nil))
			case "/panel/api/setting/all":
				_, _ = w.Write(okResponse(map[string]any{"tgRunTime": "@daily"}))
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})

		var d diag.Diagnostics
		settingsApplyTyped(context.Background(), map[string]any{"tgRunTime": "@every 6h"}, &d, client)
		if !d.HasError() {
			t.Fatal("a rejected update must be reported")
		}
	})

	t.Run("panel never comes back", func(t *testing.T) {
		// The restart is accepted, then the panel stops answering. WaitForReady
		// honours the context, so a cancelled one stands in for its 30s timeout.
		client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/login":
				_, _ = w.Write(okResponse(nil))
			case "/panel/api/setting/all":
				_, _ = w.Write(okResponse(map[string]any{"tgRunTime": "@daily"}))
			case "/panel/api/setting/update", "/panel/api/setting/restartPanel":
				_, _ = w.Write(okResponse(nil))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		var d diag.Diagnostics
		// Cancel once the restart has been accepted: Login is what WaitForReady
		// polls, and the cancelled context makes it give up immediately.
		cancel()
		settingsApplyTyped(ctx, map[string]any{"tgRunTime": "@every 6h"}, &d, client)
		if !d.HasError() {
			t.Fatal("a panel that never comes back must be reported")
		}
	})
}
