package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matondangerik/a-recon/internal/usecase/reconciliation"
	timepkg "github.com/matondangerik/a-recon/pkg/time"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestRenderResult(t *testing.T) {
	result := reconciliation.ReconciliationResult{
		TotalTx:          3,
		TotalMatched:     1,
		TotalUnmatched:   2,
		TotalDiscrepancy: decimal.RequireFromString("3"),
		UnmatchedSystem: []reconciliation.SystemData{
			{
				TrxID:           "trx-1",
				Amount:          decimal.RequireFromString("40"),
				TransactionTime: time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC),
			},
		},
		UnmatchedBank: map[string][]reconciliation.BankData{
			"bank-b": {
				{
					Unique_Identifier: "bank-b-1",
					BankName:          "bank-b",
					Amount:            decimal.RequireFromString("60"),
					Date:              dateOnly(2026, time.January, 11),
				},
			},
			"bank-a": {
				{
					Unique_Identifier: "bank-a-1",
					BankName:          "bank-a",
					Amount:            decimal.RequireFromString("20"),
					Date:              dateOnly(2026, time.January, 9),
				},
			},
		},
	}

	rendered, err := renderResult(result)

	assert.NoError(t, err)
	assert.Contains(t, rendered, "Metric")
	assert.Contains(t, rendered, "Total Transactions")
	assert.Contains(t, rendered, "Total Discrepancy")
	assert.Contains(t, rendered, "UNMATCHED")
	assert.Contains(t, rendered, "trx-1")
	assert.Contains(t, rendered, "ORPHANED")

	firstBankA := strings.Index(rendered, "bank-a")
	firstBankB := strings.Index(rendered, "bank-b")
	assert.NotEqual(t, -1, firstBankA)
	assert.NotEqual(t, -1, firstBankB)
	assert.Less(t, firstBankA, firstBankB)
}

func TestRenderResultWhenEverythingMatches(t *testing.T) {
	rendered, err := renderResult(reconciliation.ReconciliationResult{
		TotalTx:          2,
		TotalMatched:     2,
		TotalUnmatched:   0,
		TotalDiscrepancy: decimal.Zero,
		UnmatchedBank:    map[string][]reconciliation.BankData{},
	})

	assert.NoError(t, err)
	assert.Contains(t, rendered, "All transactions reconciled")
}

func TestWriteResult(t *testing.T) {
	result := reconciliation.ReconciliationResult{
		TotalTx:          1,
		TotalMatched:     1,
		TotalUnmatched:   0,
		TotalDiscrepancy: decimal.Zero,
		UnmatchedBank:    map[string][]reconciliation.BankData{},
	}
	outputPath := filepath.Join(t.TempDir(), "result.txt")

	stdout := captureStdout(t, func() {
		err := writeResult(result, outputPath)
		assert.NoError(t, err)
	})

	fileContent, err := os.ReadFile(outputPath)
	assert.NoError(t, err)
	assert.Equal(t, stdout, string(fileContent))
	assert.Contains(t, stdout, "All transactions reconciled")
}

func TestSortedBankNames(t *testing.T) {
	names := sortedBankNames(map[string][]reconciliation.BankData{
		"bank-c": nil,
		"bank-a": nil,
		"bank-b": nil,
	})

	assert.Equal(t, []string{"bank-a", "bank-b", "bank-c"}, names)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}

	output, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("failed to read stdout: %v", readErr)
	}

	return string(output)
}

func dateOnly(year int, month time.Month, day int) timepkg.DateOnly {
	return timepkg.DateOnly{
		Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
}
