package reconciliation

import (
	"context"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// match match system data with bank data
// @param Tolerance tolerance for match
// @param systemData system data
// @param bankData bank data
// @return match result
func (r *recon) match(ctx context.Context, Tolerance Tolerance, systemData []SystemData, bankData []BankData) (results []matchResult) {
	sort.Slice(bankData, func(i, j int) bool {
		if bankData[i].Date.Equal(bankData[j].Date.Time) {
			return bankData[i].Unique_Identifier < bankData[j].Unique_Identifier
		}
		return bankData[i].Date.Before(bankData[j].Date.Time)
	})
	sort.Slice(systemData, func(i, j int) bool {
		return systemData[i].TransactionTime.Before(systemData[j].TransactionTime)
	})

	for _, system := range systemData {
		bestmatch, discrepancy, remaining := r.findMatch(ctx, Tolerance, system, bankData)
		if bestmatch.Unique_Identifier != "" {
			results = append(results, matchResult{
				System:      system,
				Bank:        bestmatch,
				Match:       true,
				Discrepancy: discrepancy,
			})
		} else {
			results = append(results, matchResult{
				System:      system,
				Discrepancy: system.Amount,
			})
		}
		bankData = remaining
	}

	for _, bank := range bankData {
		matchResult := matchResult{
			Bank:        bank,
			Discrepancy: bank.Amount.Abs(),
		}

		if bank.Amount.IsNegative() {
			matchResult.System.Type = TxTypeDebit
		} else {
			matchResult.System.Type = TxTypeCredit
		}
		results = append(results, matchResult)

	}
	return results
}

// findMatch find match system data for bank data
// @param Tolerance tolerance for match
// @param system system data
// @param bank bank data sorted by transaction time
// @return match system data or empty if not found
// @return remaining bank data that not matched
func (r *recon) findMatch(_ context.Context, Tolerance Tolerance, system SystemData, bank []BankData) (bestmatch BankData, discrepancy decimal.Decimal, remaining []BankData) {
	txTime := startOfDay(system.TransactionTime)

	startTolerance := txTime.AddDate(0, 0, -Tolerance.Days)
	endTolerance := txTime.AddDate(0, 0, Tolerance.Days)
	amountTolerance := decimal.NewFromFloat(Tolerance.Amount)

	var idx = -1
	var delta int64
	for i, bank := range bank {
		if bank.Date.Before(startTolerance) {
			continue
		}

		if bank.Date.After(endTolerance) {
			break
		}

		valid, vdiscrepancy := r.getDiscrepancy(system, bank)
		if !valid {
			continue
		}

		if !vdiscrepancy.IsZero() &&
			!vdiscrepancy.LessThanOrEqual(amountTolerance) {
			continue
		}

		var d int64
		if bank.Date.Before(txTime) {
			d = txTime.Sub(bank.Date.Time).Microseconds()
		} else {
			d = bank.Date.Time.Sub(txTime).Microseconds()
		}

		if idx < 0 || d < delta {
			idx = i
			delta = d
			discrepancy = vdiscrepancy
		}
	}

	if idx < 0 {
		return bestmatch, discrepancy, bank
	}

	if idx > 0 {
		remaining = append(remaining, bank[:idx]...)
	}
	if idx < len(bank)-1 {
		remaining = append(remaining, bank[idx+1:]...)
	}

	bestmatch = bank[idx]
	return bestmatch, discrepancy, remaining
}

// getDiscrepancy get discrepancy between system data and bank data
// @param system system data
// @param bank bank data
// @return valid if system data and bank data have same direction
// @return discrepancy between system data and bank data
func (f *recon) getDiscrepancy(system SystemData, bank BankData) (valid bool, discrepancy decimal.Decimal) {
	if system.Type == TxTypeDebit && bank.Amount.IsNegative() {
		discrepancy = system.Amount.Add(bank.Amount).Abs()
		valid = true
	} else if system.Type == TxTypeCredit && bank.Amount.IsPositive() {
		discrepancy = system.Amount.Sub(bank.Amount).Abs()
		valid = true
	} else if system.Amount.IsZero() && bank.Amount.IsZero() {
		valid = true
	}
	return valid, discrepancy
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
