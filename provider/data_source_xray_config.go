package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXrayConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceXrayConfigRead,
		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceXrayConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	config, err := client.GetXrayConfig(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	payload, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(payload)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(int64(len(payload)), 10))
	return nil
}
