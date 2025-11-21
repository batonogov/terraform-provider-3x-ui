package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/batonogov/terraform-provider-3x-ui/internal/client"
)

func TestInboundPayloadValidation(t *testing.T) {
	model := inboundResourceModel{
		Remark:       types.StringValue("demo"),
		Listen:       types.StringValue("0.0.0.0"),
		Port:         types.Int64Value(443),
		Protocol:     types.StringValue("vless"),
		Enable:       types.BoolValue(true),
		SettingsJSON: types.StringValue(`{"clients":[]}`),
	}

	payload, err := model.toPayload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.Port != 443 {
		t.Fatalf("expected port 443, got %d", payload.Port)
	}
	if string(payload.Settings) != `{"clients":[]}` {
		t.Fatalf("unexpected settings payload: %s", string(payload.Settings))
	}
}

func TestInboundPayloadInvalidPort(t *testing.T) {
	model := inboundResourceModel{
		Remark:       types.StringValue("demo"),
		Port:         types.Int64Value(70000),
		Protocol:     types.StringValue("vmess"),
		SettingsJSON: types.StringValue(`{}`),
	}

	if _, err := model.toPayload(); err == nil {
		t.Fatalf("expected error for invalid port, got nil")
	}
}

type stubClient struct {
	inbounds map[int]*client.Inbound
	clients  map[string]client.InboundClient
}

func newStubClient() *stubClient {
	return &stubClient{
		inbounds: map[int]*client.Inbound{
			1: {
				ID:       1,
				Remark:   "demo",
				Protocol: "vless",
				Settings: jsonRaw(`{"clients":[{"id":"uuid","email":"user@example.com","enable":true}]}`),
			},
		},
	}
}

func (s *stubClient) Ready(ctx context.Context) error { return nil }
func (s *stubClient) ListInbounds(ctx context.Context) ([]client.Inbound, error) {
	var result []client.Inbound
	for _, inb := range s.inbounds {
		result = append(result, *inb)
	}
	return result, nil
}
func (s *stubClient) ServerStatus(ctx context.Context) (*client.ServerStatus, error) {
	return &client.ServerStatus{}, nil
}
func (s *stubClient) GetInbound(ctx context.Context, id int) (*client.Inbound, error) {
	inb, ok := s.inbounds[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return inb, nil
}
func (s *stubClient) CreateInbound(ctx context.Context, payload client.InboundPayload) (*client.Inbound, error) {
	inb := &client.Inbound{
		ID:       len(s.inbounds) + 1,
		Remark:   payload.Remark,
		Protocol: payload.Protocol,
		Settings: payload.Settings,
	}
	s.inbounds[inb.ID] = inb
	return inb, nil
}
func (s *stubClient) UpdateInbound(ctx context.Context, id int, payload client.InboundPayload) (*client.Inbound, error) {
	inb, ok := s.inbounds[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	inb.Remark = payload.Remark
	inb.Settings = payload.Settings
	return inb, nil
}
func (s *stubClient) DeleteInbound(ctx context.Context, id int) error {
	delete(s.inbounds, id)
	return nil
}
func (s *stubClient) AddClient(ctx context.Context, inboundID int, c client.InboundClient) error {
	return nil
}
func (s *stubClient) UpdateClient(ctx context.Context, inboundID int, clientID string, c client.InboundClient) error {
	return nil
}
func (s *stubClient) DeleteClient(ctx context.Context, inboundID int, clientID string) error {
	return nil
}

func jsonRaw(s string) json.RawMessage {
	return json.RawMessage([]byte(s))
}
