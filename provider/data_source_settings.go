package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSettingsRead,
		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	settings, err := client.GetSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	payload, err := json.Marshal(settings)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(payload)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(int64(len(payload)), 10))
	return nil
}
