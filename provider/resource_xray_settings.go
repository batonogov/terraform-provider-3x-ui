package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var xrayTemplateMu sync.Mutex

type xraySectionMode int

const (
	xraySectionMergeRoot xraySectionMode = iota
	xraySectionSetPath
	xraySectionReplaceAll
)

type xraySection struct {
	id   string
	mode xraySectionMode
	path []string
}

var (
	xraySectionBasics    = xraySection{id: "xray_basics", mode: xraySectionMergeRoot}
	xraySectionDNS       = xraySection{id: "xray_dns", mode: xraySectionSetPath, path: []string{"dns"}}
	xraySectionRouting   = xraySection{id: "xray_routing", mode: xraySectionSetPath, path: []string{"routing"}}
	xraySectionBalancers = xraySection{id: "xray_balancers", mode: xraySectionSetPath, path: []string{"routing", "balancers"}}
	xraySectionReverse   = xraySection{id: "xray_reverse", mode: xraySectionSetPath, path: []string{"reverse"}}
	xraySectionOutbounds = xraySection{id: "xray_outbounds", mode: xraySectionSetPath, path: []string{"outbounds"}}
	xraySectionAdvanced  = xraySection{id: "xray_advanced", mode: xraySectionReplaceAll}
)

func resourceXrayBasics() *schema.Resource {
	return resourceXraySection(xraySectionBasics)
}

func resourceXrayDNS() *schema.Resource {
	return resourceXraySection(xraySectionDNS)
}

func resourceXrayRouting() *schema.Resource {
	return resourceXraySection(xraySectionRouting)
}

func resourceXrayBalancers() *schema.Resource {
	return resourceXraySection(xraySectionBalancers)
}

func resourceXrayReverse() *schema.Resource {
	return resourceXraySection(xraySectionReverse)
}

func resourceXrayOutbounds() *schema.Resource {
	return resourceXraySection(xraySectionOutbounds)
}

func resourceXrayAdvanced() *schema.Resource {
	return resourceXraySection(xraySectionAdvanced)
}

func resourceXraySection(section xraySection) *schema.Resource {
	diffSuppress := jsonEqualDiffSuppress
	if section.mode == xraySectionMergeRoot {
		diffSuppress = jsonSubsetDiffSuppress
	}
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return resourceXraySectionApply(ctx, d, meta, section)
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return resourceXraySectionRead(ctx, d, meta, section)
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return resourceXraySectionApply(ctx, d, meta, section)
		},
		DeleteContext: resourceSettingsDelete,
		Schema: map[string]*schema.Schema{
			"json": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppress,
				StateFunc:        normalizeJSONString,
			},
		},
	}
}

func resourceXraySectionApply(ctx context.Context, d *schema.ResourceData, meta any, section xraySection) diag.Diagnostics {
	client := meta.(*Client)
	desired, ok, err := getJSONField(d, "json")
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		if d.Id() == "" {
			d.SetId(section.id)
		}
		return resourceXraySectionRead(ctx, d, meta, section)
	}

	xrayTemplateMu.Lock()
	defer xrayTemplateMu.Unlock()

	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	updated, err := applyXraySection(current, desired, section)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.UpdateXrayTemplate(ctx, updated); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(section.id)
	return resourceXraySectionRead(ctx, d, meta, section)
}

func resourceXraySectionRead(ctx context.Context, d *schema.ResourceData, meta any, section xraySection) diag.Diagnostics {
	client := meta.(*Client)
	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	value := extractXraySection(current, section)
	payload, err := json.Marshal(value)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(payload)); err != nil {
		return diag.FromErr(err)
	}
	if d.Id() == "" {
		d.SetId(section.id)
	}
	return nil
}

func getJSONField(d *schema.ResourceData, path string) (any, bool, error) {
	v, ok := d.GetOkExists(path) //nolint:staticcheck // GetOkExists needed for zero-value vs unset
	if !ok {
		return nil, false, nil
	}
	raw, _ := v.(string)
	if strings.TrimSpace(raw) == "" {
		return nil, true, fmt.Errorf("%s must be valid JSON, got empty string", path)
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, true, fmt.Errorf("%s must be valid JSON: %w", path, err)
	}
	return out, true, nil
}

func normalizeJSONString(v any) string {
	raw, _ := v.(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return raw
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return raw
	}
	return string(encoded)
}

func jsonEqualDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	old = strings.TrimSpace(old)
	new = strings.TrimSpace(new)
	if old == "" || new == "" {
		return old == new
	}
	var o any
	var n any
	if err := json.Unmarshal([]byte(old), &o); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &n); err != nil {
		return false
	}
	return deepEqualJSON(o, n)
}

func deepEqualJSON(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(ab, bb)
}

func applyXraySection(current map[string]any, desired any, section xraySection) (map[string]any, error) {
	root := cloneJSONMap(current)
	switch section.mode {
	case xraySectionMergeRoot:
		desiredMap, ok := desired.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("json must be an object for %s", section.id)
		}
		root = deepMergeJSON(root, desiredMap)
	case xraySectionReplaceAll:
		desiredMap, ok := desired.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("json must be an object for %s", section.id)
		}
		root = desiredMap
	case xraySectionSetPath:
		if len(section.path) == 0 {
			return nil, fmt.Errorf("invalid section path for %s", section.id)
		}
		setJSONPath(root, section.path, desired)
	default:
		return nil, fmt.Errorf("unknown section mode for %s", section.id)
	}
	return root, nil
}

func extractXraySection(current map[string]any, section xraySection) any {
	switch section.mode {
	case xraySectionMergeRoot:
		out := map[string]any{}
		for _, key := range []string{"log", "policy", "routing", "outbounds"} {
			if v, ok := current[key]; ok {
				out[key] = v
			}
		}
		return out
	case xraySectionReplaceAll:
		return current
	case xraySectionSetPath:
		return getJSONPath(current, section.path)
	default:
		return nil
	}
}

func cloneJSONMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func deepMergeJSON(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if existing, ok := dst[k].(map[string]any); ok {
				dst[k] = deepMergeJSON(existing, vMap)
				continue
			}
		}
		dst[k] = v
	}
	return dst
}

func setJSONPath(root map[string]any, path []string, value any) {
	current := root
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		next, ok := current[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[key] = next
		}
		current = next
	}
	current[path[len(path)-1]] = value
}

func getJSONPath(root map[string]any, path []string) any {
	current := any(root)
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current, ok = m[key]
		if !ok {
			return nil
		}
	}
	return current
}
