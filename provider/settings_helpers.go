package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSettingsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}

func mergeSettings(existing, desired map[string]any) map[string]any {
	if existing == nil && desired == nil {
		return map[string]any{}
	}
	if existing == nil {
		return desired
	}
	out := make(map[string]any, len(existing)+len(desired))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range desired {
		out[k] = v
	}
	return out
}

func getStringField(d *schema.ResourceData, path string) (string, bool) {
	v, ok := d.GetOkExists(path) //nolint:staticcheck // GetOkExists needed for zero-value vs unset
	if !ok {
		return "", false
	}
	s, _ := v.(string)
	return s, true
}

func getIntField(d *schema.ResourceData, path string) (int, bool) {
	v, ok := d.GetOkExists(path) //nolint:staticcheck // GetOkExists needed for zero-value vs unset
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, true
	}
}

func getBoolField(d *schema.ResourceData, path string) (bool, bool) {
	v, ok := d.GetOkExists(path) //nolint:staticcheck // GetOkExists needed for zero-value vs unset
	if !ok {
		return false, false
	}
	b, _ := v.(bool)
	return b, true
}

func getPortField(d *schema.ResourceData, path string) (int, bool, error) {
	v, ok := getIntField(d, path)
	if !ok {
		return 0, false, nil
	}
	if v < 1 || v > 65535 {
		return 0, true, fmt.Errorf("%s must be a valid port (1-65535), got %d", path, v)
	}
	return v, true, nil
}
