package snapshot_test

import (
	"encoding/json"
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
