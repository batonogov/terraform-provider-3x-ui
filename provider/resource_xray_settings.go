package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var xrayTemplateMu sync.Mutex

type xraySectionMode int

const (
	xraySectionMergeRoot xraySectionMode = iota
	xraySectionSetPath
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
)

// buildFunc builds a JSON-compatible value from ResourceData.
// For map-based sections it returns map[string]any, for array-based (outbounds, balancers) it returns []any.
type buildFunc func(d *schema.ResourceData) (any, error)

// flattenFunc reads API data and sets attributes on ResourceData.
type flattenFunc func(d *schema.ResourceData, data any) error

func resourceXrayBasics() *schema.Resource {
	return xrayResource(xraySectionBasics, xrayBasicsSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayBasicsJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayBasicsToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func resourceXrayDNS() *schema.Resource {
	return xrayResource(xraySectionDNS, xrayDNSSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayDNSJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayDNSToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func resourceXrayRouting() *schema.Resource {
	return xrayResource(xraySectionRouting, xrayRoutingSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayRoutingJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayRoutingToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func resourceXrayBalancers() *schema.Resource {
	return xrayResource(xraySectionBalancers, xrayBalancersSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayBalancersJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayBalancersToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func resourceXrayReverse() *schema.Resource {
	return xrayResource(xraySectionReverse, xrayReverseSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayReverseJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayReverseToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func resourceXrayOutbounds() *schema.Resource {
	return xrayResource(xraySectionOutbounds, xrayOutboundsSchema(),
		func(d *schema.ResourceData) (any, error) {
			return buildXrayOutboundsJSON(d)
		},
		func(d *schema.ResourceData, data any) error {
			flat := flattenXrayOutboundsToMap(data)
			for k, v := range flat {
				if err := d.Set(k, v); err != nil {
					return fmt.Errorf("setting %s: %w", k, err)
				}
			}
			return nil
		},
	)
}

func xrayResource(section xraySection, s map[string]*schema.Schema, build buildFunc, flatten flattenFunc) *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return xrayApply(ctx, d, meta, section, build, flatten)
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return xrayRead(ctx, d, meta, section, flatten)
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			return xrayApply(ctx, d, meta, section, build, flatten)
		},
		DeleteContext: resourceSettingsDelete,
		Schema:        s,
	}
}

func xrayApply(ctx context.Context, d *schema.ResourceData, meta any, section xraySection, build buildFunc, flatten flattenFunc) diag.Diagnostics {
	client := meta.(*Client)

	desired, err := build(d)
	if err != nil {
		return diag.FromErr(err)
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
	return xrayRead(ctx, d, meta, section, flatten)
}

func xrayRead(ctx context.Context, d *schema.ResourceData, meta any, section xraySection, flatten flattenFunc) diag.Diagnostics {
	client := meta.(*Client)
	current, err := client.GetXrayTemplate(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	value := extractXraySection(current, section)
	if err := flatten(d, value); err != nil {
		return diag.FromErr(err)
	}
	if d.Id() == "" {
		d.SetId(section.id)
	}
	return nil
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
			return nil, fmt.Errorf("value must be an object for %s", section.id)
		}
		root = deepMergeJSON(root, desiredMap)
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
		for _, key := range []string{"log", "policy", "api", "stats"} {
			if v, ok := current[key]; ok {
				out[key] = v
			}
		}
		return out
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
