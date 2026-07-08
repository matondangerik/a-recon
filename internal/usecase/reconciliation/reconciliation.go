package reconciliation

import (
	"context"
	"sync"

	"github.com/matondangerik/a-recon/pkg/errgroup"
	"github.com/shopspring/decimal"
)

func (f *recon) Reconcile(ctx context.Context, args ReconciliationArgs) (ReconciliationResult, error) {
	systemData, bankDataMap, err := f.parseReconciliationData(ctx, args)
	if err != nil {
		return ReconciliationResult{}, err
	}

	g := errgroup.New(ctx)

	var mtx sync.Mutex
	var matchResultsMap []matchResult
	for name, b := range bankDataMap {
		b := b
		s := systemData[name]

		g.Go(func(_ context.Context) (err error) {
			matchResults := f.match(ctx, args.Tolerance, s, b)

			mtx.Lock()
			defer mtx.Unlock()

			matchResultsMap = append(matchResultsMap, matchResults...)
			return
		})
	}

	err = g.Wait()
	if err != nil {
		return ReconciliationResult{}, err
	}

	return f.toReconciliationResult(matchResultsMap), nil
}

// toReconciliationResult convert match results to reconciliation result
// @param matchResults match results
// @return reconciliation result
func (f *recon) toReconciliationResult(matchResults []matchResult) (result ReconciliationResult) {
	result.TotalTx = len(matchResults)
	result.UnmatchedBank = make(map[string][]BankData)

	discrepancy := decimal.Zero
	for _, r := range matchResults {
		if r.Match {
			result.TotalMatched++
			discrepancy = discrepancy.Add(r.Discrepancy)
			continue
		}

		result.TotalUnmatched++
		if r.System.TrxID != "" {
			result.UnmatchedSystem = append(result.UnmatchedSystem, r.System)
		} else {
			result.UnmatchedBank[r.Bank.BankName] = append(result.UnmatchedBank[r.Bank.BankName], r.Bank)
		}
	}

	result.TotalDiscrepancy = discrepancy
	return
}
