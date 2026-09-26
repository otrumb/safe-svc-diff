package capture

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/otrumb/safe-svc-diff/internal/snapshot"
)

var ErrInvalidPage = errors.New("invalid upstream page")

type page struct {
	Count   int                   `json:"count"`
	Next    *string               `json:"next"`
	Results []upstreamTransaction `json:"results"`
}
type confirmation struct {
	Owner string `json:"owner"`
}
type upstreamTransaction struct {
	SafeTxHash            string         `json:"safeTxHash"`
	Safe                  string         `json:"safe"`
	To                    string         `json:"to"`
	Value                 string         `json:"value"`
	Data                  *string        `json:"data"`
	Operation             int            `json:"operation"`
	GasToken              *string        `json:"gasToken"`
	SafeTxGas             string         `json:"safeTxGas"`
	BaseGas               string         `json:"baseGas"`
	GasPrice              string         `json:"gasPrice"`
	RefundReceiver        *string        `json:"refundReceiver"`
	Nonce                 string         `json:"nonce"`
	ExecutionDate         *string        `json:"executionDate"`
	SubmissionDate        string         `json:"submissionDate"`
	Modified              string         `json:"modified"`
	BlockNumber           *int           `json:"blockNumber"`
	TransactionHash       *string        `json:"transactionHash"`
	Proposer              *string        `json:"proposer"`
	ProposedByDelegate    *string        `json:"proposedByDelegate"`
	Executor              *string        `json:"executor"`
	IsExecuted            bool           `json:"isExecuted"`
	IsSuccessful          *bool          `json:"isSuccessful"`
	EthGasPrice           *string        `json:"ethGasPrice"`
	MaxFeePerGas          *string        `json:"maxFeePerGas"`
	MaxPriorityFeePerGas  *string        `json:"maxPriorityFeePerGas"`
	GasUsed               *int           `json:"gasUsed"`
	Fee                   *string        `json:"fee"`
	Payment               *string        `json:"payment"`
	ConfirmationsRequired *int           `json:"confirmationsRequired"`
	Confirmations         []confirmation `json:"confirmations"`
	Trusted               bool           `json:"trusted"`
}

func project(input upstreamTransaction) (snapshot.Transaction, error) {
	if !hashPattern.MatchString(input.SafeTxHash) || !addressPattern.MatchString(input.Safe) || !addressPattern.MatchString(input.To) || (input.GasToken != nil && !addressPattern.MatchString(*input.GasToken)) {
		return snapshot.Transaction{}, ErrInvalidPage
	}
	for _, decimal := range []string{input.Value, input.SafeTxGas, input.BaseGas, input.GasPrice, input.Nonce} {
		if !decimalPattern.MatchString(decimal) {
			return snapshot.Transaction{}, ErrInvalidPage
		}
	}
	owners := make([]string, 0, len(input.Confirmations))
	for _, item := range input.Confirmations {
		if !addressPattern.MatchString(item.Owner) {
			return snapshot.Transaction{}, ErrInvalidPage
		}
		owners = append(owners, strings.ToLower(item.Owner))
	}
	sort.Strings(owners)
	if len(owners) > 1 {
		for index := 1; index < len(owners); index++ {
			if owners[index] == owners[index-1] {
				return snapshot.Transaction{}, ErrInvalidPage
			}
		}
	}
	var digest *string
	var length *int
	if input.Data != nil {
		raw := strings.TrimPrefix(*input.Data, "0x")
		decoded, err := hex.DecodeString(raw)
		if err != nil {
			return snapshot.Transaction{}, fmt.Errorf("decode data: %w", ErrInvalidPage)
		}
		sum := fmt.Sprintf("%x", sha256.Sum256(decoded))
		size := len(decoded)
		digest, length = &sum, &size
	}
	return snapshot.Transaction{SafeTxHash: strings.ToLower(input.SafeTxHash), Safe: strings.ToLower(input.Safe), To: strings.ToLower(input.To), Value: input.Value, DataSHA256: digest, DataLength: length, Operation: input.Operation, GasToken: lower(input.GasToken), SafeTxGas: input.SafeTxGas, BaseGas: input.BaseGas, GasPrice: input.GasPrice, RefundReceiver: lower(input.RefundReceiver), Nonce: input.Nonce, ExecutionDate: date(input.ExecutionDate), SubmissionDate: mustDate(input.SubmissionDate), Modified: mustDate(input.Modified), BlockNumber: input.BlockNumber, TransactionHash: lower(input.TransactionHash), Proposer: lower(input.Proposer), ProposedByDelegate: lower(input.ProposedByDelegate), Executor: lower(input.Executor), IsExecuted: input.IsExecuted, IsSuccessful: input.IsSuccessful, EthGasPrice: input.EthGasPrice, MaxFeePerGas: input.MaxFeePerGas, MaxPriorityFeePerGas: input.MaxPriorityFeePerGas, GasUsed: input.GasUsed, Fee: input.Fee, Payment: input.Payment, ConfirmationsRequired: input.ConfirmationsRequired, ConfirmationOwners: owners, Trusted: input.Trusted}, nil
}

func lower(value *string) *string {
	if value == nil {
		return nil
	}
	lowered := strings.ToLower(*value)
	return &lowered
}
func date(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := mustDate(*value)
	return &normalized
}
func mustDate(value string) string { normalized, _ := normalizeTime(value); return normalized }
