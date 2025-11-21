package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type clientSettingsWrapper struct {
	Clients []InboundClient `json:"clients"`
}

func (c *client) AddClient(ctx context.Context, inboundID int, inboundClient InboundClient) error {
	payload, err := wrapClientSettings(inboundID, inboundClient)
	if err != nil {
		return err
	}
	return c.postJSON(ctx, "/panel/api/inbounds/addClient", payload, nil)
}

func (c *client) UpdateClient(ctx context.Context, inboundID int, clientID string, inboundClient InboundClient) error {
	payload, err := wrapClientSettings(inboundID, inboundClient)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/panel/api/inbounds/updateClient/%s", clientID)
	return c.postJSON(ctx, path, payload, nil)
}

func (c *client) DeleteClient(ctx context.Context, inboundID int, clientID string) error {
	path := fmt.Sprintf("/panel/api/inbounds/%d/delClient/%s", inboundID, clientID)
	return c.postJSON(ctx, path, nil, nil)
}

func wrapClientSettings(inboundID int, inboundClient InboundClient) (map[string]any, error) {
	wrapper := clientSettingsWrapper{
		Clients: []InboundClient{inboundClient},
	}
	raw, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("marshal client payload: %w", err)
	}
	return map[string]any{
		"id":       inboundID,
		"settings": string(raw),
	}, nil
}
