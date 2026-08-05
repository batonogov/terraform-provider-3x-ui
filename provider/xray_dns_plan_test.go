package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func dnsServerPlanValue(t *testing.T, objType tftypes.Object, address string) tftypes.Value {
	t.Helper()
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	vals["address"] = tftypes.NewValue(tftypes.String, address)
	return tftypes.NewValue(objType, vals)
}

func dnsRichServerPlanValue(t *testing.T, objType tftypes.Object, address string) tftypes.Value {
	t.Helper()
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	vals["address"] = tftypes.NewValue(tftypes.String, address)
	vals["port"] = tftypes.NewValue(tftypes.Number, int64(5353))
	vals["domains"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "geosite:cn"),
	})
	vals["expect_ips"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "geoip:cn"),
	})
	vals["unexpected_ips"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "geoip:private"),
	})
	vals["skip_fallback"] = tftypes.NewValue(tftypes.Bool, true)
	vals["query_strategy"] = tftypes.NewValue(tftypes.String, "UseIPv4")
	vals["disable_cache"] = tftypes.NewValue(tftypes.Bool, true)
	vals["final_query"] = tftypes.NewValue(tftypes.Bool, true)
	return tftypes.NewValue(objType, vals)
}

func dnsPlanRaw(t *testing.T, schemaResp resource.SchemaResponse, servers tftypes.Value) tftypes.Value {
	t.Helper()
	ctx := context.Background()
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		switch name {
		case "id":
			vals[name] = tftypes.NewValue(tftypes.String, xraySectionDNS.id)
		case "server":
			vals[name] = servers
		default:
			vals[name] = tftypes.NewValue(attrType, nil)
		}
	}
	return tftypes.NewValue(objType, vals)
}

func TestXrayDNSResourceModifyPlanStripsStaleServerFieldsOnReorder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	rich := func(address string) tftypes.Value {
		return dnsRichServerPlanValue(t, serverObjType, address)
	}
	sparse := func(address string) tftypes.Value {
		return dnsServerPlanValue(t, serverObjType, address)
	}

	stateServers := tftypes.NewValue(serverListType, []tftypes.Value{rich("8.8.8.8"), sparse("1.1.1.1")})
	configServers := tftypes.NewValue(serverListType, []tftypes.Value{sparse("1.1.1.1"), rich("8.8.8.8")})
	pollutedPlanServers := tftypes.NewValue(serverListType, []tftypes.Value{rich("1.1.1.1"), rich("8.8.8.8")})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, pollutedPlanServers)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, configServers)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, stateServers)}
	resp := &resource.ModifyPlanResponse{
		Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objType, nil)},
	}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected ModifyPlan diagnostics: %v", resp.Diagnostics)
	}

	var got XrayDNSModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to decode reconciled plan: %v", resp.Diagnostics)
	}
	if len(got.Server) != 2 {
		t.Fatalf("expected two servers, got %d", len(got.Server))
	}
	if !got.Server[0].Port.IsNull() || !got.Server[0].Domains.IsNull() ||
		!got.Server[0].ExpectIPs.IsNull() || !got.Server[0].UnexpectedIPs.IsNull() ||
		!got.Server[0].SkipFallback.IsNull() || !got.Server[0].QueryStrategy.IsNull() ||
		!got.Server[0].DisableCache.IsNull() || !got.Server[0].FinalQuery.IsNull() {
		t.Fatalf("sparse server inherited stale fields: %#v", got.Server[0])
	}
	if got.Server[1].Port.ValueInt64() != 5353 || got.Server[1].Domains.IsNull() ||
		got.Server[1].ExpectIPs.IsNull() || got.Server[1].UnexpectedIPs.IsNull() ||
		!got.Server[1].SkipFallback.ValueBool() || got.Server[1].QueryStrategy.ValueString() != "UseIPv4" ||
		!got.Server[1].DisableCache.ValueBool() || !got.Server[1].FinalQuery.ValueBool() {
		t.Fatalf("rich server lost configured fields: %#v", got.Server[1])
	}
}

func TestXrayDNSResourceModifyPlanReconcilesKnownSiblingWithUnknownServerElement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	sparse := dnsServerPlanValue(t, serverObjType, "1.1.1.1")
	rich := dnsRichServerPlanValue(t, serverObjType, "1.1.1.1")
	unknown := tftypes.NewValue(serverObjType, tftypes.UnknownValue)
	stateServers := tftypes.NewValue(serverListType, []tftypes.Value{
		dnsRichServerPlanValue(t, serverObjType, "8.8.8.8"),
		sparse,
	})
	configServers := tftypes.NewValue(serverListType, []tftypes.Value{sparse, unknown})
	pollutedPlanServers := tftypes.NewValue(serverListType, []tftypes.Value{rich, unknown})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, pollutedPlanServers)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, configServers)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, stateServers)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan must preserve an unknown server element: %v", resp.Diagnostics)
	}

	var got types.List
	resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("server"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to read reconciled server list: %v", resp.Diagnostics)
	}
	if len(got.Elements()) != 2 || !got.Elements()[1].IsUnknown() {
		t.Fatalf("unknown server element was not preserved: %s", got)
	}
	known, ok := got.Elements()[0].(types.Object)
	if !ok {
		t.Fatalf("known server has unexpected type %T", got.Elements()[0])
	}
	if port := known.Attributes()["port"].(types.Int64); !port.IsNull() {
		t.Fatalf("known sparse sibling retained stale port: %s", port)
	}
}

func TestXrayDNSResourceModifyPlanSkipsUnknownServerCollection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	knownServers := tftypes.NewValue(serverListType, []tftypes.Value{
		dnsServerPlanValue(t, serverObjType, "8.8.8.8"),
	})
	unknownServers := tftypes.NewValue(serverListType, tftypes.UnknownValue)

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, knownServers)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, unknownServers)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, knownServers)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan must defer an unknown server collection: %v", resp.Diagnostics)
	}
}

func TestXrayDNSResourceModifyPlanSkipsNullPlanOrState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	servers := tftypes.NewValue(serverListType, []tftypes.Value{
		dnsServerPlanValue(t, serverObjType, "8.8.8.8"),
	})
	knownRaw := dnsPlanRaw(t, schemaResp, servers)
	nullRaw := tftypes.NewValue(objType, nil)

	for _, tc := range []struct {
		name     string
		planRaw  tftypes.Value
		stateRaw tftypes.Value
	}{
		{name: "null plan", planRaw: nullRaw, stateRaw: knownRaw},
		{name: "null state", planRaw: knownRaw, stateRaw: nullRaw},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tc.planRaw}
			config := tfsdk.Config{Schema: schemaResp.Schema, Raw: knownRaw}
			state := tfsdk.State{Schema: schemaResp.Schema, Raw: tc.stateRaw}
			resp := &resource.ModifyPlanResponse{Plan: plan}

			r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("ModifyPlan must skip null plan/state: %v", resp.Diagnostics)
			}
		})
	}
}

func TestXrayDNSResourceModifyPlanNoOpWhenServersAlreadyMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	servers := tftypes.NewValue(serverListType, []tftypes.Value{
		dnsServerPlanValue(t, serverObjType, "8.8.8.8"),
	})
	raw := dnsPlanRaw(t, schemaResp, servers)
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: raw}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: raw}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected ModifyPlan diagnostics: %v", resp.Diagnostics)
	}

	var got XrayDNSModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("failed to decode unchanged plan: %v", resp.Diagnostics)
	}
	if len(got.Server) != 1 || got.Server[0].Address.ValueString() != "8.8.8.8" {
		t.Fatalf("matching server plan changed unexpectedly: %#v", got.Server)
	}
}
