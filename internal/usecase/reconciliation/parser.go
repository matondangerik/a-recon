package reconciliation

import (
	"context"
	"sync"

	"github.com/matondangerik/a-recon/pkg/csv"
	"github.com/matondangerik/a-recon/pkg/errgroup"
)

// parseReconciliationData parses reconciliation data from CSV files.
// It filters out rows that are outside the specified range.
// @param ctx context.Context
// @param args ReconciliationArgs
// @return systemData []SystemData
// @return bankDataMap map[string][]BankData
// @return err error
func (f *recon) parseReconciliationData(ctx context.Context, args ReconciliationArgs) (
	systemDataMap map[string][]SystemData,
	bankDataMap map[string][]BankData, err error) {
	g := errgroup.New(ctx)
	bankDataMap = make(map[string][]BankData)

	var mtx sync.Mutex
	for bankName, bankPath := range args.File.BankPath {

		bankName := bankName
		bankPath := bankPath

		g.Go(func(_ context.Context) (err error) {
			bankData, err := f.parseBankData(ctx, args.Range, bankName, bankPath)
			if err != nil {
				return
			}
			mtx.Lock()
			defer mtx.Unlock()

			bankDataMap[bankName] = bankData
			return
		})
	}

	g.Go(func(_ context.Context) (err error) {
		systemDataMap, err = f.parseSystemData(ctx, args.Range, args.File.SystemPath, args.File.BankPath)
		return
	})

	err = g.Wait()
	return
}

// parseSystemData parses system data from a CSV file.
// It filters out rows that are outside the specified range.
// @param ctx context.Context
// @param trange Range
// @param path string
// @return data []SystemData
// @return err error
func (f *recon) parseSystemData(_ context.Context,
	trange Range,
	path string,
	bank map[string]string) (result map[string][]SystemData, err error) {
	result = make(map[string][]SystemData)

	var data []SystemData
	err = csv.ReadFile(path, &data)
	if err != nil {
		return
	}

	for _, d := range data {
		_, ok := bank[d.BankName]
		if !ok {
			continue
		}
		if d.TransactionTime.Before(trange.Start) || d.TransactionTime.After(trange.End) {
			continue
		}
		result[d.BankName] = append(result[d.BankName], d)
	}

	return result, nil
}

// parseBankData parses bank data from a CSV file.
// It filters out rows that are outside the specified range.
// @param ctx context.Context
// @param trange Range
// @param path string
// @return data []BankData
// @return err error
func (f *recon) parseBankData(_ context.Context, trange Range, bankName, path string) (data []BankData, err error) {
	err = csv.ReadFile(path, &data)
	if err != nil {
		return
	}

	var filtered []BankData
	for _, item := range data {
		if item.Date.Before(trange.Start) || item.Date.After(trange.End) {
			continue
		}

		item.BankName = bankName
		filtered = append(filtered, item)
	}
	return filtered, nil
}
