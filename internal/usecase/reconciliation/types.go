package reconciliation

import (
	"context"
	"time"

	timepkg "github.com/matondangerik/a-recon/pkg/time"
	"github.com/shopspring/decimal"
)

type BankData struct {
	Unique_Identifier string           `csv:"unique_identifier"`
	Amount            decimal.Decimal  `csv:"amount"`
	Date              timepkg.DateOnly `csv:"date"`
	BankName          string
}

type TxType string

const (
	TxTypeCredit TxType = "CREDIT"
	TxTypeDebit  TxType = "DEBIT"
)

type SystemData struct {
	TrxID           string          `csv:"trx_id"`
	Amount          decimal.Decimal `csv:"amount"`
	Type            TxType          `csv:"type"`
	TransactionTime time.Time       `csv:"transaction_time"`
	BankName        string          `csv:"bank_name"`
}

type Tolerance struct {
	Amount float64
	Days   int
}

type Range struct {
	Start time.Time
	End   time.Time
}

type File struct {
	SystemPath string
	BankPath   map[string]string
}

type matchResult struct {
	System      SystemData
	Bank        BankData
	Match       bool
	Discrepancy decimal.Decimal
}

type ReconciliationArgs struct {
	File      File
	Tolerance Tolerance
	Range     Range
}

type ReconciliationResult struct {
	TotalTx int

	TotalMatched   int
	TotalUnmatched int

	UnmatchedSystem []SystemData
	UnmatchedBank   map[string][]BankData

	TotalDiscrepancy decimal.Decimal
}

type IReconciliation interface {
	// Reconcile reconcile system data and bank data
	// @param ctx context
	// @param args reconciliation args
	// @return reconciliation result
	// @return error if reconciliation failed
	Reconcile(ctx context.Context, args ReconciliationArgs) (ReconciliationResult, error)
}

type recon struct {
}

// NewReconciliation create new reconciliation
// @return reconciliation
func NewReconciliation() IReconciliation {
	return &recon{}
}
