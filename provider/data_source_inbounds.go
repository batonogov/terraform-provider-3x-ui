package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceInbounds() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInboundsRead,
		Schema: map[string]*schema.Schema{
			"inbounds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: inboundSchemaComputed()},
			},
		},
	}
}

func dataSourceInboundsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inbounds, err := client.GetInbounds(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	items := make([]any, 0, len(inbounds))
	for _, inbound := range inbounds {
		items = append(items, flattenInbound(inbound))
	}
	if err := d.Set("inbounds", items); err != nil {
		return diag.FromErr(err)
	}
	if len(inbounds) == 0 {
		d.SetId("0")
		return nil
	}
	d.SetId(strconv.Itoa(inbounds[0].ID))
	return nil
}

func inboundSchemaComputed() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                      {Type: schema.TypeInt, Computed: true},
		"up":                      {Type: schema.TypeInt, Computed: true},
		"down":                    {Type: schema.TypeInt, Computed: true},
		"total":                   {Type: schema.TypeInt, Computed: true},
		"all_time":                {Type: schema.TypeInt, Computed: true},
		"remark":                  {Type: schema.TypeString, Computed: true},
		"enable":                  {Type: schema.TypeBool, Computed: true},
		"expiry_time":             {Type: schema.TypeInt, Computed: true},
		"traffic_reset":           {Type: schema.TypeString, Computed: true},
		"last_traffic_reset_time": {Type: schema.TypeInt, Computed: true},
		"listen":                  {Type: schema.TypeString, Computed: true},
		"port":                    {Type: schema.TypeInt, Computed: true},
		"protocol":                {Type: schema.TypeString, Computed: true},
		"settings":                {Type: schema.TypeString, Computed: true},
		"stream_settings":         {Type: schema.TypeString, Computed: true},
		"tag":                     {Type: schema.TypeString, Computed: true},
		"sniffing":                {Type: schema.TypeString, Computed: true},
	}
}

func flattenInbound(inbound Inbound) map[string]any {
	return map[string]any{
		"id":                      inbound.ID,
		"up":                      int(inbound.Up),
		"down":                    int(inbound.Down),
		"total":                   int(inbound.Total),
		"all_time":                int(inbound.AllTime),
		"remark":                  inbound.Remark,
		"enable":                  inbound.Enable,
		"expiry_time":             int(inbound.ExpiryTime),
		"traffic_reset":           inbound.TrafficReset,
		"last_traffic_reset_time": int(inbound.LastTrafficResetTime),
		"listen":                  inbound.Listen,
		"port":                    inbound.Port,
		"protocol":                inbound.Protocol,
		"settings":                inbound.Settings,
		"stream_settings":         inbound.StreamSettings,
		"tag":                     inbound.Tag,
		"sniffing":                inbound.Sniffing,
	}
}
