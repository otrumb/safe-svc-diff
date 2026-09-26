package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
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
	endpoint := path.Join("/", base.Path, "/api/v2/safes/", safe, "/multisig-transactions") + "/"
	current := *base
	current.Path = strings.TrimRight(base.Path, "/") + "/api/v2/safes/" + options.Safe + "/multisig-transactions/"
	query := current.Query()
	for _, item := range queries {
		query.Set(item.Name, item.Value)
	}
	current.RawQuery = query.Encode()
	started := client.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
	totalBytes := 0
	pages := 0
	transactions, advertised, incomplete, err := capturePass(ctx, current, client, &pages, &totalBytes, false)
	if err != nil {
		return snapshot.Document{}, err
	}
	if incomplete == nil {
		verification, verificationCount, verificationReason, _ := capturePass(ctx, current, client, &pages, &totalBytes, true)
		if verificationReason != nil {
			incomplete = verificationReason
		} else if advertised == nil || verificationCount == nil || *advertised != *verificationCount {
			reason := "unstable_pagination"
			incomplete = &reason
		} else {
			firstBytes, firstErr := projectedBytes(transactions)
			verificationBytes, verificationErr := projectedBytes(verification)
			if firstErr != nil || verificationErr != nil || !bytes.Equal(firstBytes, verificationBytes) {
				reason := "unstable_pagination"
				incomplete = &reason
			}
		}
	}
	list := sortedTransactions(transactions)
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

func capturePass(ctx context.Context, first url.URL, client Client, pages, totalBytes *int, verification bool) (map[string]snapshot.Transaction, *int, *string, error) {
	current := first
	transactions := map[string]snapshot.Transaction{}
	var advertised *int
	for {
		if *pages >= MaxPages {
			reason := "page_limit"
			return transactions, advertised, &reason, nil
		}
		result, size, fetchErr := fetchPage(ctx, &current, client)
		if fetchErr != nil {
			if *pages == 0 && !verification {
				return nil, nil, nil, fetchErr
			}
			var failure *FetchError
			reason := "network_failure"
			if errors.As(fetchErr, &failure) {
				reason = failure.Reason
			}
			return transactions, advertised, &reason, nil
		}
		*totalBytes += size
		if *totalBytes > MaxTotalBytes {
			reason := "total_size_limit"
			return transactions, advertised, &reason, nil
		}
		(*pages)++
		if advertised == nil {
			count := result.Count
			advertised = &count
		} else if *advertised != result.Count {
			reason := "unstable_pagination"
			return transactions, advertised, &reason, nil
		}
		if incomplete := mergeResults(transactions, result.Results); incomplete != nil {
			return transactions, advertised, incomplete, nil
		}
		if result.Next == nil {
			break
		}
		next, nextErr := ValidateNext(&current, *result.Next)
		if nextErr != nil {
			reason := "unsafe_next_url"
			return transactions, advertised, &reason, nil
		}
		current = *next
	}
	if advertised == nil || len(transactions) != *advertised {
		reason := "unstable_pagination"
		return transactions, advertised, &reason, nil
	}
	return transactions, advertised, nil, nil
}

func sortedTransactions(transactions map[string]snapshot.Transaction) []snapshot.Transaction {
	list := make([]snapshot.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		list = append(list, transaction)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].SafeTxHash < list[j].SafeTxHash })
	return list
}

func projectedBytes(transactions map[string]snapshot.Transaction) ([]byte, error) {
	raw, err := json.Marshal(sortedTransactions(transactions))
	if err != nil {
		return nil, fmt.Errorf("marshal projected transactions: %w", err)
	}
	return snapshot.Canonical(raw)
}

func mergeResults(transactions map[string]snapshot.Transaction, results []upstreamTransaction) *string {
	for _, input := range results {
		tx, err := project(input)
		if err != nil {
			reason := "invalid_page"
			return &reason
		}
		previous, exists := transactions[tx.SafeTxHash]
		if exists && fmt.Sprintf("%#v", previous) != fmt.Sprintf("%#v", tx) {
			reason := "pagination_conflict"
			return &reason
		}
		if !exists && len(transactions) == MaxRecords {
			reason := "record_limit"
			return &reason
		}
		transactions[tx.SafeTxHash] = tx
	}
	return nil
}
func toSnapshotQuery(items []Query) []snapshot.Query {
	result := make([]snapshot.Query, len(items))
	for index, item := range items {
		result[index] = snapshot.Query{Name: item.Name, Value: item.Value}
	}
	return result
}
