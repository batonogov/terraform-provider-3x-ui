package provider

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginWithTwoFactor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("twoFactorCode") != "123456" {
			w.Write(failResponse("missing 2fa"))
			return
		}
		w.Write(okResponse(nil))
	}))
	defer srv.Close()

	client, err := NewClient(ClientConfig{
		Endpoint:           srv.URL,
		BasePath:           "/",
		Username:           "admin",
		Password:           "admin",
		TwoFactorCode:      "123456",
		InsecureSkipVerify: true,
		Timeout:            2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("login failed: %v", err)
	}
}

func TestDecodeAPIResponse_EmptyBodyStatusError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioNopCloser(bytes.NewBuffer(nil)),
	}
	if err := decodeAPIResponse(resp, nil); err == nil {
		t.Fatalf("expected error for empty body with 400")
	}
}

func TestDecodeAPIResponse_EmptyBodyOK(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioNopCloser(bytes.NewBuffer(nil)),
	}
	if err := decodeAPIResponse(resp, nil); err != nil {
		t.Fatalf("unexpected error for empty body with 200: %v", err)
	}
}

func TestDecodeAPIResponse_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioNopCloser(bytes.NewBufferString("{")),
	}
	if err := decodeAPIResponse(resp, nil); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestDecodeAPIResponse_InvalidJSONWithStatusError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioNopCloser(bytes.NewBufferString("{")),
	}
	if err := decodeAPIResponse(resp, nil); err == nil {
		t.Fatalf("expected error for invalid json with status error")
	}
}

func TestDecodeAPIResponse_FailMsg(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioNopCloser(bytes.NewBuffer(failResponse("boom"))),
	}
	if err := decodeAPIResponse(resp, nil); err == nil {
		t.Fatalf("expected error for success=false with msg")
	}
}

func TestDecodeAPIResponse_FailNoMsg(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioNopCloser(bytes.NewBuffer([]byte(`{"success":false}`))),
	}
	if err := decodeAPIResponse(resp, nil); err == nil {
		t.Fatalf("expected error for success=false without msg")
	}
}

func ioNopCloser(buf *bytes.Buffer) *readNopCloser {
	return &readNopCloser{buf: buf}
}

type readNopCloser struct {
	buf *bytes.Buffer
}

func (r *readNopCloser) Read(p []byte) (int, error) {
	return r.buf.Read(p)
}

func (r *readNopCloser) Close() error {
	return nil
}
