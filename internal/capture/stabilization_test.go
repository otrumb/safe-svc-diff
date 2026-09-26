package capture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type scriptedResponse struct {
	body string
	err  error
}

func TestCapture_marks_stable_two_pass_capture_complete(t *testing.T) {
	client, calls := sequenceClient(t, scriptedResponse{body: transactionPage("a")}, scriptedResponse{body: transactionPage("a")})

	document, err := Capture(context.Background(), captureOptions(), client)

	if err != nil || !document.Capture.Complete || *calls != 2 {
		t.Fatalf("Capture() complete=%v calls=%d error=%v", document.Capture.Complete, *calls, err)
	}
}

func TestCapture_marks_same_count_replacement_incomplete(t *testing.T) {
	client, calls := sequenceClient(t, scriptedResponse{body: transactionPage("a")}, scriptedResponse{body: transactionPage("b")})

	document, err := Capture(context.Background(), captureOptions(), client)

	if err != nil || document.Capture.Complete || document.Capture.IncompleteReason == nil || *document.Capture.IncompleteReason != "unstable_pagination" || *calls != 2 {
		t.Fatalf("Capture() complete=%v reason=%v calls=%d error=%v", document.Capture.Complete, document.Capture.IncompleteReason, *calls, err)
	}
}

func TestCapture_accepts_reordered_equivalent_verification_set(t *testing.T) {
	client, calls := sequenceClient(t, scriptedResponse{body: transactionPairPage("a", "b")}, scriptedResponse{body: transactionPairPage("b", "a")})

	document, err := Capture(context.Background(), captureOptions(), client)

	if err != nil || !document.Capture.Complete || *calls != 2 {
		t.Fatalf("Capture() complete=%v calls=%d error=%v", document.Capture.Complete, *calls, err)
	}
}

func TestCapture_marks_verification_failure_incomplete_without_retry_loop(t *testing.T) {
	client, calls := sequenceClient(t, scriptedResponse{body: transactionPage("a")}, scriptedResponse{err: errors.New("verification offline")})

	document, err := Capture(context.Background(), captureOptions(), client)

	if err != nil || document.Capture.Complete || document.Capture.IncompleteReason == nil || *document.Capture.IncompleteReason != "network_failure" || *calls != 2 {
		t.Fatalf("Capture() complete=%v reason=%v calls=%d error=%v", document.Capture.Complete, document.Capture.IncompleteReason, *calls, err)
	}
}

func captureOptions() Options {
	return Options{BaseURL: "https://safe.example", Safe: "0x1111111111111111111111111111111111111111"}
}

func sequenceClient(t *testing.T, responses ...scriptedResponse) (Client, *int) {
	t.Helper()
	calls := 0
	client := Client{HTTP: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		if calls == len(responses) {
			t.Fatalf("unexpected request %d", calls+1)
		}
		response := responses[calls]
		calls++
		if response.err != nil {
			return nil, response.err
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response.body))}, nil
	})}, Now: func() time.Time { return time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC) }, Sleep: func(context.Context, time.Duration) error { return errors.New("stop retries") }}
	return client, &calls
}

func transactionPage(hash string) string {
	return fmt.Sprintf(`{"count":1,"next":null,"results":[%s]}`, transactionJSON(hash))
}

func transactionPairPage(first, second string) string {
	return fmt.Sprintf(`{"count":2,"next":null,"results":[%s,%s]}`, transactionJSON(first), transactionJSON(second))
}

func transactionJSON(hash string) string {
	return fmt.Sprintf(`{"safeTxHash":"0x%s","safe":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222","value":"0","data":null,"operation":0,"gasToken":"0x0000000000000000000000000000000000000000","safeTxGas":"0","baseGas":"0","gasPrice":"0","refundReceiver":null,"nonce":"0","executionDate":null,"submissionDate":"2026-09-25T01:00:00Z","modified":"2026-09-25T01:00:00Z","blockNumber":null,"transactionHash":null,"proposer":null,"proposedByDelegate":null,"executor":null,"isExecuted":false,"isSuccessful":null,"ethGasPrice":null,"maxFeePerGas":null,"maxPriorityFeePerGas":null,"gasUsed":null,"fee":null,"payment":null,"confirmationsRequired":null,"confirmations":[],"trusted":true}`, strings.Repeat(hash, 64))
}
