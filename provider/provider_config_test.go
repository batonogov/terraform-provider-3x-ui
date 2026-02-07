package provider

import (
	"testing"
	"time"
)

func TestNewClientInvalidEndpoint(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Endpoint: "not-a-url",
		Username: "admin",
		Password: "admin",
		Timeout:  time.Second,
	})
	if err == nil {
		t.Fatalf("expected error for invalid endpoint")
	}
}

func TestNewClientBasePathNormalization(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Endpoint: "http://example.com",
		BasePath: "xui",
		Username: "admin",
		Password: "admin",
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if client.basePath != "/xui/" {
		t.Fatalf("expected normalized base path '/xui/', got %q", client.basePath)
	}
}
