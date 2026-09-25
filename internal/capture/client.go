package capture

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var ErrUnsafe = errors.New("unsafe request")
var addressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var hashPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

var allowedQuery = map[string]string{
	"failed": "bool", "modified__lt": "time", "modified__gt": "time", "modified__lte": "time", "modified__gte": "time",
	"nonce__lt": "decimal", "nonce__gt": "decimal", "nonce__lte": "decimal", "nonce__gte": "decimal", "nonce": "decimal",
	"safe_tx_hash": "hash", "to": "address", "value__lt": "decimal", "value__gt": "decimal", "value": "decimal",
	"executed": "bool", "has_confirmations": "bool", "trusted": "bool", "execution_date__gte": "time", "execution_date__lte": "time",
	"submission_date__gte": "time", "submission_date__lte": "time", "transaction_hash": "hash", "ordering": "ordering",
}

func ParseQuery(items []string) ([]Query, error) {
	values := map[string]string{"limit": "200", "offset": "0", "ordering": "-nonce,-created", "trusted": "false"}
	for _, item := range items {
		name, value, found := strings.Cut(item, "=")
		kind, allowed := allowedQuery[name]
		if !found || !allowed || name == "limit" || name == "offset" {
			return nil, fmt.Errorf("query %q: %w", name, ErrUnsafe)
		}
		if _, duplicate := values[name]; duplicate && name != "ordering" && name != "trusted" {
			return nil, fmt.Errorf("duplicate query %q: %w", name, ErrUnsafe)
		}
		canonical, err := canonicalQuery(kind, value)
		if err != nil {
			return nil, fmt.Errorf("query %s: %w", name, err)
		}
		values[name] = canonical
	}
	result := make([]Query, 0, len(values))
	for name, value := range values {
		result = append(result, Query{Name: name, Value: value})
	}
	slices.SortFunc(result, func(left, right Query) int {
		return strings.Compare(left.Name+"\x00"+left.Value, right.Name+"\x00"+right.Value)
	})
	return result, nil
}

func canonicalQuery(kind, value string) (string, error) {
	switch kind {
	case "bool":
		value = strings.ToLower(value)
		if value == "true" || value == "false" {
			return value, nil
		}
	case "decimal":
		if decimalPattern.MatchString(value) {
			return value, nil
		}
	case "address":
		if addressPattern.MatchString(value) {
			return strings.ToLower(value), nil
		}
	case "hash":
		if hashPattern.MatchString(value) {
			return strings.ToLower(value), nil
		}
	case "ordering":
		for _, field := range strings.Split(value, ",") {
			if !regexp.MustCompile(`^-?(nonce|created|modified)$`).MatchString(field) {
				return "", ErrUnsafe
			}
		}
		return value, nil
	case "time":
		if normalized, err := normalizeTime(value); err == nil {
			return normalized, nil
		}
	}
	return "", ErrUnsafe
}

func ValidateNext(current *url.URL, raw string) (*url.URL, error) {
	next, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse next: %w", err)
	}
	next = current.ResolveReference(next)
	if next.Scheme != "https" || next.User != nil || next.Fragment != "" || !strings.EqualFold(next.Host, current.Host) || next.Path != current.Path {
		return nil, ErrUnsafe
	}
	currentQuery, nextQuery := current.Query(), next.Query()
	for name, values := range nextQuery {
		if len(values) != 1 || (name != "offset" && currentQuery.Get(name) != values[0]) {
			return nil, ErrUnsafe
		}
	}
	for name := range currentQuery {
		if name != "offset" && nextQuery.Get(name) != currentQuery.Get(name) {
			return nil, ErrUnsafe
		}
	}
	oldOffset, oldErr := strconv.Atoi(currentQuery.Get("offset"))
	newOffset, newErr := strconv.Atoi(nextQuery.Get("offset"))
	if oldErr != nil || newErr != nil || newOffset <= oldOffset {
		return nil, ErrUnsafe
	}
	return next, nil
}
