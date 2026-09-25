package capture

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (function roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestCapture_projects_single_page(t *testing.T) {
	// Given
	body := `{"count":0,"next":null,"results":[]}`
	client := Client{HTTP: &http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: http.NoBody}, nil
	})}, Now: func() time.Time { return time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC) }, Sleep: func(context.Context, time.Duration) error { return nil }}
	client.HTTP.Transport = roundTrip(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: ioNopCloser{strings.NewReader(body)}}, nil
	})
	// When
	document, err := Capture(context.Background(), Options{BaseURL: "https://safe.example", Safe: "0x1111111111111111111111111111111111111111"}, client)
	// Then
	if err != nil || !document.Capture.Complete || document.Capture.RecordsFetched != 0 {
		t.Fatalf("Capture()=%+v, %v", document, err)
	}
}

func TestCapture_preserves_checksum_address_in_request_path(t *testing.T) {
	// Given
	safe := "0x5298A93734C3D979eF1f23F78eBB871879A21F22"
	var requestedPath string
	client := Client{HTTP: &http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		requestedPath = request.URL.Path
		body := `{"count":0,"next":null,"results":[]}`
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: ioNopCloser{strings.NewReader(body)}}, nil
	})}, Now: func() time.Time { return time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC) }, Sleep: func(context.Context, time.Duration) error { return nil }}

	// When
	document, err := Capture(context.Background(), Options{BaseURL: "https://safe.example", Safe: safe}, client)

	// Then
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	want := "/api/v2/safes/" + safe + "/multisig-transactions/"
	if requestedPath != want {
		t.Fatalf("request path = %q, want checksum-preserving %q", requestedPath, want)
	}
	wantEndpoint := "/api/v2/safes/" + strings.ToLower(safe) + "/multisig-transactions/"
	if document.Capture.EndpointPath != wantEndpoint {
		t.Fatalf("snapshot endpoint = %q, want canonical %q", document.Capture.EndpointPath, wantEndpoint)
	}
}

type ioNopCloser struct{ *strings.Reader }

func (ioNopCloser) Close() error { return nil }
