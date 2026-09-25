package capture

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestParseQuery_rejects_reserved_parameter(t *testing.T) {
	// Given
	values := []string{"limit=10"}
	// When
	_, err := ParseQuery(values)
	// Then
	if err == nil {
		t.Fatal("ParseQuery() accepted reserved parameter")
	}
}

func TestParseQuery_defaults_to_untrusted_inclusive(t *testing.T) {
	queries, err := ParseQuery(nil)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}
	for _, query := range queries {
		if query.Name == "trusted" && query.Value == "false" {
			return
		}
	}
	t.Fatal("ParseQuery() did not default trusted=false")
}

func TestFetchPage_does_not_send_ambient_API_key(t *testing.T) {
	t.Setenv("SAFE_API_KEY", "must-not-leak")
	target, _ := url.Parse("https://attacker.example/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/")
	client := Client{HTTP: &http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "" {
			t.Fatal("ambient API key leaked")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"count":0,"next":null,"results":[]}`))}, nil
	})}, Sleep: func(context.Context, time.Duration) error { return nil }}

	_, _, err := fetchPage(context.Background(), target, client)
	if err != nil {
		t.Fatalf("fetchPage() error = %v", err)
	}
}

func TestFetchPage_classifies_failures(t *testing.T) {
	tests := []struct {
		name   string
		client Client
		want   string
	}{
		{"http failure", Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusBadRequest, Body: http.NoBody}, nil
		})}, Sleep: func(context.Context, time.Duration) error { return nil }}, "http_failure"},
		{"network failure", Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })}, Sleep: func(context.Context, time.Duration) error { return errors.New("stop") }}, "network_failure"},
		{"invalid page", Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader("{"))}, nil
		})}, Sleep: func(context.Context, time.Duration) error { return nil }}, "invalid_page"},
		{"page size limit", Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", MaxPageBytes+1)))}, nil
		})}, Sleep: func(context.Context, time.Duration) error { return nil }}, "page_size_limit"},
		{"retry exhausted", Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: http.NoBody}, nil
		})}, Sleep: func(context.Context, time.Duration) error { return nil }}, "retry_exhausted"},
	}
	target, _ := url.Parse("https://safe.example/api/v2/test")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := fetchPage(context.Background(), target, test.client)
			var failure *FetchError
			if !errors.As(err, &failure) || failure.Reason != test.want {
				t.Fatalf("fetchPage() error = %v, want reason %q", err, test.want)
			}
		})
	}
}

func TestValidateNext_rejects_cross_origin(t *testing.T) {
	// Given
	current, _ := url.Parse("https://safe.example/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/?limit=200&offset=0&trusted=true")
	// When
	_, err := ValidateNext(current, "https://evil.example/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/?limit=200&offset=200&trusted=true")
	// Then
	if err == nil {
		t.Fatal("ValidateNext() accepted cross origin")
	}
}
