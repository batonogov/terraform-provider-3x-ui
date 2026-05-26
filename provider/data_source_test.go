package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---------------------------------------------------------------------------
// ClientTraffics data source
// ---------------------------------------------------------------------------

func TestClientTrafficsDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/getClientTraffics/test@example.com" {
			w.Write(okResponse(ClientTraffic{
				ID: 1, InboundID: 5, Email: "test@example.com",
				Up: 100, Down: 200, Total: 1000, ExpiryTime: 9999, Enable: true,
			}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	ds := &ClientTrafficsDataSource{client: client}

	// Build ReadRequest with email in Config
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	configVals := map[string]tftypes.Value{
		"email":       tftypes.NewValue(tftypes.String, "test@example.com"),
		"id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"inbound_id":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"enable":      tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		"up":          tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"down":        tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"total":       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"expiry_time": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	}
	configVal := tftypes.NewValue(objType, configVals)
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configVal,
		},
	}
	resp := newDSReadResponse(t, ds)
	ds.Read(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state ClientTrafficsDataSourceModel
	resp.State.Get(ctx, &state)
	if state.Email.ValueString() != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", state.Email.ValueString())
	}
	if state.InboundID.ValueInt64() != 5 {
		t.Fatalf("expected inbound_id 5, got %d", state.InboundID.ValueInt64())
	}
	if state.Up.ValueInt64() != 100 {
		t.Fatalf("expected up 100, got %d", state.Up.ValueInt64())
	}
}

func TestClientTrafficsDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/getClientTraffics/test@example.com" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	ds := &ClientTrafficsDataSource{client: client}

	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	configVals := map[string]tftypes.Value{
		"email":       tftypes.NewValue(tftypes.String, "test@example.com"),
		"id":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"inbound_id":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"enable":      tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		"up":          tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"down":        tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"total":       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"expiry_time": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
	}
	configVal := tftypes.NewValue(objType, configVals)
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configVal,
		},
	}
	resp := newDSReadResponse(t, ds)
	ds.Read(ctx, req, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// Helpers for data source unit tests
// ---------------------------------------------------------------------------

// newDSReadResponse creates a datasource.ReadResponse with a properly
// initialised State so that resp.State.Set works in unit tests.
func newDSReadResponse(t *testing.T, ds datasource.DataSource) datasource.ReadResponse {
	t.Helper()
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)

	ctx := context.Background()
	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
	}
	return datasource.ReadResponse{State: state}
}

// dsHandler creates a httptest handler that logs in and then delegates to
// the provided handler for all non-login requests.
func dsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/login" {
		w.Write(okResponse(nil))
		return
	}
}

// ---------------------------------------------------------------------------
// ServerStatus data source
// ---------------------------------------------------------------------------

func TestServerStatusDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/status" {
			w.Write(okResponse(map[string]any{"cpu": 42.0, "mem": 60.0}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &ServerStatusDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state ServerStatusDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.JSON.IsUnknown() || state.JSON.IsNull() {
		t.Fatal("expected json to be set")
	}
	if state.ID.IsUnknown() || state.ID.IsNull() {
		t.Fatal("expected id to be set")
	}
}

func TestServerStatusDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/status" {
			w.Write(failResponse("internal error"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &ServerStatusDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// OnlineClients data source
// ---------------------------------------------------------------------------

func TestOnlineClientsDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/onlines" {
			w.Write(okResponse([]string{"a@test.com", "b@test.com"}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	ds := &OnlineClientsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state OnlineClientsDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.ID.ValueString() != "2" {
		t.Fatalf("expected id=2, got %s", state.ID.ValueString())
	}
	if len(state.Clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(state.Clients))
	}
}

func TestOnlineClientsDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/onlines" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	ds := &OnlineClientsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// XrayVersions data source
// ---------------------------------------------------------------------------

func TestXrayVersionsDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/getXrayVersion" {
			w.Write(okResponse([]string{"v1.8.0", "v1.8.1"}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &XrayVersionsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state XrayVersionsDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.ID.ValueString() != "2" {
		t.Fatalf("expected id=2, got %s", state.ID.ValueString())
	}
}

func TestXrayVersionsDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/getXrayVersion" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &XrayVersionsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// Inbounds data source
// ---------------------------------------------------------------------------

func TestInboundsDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/list" {
			w.Write(okResponse([]Inbound{{ID: 7, Remark: "test", Settings: "{}"}}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &InboundsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state InboundsDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.ID.ValueString() != "7" {
		t.Fatalf("expected id=7, got %s", state.ID.ValueString())
	}
	if state.Inbounds.IsUnknown() || state.Inbounds.IsNull() {
		t.Fatal("expected inbounds to be set")
	}
}

func TestInboundsDataSource_Read_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/list" {
			w.Write(okResponse([]Inbound{}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &InboundsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state InboundsDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.ID.ValueString() != "0" {
		t.Fatalf("expected id=0 for empty, got %s", state.ID.ValueString())
	}
}

func TestInboundsDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/inbounds/list" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &InboundsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// Settings data source
// ---------------------------------------------------------------------------

func TestSettingsDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/setting/all" {
			w.Write(okResponse(map[string]any{"webPort": 2053.0}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &SettingsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state SettingsDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.JSON.IsUnknown() || state.JSON.IsNull() {
		t.Fatal("expected json to be set")
	}
}

func TestSettingsDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/setting/all" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &SettingsDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// XrayConfig data source
// ---------------------------------------------------------------------------

func TestXrayConfigDataSource_Read(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/getConfigJson" {
			w.Write(okResponse(map[string]any{"inbounds": []any{}}))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &XrayConfigDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	var state XrayConfigDataSourceModel
	resp.State.Get(context.Background(), &state)
	if state.JSON.IsUnknown() || state.JSON.IsNull() {
		t.Fatal("expected json to be set")
	}
}

func TestXrayConfigDataSource_Read_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panel/api/server/getConfigJson" {
			w.Write(failResponse("fail"))
			return
		}
		dsHandler(w, r)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ds := &XrayConfigDataSource{client: client}
	resp := newDSReadResponse(t, ds)
	ds.Read(context.Background(), datasource.ReadRequest{}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error, got none")
	}
}

// ---------------------------------------------------------------------------
// Metadata tests
// ---------------------------------------------------------------------------

func TestDataSourceMetadata(t *testing.T) {
	cases := []struct {
		ds       datasource.DataSource
		expected string
	}{
		{NewClientTrafficsDataSource(), "threexui_client_traffics"},
		{NewOnlineClientsDataSource(), "threexui_online_clients"},
		{NewServerStatusDataSource(), "threexui_server_status"},
		{NewXrayVersionsDataSource(), "threexui_xray_versions"},
		{NewInboundsDataSource(), "threexui_inbounds"},
		{NewSettingsDataSource(), "threexui_settings"},
		{NewXrayConfigDataSource(), "threexui_xray_config"},
	}
	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			var resp datasource.MetadataResponse
			tc.ds.Metadata(context.Background(), datasource.MetadataRequest{
				ProviderTypeName: "threexui",
			}, &resp)
			if resp.TypeName != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, resp.TypeName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Configure tests
// ---------------------------------------------------------------------------

func TestDataSourceConfigure(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")

	configure := func(ds datasource.DataSource, req datasource.ConfigureRequest) datasource.ConfigureResponse {
		var resp datasource.ConfigureResponse
		switch d := ds.(type) {
		case *ClientTrafficsDataSource:
			d.Configure(context.Background(), req, &resp)
		case *OnlineClientsDataSource:
			d.Configure(context.Background(), req, &resp)
		case *ServerStatusDataSource:
			d.Configure(context.Background(), req, &resp)
		case *XrayVersionsDataSource:
			d.Configure(context.Background(), req, &resp)
		case *InboundsDataSource:
			d.Configure(context.Background(), req, &resp)
		case *SettingsDataSource:
			d.Configure(context.Background(), req, &resp)
		case *XrayConfigDataSource:
			d.Configure(context.Background(), req, &resp)
		default:
			t.Fatalf("unhandled data source type: %T", ds)
		}
		return resp
	}

	cases := []struct {
		name string
		ds   datasource.DataSource
	}{
		{"client_traffics", NewClientTrafficsDataSource()},
		{"online_clients", NewOnlineClientsDataSource()},
		{"server_status", NewServerStatusDataSource()},
		{"xray_versions", NewXrayVersionsDataSource()},
		{"inbounds", NewInboundsDataSource()},
		{"settings", NewSettingsDataSource()},
		{"xray_config", NewXrayConfigDataSource()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// nil ProviderData — should be a no-op, no error
			resp := configure(tc.ds, datasource.ConfigureRequest{})
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error on nil ProviderData: %v", resp.Diagnostics)
			}

			// Valid client
			resp = configure(tc.ds, datasource.ConfigureRequest{ProviderData: client})
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error with valid client: %v", resp.Diagnostics)
			}

			// Wrong type
			resp = configure(tc.ds, datasource.ConfigureRequest{ProviderData: "not a client"})
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected error with wrong type, got none")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Client-level tests (edge cases not covered by data source Read tests)
// ---------------------------------------------------------------------------

func TestClientGetClientTraffics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/getClientTraffics/test@example.com" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(ClientTraffic{
			ID: 1, InboundID: 5, Email: "test@example.com",
			Up: 100, Down: 200, Total: 1000, ExpiryTime: 9999, Enable: true,
		}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	traffic, err := client.GetClientTraffics(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("GetClientTraffics failed: %v", err)
	}
	if traffic.ID != 1 || traffic.Email != "test@example.com" || traffic.Up != 100 || traffic.InboundID != 5 {
		t.Fatalf("unexpected traffic: %#v", traffic)
	}
}

func TestClientGetClientTrafficsPathEscape(t *testing.T) {
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RawPath
		if receivedPath == "" {
			receivedPath = r.URL.Path
		}
		w.Write(okResponse(ClientTraffic{ID: 1, Email: "user/slash"}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	setLegacyClientAPI(client)
	_, err := client.GetClientTraffics(context.Background(), "user/slash")
	if err != nil {
		t.Fatalf("GetClientTraffics failed: %v", err)
	}
	expected := "/panel/api/inbounds/getClientTraffics/user%2Fslash"
	if receivedPath != expected {
		t.Fatalf("path not escaped: got %q, want %q", receivedPath, expected)
	}
}

func TestClientGetClientTrafficsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetClientTraffics(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent client, got nil")
	}
}

func TestClientGetClientTrafficsEmptyEmail(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	_, err := client.GetClientTraffics(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
}

func TestClientGetInbounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/inbounds/list" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse([]Inbound{{ID: 1, Port: 1234, Protocol: "vmess", Settings: "{}"}}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	inbounds, err := client.GetInbounds(context.Background())
	if err != nil {
		t.Fatalf("GetInbounds failed: %v", err)
	}
	if len(inbounds) != 1 || inbounds[0].ID != 1 {
		t.Fatalf("unexpected inbounds: %#v", inbounds)
	}
}

func TestClientGetServerStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"cpu": 12}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	status, err := client.GetServerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetServerStatus failed: %v", err)
	}
	if status["cpu"] != float64(12) {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestClientGetXrayVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getXrayVersion" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse([]string{"v1", "v2"}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	versions, err := client.GetXrayVersions(context.Background())
	if err != nil {
		t.Fatalf("GetXrayVersions failed: %v", err)
	}
	if len(versions) != 2 || versions[0] != "v1" {
		t.Fatalf("unexpected versions: %#v", versions)
	}
}

func TestClientGetXrayConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getConfigJson" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"inbounds": []any{}}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	config, err := client.GetXrayConfig(context.Background())
	if err != nil {
		t.Fatalf("GetXrayConfig failed: %v", err)
	}
	if _, ok := config["inbounds"]; !ok {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestClientGetSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/setting/all" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(okResponse(map[string]any{"webPort": 2053}))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	settings, err := client.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings["webPort"] != float64(2053) {
		t.Fatalf("unexpected settings: %#v", settings)
	}
}
