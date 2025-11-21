package client

import (
	"encoding/json"
	"fmt"
)

type responseEnvelope struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}

// APIError represents an error returned by the 3x-ui API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("3x-ui API error (%d): %s", e.StatusCode, e.Message)
}

// Inbound represents an inbound definition returned by the panel.
type Inbound struct {
	ID                   int             `json:"id"`
	UserID               int             `json:"userId"`
	Remark               string          `json:"remark"`
	Listen               string          `json:"listen"`
	Port                 int             `json:"port"`
	Protocol             string          `json:"protocol"`
	Settings             json.RawMessage `json:"settings"`
	StreamSettings       json.RawMessage `json:"stream"`
	Sniffing             json.RawMessage `json:"sniffing"`
	Allocate             json.RawMessage `json:"allocate"`
	Tag                  string          `json:"tag"`
	Enable               bool            `json:"enable"`
	Up                   int64           `json:"up"`
	Down                 int64           `json:"down"`
	Total                int64           `json:"total"`
	AllTime              int64           `json:"allTime"`
	ExpiryTime           int64           `json:"expiryTime"`
	TrafficReset         string          `json:"trafficReset"`
	LastTrafficResetTime int64           `json:"lastTrafficResetTime"`
	Clients              json.RawMessage `json:"clients"`
}

// InboundClient captures client info inside inbound settings.
type InboundClient struct {
	ID         string `json:"id,omitempty"`
	Security   string `json:"security,omitempty"`
	Password   string `json:"password,omitempty"`
	Flow       string `json:"flow,omitempty"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp,omitempty"`
	TotalGB    int64  `json:"totalGB,omitempty"`
	ExpiryTime int64  `json:"expiryTime,omitempty"`
	Enable     bool   `json:"enable"`
	TgID       int64  `json:"tgId,omitempty"`
	SubID      string `json:"subId,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Reset      int    `json:"reset,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
}

// ServerStatus mirrors web/service/server.go:Status.
type ServerStatus struct {
	CPU         float64   `json:"cpu"`
	CPUCores    int       `json:"cpuCores"`
	LogicalPro  int       `json:"logicalPro"`
	CPUSpeedMHz float64   `json:"cpuSpeedMhz"`
	Mem         Capacity  `json:"mem"`
	Swap        Capacity  `json:"swap"`
	Disk        Capacity  `json:"disk"`
	Xray        XrayState `json:"xray"`
	Uptime      uint64    `json:"uptime"`
	Loads       []float64 `json:"loads"`
	TCPCount    int       `json:"tcpCount"`
	UDPCount    int       `json:"udpCount"`
	NetIO       struct {
		Up   uint64 `json:"up"`
		Down uint64 `json:"down"`
	} `json:"netIO"`
	NetTraffic struct {
		Sent uint64 `json:"sent"`
		Recv uint64 `json:"recv"`
	} `json:"netTraffic"`
	PublicIP struct {
		IPv4 string `json:"ipv4"`
		IPv6 string `json:"ipv6"`
	} `json:"publicIP"`
	AppStats struct {
		Threads uint32 `json:"threads"`
		Mem     uint64 `json:"mem"`
		Uptime  uint64 `json:"uptime"`
	} `json:"appStats"`
}

// Capacity describes current/total usage metrics.
type Capacity struct {
	Current uint64 `json:"current"`
	Total   uint64 `json:"total"`
}

// XrayState represents the Xray service health.
type XrayState struct {
	State    string `json:"state"`
	ErrorMsg string `json:"errorMsg"`
	Version  string `json:"version"`
}
