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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestApplySubscription_RestartsOnServerKeyChange is the regression guard for #291.
//
// It runs applySubscription against a stub 3x-ui (httptest) that records
// /setting/restartPanel calls, and asserts that:
//
//   - changing a subscription SERVER-BINDING key (subEnable/subListen/subDomain/
//     subPort/subPath/subCertFile/subKeyFile) triggers exactly one panel restart;
//   - changing a LINK-GENERATION key only (subURI) triggers NO restart.
//
// Before the fix, applySubscription never called SendRestart on the subscription
// path, so the subscription server did not rebind and the subscription URL kept
// 404-ing until a manual panel restart (#291). The unit tests on
// panelSettingsNeedRestart protect the key LIST; this test protects the
// end-to-end "restart is actually invoked" contract.
func TestApplySubscription_RestartsOnServerKeyChange(t *testing.T) {
	t.Parallel()

	type tc struct {
		name        string
		existing    map[string]any
		plan        PanelSubscriptionModel
		wantRestart int32
	}

	// Build a plan model from a partial settings payload. Only the keys present
	// in the map are set on the model; everything else stays null/unknown so
	// expandPanelSubscription emits exactly those keys.
	mkPlan := func(payload map[string]any) PanelSubscriptionModel {
		m := PanelSubscriptionModel{}
		if v, ok := payload["subEnable"]; ok {
			m.SubEnable = types.BoolValue(v.(bool))
		}
		if v, ok := payload["subEnableRouting"]; ok {
			m.SubEnableRouting = types.BoolValue(v.(bool))
		}
		if v, ok := payload["subListen"]; ok {
			m.SubListen = types.StringValue(v.(string))
		}
		if v, ok := payload["subPort"]; ok {
			m.SubPort = types.Int64Value(int64(v.(int)))
		}
		if v, ok := payload["subPath"]; ok {
			m.SubPath = types.StringValue(v.(string))
		}
		if v, ok := payload["subDomain"]; ok {
			m.SubDomain = types.StringValue(v.(string))
		}
		if v, ok := payload["subCertFile"]; ok {
			m.SubCertFile = types.StringValue(v.(string))
		}
		if v, ok := payload["subKeyFile"]; ok {
			m.SubKeyFile = types.StringValue(v.(string))
		}
		if v, ok := payload["subURI"]; ok {
			m.SubURI = types.StringValue(v.(string))
		}
		if v, ok := payload["subJsonMux"]; ok {
			m.SubJsonMux = types.StringValue(v.(string))
		}
		if v, ok := payload["subJsonPath"]; ok {
			m.SubJsonPath = types.StringValue(v.(string))
		}
		if v, ok := payload["subJsonEnable"]; ok {
			m.SubJsonEnable = types.BoolValue(v.(bool))
		}
		if v, ok := payload["subTitle"]; ok {
			m.SubTitle = types.StringValue(v.(string))
		}
		return m
	}

	cases := []tc{
		{
			// /setting/all serialises AllSetting whole, so a real panel always
			// reports subEnable — with its default, false, before the first apply.
			// (An empty `existing` would mean a panel that does not have the key at
			// all, which is a different case: see
			// TestPanelSettingsNeedRestart_KeyUnknownToPanel.)
			name:        "enable subscription server (first apply)",
			existing:    map[string]any{"subEnable": false, "subPort": float64(2096), "subPath": "/sub/"},
			plan:        mkPlan(map[string]any{"subEnable": true, "subPort": 2096, "subPath": "/sub/"}),
			wantRestart: 1,
		},
		{
			name:        "change subscription port",
			existing:    map[string]any{"subEnable": true, "subPort": float64(2096)},
			plan:        mkPlan(map[string]any{"subPort": 2097}),
			wantRestart: 1,
		},
		{
			name:        "change subscription path",
			existing:    map[string]any{"subPath": "/sub/"},
			plan:        mkPlan(map[string]any{"subPath": "/sub2/"}),
			wantRestart: 1,
		},
		{
			name:        "change subscription listen address",
			existing:    map[string]any{"subEnable": true, "subListen": "0.0.0.0"},
			plan:        mkPlan(map[string]any{"subListen": "127.0.0.1"}),
			wantRestart: 1,
		},
		{
			name:        "disable subscription server",
			existing:    map[string]any{"subEnable": true},
			plan:        mkPlan(map[string]any{"subEnable": false}),
			wantRestart: 1,
		},
		{
			// #443: the JSON body settings are read once in initRouter, so without a
			// restart the panel keeps serving the old blob forever.
			name:        "change JSON subscription body (subJsonMux)",
			existing:    map[string]any{"subJsonMux": `{"enabled":false}`},
			plan:        mkPlan(map[string]any{"subJsonMux": `{"enabled":true}`}),
			wantRestart: 1,
		},
		{
			// Route registration: a changed path 404s until the engine is rebuilt.
			name:        "change JSON subscription path",
			existing:    map[string]any{"subJsonPath": "/json/"},
			plan:        mkPlan(map[string]any{"subJsonPath": "/j/"}),
			wantRestart: 1,
		},
		{
			name:        "toggle JSON subscription route",
			existing:    map[string]any{"subJsonEnable": false},
			plan:        mkPlan(map[string]any{"subJsonEnable": true}),
			wantRestart: 1,
		},
		{
			// Looks like a link-generation field but is frozen at startup too.
			name:        "change subscription page title",
			existing:    map[string]any{"subTitle": "old"},
			plan:        mkPlan(map[string]any{"subTitle": "new"}),
			wantRestart: 1,
		},
		{
			name:        "unchanged JSON body does NOT restart",
			existing:    map[string]any{"subJsonMux": `{"enabled":true}`, "subTitle": "same"},
			plan:        mkPlan(map[string]any{"subJsonMux": `{"enabled":true}`, "subTitle": "same"}),
			wantRestart: 0,
		},
		{
			name:        "link-generation key only (subURI) does NOT restart",
			existing:    map[string]any{"subURI": "https://old.example/sub/"},
			plan:        mkPlan(map[string]any{"subURI": "https://new.example/sub/"}),
			wantRestart: 0,
		},
		{
			name:        "no-op plan (identical) does NOT restart",
			existing:    map[string]any{"subEnable": true, "subPort": float64(2096)},
			plan:        mkPlan(map[string]any{"subEnable": true, "subPort": 2096}),
			wantRestart: 0,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var restarts int32
			mu := sync.Mutex{}
			server := make(map[string]any)
			for k, v := range c.existing {
				server[k] = v
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/login":
					w.Write(okResponse(nil))
					return
				case "/panel/api/setting/all":
					// GetSettings: return the current server settings (new-API surface;
					// also satisfies the one-shot useSettingsAPI probe).
					mu.Lock()
					snapshot := make(map[string]any, len(server))
					for k, v := range server {
						snapshot[k] = v
					}
					mu.Unlock()
					w.Write(okResponse(snapshot))
					return
				case "/panel/api/setting/update":
					var in map[string]any
					_ = json.NewDecoder(r.Body).Decode(&in)
					mu.Lock()
					for k, v := range in {
						server[k] = v
					}
					mu.Unlock()
					w.Write(okResponse(nil))
					return
				case "/panel/api/setting/restartPanel":
					atomic.AddInt32(&restarts, 1)
					w.Write(okResponse(nil))
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}))
			defer srv.Close()

			client := newTestClient(t, srv.URL)
			// Pin the new-API surface so settingPath() resolves to /panel/api/setting/*
			// without a probe round-trip (matches TestSendRestartFallsBackOn404).
			client.settingsAPIMu.Lock()
			v := true
			client.settingsUnderAPI = &v
			client.settingsAPIMu.Unlock()

			r := &PanelSubscriptionResource{client: client}
			var d diag.Diagnostics
			r.applySubscription(context.Background(), &c.plan, &d)

			if d.HasError() {
				t.Fatalf("applySubscription returned errors: %v", d.Errors())
			}
			if got := atomic.LoadInt32(&restarts); got != c.wantRestart {
				t.Fatalf("restartPanel calls = %d, want %d", got, c.wantRestart)
			}
		})
	}
}
