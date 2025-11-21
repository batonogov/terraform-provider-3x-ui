package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// InboundPayload represents the payload to create/update an inbound.
type InboundPayload struct {
	Remark               string          `json:"remark"`
	Listen               string          `json:"listen"`
	Port                 int             `json:"port"`
	Protocol             string          `json:"protocol"`
	Enable               bool            `json:"enable"`
	Settings             json.RawMessage `json:"settings"`
	StreamSettings       json.RawMessage `json:"streamSettings,omitempty"`
	Sniffing             json.RawMessage `json:"sniffing,omitempty"`
	Up                   int64           `json:"up"`
	Down                 int64           `json:"down"`
	Total                int64           `json:"total"`
	ExpiryTime           int64           `json:"expiryTime"`
	TrafficReset         string          `json:"trafficReset"`
	LastTrafficResetTime int64           `json:"lastTrafficResetTime"`
}

func (c *client) CreateInbound(ctx context.Context, payload InboundPayload) (*Inbound, error) {
	var inbound Inbound
	if err := c.postJSON(ctx, "/panel/api/inbounds/add", payload, &inbound); err != nil {
		return nil, err
	}
	return &inbound, nil
}

func (c *client) UpdateInbound(ctx context.Context, id int, payload InboundPayload) (*Inbound, error) {
	var inbound Inbound
	path := fmt.Sprintf("/panel/api/inbounds/update/%d", id)
	if err := c.postJSON(ctx, path, payload, &inbound); err != nil {
		return nil, err
	}
	return &inbound, nil
}

func (c *client) DeleteInbound(ctx context.Context, id int) error {
	path := fmt.Sprintf("/panel/api/inbounds/del/%d", id)
	return c.postJSON(ctx, path, nil, nil)
}

func (c *client) GetInbound(ctx context.Context, id int) (*Inbound, error) {
	var inbound Inbound
	path := fmt.Sprintf("/panel/api/inbounds/get/%d", id)
	if err := c.getJSON(ctx, path, url.Values{}, &inbound); err != nil {
		return nil, err
	}
	return &inbound, nil
}
