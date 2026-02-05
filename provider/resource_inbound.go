package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem:     &schema.Resource{Schema: settingsSchema()},
			},
			"stream_settings": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem:     &schema.Resource{Schema: streamSettingsSchema()},
			},
			"sniffing": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"dest_override": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"metadata_only": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"route_only": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				}},
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
	if err := ensureVlessEncFromAuth(ctx, client, d, inbound); err != nil {
		return diag.FromErr(err)
	}
	if err := applyDefaultInboundSettings(inbound); err != nil {
		return diag.FromErr(err)
	}
	if err := ensureRealityKeys(ctx, client, inbound, nil); err != nil {
		return diag.FromErr(err)
	}
	if err := ensureInboundClientIDs(inbound); err != nil {
		return diag.FromErr(err)
	}

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
	existing, err := client.GetInbound(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	inbound := expandInbound(d)
	if err := ensureVlessEncFromAuth(ctx, client, d, inbound); err != nil {
		return diag.FromErr(err)
	}
	if err := applyDefaultInboundSettings(inbound); err != nil {
		return diag.FromErr(err)
	}
	if err := ensureRealityKeys(ctx, client, inbound, existing); err != nil {
		return diag.FromErr(err)
	}
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
		Settings:             buildSettingsJSON(d),
		StreamSettings:       buildStreamSettingsJSON(d),
		Sniffing:             buildSniffingJSON(d),
	}
}

func settingsKeyMissing(d *schema.ResourceData, key string) bool {
	raw, ok := d.GetOk("settings")
	if !ok {
		return true
	}
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return true
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		return true
	}
	_, ok = item[key]
	return !ok
}

func ensureVlessEncFromAuth(ctx context.Context, client *Client, d *schema.ResourceData, inbound *Inbound) error {
	if inbound == nil || inbound.Protocol != "vless" || client == nil {
		return nil
	}
	if strings.TrimSpace(inbound.Settings) == "" {
		return nil
	}
	settings, err := ParseJSONField(inbound.Settings)
	if err != nil {
		return err
	}
	selected := stringValue(settings["selectedAuth"])
	if selected == "" {
		return nil
	}

	decryptionMissing := settingsKeyMissing(d, "decryption") || stringValue(settings["decryption"]) == ""
	encryptionMissing := settingsKeyMissing(d, "encryption") || stringValue(settings["encryption"]) == ""
	if !decryptionMissing && !encryptionMissing {
		return nil
	}

	auths, err := client.GetNewVlessEnc(ctx)
	if err != nil {
		return err
	}
	var match *VlessEncAuth
	for i := range auths {
		if auths[i].Label == selected {
			match = &auths[i]
			break
		}
	}
	if match == nil {
		return fmt.Errorf("no auth block for selected_auth %q", selected)
	}
	if decryptionMissing {
		settings["decryption"] = match.Decryption
	}
	if encryptionMissing {
		settings["encryption"] = match.Encryption
	}
	updated, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(updated)
	return nil
}

func ensureInboundClientIDs(inbound *Inbound) error {
	if inbound == nil {
		return nil
	}
	settings, err := ParseJSONField(inbound.Settings)
	if err != nil {
		return err
	}
	clientsRaw, ok := settings["clients"]
	if !ok {
		return nil
	}
	clients, ok := clientsRaw.([]any)
	if !ok {
		return nil
	}
	changed := false
	for i := range clients {
		clientMap, ok := clients[i].(map[string]any)
		if !ok {
			continue
		}
		id, _ := clientMap["id"].(string)
		if id == "" {
			newID, err := newUUID()
			if err != nil {
				return err
			}
			clientMap["id"] = newID
			clients[i] = clientMap
			changed = true
		}
	}
	if !changed {
		return nil
	}
	settings["clients"] = clients
	updated, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound.Settings = string(updated)
	return nil
}

func ensureRealityKeys(ctx context.Context, client *Client, inbound *Inbound, existing *Inbound) error {
	if inbound == nil || inbound.StreamSettings == "" {
		return nil
	}
	payload, err := ParseJSONField(inbound.StreamSettings)
	if err != nil {
		return err
	}
	security := stringValue(payload["security"])
	if security != "reality" {
		return nil
	}
	rs, _ := payload["realitySettings"].(map[string]any)
	if rs == nil {
		rs = map[string]any{}
	}
	mergeRealityFromExisting(existing, rs)
	ensureRealityDefaults(rs)
	if !hasRealityShortIDs(rs) {
		rs["shortIds"] = randomShortIDs()
	}
	if pk, ok := rs["privateKey"].(string); ok && pk != "" {
		return nil
	}
	cert, err := client.GetNewX25519Cert(ctx)
	if err != nil {
		return err
	}
	privateKey := stringValue(cert["privateKey"])
	publicKey := stringValue(cert["publicKey"])
	if privateKey == "" {
		return fmt.Errorf("generated reality privateKey is empty")
	}
	rs["privateKey"] = privateKey
	settings, _ := rs["settings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
	}
	if settings["publicKey"] == nil || stringValue(settings["publicKey"]) == "" {
		settings["publicKey"] = publicKey
	}
	rs["settings"] = settings
	payload["realitySettings"] = rs
	updated, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	inbound.StreamSettings = string(updated)
	return nil
}

func ensureRealityDefaults(reality map[string]any) {
	if reality == nil {
		return
	}
	if hasStringListValues(reality["serverNames"]) {
		return
	}
	target := stringValue(reality["target"])
	if target != "" {
		host := strings.Split(target, ":")[0]
		if host != "" {
			reality["serverNames"] = []any{host}
			return
		}
	}
	reality["target"] = "www.apple.com:443"
	reality["serverNames"] = []any{"www.apple.com", "apple.com"}
}

func mergeRealityFromExisting(existing *Inbound, reality map[string]any) {
	if existing == nil || existing.StreamSettings == "" {
		return
	}
	payload, err := ParseJSONField(existing.StreamSettings)
	if err != nil {
		return
	}
	rs, _ := payload["realitySettings"].(map[string]any)
	if rs == nil {
		return
	}
	if stringValue(reality["privateKey"]) == "" {
		if pk := stringValue(rs["privateKey"]); pk != "" {
			reality["privateKey"] = pk
		}
	}
	if !hasRealityShortIDs(reality) {
		if raw, ok := rs["shortIds"]; ok {
			if list, ok := raw.([]any); ok && len(list) > 0 {
				reality["shortIds"] = list
			}
		}
	}
	settings, _ := reality["settings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
	}
	if stringValue(settings["publicKey"]) == "" {
		if s, ok := rs["settings"].(map[string]any); ok {
			if pk := stringValue(s["publicKey"]); pk != "" {
				settings["publicKey"] = pk
			}
		}
	}
	reality["settings"] = settings
}

func hasRealityShortIDs(reality map[string]any) bool {
	if reality == nil {
		return false
	}
	return hasStringListValues(reality["shortIds"])
}

func hasRealityServerNames(reality map[string]any) bool {
	if reality == nil {
		return false
	}
	return hasStringListValues(reality["serverNames"])
}

func hasStringListValues(raw any) bool {
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s != "" {
				return true
			}
		}
	}
	return false
}

func randomHex(length int) string {
	if length <= 0 {
		return ""
	}
	buf := make([]byte, (length+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	out := hex.EncodeToString(buf)
	if len(out) > length {
		out = out[:length]
	}
	return out
}

func randomShortIDs() []any {
	lengths := []int{2, 4, 6, 8, 10, 12, 14, 16}
	out := make([]any, 0, len(lengths))
	for _, l := range lengths {
		out = append(out, randomHex(l))
	}
	return out
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
	set("settings", flattenSettings(inbound.Settings))
	set("stream_settings", flattenStreamSettings(inbound.StreamSettings))
	set("sniffing", flattenSniffing(inbound.Sniffing))
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
