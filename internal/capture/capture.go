package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/otrumb/safe-svc-diff/internal/snapshot"
)

type Client struct {
	HTTP  *http.Client
	Now   func() time.Time
	Sleep func(context.Context, time.Duration) error
}

func Capture(ctx context.Context, options Options, client Client) (snapshot.Document, error) {
	base, err := url.Parse(options.BaseURL)
	if err != nil || base.Scheme != "https" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return snapshot.Document{}, ErrUnsafe
	}
	if !addressPattern.MatchString(options.Safe) {
		return snapshot.Document{}, ErrUnsafe
	}
	queries, err := ParseQuery(options.Queries)
	if err != nil {
		return snapshot.Document{}, err
	}
	safe := strings.ToLower(options.Safe)
	endpoint := "/api/v2/safes/" + safe + "/multisig-transactions/"
	current := *base
	current.Path = strings.TrimRight(base.Path, "/") + endpoint
	query := current.Query()
	for _, item := range queries {
		query.Set(item.Name, item.Value)
	}
	current.RawQuery = query.Encode()
	started := client.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
	transactions := map[string]snapshot.Transaction{}
	totalBytes := 0
	pages := 0
	var advertised *int
	var incomplete *string
	for {
		if pages >= MaxPages {
			reason := "page_limit"
			incomplete = &reason
			break
		}
		result, size, fetchErr := fetchPage(ctx, &current, client)
		if fetchErr != nil {
			if pages == 0 {
				return snapshot.Document{}, fetchErr
			}
			reason := "retry_exhausted"
			incomplete = &reason
			break
		}
		totalBytes += size
		if totalBytes > MaxTotalBytes {
			reason := "total_size_limit"
			incomplete = &reason
			break
		}
		pages++
		if advertised == nil {
			count := result.Count
			advertised = &count
		} else if *advertised != result.Count {
			reason := "unstable_pagination"
			incomplete = &reason
			break
		}
		for _, input := range result.Results {
			tx, projectErr := project(input)
			if projectErr != nil {
				reason := "invalid_page"
				incomplete = &reason
				break
			}
			previous, exists := transactions[tx.SafeTxHash]
			if exists && fmt.Sprintf("%#v", previous) != fmt.Sprintf("%#v", tx) {
				reason := "pagination_conflict"
				incomplete = &reason
				break
			}
			transactions[tx.SafeTxHash] = tx
		}
		if incomplete != nil {
			break
		}
		if len(transactions) > MaxRecords {
			reason := "record_limit"
			incomplete = &reason
			break
		}
		if result.Next == nil {
			break
		}
		next, nextErr := ValidateNext(&current, *result.Next)
		if nextErr != nil {
			reason := "unsafe_next_url"
			incomplete = &reason
			break
		}
		current = *next
	}
	list := make([]snapshot.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		list = append(list, transaction)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].SafeTxHash < list[j].SafeTxHash })
	if incomplete == nil && (advertised == nil || len(list) != *advertised) {
		reason := "unstable_pagination"
		incomplete = &reason
	}
	completed := client.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
	if incomplete != nil {
		completed = ""
	}
	document := snapshot.Document{SchemaVersion: "safe-svc-diff.snapshot/v1", ProjectionVersion: "safe-multisig/v1", Producer: snapshot.Producer{Name: "safe-svc-capture", Version: Version}, Capture: snapshot.Capture{SourceOrigin: strings.ToLower(base.Scheme + "://" + base.Host), EndpointPath: endpoint, Safe: safe, Query: toSnapshotQuery(queries), StartedAt: started, Complete: incomplete == nil, IncompleteReason: incomplete, PagesFetched: pages, RecordsFetched: len(list), AdvertisedCount: advertised, MaxPages: MaxPages, MaxRecords: MaxRecords, MaxPageBytes: MaxPageBytes, MaxTotalBytes: MaxTotalBytes}, Transactions: list}
	if incomplete == nil {
		document.Capture.CompletedAt = &completed
	}
	return document, nil
}

func fetchPage(ctx context.Context, target *url.URL, client Client) (page, int, error) {
	for attempt := 0; attempt < 4; attempt++ {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
		if key := os.Getenv("SAFE_API_KEY"); key != "" {
			request.Header.Set("Authorization", "Bearer "+key)
		}
		response, err := client.HTTP.Do(request)
		if err == nil {
			body, readErr := io.ReadAll(io.LimitReader(response.Body, MaxPageBytes+1))
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil {
				return page{}, 0, ErrInvalidPage
			}
			if len(body) > MaxPageBytes {
				return page{}, 0, ErrInvalidPage
			}
			if response.StatusCode == http.StatusOK && strings.Contains(response.Header.Get("Content-Type"), "json") {
				var result page
				if json.Unmarshal(body, &result) == nil {
					return result, len(body), nil
				}
				return page{}, 0, ErrInvalidPage
			}
			if !retryable(response.StatusCode) {
				return page{}, 0, fmt.Errorf("HTTP %d", response.StatusCode)
			}
		}
		if attempt < 3 {
			if sleepErr := client.Sleep(ctx, time.Duration(1<<attempt)*500*time.Millisecond); sleepErr != nil {
				return page{}, 0, sleepErr
			}
		}
	}
	return page{}, 0, errors.New("retry exhausted")
}
func retryable(status int) bool {
	return status == 408 || status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}
func toSnapshotQuery(items []Query) []snapshot.Query {
	result := make([]snapshot.Query, len(items))
	for index, item := range items {
		result[index] = snapshot.Query{Name: item.Name, Value: item.Value}
	}
	return result
}

func WriteAtomic(path string, raw []byte) (err error) {
	if _, statErr := os.Stat(path); statErr == nil {
		return fmt.Errorf("destination exists")
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".safe-svc-capture-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(name)
		}
	}()
	if err = temp.Chmod(0o600); err != nil {
		return err
	}
	if _, err = temp.Write(raw); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
