package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServerStatus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerStatusRead,
		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServerStatusRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	status, err := client.GetServerStatus(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	payload, err := json.Marshal(status)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(payload)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(int64(len(payload)), 10))
	return nil
}
