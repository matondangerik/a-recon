package reconciliation

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestToReconciliationResult(t *testing.T) {
	r := &recon{}

	result := r.toReconciliationResult([]matchResult{
		{
			System:      SystemData{TrxID: "trx-match"},
			Bank:        BankData{Unique_Identifier: "bank-match", BankName: "bank-a"},
			Match:       true,
			Discrepancy: decimal.RequireFromString("3"),
		},
		{
			System:      SystemData{TrxID: "trx-unmatched"},
			Discrepancy: decimal.RequireFromString("40"),
		},
		{
			Bank:        BankData{Unique_Identifier: "bank-unmatched-a", BankName: "bank-a"},
			Discrepancy: decimal.RequireFromString("60"),
		},
		{
			Bank:        BankData{Unique_Identifier: "bank-unmatched-b", BankName: "bank-b"},
			Discrepancy: decimal.RequireFromString("20"),
		},
	})

	assert.Equal(t, 4, result.TotalTx)
	assert.Equal(t, 1, result.TotalMatched)
	assert.Equal(t, 3, result.TotalUnmatched)
	assert.True(t, result.TotalDiscrepancy.Equal(decimal.RequireFromString("3")))
	if assert.Len(t, result.UnmatchedSystem, 1) {
		assert.Equal(t, "trx-unmatched", result.UnmatchedSystem[0].TrxID)
	}
	if assert.Len(t, result.UnmatchedBank, 2) {
		assert.Len(t, result.UnmatchedBank["bank-a"], 1)
		assert.Len(t, result.UnmatchedBank["bank-b"], 1)
	}
}

func TestReconcile(t *testing.T) {
	r := &recon{}
	dir := t.TempDir()

	systemPath := writeTempCSV(t, dir, "system.csv", ""+
		"trx_id,amount,type,transaction_time,bank_name\n"+
		"trx-credit,100,CREDIT,2026-01-10T10:00:00Z,bank-a\n"+
		"trx-debit,100,DEBIT,2026-01-10T11:00:00Z,bank-b\n"+
		"trx-unmatched,40,CREDIT,2026-01-10T12:00:00Z,bank-a\n",
	)
	bankAPath := writeTempCSV(t, dir, "bank-a.csv", ""+
		"unique_identifier,amount,date\n"+
		"bank-credit,103,2026-01-10\n"+
		"bank-unmatched,60,2026-01-10\n",
	)
	bankBPath := writeTempCSV(t, dir, "bank-b.csv", ""+
		"unique_identifier,amount,date\n"+
		"bank-debit,-100,2026-01-10\n",
	)

	result, err := r.Reconcile(context.Background(), ReconciliationArgs{
		File: File{
			SystemPath: systemPath,
			BankPath: map[string]string{
				"bank-a": bankAPath,
				"bank-b": bankBPath,
			},
		},
		Tolerance: Tolerance{
			Amount: 5,
			Days:   1,
		},
		Range: Range{
			Start: at(2026, time.January, 10, 0, 0),
			End:   at(2026, time.January, 11, 0, 0),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 4, result.TotalTx)
	assert.Equal(t, 2, result.TotalMatched)
	assert.Equal(t, 2, result.TotalUnmatched)
	assert.True(t, result.TotalDiscrepancy.Equal(decimal.RequireFromString("3")))
	if assert.Len(t, result.UnmatchedSystem, 1) {
		assert.Equal(t, "trx-unmatched", result.UnmatchedSystem[0].TrxID)
	}
	if assert.Len(t, result.UnmatchedBank["bank-a"], 1) {
		assert.Equal(t, "bank-unmatched", result.UnmatchedBank["bank-a"][0].Unique_Identifier)
	}
}
