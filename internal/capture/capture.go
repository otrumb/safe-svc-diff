package capture

import (
	"context"
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
			var failure *FetchError
			reason := "network_failure"
			if errors.As(fetchErr, &failure) {
				reason = failure.Reason
			}
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
		incomplete = mergeResults(transactions, result.Results)
		if incomplete != nil {
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
