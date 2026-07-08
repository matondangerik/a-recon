package reconciliation

import (
	"time"

	"github.com/shopspring/decimal"
)

type BankData struct {
	Unique_Identifier string          `cvs:"unique_identifier"`
	Amount            decimal.Decimal `cvs:"amount"`
	Date              time.Time       `cvs:"date"`
}

type TxType string

const (
	TxTypeCredit TxType = "CREDIT"
	TxTypeDebit  TxType = "DEBIT"
)

type SystemData struct {
	TrxID           string          `cvs:"trx_id"`
	Amount          decimal.Decimal `cvs:"amount"`
	Type            TxType          `cvs:"type"`
	TransactionTime time.Time       `cvs:"transaction_time"`
}

type Tolerance struct {
	Amount float64
	Days   int
}

type matchResult struct {
	System      SystemData
	Bank        BankData
	Match       bool
	Discrepancy decimal.Decimal
}

type recon struct {
}
