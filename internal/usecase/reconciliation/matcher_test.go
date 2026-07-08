package reconciliation

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	t.Run("returns mixed matched and unmatched results in system order", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-later",
				Amount:          decimal.RequireFromString("100"),
				Type:            TxTypeCredit,
				TransactionTime: at(2026, time.January, 10, 10, 0),
			},
			{
				TrxID:           "trx-earlier",
				Amount:          decimal.RequireFromString("40"),
				Type:            TxTypeCredit,
				TransactionTime: at(2026, time.January, 9, 9, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-unmatched",
				Amount:            decimal.RequireFromString("-25"),
				Date:              at(2026, time.January, 11, 0, 0),
			},
			{
				Unique_Identifier: "bank-match",
				Amount:            decimal.RequireFromString("103"),
				Date:              at(2026, time.January, 10, 0, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 5, Days: 1}, systemData, bankData)

		if !assert.Len(t, results, 3) {
			return
		}

		assert.False(t, results[0].Match)
		assert.Equal(t, "trx-earlier", results[0].System.TrxID)
		assert.True(t, results[0].Discrepancy.Equal(decimal.RequireFromString("40")))

		assert.True(t, results[1].Match)
		assert.Equal(t, "trx-later", results[1].System.TrxID)
		assert.Equal(t, "bank-match", results[1].Bank.Unique_Identifier)
		assert.True(t, results[1].Discrepancy.Equal(decimal.RequireFromString("3")))

		assert.False(t, results[2].Match)
		assert.Equal(t, "bank-unmatched", results[2].Bank.Unique_Identifier)
		assert.Equal(t, TxTypeDebit, results[2].System.Type)
		assert.True(t, results[2].Discrepancy.Equal(decimal.RequireFromString("25")))
	})

	t.Run("selects the closest bank transaction within tolerance", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-closest",
				Amount:          decimal.RequireFromString("100"),
				Type:            TxTypeCredit,
				TransactionTime: at(2026, time.January, 10, 12, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-far",
				Amount:            decimal.RequireFromString("100"),
				Date:              at(2026, time.January, 9, 0, 0),
			},
			{
				Unique_Identifier: "bank-near",
				Amount:            decimal.RequireFromString("100"),
				Date:              at(2026, time.January, 10, 11, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 0, Days: 2}, systemData, bankData)

		if !assert.Len(t, results, 2) {
			return
		}
		assert.True(t, results[0].Match)
		assert.Equal(t, "bank-near", results[0].Bank.Unique_Identifier)
		assert.False(t, results[1].Match)
		assert.Equal(t, "bank-far", results[1].Bank.Unique_Identifier)
	})

	t.Run("matches zero amount transactions", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-zero",
				Amount:          decimal.Zero,
				Type:            TxTypeCredit,
				TransactionTime: at(2026, time.January, 10, 8, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-zero",
				Amount:            decimal.Zero,
				Date:              at(2026, time.January, 10, 0, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 0, Days: 0}, systemData, bankData)

		if !assert.Len(t, results, 1) {
			return
		}
		assert.True(t, results[0].Match)
		assert.Equal(t, "bank-zero", results[0].Bank.Unique_Identifier)
		assert.True(t, results[0].Discrepancy.Equal(decimal.Zero))
	})

	t.Run("returns both sides as unmatched when outside tolerance", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-outside",
				Amount:          decimal.RequireFromString("100"),
				Type:            TxTypeCredit,
				TransactionTime: at(2026, time.January, 10, 8, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-outside",
				Amount:            decimal.RequireFromString("106"),
				Date:              at(2026, time.January, 12, 0, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 5, Days: 1}, systemData, bankData)

		if !assert.Len(t, results, 2) {
			return
		}
		assert.False(t, results[0].Match)
		assert.Equal(t, "trx-outside", results[0].System.TrxID)
		assert.True(t, results[0].Discrepancy.Equal(decimal.RequireFromString("100")))

		assert.False(t, results[1].Match)
		assert.Equal(t, "bank-outside", results[1].Bank.Unique_Identifier)
		assert.Equal(t, TxTypeCredit, results[1].System.Type)
		assert.True(t, results[1].Discrepancy.Equal(decimal.RequireFromString("106")))
	})

	t.Run("skips out of range rows and matches debit after mismatches", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-debit",
				Amount:          decimal.RequireFromString("100"),
				Type:            TxTypeDebit,
				TransactionTime: at(2026, time.January, 10, 12, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-too-old",
				Amount:            decimal.RequireFromString("-100"),
				Date:              at(2026, time.January, 8, 0, 0),
			},
			{
				Unique_Identifier: "bank-wrong-amount",
				Amount:            decimal.RequireFromString("-130"),
				Date:              at(2026, time.January, 10, 11, 0),
			},
			{
				Unique_Identifier: "bank-debit-match",
				Amount:            decimal.RequireFromString("-100"),
				Date:              at(2026, time.January, 10, 13, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 5, Days: 1}, systemData, bankData)

		if !assert.Len(t, results, 3) {
			return
		}

		assert.True(t, results[0].Match)
		assert.Equal(t, "trx-debit", results[0].System.TrxID)
		assert.Equal(t, "bank-debit-match", results[0].Bank.Unique_Identifier)
		assert.True(t, results[0].Discrepancy.Equal(decimal.Zero))

		assert.False(t, results[1].Match)
		assert.Equal(t, "bank-too-old", results[1].Bank.Unique_Identifier)
		assert.Equal(t, TxTypeDebit, results[1].System.Type)

		assert.False(t, results[2].Match)
		assert.Equal(t, "bank-wrong-amount", results[2].Bank.Unique_Identifier)
		assert.Equal(t, TxTypeDebit, results[2].System.Type)
		assert.True(t, results[2].Discrepancy.Equal(decimal.RequireFromString("130")))
	})

	t.Run("matches debit transactions using current negative bank semantics", func(t *testing.T) {
		r := &recon{}

		systemData := []SystemData{
			{
				TrxID:           "trx-debit-match",
				Amount:          decimal.RequireFromString("100"),
				Type:            TxTypeDebit,
				TransactionTime: at(2026, time.January, 10, 12, 0),
			},
		}
		bankData := []BankData{
			{
				Unique_Identifier: "bank-debit-match",
				Amount:            decimal.RequireFromString("-100"),
				Date:              at(2026, time.January, 10, 13, 0),
			},
		}

		results := r.match(context.Background(), Tolerance{Amount: 0, Days: 1}, systemData, bankData)

		if !assert.Len(t, results, 1) {
			return
		}

		assert.True(t, results[0].Match)
		assert.Equal(t, "trx-debit-match", results[0].System.TrxID)
		assert.Equal(t, "bank-debit-match", results[0].Bank.Unique_Identifier)
		assert.True(t, results[0].Discrepancy.Equal(decimal.RequireFromString("0")))
	})
}

func at(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
