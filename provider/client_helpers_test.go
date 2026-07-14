package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// Tests for Client.GetNewX25519Cert and Client.GetNewVlessEnc via httptest.
// ---------------------------------------------------------------------------

func TestGetNewX25519Cert_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getNewX25519Cert" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		certObj := map[string]any{
			"privateKey": "priv-xyz",
			"publicKey":  "pub-xyz",
		}
		w.Write(okResponse(certObj))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	result, err := client.GetNewX25519Cert(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["privateKey"] != "priv-xyz" {
		t.Fatalf("unexpected privateKey: %v", result["privateKey"])
	}
	if result["publicKey"] != "pub-xyz" {
		t.Fatalf("unexpected publicKey: %v", result["publicKey"])
	}
}

func TestGetNewX25519Cert_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(failResponse("internal error"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	_, err := client.GetNewX25519Cert(context.Background())
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestGetNewX25519Cert_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return completely invalid JSON body (not wrapped in apiResponse)
		w.Write([]byte(`not-json-at-all`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	_, err := client.GetNewX25519Cert(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestGetNewVlessEnc_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/panel/api/server/getNewVlessEnc" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		authsObj := map[string]any{
			"auths": []any{
				map[string]any{
					"label":      "label1",
					"decryption": "none",
					"encryption": "none",
				},
				map[string]any{
					"label":      "label2",
					"decryption": "none",
					"encryption": "aes-128-gcm",
				},
			},
		}
		w.Write(okResponse(authsObj))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	auths, err := client.GetNewVlessEnc(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auths) != 2 {
		t.Fatalf("expected 2 auths, got %d", len(auths))
	}
	if auths[0].Label != "label1" {
		t.Fatalf("unexpected label[0]: %v", auths[0].Label)
	}
	if auths[0].Encryption != "none" {
		t.Fatalf("unexpected encryption[0]: %v", auths[0].Encryption)
	}
	if auths[1].Label != "label2" {
		t.Fatalf("unexpected label[1]: %v", auths[1].Label)
	}
	if auths[1].Encryption != "aes-128-gcm" {
		t.Fatalf("unexpected encryption[1]: %v", auths[1].Encryption)
	}
}

func TestGetNewVlessEnc_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authsObj := map[string]any{
			"auths": []any{},
		}
		w.Write(okResponse(authsObj))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	auths, err := client.GetNewVlessEnc(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auths) != 0 {
		t.Fatalf("expected 0 auths, got %d", len(auths))
	}
}

func TestGetNewVlessEnc_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(failResponse("denied"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	_, err := client.GetNewVlessEnc(context.Background())
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestGetNewVlessEnc_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return completely invalid JSON body
		w.Write([]byte(`not-json-at-all`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)

	_, err := client.GetNewVlessEnc(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestVlessEncAuth_JSONRoundTrip(t *testing.T) {
	// Verify the struct unmarshals correctly from real 3x-ui JSON.
	raw := `[{"label":"l1","decryption":"none","encryption":"none"},{"label":"l2","decryption":"none","encryption":"xchacha20"}]`
	var auths []VlessEncAuth
	if err := json.Unmarshal([]byte(raw), &auths); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auths) != 2 {
		t.Fatalf("expected 2 auths, got %d", len(auths))
	}
	if auths[0].Label != "l1" || auths[1].Encryption != "xchacha20" {
		t.Fatalf("unexpected values: %+v", auths)
	}
}
