package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSettingsMu_NoConcurrentReadModifyWrite verifies that settingsMu
// serializes concurrent settings updates from applyPanelGeneral and
// settingsApplyTyped, preventing lost updates.
//
// The fake server replaces the entire settings map on each update (as
// the real 3x-ui API does). Without settingsMu, the second writer reads
// a stale snapshot and overwrites the first writer's changes.
func TestSettingsMu_NoConcurrentReadModifyWrite(t *testing.T) {
	var mu sync.Mutex
	// Pre-seed with baseline keys so both writers include them in their
	// read-modify-write cycle. A writer that reads a stale snapshot will
	// post back stale values for the key the other writer changed.
	state := map[string]any{
		"pageSize":    float64(25),
		"remarkModel": "-io",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/setting/all":
			// Probe from useSettingsAPI: return 404 → legacy paths.
			w.WriteHeader(http.StatusNotFound)
		case "/panel/setting/all":
			mu.Lock()
			snapshot := make(map[string]any, len(state))
			for k, v := range state {
				snapshot[k] = v
			}
			mu.Unlock()
			// Artificial delay to widen the race window.
			time.Sleep(50 * time.Millisecond)
			w.Write(okResponse(snapshot))
		case "/panel/setting/update":
			// Real API replaces settings entirely — do the same here.
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			state = body
			mu.Unlock()
			w.Write(okResponse(nil))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Writer 1: applyPanelGeneral sets pageSize to 50.
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := &PanelGeneralResource{client: client}
		plan := &PanelGeneralModel{
			PageSize: types.Int64Value(50),
		}
		var d diag.Diagnostics
		r.applyPanelGeneral(ctx, plan, &d)
		if d.HasError() {
			t.Errorf("applyPanelGeneral: %s", d.Errors()[0].Detail())
		}
	}()

	// Writer 2: settingsApplyTyped sets remarkModel to "-ieo".
	wg.Add(1)
	go func() {
		defer wg.Done()
		desired := map[string]any{"remarkModel": "-ieo"}
		var d diag.Diagnostics
		settingsApplyTyped(ctx, desired, &d, client)
		if d.HasError() {
			t.Errorf("settingsApplyTyped: %s", d.Errors()[0].Detail())
		}
	}()

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	// Both keys must reflect the latest writes — no lost update.
	if state["pageSize"] != float64(50) {
		t.Errorf("pageSize lost: got %v, want 50", state["pageSize"])
	}
	if state["remarkModel"] != "-ieo" {
		t.Errorf("remarkModel lost: got %v, want -ieo", state["remarkModel"])
	}
}

func TestApplyPanelGeneralPreservesCachedRedactedSettingSecrets(t *testing.T) {
	var updateBody map[string]any
	state := map[string]any{
		"pageSize":       float64(25),
		"tgBotToken":     "",
		"twoFactorToken": "********",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/setting/all":
			// Probe from useSettingsAPI: return 404 → legacy paths.
			w.WriteHeader(http.StatusNotFound)
		case "/panel/setting/all":
			w.Write(okResponse(state))
		case "/panel/setting/update":
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Errorf("decode update body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			state = updateBody
			w.Write(okResponse(nil))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	client.rememberConfiguredSettingSecrets(map[string]any{
		"tgBotToken":     "telegram-token",
		"twoFactorToken": "totp-secret",
	})

	r := &PanelGeneralResource{client: client}
	plan := &PanelGeneralModel{PageSize: types.Int64Value(50)}
	var d diag.Diagnostics
	r.applyPanelGeneral(context.Background(), plan, &d)
	if d.HasError() {
		t.Fatalf("applyPanelGeneral: %s", d.Errors()[0].Detail())
	}

	if updateBody["pageSize"] != float64(50) {
		t.Fatalf("pageSize = %v, want 50", updateBody["pageSize"])
	}
	if updateBody["tgBotToken"] != "telegram-token" {
		t.Fatalf("tgBotToken = %v, want cached token", updateBody["tgBotToken"])
	}
	if updateBody["twoFactorToken"] != "totp-secret" {
		t.Fatalf("twoFactorToken = %v, want cached token", updateBody["twoFactorToken"])
	}
}

func TestSettingsApplyTypedAllowsExplicitSecretClear(t *testing.T) {
	var updateBody map[string]any
	state := map[string]any{
		"remarkModel": "-io",
		"tgBotToken":  "",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/setting/all":
			// Probe from useSettingsAPI: return 404 → legacy paths.
			w.WriteHeader(http.StatusNotFound)
		case "/panel/setting/all":
			w.Write(okResponse(state))
		case "/panel/setting/update":
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Errorf("decode update body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			state = updateBody
			w.Write(okResponse(nil))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	client.rememberConfiguredSettingSecrets(map[string]any{"tgBotToken": "telegram-token"})

	var d diag.Diagnostics
	settingsApplyTyped(context.Background(), map[string]any{"tgBotToken": ""}, &d, client)
	if d.HasError() {
		t.Fatalf("settingsApplyTyped: %s", d.Errors()[0].Detail())
	}

	if updateBody["tgBotToken"] != "" {
		t.Fatalf("tgBotToken = %v, want explicit empty value", updateBody["tgBotToken"])
	}
}

// TestXrayTemplateMu_NoConcurrentReadModifyWrite verifies that
// xrayTemplateMu serializes concurrent xray template updates from
// SetXrayOutboundTestURL (called by applyPanelGeneral) and
// xrayApplyTyped, preventing lost updates.
//
// The fake server replaces xrayState entirely on update (as the real
// API does). Without xrayTemplateMu, a stale write from one goroutine
// would drop the other's changes.
func TestXrayTemplateMu_NoConcurrentReadModifyWrite(t *testing.T) {
	var mu sync.Mutex
	xrayState := map[string]any{
		"log":      map[string]any{"loglevel": "warning"},
		"inbounds": []any{},
	}
	var outboundTestURL atomic.Value
	outboundTestURL.Store("")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
		case "/panel/api/setting/all":
			// Probe from useSettingsAPI: return 404 so the client
			// falls back to legacy /panel/xray/* paths.
			w.WriteHeader(http.StatusNotFound)
		case "/panel/xray":
			mu.Lock()
			snapshot := make(map[string]any, len(xrayState))
			for k, v := range xrayState {
				snapshot[k] = v
			}
			mu.Unlock()
			// Artificial delay to widen the race window.
			time.Sleep(50 * time.Millisecond)
			// The endpoint wraps xraySetting + outboundTestUrl.
			wrapper := map[string]any{
				"xraySetting":     snapshot,
				"outboundTestUrl": outboundTestURL.Load().(string),
			}
			raw, _ := json.Marshal(wrapper)
			w.Write(okResponse(string(raw)))
		case "/panel/xray/update":
			r.ParseForm()
			xraySetting := r.FormValue("xraySetting")
			testURL := r.FormValue("outboundTestUrl")
			// Real API replaces the entire xray template — do the same.
			var parsed map[string]any
			json.Unmarshal([]byte(xraySetting), &parsed)
			mu.Lock()
			xrayState = parsed
			mu.Unlock()
			if testURL != "" {
				outboundTestURL.Store(testURL)
			}
			w.Write(okResponse(nil))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Writer 1: applyPanelGeneral sets xray_outbound_test_url.
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := &PanelGeneralResource{client: client}
		plan := &PanelGeneralModel{
			XrayOutboundTestURL: types.StringValue("https://test.example.com/generate_204"),
		}
		var d diag.Diagnostics
		r.applyPanelGeneral(ctx, plan, &d)
		if d.HasError() {
			t.Errorf("applyPanelGeneral xray: %s", d.Errors()[0].Detail())
		}
	}()

	// Writer 2: xrayApplyTyped sets dns section.
	wg.Add(1)
	go func() {
		defer wg.Done()
		desired := map[string]any{
			"dns": map[string]any{
				"servers": []any{"8.8.8.8"},
			},
		}
		var d diag.Diagnostics
		xrayApplyTyped(ctx, desired, &d, client, xraySectionDNS)
		if d.HasError() {
			t.Errorf("xrayApplyTyped: %s", d.Errors()[0].Detail())
		}
	}()

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	// The log section must survive — no lost update from either writer.
	if _, ok := xrayState["log"]; !ok {
		t.Errorf("log section lost from xray state")
	}

	// The dns section must survive — no lost update from SetXrayOutboundTestURL.
	if _, ok := xrayState["dns"]; !ok {
		t.Errorf("dns section lost from xray state")
	}

	// outboundTestUrl must be set.
	if outboundTestURL.Load().(string) != "https://test.example.com/generate_204" {
		t.Errorf("outboundTestUrl lost: got %v", outboundTestURL.Load())
	}
}
