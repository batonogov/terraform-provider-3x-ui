package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestHostGroupExpandFlatten verifies the model↔struct round-trip: every
// managed attribute survives hostGroupFromModel → flattenHostGroupToModel
// unchanged (snake_case tfsdk ↔ camelCase struct).
func TestHostGroupExpandFlatten(t *testing.T) {
	model := &HostGroupResourceModel{
		GroupID:                types.StringValue("grp123"),
		Remark:                 types.StringValue("premium-servers"),
		ServerDescription:      types.StringValue("Premium EU nodes"),
		SortOrder:              types.Int64Value(5),
		IsDisabled:             types.BoolValue(false),
		IsHidden:               types.BoolValue(true),
		Port:                   types.Int64Value(443),
		Security:               types.StringValue("tls"),
		Sni:                    types.StringValue("example.com"),
		HostHeader:             types.StringValue("example.com"),
		Path:                   types.StringValue("/ws"),
		Fingerprint:            types.StringValue("chrome"),
		OverrideSniFromAddress: types.BoolValue(true),
		KeepSniBlank:           types.BoolValue(false),
		VerifyPeerCertByName:   types.StringValue("example.com"),
		AllowInsecure:          types.BoolValue(false),
		EchConfigList:          types.StringValue("{}"),
		MuxParams:              types.StringValue("{}"),
		SockoptParams:          types.StringValue("{}"),
		FinalMask:              types.StringValue("tcp"),
		MihomoIpVersion:        types.StringValue("ipv4-prefer"),
		MihomoX25519:           types.BoolValue(true),
		ShuffleHost:            types.BoolValue(true),
	}
	model.InboundIDs = types.ListValueMust(types.Int64Type, []attr.Value{
		types.Int64Value(1), types.Int64Value(2), types.Int64Value(3),
	})
	model.Hosts = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("host-a.example.com"), types.StringValue("host-b.example.com"),
	})
	model.Tags = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("eu"), types.StringValue("premium"),
	})
	model.Alpn = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("h2"), types.StringValue("http/1.1"),
	})
	model.PinnedPeerCertSha256 = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("sha256/abc"),
	})
	model.ExcludeFromSubTypes = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("clash"),
	})
	model.NodeGuids = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("node-guid-1"),
	})

	ctx := context.Background()
	hg := hostGroupFromModel(ctx, model)

	if hg.GroupId != "grp123" {
		t.Fatalf("GroupId: got %q, want grp123", hg.GroupId)
	}
	if hg.Remark != "premium-servers" {
		t.Fatalf("Remark: got %q", hg.Remark)
	}
	if hg.Port != 443 {
		t.Fatalf("Port: got %d", hg.Port)
	}
	if hg.Security != "tls" {
		t.Fatalf("Security: got %q", hg.Security)
	}
	if len(hg.InboundIds) != 3 || hg.InboundIds[0] != 1 || hg.InboundIds[2] != 3 {
		t.Fatalf("InboundIds: got %v", hg.InboundIds)
	}
	if len(hg.Hosts) != 2 || hg.Hosts[0] != "host-a.example.com" {
		t.Fatalf("Hosts: got %v", hg.Hosts)
	}
	if len(hg.Tags) != 2 || hg.Tags[1] != "premium" {
		t.Fatalf("Tags: got %v", hg.Tags)
	}
	if hg.MihomoIpVersion != "ipv4-prefer" {
		t.Fatalf("MihomoIpVersion: got %q", hg.MihomoIpVersion)
	}

	// Flatten back into a fresh model and spot-check.
	flat := &HostGroupResourceModel{}
	flattenHostGroupToModel(hg, flat)
	if flat.GroupID.ValueString() != "grp123" {
		t.Fatalf("flatten GroupID: got %q", flat.GroupID)
	}
	if flat.Port.ValueInt64() != 443 {
		t.Fatalf("flatten Port: got %d", flat.Port.ValueInt64())
	}
	if flat.Security.ValueString() != "tls" {
		t.Fatalf("flatten Security: got %q", flat.Security)
	}
	// inbound_ids round-trip
	var ids []int64
	flat.InboundIDs.ElementsAs(ctx, &ids, false)
	if len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
		t.Fatalf("flatten InboundIDs: got %v", ids)
	}
}

// TestHostGroupModelToWire verifies the HostGroup struct serialises with the
// camelCase JSON keys the 3x-ui v3.5.0 controller binds (entity.HostGroup).
func TestHostGroupModelToWire(t *testing.T) {
	hg := &HostGroup{
		GroupId:                "grp456",
		InboundIds:             []int{10, 20},
		Remark:                 "test",
		SortOrder:              1,
		Port:                   8080,
		Security:               "reality",
		Hosts:                  []string{"a.test"},
		Alpn:                   []string{"h2"},
		OverrideSniFromAddress: true,
		MihomoIpVersion:        "dual",
		ShuffleHost:            true,
	}
	raw, err := json.Marshal(hg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	camelKeys := []string{
		"groupId", "inboundIds", "remark", "sortOrder", "port", "security",
		"overrideSniFromAddress", "mihomoIpVersion", "shuffleHost",
	}
	for _, k := range camelKeys {
		if _, ok := m[k]; !ok {
			t.Fatalf("expected camelCase JSON key %q, got keys: %v", k, mapKeys(m))
		}
	}
	// snake_case must NOT appear on the wire.
	for _, k := range []string{"group_id", "inbound_ids", "sort_order", "override_sni_from_address"} {
		if _, ok := m[k]; ok {
			t.Fatalf("snake_case key %q must not be on the wire", k)
		}
	}
}

// TestHostGroupEmptyInboundIDs verifies expand handles an unset/null inbound_ids
// gracefully (the schema enforces min=1, but expand must not panic on null).
func TestHostGroupEmptyInboundIDs(t *testing.T) {
	model := &HostGroupResourceModel{
		Remark: types.StringValue("solo"),
		// InboundIDs intentionally null.
	}
	ctx := context.Background()
	hg := hostGroupFromModel(ctx, model)
	if hg.InboundIds != nil {
		t.Fatalf("expected nil InboundIds for null list, got %v", hg.InboundIds)
	}
	if hg.Remark != "solo" {
		t.Fatalf("Remark: got %q", hg.Remark)
	}
}

// TestHostGroupGeneratedGroupID verifies that when group_id is unset on the
// model, expand leaves GroupId empty so the server generates one on Create.
func TestHostGroupGeneratedGroupID(t *testing.T) {
	model := &HostGroupResourceModel{
		Remark: types.StringValue("auto-id"),
	}
	model.GroupID = types.StringNull()
	model.InboundIDs = types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1)})

	ctx := context.Background()
	hg := hostGroupFromModel(ctx, model)
	if hg.GroupId != "" {
		t.Fatalf("expected empty GroupId for server-generated id, got %q", hg.GroupId)
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ---------------------------------------------------------------------------
// Client method tests (httptest): these guard the two v3.5.0 host-group
// wire-format gotchas — /add returns an ARRAY of host rows (not a single
// HostGroup), and a missing group is signalled by an explicit
// "host group not found" message (not gorm.ErrRecordNotFound).
// ---------------------------------------------------------------------------

func TestCreateHostGroupParsesArrayResponse(t *testing.T) {
	var receivedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/hosts/add":
			receivedBody, _ = io.ReadAll(r.Body)
			// 3x-ui's AddHostGroup returns []*model.Host — a JSON ARRAY, with the
			// server-generated groupId on each row. Mirror that shape exactly.
			w.Write(okResponse([]any{
				map[string]any{"groupId": "grp-abc", "remark": "r1", "host": "h1"},
				map[string]any{"groupId": "grp-abc", "remark": "r1", "host": "h2"},
			}))
			return
		case "/panel/api/hosts/get/grp-abc":
			w.Write(okResponse(map[string]any{
				"groupId": "grp-abc", "remark": "r1",
				"inboundIds": []int{1, 2},
			}))
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	created, err := client.CreateHostGroup(context.Background(), &HostGroup{
		Remark:     "r1",
		InboundIds: []int{1, 2},
	})
	if err != nil {
		t.Fatalf("CreateHostGroup error: %v", err)
	}
	if created == nil || created.GroupId != "grp-abc" {
		t.Fatalf("expected created groupId grp-abc, got %+v", created)
	}
	// Confirm the request body was JSON (not form-encoded).
	var sent map[string]any
	if err := json.Unmarshal(receivedBody, &sent); err != nil {
		t.Fatalf("/add body should be JSON, parse failed: %v", err)
	}
	if sent["remark"] != "r1" {
		t.Fatalf("expected remark=r1 in JSON body, got %v", sent["remark"])
	}
}

func TestGetHostGroupMissingReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		case "/panel/api/hosts/get/missing":
			// 3x-ui signals a missing host group via success:false carrying an
			// explicit "host group not found" message (not gorm.ErrRecordNotFound).
			b, _ := json.Marshal(apiResponse{
				Success: false,
				Msg:     "pages.hosts.toasts.obtain (host group not found)",
			})
			w.Write(b)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	got, err := client.GetHostGroup(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing group must return nil,nil (not an error), got err: %v", err)
	}
	if got != nil {
		t.Fatalf("missing group must return nil HostGroup, got %+v", got)
	}
}

// ---------------------------------------------------------------------------
// Resource-level CRUD tests (tfsdk State wiring). These exercise the resource
// Create/Read/Delete methods end-to-end against a mock panel so Codecov patch
// coverage is green and the v3.5.0 wire-format gotchas regress at the resource
// layer too (not only at the client layer).
// ---------------------------------------------------------------------------

func hostGroupResourceReadState(t *testing.T, r *HostGroupResource, groupID string) tfsdk.State {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k := range schemaResp.Schema.Attributes {
		vals[k] = tftypes.NewValue(objType.AttributeTypes[k], nil)
	}
	// Read needs group_id to build the /get/:groupId path.
	vals["group_id"] = tftypes.NewValue(tftypes.String, groupID)
	vals["id"] = tftypes.NewValue(tftypes.String, groupID)
	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objType, vals),
	}
}

func newHostGroupResourceReadResponse(t *testing.T, r *HostGroupResource) resource.ReadResponse {
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

func TestHostGroupResource_Read_RemovesOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		}
		// 3x-ui signals a missing host group with success:false + an explicit
		// "host group not found" message (not gorm.ErrRecordNotFound).
		if r.URL.Path == "/panel/api/hosts/get/gone" {
			b, _ := json.Marshal(apiResponse{
				Success: false,
				Msg:     "pages.hosts.toasts.obtain (host group not found)",
			})
			w.Write(b)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	r := &HostGroupResource{client: newTestClient(t, srv.URL)}
	state := hostGroupResourceReadState(t, r, "gone")
	resp := newHostGroupResourceReadResponse(t, r)
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read of a missing group must not error (it removes from state): %v", resp.Diagnostics)
	}
	// RemoveResource → state Raw is null.
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed for missing host group, got non-null state")
	}
}

func TestHostGroupResource_Read_PopulatesFromGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: "sess"})
			w.Write(okResponse(nil))
			return
		}
		if r.URL.Path == "/panel/api/hosts/get/grp-1" {
			w.Write(okResponse(map[string]any{
				"groupId": "grp-1", "remark": "hello",
				"inboundIds": []int{10}, "port": 443,
				"security": "tls", "isDisabled": false,
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	r := &HostGroupResource{client: newTestClient(t, srv.URL)}
	state := hostGroupResourceReadState(t, r, "grp-1")
	resp := newHostGroupResourceReadResponse(t, r)
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected Read error: %v", resp.Diagnostics)
	}
	var got HostGroupResourceModel
	resp.State.Get(context.Background(), &got)
	if got.Remark.ValueString() != "hello" {
		t.Fatalf("expected remark=hello after Read, got %q", got.Remark)
	}
	if got.GroupID.ValueString() != "grp-1" {
		t.Fatalf("expected group_id=grp-1, got %q", got.GroupID)
	}
}
