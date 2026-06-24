package provider

import (
	"encoding/json"
	"strings"
)

// Inbound represents a 3x-ui inbound (model.Inbound).
// Note: settings/streamSettings/sniffing are JSON strings internally.
// Custom UnmarshalJSON handles both the legacy format (escaped JSON strings)
// used by v2.9.x/v3.0.x and the modern format (nested JSON objects) used by v3.1.0+.
type Inbound struct {
	ID                   int             `json:"id"`
	Up                   int64           `json:"up"`
	Down                 int64           `json:"down"`
	Total                int64           `json:"total"`
	AllTime              int64           `json:"allTime"`
	Remark               string          `json:"remark"`
	Enable               bool            `json:"enable"`
	ExpiryTime           int64           `json:"expiryTime"`
	TrafficReset         string          `json:"trafficReset"`
	LastTrafficResetTime int64           `json:"lastTrafficResetTime"`
	ClientStats          []ClientTraffic `json:"clientStats"`

	Listen            string `json:"listen"`
	Port              int    `json:"port"`
	Protocol          string `json:"protocol"`
	Settings          string `json:"settings"`
	StreamSettings    string `json:"streamSettings"`
	Tag               string `json:"tag"`
	SubSortIndex      int    `json:"subSortIndex,omitempty"`
	ShareAddr         string `json:"shareAddr,omitempty"`
	ShareAddrStrategy string `json:"shareAddrStrategy,omitempty"`
	Sniffing          string `json:"sniffing"`
	NodeID            *int   `json:"nodeId,omitempty"`
}

func (i *Inbound) UnmarshalJSON(data []byte) error {
	type Alias Inbound
	aux := &struct {
		Settings       json.RawMessage `json:"settings"`
		StreamSettings json.RawMessage `json:"streamSettings"`
		Sniffing       json.RawMessage `json:"sniffing"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	i.Settings = rawJSONToString(aux.Settings)
	i.StreamSettings = rawJSONToString(aux.StreamSettings)
	i.Sniffing = rawJSONToString(aux.Sniffing)
	return nil
}

// rawJSONToString normalises a JSON field that may be either a JSON string
// (legacy v2.9.x/v3.0.x: "settings":"{\"clients\":[]}") or a raw JSON
// object/array (v3.1.0+: "settings":{"clients":[]}) back to the plain
// string the provider uses internally.
func rawJSONToString(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	if len(trimmed) >= 2 && trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
	}
	return trimmed
}

// ClientTraffic represents traffic statistics for a client (xray.ClientTraffic).
type ClientTraffic struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	UUID       string `json:"uuid"`
	SubID      string `json:"subId"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	AllTime    int64  `json:"allTime"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int    `json:"reset"`
	LastOnline int64  `json:"lastOnline"`
}

// ParseJSONField decodes a JSON string into a generic map.
func ParseJSONField(value string) (map[string]any, error) {
	var out map[string]any
	if value == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, err
	}
	return out, nil
}
