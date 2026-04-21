package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// List conversion helpers (types.List <-> []any)
// ---------------------------------------------------------------------------

// typesListInt64ToAnySlice converts a types.List of Int64Type to []any for the
// untyped map format (e.g. WireGuard MTU).
func typesListInt64ToAnySlice(l types.List) []any {
	elems := l.Elements()
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		if iv, ok := e.(types.Int64); ok && !iv.IsNull() && !iv.IsUnknown() {
			out = append(out, int(iv.ValueInt64()))
		}
	}
	return out
}

// typesListToAnySlice converts a types.List of StringType to []any for the
// untyped map format.
func typesListToAnySlice(l types.List) []any {
	elems := l.Elements()
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		if sv, ok := e.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
			out = append(out, sv.ValueString())
		}
	}
	return out
}

// anySliceToTypesList converts a []any of strings to a types.List of
// StringType. Returns types.ListNull if the slice is nil or empty.
func anySliceToTypesList(v any) types.List {
	slice, ok := v.([]any)
	if !ok || len(slice) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(slice))
	for _, item := range slice {
		switch s := item.(type) {
		case string:
			elems = append(elems, types.StringValue(s))
		default:
			// best-effort: skip non-string entries
			continue
		}
	}
	if len(elems) == 0 {
		return types.ListNull(types.StringType)
	}
	return types.ListValueMust(types.StringType, elems)
}
