package provider

import (
	"testing"
)

func TestExpandAndFlattenMetrics(t *testing.T) {
	expanded := expandBasicsMetrics(map[string]any{
		"tag":    "metrics_out",
		"listen": "127.0.0.1:11111",
	})
	if expanded["tag"] != "metrics_out" {
		t.Fatalf("unexpected tag: %v", expanded["tag"])
	}
	if expanded["listen"] != "127.0.0.1:11111" {
		t.Fatalf("unexpected listen: %v", expanded["listen"])
	}

	flattened := flattenBasicsMetrics(map[string]any{
		"tag":    "metrics_out",
		"listen": "127.0.0.1:11111",
	})
	if flattened["tag"] != "metrics_out" {
		t.Fatalf("unexpected tag: %v", flattened["tag"])
	}
	if flattened["listen"] != "127.0.0.1:11111" {
		t.Fatalf("unexpected listen: %v", flattened["listen"])
	}
}

func TestExpandBasicsMetricsEmpty(t *testing.T) {
	if expandBasicsMetrics(map[string]any{}) != nil {
		t.Fatal("expected nil for empty metrics")
	}
	if flattenBasicsMetrics(map[string]any{}) != nil {
		t.Fatal("expected nil for empty metrics")
	}
}
