package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInbound() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInboundCreate,
		ReadContext:   resourceInboundRead,
		UpdateContext: resourceInboundUpdate,
		DeleteContext: resourceInboundDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"up": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"down": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"total": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"all_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"remark": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"expiry_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"traffic_reset": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "never",
			},
			"last_traffic_reset_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"listen": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateJSONString,
				DiffSuppressFunc: jsonSubsetDiffSuppress,
			},
			"stream_settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONString,
				DiffSuppressFunc: jsonSubsetDiffSuppress,
			},
			"sniffing": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONString,
				DiffSuppressFunc: jsonSubsetDiffSuppress,
			},
			"tag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateJSONString(v interface{}, k string) (ws []string, es []error) {
	s, ok := v.(string)
	if !ok || s == "" {
		return ws, es
	}
	if _, err := ParseJSONField(s); err != nil {
		es = append(es, fmt.Errorf("%s must be valid JSON: %v", k, err))
	}
	return ws, es
}

func jsonSubsetDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	if strings.TrimSpace(new) == "" {
		return true
	}
	var desired any
	if err := json.Unmarshal([]byte(new), &desired); err != nil {
		return false
	}
	var actual any
	if err := json.Unmarshal([]byte(old), &actual); err != nil {
		return false
	}
	return isSubset(desired, actual)
}

func isSubset(desired, actual any) bool {
	switch dv := desired.(type) {
	case map[string]any:
		av, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for k, dval := range dv {
			aval, ok := av[k]
			if !ok {
				return false
			}
			if !isSubset(dval, aval) {
				return false
			}
		}
		return true
	case []any:
		av, ok := actual.([]any)
		if !ok {
			return false
		}
		if len(dv) == 0 {
			return true
		}
		if len(dv) > len(av) {
			return false
		}
		for i := range dv {
			found := false
			for j := range av {
				if isSubset(dv[i], av[j]) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(desired, actual)
	}
}

func resourceInboundCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inbound := expandInbound(d)

	created, err := client.AddInbound(ctx, inbound)
	if err != nil {
		return diag.FromErr(err)
	}
	if created != nil {
		d.SetId(fmt.Sprintf("%d", created.ID))
		return setInboundState(d, created)
	}
	return diag.Errorf("empty response from API")
}

func resourceInboundRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	inbound, err := client.GetInbound(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return setInboundState(d, inbound)
}

func resourceInboundUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	inbound := expandInbound(d)
	inbound.ID = id

	updated, err := client.UpdateInbound(ctx, inbound)
	if err != nil {
		return diag.FromErr(err)
	}
	return setInboundState(d, updated)
}

func resourceInboundDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.DeleteInbound(ctx, id); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func expandInbound(d *schema.ResourceData) *Inbound {
	return &Inbound{
		Up:                   int64(d.Get("up").(int)),
		Down:                 int64(d.Get("down").(int)),
		Total:                int64(d.Get("total").(int)),
		Remark:               d.Get("remark").(string),
		Enable:               d.Get("enable").(bool),
		ExpiryTime:           int64(d.Get("expiry_time").(int)),
		TrafficReset:         d.Get("traffic_reset").(string),
		LastTrafficResetTime: int64(d.Get("last_traffic_reset_time").(int)),
		Listen:               d.Get("listen").(string),
		Port:                 d.Get("port").(int),
		Protocol:             d.Get("protocol").(string),
		Settings:             d.Get("settings").(string),
		StreamSettings:       d.Get("stream_settings").(string),
		Sniffing:             d.Get("sniffing").(string),
	}
}

func setInboundState(d *schema.ResourceData, inbound *Inbound) diag.Diagnostics {
	if inbound == nil {
		return diag.Errorf("empty inbound")
	}
	set := func(key string, value any) {
		_ = d.Set(key, value)
	}
	set("up", int(inbound.Up))
	set("down", int(inbound.Down))
	set("total", int(inbound.Total))
	set("all_time", int(inbound.AllTime))
	set("remark", inbound.Remark)
	set("enable", inbound.Enable)
	set("expiry_time", int(inbound.ExpiryTime))
	set("traffic_reset", inbound.TrafficReset)
	set("last_traffic_reset_time", int(inbound.LastTrafficResetTime))
	set("listen", inbound.Listen)
	set("port", inbound.Port)
	set("protocol", inbound.Protocol)
	set("settings", inbound.Settings)
	set("stream_settings", inbound.StreamSettings)
	set("sniffing", inbound.Sniffing)
	set("tag", inbound.Tag)
	return nil
}

func parseID(id string) (int, error) {
	var parsed int
	_, err := fmt.Sscanf(id, "%d", &parsed)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid id: %s", id)
	}
	return parsed, nil
}
