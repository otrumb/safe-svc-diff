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

type ioNopCloser struct{ *strings.Reader }

func (ioNopCloser) Close() error { return nil }
