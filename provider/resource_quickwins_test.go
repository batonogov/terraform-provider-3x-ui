package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// list_helpers.go
// ---------------------------------------------------------------------------

func TestTypesListInt64ToAnySlice(t *testing.T) {
	lst, _ := types.ListValueFrom(context.Background(), types.Int64Type, []types.Int64{
		types.Int64Value(1), types.Int64Value(2), types.Int64Value(3),
	})
	result := typesListInt64ToAnySlice(lst)
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0].(int) != 1 || result[2].(int) != 3 {
		t.Fatalf("unexpected values: %v", result)
	}
}

func TestTypesListInt64ToAnySlice_Empty(t *testing.T) {
	lst := types.ListNull(types.Int64Type)
	result := typesListInt64ToAnySlice(lst)
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestTypesListToAnySlice(t *testing.T) {
	lst, _ := types.ListValueFrom(context.Background(), types.StringType, []types.String{
		types.StringValue("a"), types.StringValue("b"),
	})
	result := typesListToAnySlice(lst)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0].(string) != "a" {
		t.Fatalf("unexpected values: %v", result)
	}
}

func TestAnySliceToTypesList(t *testing.T) {
	result := anySliceToTypesList([]any{"x", "y"})
	if result.IsNull() {
		t.Fatal("expected non-null list")
	}
	elems := result.Elements()
	if len(elems) != 2 {
		t.Fatalf("expected 2, got %d", len(elems))
	}
}

func TestAnySliceToTypesList_Nil(t *testing.T) {
	result := anySliceToTypesList(nil)
	if !result.IsNull() {
		t.Fatal("expected null list for nil input")
	}
}

func TestAnySliceToTypesList_Empty(t *testing.T) {
	result := anySliceToTypesList([]any{})
	if !result.IsNull() {
		t.Fatal("expected null list for empty input")
	}
}

func TestAnySliceToTypesList_NonStrings(t *testing.T) {
	result := anySliceToTypesList([]any{42, true})
	if !result.IsNull() {
		t.Fatal("expected null list when all entries are non-strings")
	}
}

func TestAnySliceToTypesList_NonSlice(t *testing.T) {
	result := anySliceToTypesList("not a slice")
	if !result.IsNull() {
		t.Fatal("expected null list for non-slice input")
	}
}

// ---------------------------------------------------------------------------
// resource_panel_user.go
// ---------------------------------------------------------------------------

func TestPanelUserResource_Metadata(t *testing.T) {
	r := &PanelUserResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "threexui",
	}, &resp)
	if resp.TypeName != "threexui_panel_user" {
		t.Fatalf("expected threexui_panel_user, got %s", resp.TypeName)
	}
}

func TestPanelUserResource_Configure(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	r := &PanelUserResource{}

	// nil ProviderData — no-op
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on nil: %v", resp.Diagnostics)
	}

	// Valid client
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error with client: %v", resp.Diagnostics)
	}

	// Wrong type
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "wrong"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error with wrong type")
	}
}

func TestPanelUserResource_Read(t *testing.T) {
	r := &PanelUserResource{}
	var resp resource.ReadResponse
	r.Read(context.Background(), resource.ReadRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no-op Read, got errors: %v", resp.Diagnostics)
	}
}

func TestPanelUserResource_Delete(t *testing.T) {
	r := &PanelUserResource{}
	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)

	var hasWarning bool
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Fatal("expected warning diagnostic on Delete")
	}
}

// ---------------------------------------------------------------------------
// resource_xray_version.go
// ---------------------------------------------------------------------------

func TestXrayVersionResource_Metadata(t *testing.T) {
	r := &XrayVersionResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "threexui",
	}, &resp)
	if resp.TypeName != "threexui_xray_version" {
		t.Fatalf("expected threexui_xray_version, got %s", resp.TypeName)
	}
}

func TestXrayVersionResource_Configure(t *testing.T) {
	client := newTestClient(t, "http://localhost:0")
	r := &XrayVersionResource{}

	// nil ProviderData — no-op
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error on nil: %v", resp.Diagnostics)
	}

	// Valid client
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error with client: %v", resp.Diagnostics)
	}

	// Wrong type
	resp = resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: 123}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error with wrong type")
	}
}

func TestXrayVersionResource_Delete(t *testing.T) {
	r := &XrayVersionResource{}
	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)

	var hasWarning bool
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Fatal("expected warning diagnostic on Delete")
	}
}
