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
// The test creates a fake 3x-ui server with an artificial delay between
// Get and Update. Two goroutines simultaneously apply different settings
// keys. Without serialization, the second writer would overwrite the
// first writer's changes (lost update). With settingsMu, both keys
// survive.
func TestSettingsMu_NoConcurrentReadModifyWrite(t *testing.T) {
	var mu sync.Mutex
	state := map[string]any{} // simulated server-side settings

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
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
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			mu.Lock()
			for k, v := range body {
				state[k] = v
			}
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

	// Writer 1: applyPanelGeneral sets pageSize.
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

	// Writer 2: settingsApplyTyped sets remarkModel.
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

	// Both keys must be present — no lost update.
	if state["pageSize"] != float64(50) {
		t.Errorf("pageSize lost: got %v", state["pageSize"])
	}
	if state["remarkModel"] != "-ieo" {
		t.Errorf("remarkModel lost: got %v", state["remarkModel"])
	}
}

// TestXrayTemplateMu_NoConcurrentReadModifyWrite verifies that
// xrayTemplateMu serializes concurrent xray template updates from
// SetXrayOutboundTestURL (called by applyPanelGeneral) and
// xrayApplyTyped, preventing lost updates.
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
			var parsed map[string]any
			json.Unmarshal([]byte(xraySetting), &parsed)
			mu.Lock()
			for k, v := range parsed {
				xrayState[k] = v
			}
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

	// outboundTestUrl must be set.
	if outboundTestURL.Load().(string) != "https://test.example.com/generate_204" {
		t.Errorf("outboundTestUrl lost: got %v", outboundTestURL.Load())
	}
}
