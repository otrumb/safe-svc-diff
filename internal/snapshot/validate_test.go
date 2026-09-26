package snapshot_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/otrumb/safe-svc-diff/internal/snapshot"
)

func TestValidate_accepts_canonical_fixture(t *testing.T) {
	// Given
	raw := fixtureSnapshot(t, "01-identical.json", "before")

	// When
	err := snapshot.Validate(raw)

	// Then
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidate_rejects_duplicate_identity(t *testing.T) {
	// Given
	raw := fixtureSnapshot(t, "10-invalid-contract.json", "before")

	// When
	err := snapshot.Validate(raw)

	// Then
	if err == nil {
		t.Fatal("Validate() accepted duplicate identity")
	}
}

func TestCanonical_returns_fixture_bytes(t *testing.T) {
	// Given
	raw := fixtureSnapshot(t, "01-identical.json", "before")

	// When
	got, err := snapshot.Canonical(raw)

	// Then
	if err != nil {
		t.Fatalf("Canonical() error = %v", err)
	}
	if string(got) != string(raw) {
		t.Fatal("Canonical() changed canonical fixture")
	}
}

func TestCanonical_escapes_line_and_paragraph_separators_like_shared_fixture(t *testing.T) {
	// Given
	want, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "canonical", "u2028-u2029.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("{\"text\":\"line" + string(rune(0x2028)) + "paragraph" + string(rune(0x2029)) + "\"}")

	// When
	got, err := snapshot.Canonical(raw)

	// Then
	if err != nil {
		t.Fatalf("Canonical() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Canonical() = %q, want shared bytes %q", got, want)
	}
}

func TestValidate_accepts_zero_to_many_sorted_unique_confirmation_owners(t *testing.T) {
	for count := 0; count <= 4; count++ {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			raw := fixtureSnapshot(t, "01-identical.json", "before")
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			transactions := document["transactions"].([]any)
			transaction := transactions[0].(map[string]any)
			owners := make([]string, count)
			for index := range owners {
				owners[index] = fmt.Sprintf("0x%040x", index+1)
			}
			transaction["confirmationOwners"] = owners
			encoded, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := snapshot.Canonical(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if err := snapshot.Validate(canonical); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestValidate_rejects_record_count_over_declared_limit(t *testing.T) {
	raw := fixtureSnapshot(t, "02-added.json", "after")
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	capture := document["capture"].(map[string]any)
	transactions := document["transactions"].([]any)
	copy := transactions[0].(map[string]any)
	copy["safeTxHash"] = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	document["transactions"] = append(transactions, copy)
	capture["recordsFetched"] = float64(2)
	capture["advertisedCount"] = float64(2)
	capture["maxRecords"] = float64(1)
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := snapshot.Canonical(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := snapshot.Validate(canonical); err == nil {
		t.Fatal("Validate() accepted recordsFetched > maxRecords")
	}
}

func fixtureSnapshot(t *testing.T, name, side string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "tests", "fixtures", "pairs", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var pair struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
	}
	if err := json.Unmarshal(raw, &pair); err != nil {
		t.Fatalf("extract fixture: %v", err)
	}
	if side == "before" {
		return append(pair.Before, '\n')
	}
	return append(pair.After, '\n')
}
