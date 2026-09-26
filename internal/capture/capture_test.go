package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/otrumb/safe-svc-diff/internal/snapshot"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (function roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestCapture_includes_normalized_base_path_in_endpoint_identity(t *testing.T) {
	client := emptyClient()
	document, err := Capture(context.Background(), Options{BaseURL: "https://safe.example/proxy/tx-service/", Safe: "0x1111111111111111111111111111111111111111"}, client)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	want := "/proxy/tx-service/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/"
	if document.Capture.EndpointPath != want {
		t.Fatalf("endpointPath = %q, want %q", document.Capture.EndpointPath, want)
	}
}

func TestWriteAtomic_never_replaces_concurrently_claimed_destination(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "snapshot.json")
	start := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(2)
	results := make(chan error, 2)
	for _, value := range []string{"first", "second"} {
		go func() {
			defer wait.Done()
			<-start
			results <- WriteAtomic(path, []byte(value))
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful writes = %d, want 1", successes)
	}
	raw, err := os.ReadFile(path)
	if err != nil || (string(raw) != "first" && string(raw) != "second") {
		t.Fatalf("destination = %q, %v", raw, err)
	}
}

func TestWriteAtomic_does_not_replace_existing_destination(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("replacement")); err == nil {
		t.Fatal("WriteAtomic() replaced existing destination")
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "existing" {
		t.Fatalf("destination = %q, %v", raw, err)
	}
}

func TestMergeResults_accepts_exact_limit_and_rejects_excess(t *testing.T) {
	transactions := make(map[string]snapshot.Transaction, MaxRecords)
	for index := range MaxRecords - 1 {
		hash := fmt.Sprintf("0x%064x", index+1)
		transactions[hash] = snapshot.Transaction{SafeTxHash: hash}
	}
	last := validUpstreamTransaction(fmt.Sprintf("0x%064x", MaxRecords))
	if reason := mergeResults(transactions, []upstreamTransaction{last}); reason != nil || len(transactions) != MaxRecords {
		t.Fatalf("exact limit rejected: reason %v, records %d", reason, len(transactions))
	}
	excess := validUpstreamTransaction(fmt.Sprintf("0x%064x", MaxRecords+1))
	reason := mergeResults(transactions, []upstreamTransaction{excess})
	if reason == nil || *reason != "record_limit" || len(transactions) != MaxRecords {
		t.Fatalf("excess retained: reason %v, records %d", reason, len(transactions))
	}
}

func validUpstreamTransaction(hash string) upstreamTransaction {
	gasToken := "0x0000000000000000000000000000000000000000"
	return upstreamTransaction{SafeTxHash: hash, Safe: "0x1111111111111111111111111111111111111111", To: "0x2222222222222222222222222222222222222222", Value: "0", GasToken: &gasToken, SafeTxGas: "0", BaseGas: "0", GasPrice: "0", Nonce: "0", SubmissionDate: "2026-09-25T01:00:00Z", Modified: "2026-09-25T01:00:00Z"}
}

func emptyClient() Client { return responseClient(`{"count":0,"next":null,"results":[]}`) }

func responseClient(body string) Client {
	return Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: ioNopCloser{strings.NewReader(body)}}, nil
	})}, Now: func() time.Time { return time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC) }, Sleep: func(context.Context, time.Duration) error { return nil }}
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

func TestCapture_projects_null_gas_token_from_valid_upstream_fixture(t *testing.T) {
	// Given
	body, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", "gas-token-null.json"))
	if err != nil {
		t.Fatal(err)
	}

	// When
	document, err := Capture(context.Background(), captureOptions(), responseClient(string(body)))

	// Then
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"gasToken":null`) {
		t.Fatalf("snapshot gasToken is not null: %s", raw)
	}
	canonical, err := snapshot.Canonical(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := snapshot.Validate(canonical); err != nil {
		t.Fatalf("nullable gasToken snapshot invalid: %v", err)
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
