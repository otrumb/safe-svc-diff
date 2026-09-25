package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/otrumb/safe-svc-diff/internal/capture"
	"github.com/otrumb/safe-svc-diff/internal/snapshot"
)

type queryFlags []string

func (values *queryFlags) String() string         { return fmt.Sprint([]string(*values)) }
func (values *queryFlags) Set(value string) error { *values = append(*values, value); return nil }

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: safe-svc-capture capture|validate|version")
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Println(capture.Version)
		return 0
	case "validate":
		if len(args) != 2 {
			return 2
		}
		raw, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		if err = snapshot.Validate(raw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 4
		}
		var doc snapshot.Document
		_ = json.Unmarshal(raw, &doc)
		if !doc.Capture.Complete {
			return 3
		}
		fmt.Println("valid")
		return 0
	case "capture":
		return runCapture(args[1:])
	default:
		return 2
	}
}
func runCapture(args []string) int {
	set := flag.NewFlagSet("capture", flag.ContinueOnError)
	base := set.String("base-url", "", "service base URL")
	safe := set.String("safe", "", "Safe address")
	out := set.String("out", "", "destination")
	var queries queryFlags
	set.Var(&queries, "query", "filter NAME=VALUE")
	if set.Parse(args) != nil || *base == "" || *safe == "" || *out == "" {
		return 2
	}
	client := capture.Client{HTTP: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("redirect rejected") }}, Now: time.Now, Sleep: func(ctx context.Context, d time.Duration) error {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}}
	document, err := capture.Capture(context.Background(), capture.Options{BaseURL: *base, Safe: *safe, Out: *out, Queries: queries}, client)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return 70
	}
	raw, err = snapshot.Canonical(raw)
	if err != nil {
		return 70
	}
	if err = snapshot.Validate(raw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	if err = capture.WriteAtomic(*out, raw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !document.Capture.Complete {
		return 3
	}
	return 0
}
