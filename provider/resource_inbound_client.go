package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInboundClient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInboundClientCreate,
		ReadContext:   resourceInboundClientRead,
		UpdateContext: resourceInboundClientUpdate,
		DeleteContext: resourceInboundClientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"inbound_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"security": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"flow": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"limit_ip": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"total_gb": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"expiry_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"tg_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"sub_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"reset": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceInboundClientCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inboundID := d.Get("inbound_id").(int)
	clientData := expandInboundClient(d)
	clientID := getClientID(d, clientData)
	if clientID == "" {
		return diag.Errorf("client_id is required")
	}
	clientData["id"] = clientID

	if err := client.AddInboundClient(ctx, inboundID, clientData); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(makeInboundClientID(inboundID, clientID))
	return resourceInboundClientRead(ctx, d, meta)
}

func resourceInboundClientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inboundID, clientID, err := splitInboundClientID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	inbound, err := client.GetInbound(ctx, inboundID)
	if err != nil {
		return diag.FromErr(err)
	}

	settings, err := parseInboundSettings(inbound.Settings)
	if err != nil {
		return diag.FromErr(err)
	}

	found := findClientByID(settings.Clients, clientID)
	if found == nil {
		d.SetId("")
		return nil
	}

	setInboundClientState(d, inboundID, clientID, found)
	return nil
}

func resourceInboundClientUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inboundID, clientID, err := splitInboundClientID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	clientData := expandInboundClient(d)
	clientData["id"] = clientID

	if err := client.UpdateInboundClient(ctx, inboundID, clientID, clientData); err != nil {
		return diag.FromErr(err)
	}
	return resourceInboundClientRead(ctx, d, meta)
}

func resourceInboundClientDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Client)
	inboundID, clientID, err := splitInboundClientID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.DeleteInboundClient(ctx, inboundID, clientID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func expandInboundClient(d *schema.ResourceData) map[string]any {
	client := map[string]any{}
	if v, ok := d.GetOk("email"); ok {
		client["email"] = v.(string)
	}
	if v, ok := d.GetOk("security"); ok {
		client["security"] = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		client["password"] = v.(string)
	}
	if v, ok := d.GetOk("flow"); ok {
		client["flow"] = v.(string)
	}
	if v, ok := d.GetOk("limit_ip"); ok {
		client["limitIp"] = v.(int)
	}
	if v, ok := d.GetOk("total_gb"); ok {
		client["totalGB"] = v.(int)
	}
	if v, ok := d.GetOk("expiry_time"); ok {
		client["expiryTime"] = v.(int)
	}
	if v, ok := d.GetOkExists("enable"); ok {
		client["enable"] = v.(bool)
	}
	if v, ok := d.GetOk("tg_id"); ok {
		client["tgId"] = v.(int)
	}
	if v, ok := d.GetOk("sub_id"); ok {
		client["subId"] = v.(string)
	}
	if v, ok := d.GetOk("comment"); ok {
		client["comment"] = v.(string)
	}
	if v, ok := d.GetOk("reset"); ok {
		client["reset"] = v.(int)
	}
	if v, ok := d.GetOk("client_id"); ok {
		client["id"] = v.(string)
	}
	return client
}

func setInboundClientState(d *schema.ResourceData, inboundID int, clientID string, client map[string]any) {
	_ = d.Set("inbound_id", inboundID)
	_ = d.Set("client_id", clientID)
	_ = d.Set("email", stringValue(client["email"]))
	_ = d.Set("security", stringValue(client["security"]))
	_ = d.Set("password", stringValue(client["password"]))
	_ = d.Set("flow", stringValue(client["flow"]))
	_ = d.Set("limit_ip", intValue(client["limitIp"]))
	_ = d.Set("total_gb", intValue(client["totalGB"]))
	_ = d.Set("expiry_time", intValue(client["expiryTime"]))
	_ = d.Set("enable", boolValue(client["enable"]))
	_ = d.Set("tg_id", intValue(client["tgId"]))
	_ = d.Set("sub_id", stringValue(client["subId"]))
	_ = d.Set("comment", stringValue(client["comment"]))
	_ = d.Set("reset", intValue(client["reset"]))
}

type inboundSettings struct {
	Clients []map[string]any `json:"clients"`
}

func parseInboundSettings(settings string) (*inboundSettings, error) {
	if strings.TrimSpace(settings) == "" {
		return &inboundSettings{}, nil
	}
	var out inboundSettings
	if err := json.Unmarshal([]byte(settings), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func findClientByID(clients []map[string]any, clientID string) map[string]any {
	for _, client := range clients {
		if stringValue(client["id"]) == clientID || stringValue(client["password"]) == clientID || stringValue(client["email"]) == clientID {
			return client
		}
	}
	return nil
}

func getClientID(d *schema.ResourceData, client map[string]any) string {
	if v, ok := d.GetOk("client_id"); ok {
		return v.(string)
	}
	if v := stringValue(client["id"]); v != "" {
		return v
	}
	if v := stringValue(client["password"]); v != "" {
		return v
	}
	if v := stringValue(client["email"]); v != "" {
		return v
	}
	return ""
}

func makeInboundClientID(inboundID int, clientID string) string {
	return fmt.Sprintf("%d:%s", inboundID, clientID)
}

func splitInboundClientID(id string) (int, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid inbound_client id: %s", id)
	}
	var inboundID int
	if _, err := fmt.Sscanf(parts[0], "%d", &inboundID); err != nil || inboundID == 0 {
		return 0, "", fmt.Errorf("invalid inbound id: %s", id)
	}
	if parts[1] == "" {
		return 0, "", fmt.Errorf("invalid client id: %s", id)
	}
	return inboundID, parts[1], nil
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func intValue(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		return 0
	}
}

func boolValue(v any) bool {
	b, _ := v.(bool)
	return b
}
