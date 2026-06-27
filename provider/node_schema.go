package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NodeResourceModel is the Terraform-level typed model for threexui_node.
// Managed attributes map to the user-editable fields of model.Node; the rest
// are observed state populated by the central panel's heartbeat probes
// (Computed). See .pi/m2-contract.md and 3x-ui-3.4.1/internal/database/model/model.go.
type NodeResourceModel struct {
	// Managed (user-editable).
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Remark              types.String   `tfsdk:"remark"`
	Scheme              types.String   `tfsdk:"scheme"`
	Address             types.String   `tfsdk:"address"`
	Port                types.Int64    `tfsdk:"port"`
	BasePath            types.String   `tfsdk:"base_path"`
	ApiToken            types.String   `tfsdk:"api_token"`
	Enable              types.Bool     `tfsdk:"enable"`
	AllowPrivateAddress types.Bool     `tfsdk:"allow_private_address"`
	TlsVerifyMode       types.String   `tfsdk:"tls_verify_mode"`
	PinnedCertSha256    types.String   `tfsdk:"pinned_cert_sha256"`
	InboundSyncMode     types.String   `tfsdk:"inbound_sync_mode"`
	InboundTags         []types.String `tfsdk:"inbound_tags"`
	OutboundTag         types.String   `tfsdk:"outbound_tag"`

	// Observed identity / heartbeat state (Computed).
	Guid          types.String  `tfsdk:"guid"`
	Status        types.String  `tfsdk:"status"`
	LastHeartbeat types.Int64   `tfsdk:"last_heartbeat"`
	LatencyMs     types.Int64   `tfsdk:"latency_ms"`
	XrayVersion   types.String  `tfsdk:"xray_version"`
	PanelVersion  types.String  `tfsdk:"panel_version"`
	CpuPct        types.Float64 `tfsdk:"cpu_pct"`
	MemPct        types.Float64 `tfsdk:"mem_pct"`
	UptimeSecs    types.Int64   `tfsdk:"uptime_secs"`
	NetUp         types.Int64   `tfsdk:"net_up"`
	NetDown       types.Int64   `tfsdk:"net_down"`
	LastError     types.String  `tfsdk:"last_error"`
	XrayState     types.String  `tfsdk:"xray_state"`
	XrayError     types.String  `tfsdk:"xray_error"`
	ConfigDirty   types.Bool    `tfsdk:"config_dirty"`
	ConfigDirtyAt types.Int64   `tfsdk:"config_dirty_at"`
	InboundCount  types.Int64   `tfsdk:"inbound_count"`
	ClientCount   types.Int64   `tfsdk:"client_count"`
	OnlineCount   types.Int64   `tfsdk:"online_count"`
	ActiveCount   types.Int64   `tfsdk:"active_count"`
	DisabledCount types.Int64   `tfsdk:"disabled_count"`
	DepletedCount types.Int64   `tfsdk:"depleted_count"`
	ParentGuid    types.String  `tfsdk:"parent_guid"`
	Transitive    types.Bool    `tfsdk:"transitive"`
	CreatedAt     types.Int64   `tfsdk:"created_at"`
	UpdatedAt     types.Int64   `tfsdk:"updated_at"`
}

// nodeResourceSchema returns the schema for the threexui_node resource.
// Kept in its own *_schema.go file per the provider's file-naming convention.
func nodeResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a remote 3x-ui panel registered as a cluster node (multi-node surface, /panel/api/nodes). Available since 3x-ui v3.0.2.\n\n" +
			"The central panel probes the node for reachability (ensureReachable) during create/update, so the node's web API must be reachable from the central panel at apply time.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Numeric node id (as a string). Import key.",
			},
			// --- Managed attributes ---
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique node name (upstream uniqueIndex). Changing this forces a new resource because the node is identified by name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remark": schema.StringAttribute{
				Optional: true,
			},
			"scheme": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https"),
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https"),
				},
				Description: "Node web API scheme. Defaults to https.",
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "Node address (host). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				Description: "Node web API port (1-65535).",
			},
			"base_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
				Description: "Node web API base path. Defaults to '/'.",
			},
			"api_token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Bearer API token used by the central panel to authenticate to the node's web API. " +
					"Required unless tls_verify_mode is 'mtls' (mTLS nodes authenticate via client certificate). " +
					"The panel returns this value raw without redaction, so it is marked Sensitive.",
			},
			"enable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the node is enabled. Defaults to true.",
			},
			"allow_private_address": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Allow the node address to resolve to a private IP.",
			},
			"tls_verify_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("verify"),
				Validators: []validator.String{
					stringvalidator.OneOf("verify", "skip", "pin", "mtls"),
				},
				Description: "TLS verification mode for the node web API. One of verify, skip, pin, mtls. Defaults to verify.",
			},
			"pinned_cert_sha256": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Pinned certificate fingerprint (SHA-256) required when tls_verify_mode is 'pin'. Returned raw by the panel, hence Sensitive.",
			},
			"inbound_sync_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "selected"),
				},
				Description: "Which inbounds to sync to the node: 'all' or 'selected' (filtered by inbound_tags). Defaults to all.",
			},
			"inbound_tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Inbound tags to sync when inbound_sync_mode is 'selected'.",
			},
			"outbound_tag": schema.StringAttribute{
				Optional:    true,
				Description: "Xray outbound/balancer tag bridging this node. Changing it causes the central panel to restart the Xray core. Leave empty to disable the bridge.",
			},

			// --- Observed state (Computed) ---
			"guid":           computedString("Remote panel stable GUID."),
			"status":         computedString("Node status: online, offline, or unknown."),
			"last_heartbeat": computedInt64("Unix seconds of the last successful heartbeat probe (0 = never)."),
			"latency_ms":     computedInt64("Last heartbeat latency in milliseconds."),
			"xray_version":   computedString("Xray version reported by the node."),
			"panel_version":  computedString("3x-ui panel version reported by the node."),
			"cpu_pct": schema.Float64Attribute{
				Computed: true,
			},
			"mem_pct": schema.Float64Attribute{
				Computed: true,
			},
			"uptime_secs":     computedInt64("Node uptime in seconds."),
			"net_up":          computedInt64("Network upload bytes."),
			"net_down":        computedInt64("Network download bytes."),
			"last_error":      computedString("Last heartbeat error message."),
			"xray_state":      computedString("Xray core state reported by the node."),
			"xray_error":      computedString("Xray core error reported by the node."),
			"config_dirty":    computedBool("Whether the node config differs from the central panel (pending sync)."),
			"config_dirty_at": computedInt64("Unix millis when config_dirty was last set."),
			"inbound_count":   computedInt64("Number of inbounds on the node."),
			"client_count":    computedInt64("Number of clients on the node."),
			"online_count":    computedInt64("Number of online clients on the node."),
			"active_count":    computedInt64("Number of active clients on the node."),
			"disabled_count":  computedInt64("Number of disabled clients on the node."),
			"depleted_count":  computedInt64("Number of depleted clients on the node."),
			"parent_guid":     computedString("Parent node GUID when surfaced as part of a node tree."),
			"transitive":      computedBool("True for read-only transitive sub-nodes surfaced from a downstream panel."),
			"created_at":      computedInt64("Node creation time (unix millis)."),
			"updated_at":      computedInt64("Node last update time (unix millis)."),
		},
	}
}

// computedString / computedInt64 / computedBool are small helpers that keep the
// observed-state block above compact. They carry UseStateForUnknown so a
// refresh after import/apply does not produce false drift.
func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: desc,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func computedInt64(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:    true,
		Description: desc,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}
}

func computedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Computed:    true,
		Description: desc,
	}
}

// (path retained for future ConfigValidators; see G6 in .pi/m2-contract.md)
var _ = path.Root
