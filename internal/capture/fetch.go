package capture

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FetchError struct {
	Reason string
	Detail string
}

func (failure *FetchError) Error() string { return failure.Reason + ": " + failure.Detail }

func fetchPage(ctx context.Context, target *url.URL, client Client) (page, int, error) {
	lastReason := "network_failure"
	for attempt := 0; attempt < 4; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
		if err != nil {
			return page{}, 0, &FetchError{Reason: "invalid_page", Detail: err.Error()}
		}
		response, requestErr := client.HTTP.Do(request)
		if requestErr == nil {
			result, size, reason := decodeResponse(response)
			if reason == "" {
				return result, size, nil
			}
			if reason != "retry_exhausted" {
				return page{}, 0, &FetchError{Reason: reason, Detail: http.StatusText(response.StatusCode)}
			}
			lastReason = "retry_exhausted"
		}
		if attempt < 3 {
			if sleepErr := client.Sleep(ctx, time.Duration(1<<attempt)*500*time.Millisecond); sleepErr != nil {
				return page{}, 0, &FetchError{Reason: "network_failure", Detail: sleepErr.Error()}
			}
		}
	}
	return page{}, 0, &FetchError{Reason: lastReason, Detail: "attempts exhausted"}
}

func decodeResponse(response *http.Response) (page, int, string) {
	body, readErr := io.ReadAll(io.LimitReader(response.Body, MaxPageBytes+1))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		return page{}, 0, "network_failure"
	}
	if len(body) > MaxPageBytes {
		return page{}, 0, "page_size_limit"
	}
	if response.StatusCode != http.StatusOK {
		if retryable(response.StatusCode) {
			return page{}, 0, "retry_exhausted"
		}
		return page{}, 0, "http_failure"
	}
	if !strings.Contains(response.Header.Get("Content-Type"), "json") {
		return page{}, 0, "invalid_page"
	}
	var result page
	if err := json.Unmarshal(body, &result); err != nil {
		return page{}, 0, "invalid_page"
	}
	return result, len(body), ""
}

func retryable(status int) bool {
	return status == 408 || status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}
