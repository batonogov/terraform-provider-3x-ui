package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXrayVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceXrayVersionsRead,
		Schema: map[string]*schema.Schema{
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceXrayVersionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	versions, err := client.GetXrayVersions(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	items := make([]any, 0, len(versions))
	for _, v := range versions {
		items = append(items, v)
	}
	if err := d.Set("versions", items); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(len(versions)))
	return nil
}
