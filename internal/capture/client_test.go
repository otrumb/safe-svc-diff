package capture

import (
	"net/url"
	"testing"
)

func TestParseQuery_rejects_reserved_parameter(t *testing.T) {
	// Given
	values := []string{"limit=10"}
	// When
	_, err := ParseQuery(values)
	// Then
	if err == nil {
		t.Fatal("ParseQuery() accepted reserved parameter")
	}
}

func TestValidateNext_rejects_cross_origin(t *testing.T) {
	// Given
	current, _ := url.Parse("https://safe.example/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/?limit=200&offset=0&trusted=true")
	// When
	_, err := ValidateNext(current, "https://evil.example/api/v2/safes/0x1111111111111111111111111111111111111111/multisig-transactions/?limit=200&offset=200&trusted=true")
	// Then
	if err == nil {
		t.Fatal("ValidateNext() accepted cross origin")
	}
}
