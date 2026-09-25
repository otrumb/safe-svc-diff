package snapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/otrumb/safe-svc-diff/contract"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var ErrInvalid = errors.New("invalid snapshot")

func Validate(raw []byte) error {
	canonical, err := Canonical(raw)
	if err != nil {
		return fmt.Errorf("canonicalize: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return fmt.Errorf("noncanonical JSON: %w", ErrInvalid)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	schemaValue, err := jsonschema.UnmarshalJSON(bytes.NewReader(contract.SnapshotV1))
	if err != nil {
		return fmt.Errorf("decode schema: %w", err)
	}
	if err := compiler.AddResource("snapshot-v1.schema.json", schemaValue); err != nil {
		return fmt.Errorf("add schema: %w", err)
	}
	schema, err := compiler.Compile("snapshot-v1.schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("decode value: %w", err)
	}
	if err := schema.Validate(value); err != nil {
		return fmt.Errorf("schema validation: %w", errors.Join(ErrInvalid, err))
	}
	var document Document
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("decode model: %w", err)
	}
	return validateInvariants(document)
}

func validateInvariants(document Document) error {
	hashes := make([]string, 0, len(document.Transactions))
	for _, transaction := range document.Transactions {
		if transaction.Safe != document.Capture.Safe {
			return fmt.Errorf("transaction safe mismatch: %w", ErrInvalid)
		}
		hashes = append(hashes, transaction.SafeTxHash)
		if !strictlySorted(transaction.ConfirmationOwners) {
			return fmt.Errorf("confirmation owners not sorted unique: %w", ErrInvalid)
		}
	}
	if !strictlySorted(hashes) {
		return fmt.Errorf("transactions not sorted unique: %w", ErrInvalid)
	}
	if document.Capture.RecordsFetched != len(document.Transactions) {
		return fmt.Errorf("record count mismatch: %w", ErrInvalid)
	}
	if document.Capture.RecordsFetched > document.Capture.MaxRecords {
		return fmt.Errorf("record count exceeds limit: %w", ErrInvalid)
	}
	if document.Capture.Complete {
		if document.Capture.CompletedAt == nil || document.Capture.IncompleteReason != nil || document.Capture.PagesFetched < 1 || document.Capture.AdvertisedCount == nil || *document.Capture.AdvertisedCount != document.Capture.RecordsFetched {
			return fmt.Errorf("complete metadata inconsistent: %w", ErrInvalid)
		}
	} else if document.Capture.IncompleteReason == nil {
		return fmt.Errorf("incomplete reason missing: %w", ErrInvalid)
	}
	return nil
}

func strictlySorted(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}
