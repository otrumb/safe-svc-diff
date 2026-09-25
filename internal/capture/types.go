package capture

import (
	"fmt"
	"time"
)

const Version = "0.1.0"
const MaxPages = 100
const MaxRecords = 20000
const MaxPageBytes = 8 << 20
const MaxTotalBytes = 64 << 20

type Query struct {
	Name  string
	Value string
}
type Options struct {
	BaseURL string
	Safe    string
	Out     string
	Queries []string
}

func normalizeTime(value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", fmt.Errorf("timestamp: %w", err)
	}
	return parsed.UTC().Format("2006-01-02T15:04:05.000000Z"), nil
}
