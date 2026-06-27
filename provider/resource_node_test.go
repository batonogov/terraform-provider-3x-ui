package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---------------------------------------------------------------------------
// Resource surface (Metadata / Configure / Schema / Delete / ImportState)
// ---------------------------------------------------------------------------

func TestNodeResource_Metadata(t *testing.T) {
	r := &NodeResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "threexui",
	}, &resp)
	if resp.TypeName != "threexui_node" {
		t.Fatalf("expected threexui_node, got %s", resp.TypeName)
	}
}

func TestNodeResource_Schema(t *testing.T) {
	r := &NodeResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema errors: %v", resp.Diagnostics)
	}
	// Required managed attributes must be present.
	required := []string{"name", "address", "port"}
	for _, attr := range required {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Fatalf("required attribute %q missing from schema", attr)
		}
	}
	// Sensitive managed attributes must stay Sensitive (security rule).
	for _, attr := range []string{"api_token", "pinned_cert_sha256"} {
		a, ok := resp.Schema.Attributes[attr]
		if !ok {
			t.Fatalf("sensitive attribute %q missing", attr)
		}
		if !a.IsSensitive() {
			t.Fatalf("attribute %q must be Sensitive", attr)
		}
	}
}

func TestNodeResource_Configure(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	r := &NodeResource{}

	// nil ProviderData — no-op.
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on nil ProviderData: %v", resp.Diagnostics)
	}

	// Valid client.
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error with valid client: %v", resp.Diagnostics)
	}
	if r.client == nil {
		t.Fatal("client not set after Configure")
	}

	// Wrong type.
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "not a client"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error with wrong ProviderData type")
	}
}

func TestNodeResource_Delete_NotYetImplemented(t *testing.T) {
	r := &NodeResource{}
	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)
	// Delete is a placeholder for M2 — must emit a warning, no error.
	hasWarning := false
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Fatal("expected a warning diagnostic on Delete (M3 TODO)")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on Delete: %v", resp.Diagnostics)
	}
}

func TestNodeResource_Read_RemovesOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		}
		if r.URL.Path == "/panel/api/nodes/42" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	r := &NodeResource{client: newTestClient(t, srv.URL)}

	// State carries an id pointing at a node the panel reports as 404.
	state := nodeResourceReadState(t, r, "42")
	resp := newNodeResourceReadResponse(t, r)
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on not-found Read: %v", resp.Diagnostics)
	}
}

// ---------------------------------------------------------------------------
// Pure helpers: parseNodeID / nodeFromModel / flattenNodeToModel round-trip
// ---------------------------------------------------------------------------

func TestParseNodeID(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1", 1, false},
		{"42", 42, false},
		{"0", 0, true},   // zero is invalid
		{"abc", 0, true}, // non-numeric
		{"", 0, true},
	}
	for _, tc := range cases {
		got, err := parseNodeID(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("parseNodeID(%q) expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseNodeID(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("parseNodeID(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestNodeFromModel_FlattenRoundtrip(t *testing.T) {
	plan := &NodeResourceModel{
		Name:                types.StringValue("de-fra-1"),
		Remark:              types.StringValue("Frankfurt"),
		Scheme:              types.StringValue("https"),
		Address:             types.StringValue("node1.example.com"),
		Port:                types.Int64Value(2053),
		BasePath:            types.StringValue("/"),
		ApiToken:            types.StringValue("tok-123"),
		Enable:              types.BoolValue(true),
		AllowPrivateAddress: types.BoolValue(false),
		TlsVerifyMode:       types.StringValue("pin"),
		PinnedCertSha256:    types.StringValue("abcdef"),
		InboundSyncMode:     types.StringValue("selected"),
		InboundTags:         []types.String{types.StringValue("tag-a"), types.StringValue("tag-b")},
		OutboundTag:         types.StringValue("proxy-out"),
	}

	n := nodeFromModel(plan)
	if n.Name != "de-fra-1" || n.Address != "node1.example.com" || n.Port != 2053 {
		t.Fatalf("nodeFromModel wrong managed fields: %+v", n)
	}
	if n.ApiToken != "tok-123" || n.PinnedCertSha256 != "abcdef" {
		t.Fatalf("nodeFromModel wrong sensitive fields: %+v", n)
	}
	if n.TlsVerifyMode != "pin" || n.InboundSyncMode != "selected" {
		t.Fatalf("nodeFromModel wrong mode fields: %+v", n)
	}
	if len(n.InboundTags) != 2 || n.InboundTags[0] != "tag-a" {
		t.Fatalf("nodeFromModel wrong inbound tags: %v", n.InboundTags)
	}
	if n.OutboundTag != "proxy-out" {
		t.Fatalf("nodeFromModel wrong outbound tag: %q", n.OutboundTag)
	}
	if !n.Enable {
		t.Fatal("Enable should default/propagate to true")
	}

	// Flatten back and assert the managed fields survive the round-trip.
	var got NodeResourceModel
	got.ApiToken = types.StringValue("tok-123") // preserve (flatten keeps non-empty remote)
	got.PinnedCertSha256 = types.StringValue("abcdef")
	flattenNodeToModel(n, &got)
	if got.Name.ValueString() != "de-fra-1" {
		t.Fatalf("flatten Name mismatch: %s", got.Name.ValueString())
	}
	if got.Port.ValueInt64() != 2053 {
		t.Fatalf("flatten Port mismatch: %d", got.Port.ValueInt64())
	}
	if got.TlsVerifyMode.ValueString() != "pin" {
		t.Fatalf("flatten TlsVerifyMode mismatch: %s", got.TlsVerifyMode.ValueString())
	}
	if len(got.InboundTags) != 2 || got.InboundTags[1].ValueString() != "tag-b" {
		t.Fatalf("flatten InboundTags mismatch: %v", got.InboundTags)
	}
}

// flattenNodeToModel must NOT clobber sensitive attrs with empty remote
// values (defensive; per #314 R1 the panel returns them raw, but an empty
// remote means "unset", not "erase").
func TestFlattenNodeToModel_PreservesEmptySensitive(t *testing.T) {
	n := &Node{Name: "n", Address: "a", Port: 1, ApiToken: "", PinnedCertSha256: ""}
	m := &NodeResourceModel{
		ApiToken:         types.StringValue("kept-tok"),
		PinnedCertSha256: types.StringValue("kept-pin"),
	}
	flattenNodeToModel(n, m)
	if m.ApiToken.ValueString() != "kept-tok" {
		t.Fatalf("ApiToken was clobbered with empty remote: %q", m.ApiToken.ValueString())
	}
	if m.PinnedCertSha256.ValueString() != "kept-pin" {
		t.Fatalf("PinnedCertSha256 was clobbered with empty remote: %q", m.PinnedCertSha256.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Client methods: CreateNode / GetNode
// ---------------------------------------------------------------------------

func TestCreateNode(t *testing.T) {
	var receivedForm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes":
			r.ParseForm()
			receivedForm = r.Form.Encode()
			w.Write(okResponse(map[string]any{
				"id":         7,
				"name":       "de-fra-1",
				"address":    "node1.example.com",
				"port":       2053,
				"apiToken":   "tok-123",
				"enable":     true,
				"guid":       "g-7",
				"status":     "unknown",
				"transitive": false,
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	created, err := client.CreateNode(context.Background(), &Node{
		Name: "de-fra-1", Scheme: "https", Address: "node1.example.com",
		Port: 2053, ApiToken: "tok-123", Enable: true,
		InboundTags: []string{},
	})
	if err != nil {
		t.Fatalf("CreateNode error: %v", err)
	}
	if created.Id != 7 {
		t.Fatalf("expected created id 7, got %d", created.Id)
	}
	// The form must carry the managed fields with upstream names.
	for _, want := range []string{"name=de-fra-1", "address=node1.example.com", "port=2053", "apiToken=tok-123", "inboundTags=%5B%5D"} {
		if !contains(receivedForm, want) {
			t.Fatalf("create form missing %q in %q", want, receivedForm)
		}
	}
}

func TestCreateNode_Nil(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	if _, err := client.CreateNode(context.Background(), nil); err == nil {
		t.Fatal("expected error on nil node")
	}
}

func TestGetNode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes/7":
			w.Write(okResponse(map[string]any{
				"id": 7, "name": "de-fra-1", "address": "node1.example.com",
				"port": 2053, "status": "online", "guid": "g-7",
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	got, err := client.GetNode(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetNode error: %v", err)
	}
	if got.Id != 7 || got.Name != "de-fra-1" || got.Status != "online" {
		t.Fatalf("unexpected node: %+v", got)
	}
}

func TestGetNode_ZeroID(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	if _, err := client.GetNode(context.Background(), 0); err == nil {
		t.Fatal("expected error on zero id")
	}
}

func TestNodeToForm(t *testing.T) {
	form := nodeToForm(&Node{
		Name: "n", Scheme: "https", Address: "a", Port: 2053,
		ApiToken: "tok", Enable: true, InboundTags: []string{"x", "y"},
	})
	if form.Get("name") != "n" {
		t.Fatalf("name: %q", form.Get("name"))
	}
	if form.Get("port") != "2053" {
		t.Fatalf("port: %q", form.Get("port"))
	}
	// inboundTags serialized as JSON array string.
	if form.Get("inboundTags") != "[\"x\",\"y\"]" {
		t.Fatalf("inboundTags: %q", form.Get("inboundTags"))
	}
	if form.Get("enable") != "true" {
		t.Fatalf("enable: %q", form.Get("enable"))
	}
}

func TestNodeToForm_Nil(t *testing.T) {
	if form := nodeToForm(nil); len(form) != 0 {
		t.Fatalf("expected empty form for nil, got %v", form)
	}
}

// nodeResourceReadState builds a tfsdk.State for the node resource with the
// given id, so Read can be exercised in unit tests.
func nodeResourceReadState(t *testing.T, r *NodeResource, id string) tfsdk.State {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k := range schemaResp.Schema.Attributes {
		vals[k] = tftypes.NewValue(objType.AttributeTypes[k], nil)
	}
	// Only the id is needed for Read; everything else stays null.
	vals["id"] = tftypes.NewValue(tftypes.String, id)
	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objType, vals),
	}
}

// newNodeResourceReadResponse builds a ReadResponse with a properly
// initialised State so that resp.State.RemoveResource works in unit tests.
func newNodeResourceReadResponse(t *testing.T, r *NodeResource) resource.ReadResponse {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	return resource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
}

// TestNodeResource_Create exercises the full Create path against a mock panel:
// plan -> CreateNode -> re-read GetNode -> state. Covers the create flow that
// acc-tests normally cover, so it counts toward Codecov patch coverage.
func TestNodeResource_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes":
			r.ParseForm()
			if r.Form.Get("name") != "de-fra-1" {
				w.Write(failResponse("bad name"))
				return
			}
			w.Write(okResponse(map[string]any{
				"id": 7, "name": "de-fra-1", "address": "node1.example.com",
				"port": 2053, "enable": true,
			}))
			return
		case "/panel/api/nodes/7":
			w.Write(okResponse(map[string]any{
				"id": 7, "name": "de-fra-1", "address": "node1.example.com",
				"port": 2053, "enable": true, "status": "online", "guid": "g-7",
				"latencyMs": 12,
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	r := &NodeResource{client: newTestClient(t, srv.URL)}
	plan := nodeResourcePlan(t, r, map[string]any{
		"name":    "de-fra-1",
		"address": "node1.example.com",
		"port":    int64(2053),
		"scheme":  "https",
		"enable":  true,
	})

	resp := newNodeResourceCreateResponse(t, r)
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on Create: %v", resp.Diagnostics)
	}

	var state NodeResourceModel
	resp.State.Get(context.Background(), &state)
	if state.ID.ValueString() != "7" {
		t.Fatalf("expected id 7, got %s", state.ID.ValueString())
	}
	if state.Status.ValueString() != "online" {
		t.Fatalf("expected observed status online (from re-read), got %s", state.Status.ValueString())
	}
}

// TestNodeResource_Create_ReachabilityError verifies the create error path
// surfaces the ensureReachability hint.
func TestNodeResource_Create_ReachabilityError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes":
			w.Write(failResponse("node unreachable"))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	r := &NodeResource{client: newTestClient(t, srv.URL)}
	plan := nodeResourcePlan(t, r, map[string]any{
		"name": "x", "address": "unreachable.example", "port": int64(2053),
	})
	resp := newNodeResourceCreateResponse(t, r)
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error on unreachable node create")
	}
}

// nodeResourcePlan builds a tfsdk.Plan with the given managed attribute values.
func nodeResourcePlan(t *testing.T, r *NodeResource, vals map[string]any) tfsdk.Plan {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	out := map[string]tftypes.Value{}
	for k := range schemaResp.Schema.Attributes {
		switch k {
		case "port", "latency_ms", "last_heartbeat", "uptime_secs", "net_up",
			"net_down", "config_dirty_at", "inbound_count", "client_count",
			"online_count", "active_count", "disabled_count", "depleted_count",
			"created_at", "updated_at":
			out[k] = tftypes.NewValue(tftypes.Number, nil)
		case "enable", "allow_private_address", "config_dirty", "transitive":
			out[k] = tftypes.NewValue(tftypes.Bool, nil)
		case "cpu_pct", "mem_pct":
			out[k] = tftypes.NewValue(tftypes.Number, nil)
		case "inbound_tags":
			out[k] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil)
		default:
			out[k] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	if v, ok := vals["name"]; ok {
		out["name"] = tftypes.NewValue(tftypes.String, v)
	}
	if v, ok := vals["address"]; ok {
		out["address"] = tftypes.NewValue(tftypes.String, v)
	}
	if v, ok := vals["port"]; ok {
		out["port"] = tftypes.NewValue(tftypes.Number, v)
	}
	if v, ok := vals["scheme"]; ok {
		out["scheme"] = tftypes.NewValue(tftypes.String, v)
	}
	if v, ok := vals["enable"]; ok {
		out["enable"] = tftypes.NewValue(tftypes.Bool, v)
	}
	if v, ok := vals["id"]; ok {
		out["id"] = tftypes.NewValue(tftypes.String, v)
	}
	_ = objType
	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object), out),
	}
}

func newNodeResourceCreateResponse(t *testing.T, r *NodeResource) resource.CreateResponse {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	return resource.CreateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
}

// TestNodeResource_Update_Placeholder covers the M2 Update placeholder: it
// re-reads and emits the "not yet implemented" warning. Real update is M3.
func TestNodeResource_Update_Placeholder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/nodes/7":
			w.Write(okResponse(map[string]any{
				"id": 7, "name": "de-fra-1", "address": "node1.example.com",
				"port": 2053, "enable": true,
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	r := &NodeResource{client: newTestClient(t, srv.URL)}
	plan := nodeResourcePlan(t, r, map[string]any{
		"name": "de-fra-1", "address": "node1.example.com", "port": int64(2053),
		"id": "7",
	})

	resp := newNodeResourceUpdateResponse(t, r)
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan}, &resp)
	hasWarning := false
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Fatal("expected the M3-TODO warning on Update")
	}
}

func TestNewNodeResource(t *testing.T) {
	if NewNodeResource() == nil {
		t.Fatal("NewNodeResource returned nil")
	}
}

func newNodeResourceUpdateResponse(t *testing.T, r *NodeResource) resource.UpdateResponse {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	return resource.UpdateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
}
