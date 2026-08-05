package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func dnsServerPlanValue(
	t *testing.T,
	objType tftypes.Object,
	address string,
	port *int64,
	domains []string,
	skipFallback *bool,
) tftypes.Value {
	t.Helper()
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	vals["address"] = tftypes.NewValue(tftypes.String, address)
	if port != nil {
		vals["port"] = tftypes.NewValue(tftypes.Number, *port)
	}
	if len(domains) > 0 {
		elements := make([]tftypes.Value, len(domains))
		for i, domain := range domains {
			elements[i] = tftypes.NewValue(tftypes.String, domain)
		}
		vals["domains"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elements)
	}
	if skipFallback != nil {
		vals["skip_fallback"] = tftypes.NewValue(tftypes.Bool, *skipFallback)
	}
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
	port := int64(5353)
	skipFallback := true
	rich := func(address string) tftypes.Value {
		return dnsServerPlanValue(t, serverObjType, address, &port, []string{"geosite:cn"}, &skipFallback)
	}
	sparse := func(address string) tftypes.Value {
		return dnsServerPlanValue(t, serverObjType, address, nil, nil, nil)
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
	if !got.Server[0].Port.IsNull() || !got.Server[0].Domains.IsNull() || !got.Server[0].SkipFallback.IsNull() {
		t.Fatalf("sparse server inherited stale fields: %#v", got.Server[0])
	}
	if got.Server[1].Port.ValueInt64() != port || got.Server[1].Domains.IsNull() || !got.Server[1].SkipFallback.ValueBool() {
		t.Fatalf("rich server lost configured fields: %#v", got.Server[1])
	}
}

func TestXrayDNSResourceModifyPlanSkipsUnknownServerElement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &XrayDNSResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	serverListType := objType.AttributeTypes["server"].(tftypes.List)
	serverObjType := serverListType.ElementType.(tftypes.Object)
	knownServers := tftypes.NewValue(serverListType, []tftypes.Value{
		dnsServerPlanValue(t, serverObjType, "8.8.8.8", nil, nil, nil),
	})
	configServers := tftypes.NewValue(serverListType, []tftypes.Value{
		tftypes.NewValue(serverObjType, tftypes.UnknownValue),
	})

	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, knownServers)}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, configServers)}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: dnsPlanRaw(t, schemaResp, knownServers)}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, Config: config, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan must defer an unknown server element: %v", resp.Diagnostics)
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
		dnsServerPlanValue(t, serverObjType, "8.8.8.8", nil, nil, nil),
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
