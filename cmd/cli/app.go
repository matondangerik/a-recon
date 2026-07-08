package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/matondangerik/a-recon/internal/usecase/reconciliation"
)

func main() {
	// Command line flags
	var systemTxPath = flag.String("system", "", "Path to system transactions CSV file (required)")
	var bankStatementPaths = flag.String("banks", "", "Comma-separated list of bank statement paths in format path1, path2, file name will be used as bank name (required)")
	var startDate = flag.String("start", "", "Start date for reconciliation (YYYY-MM-DD format) (required)")
	var endDate = flag.String("end", "", "End date for reconciliation (YYYY-MM-DD format) (required)")
	var daysTolerance = flag.Int("tolerance", 1, "Days tolerance for reconciliation (default: 1)")
	var amountTolerance = flag.Float64("amount-tolerance", 5, "Amount tolerance for reconciliation (default: 5)")
	var outputPath = flag.String("output", "", "Output file path (default: stdout)")

	flag.Parse()

	// Validate required parameters
	if *systemTxPath == "" {
		log.Fatal("System transactions file path is required")
	}

	if *bankStatementPaths == "" {
		log.Fatal("Bank statement paths are required")
	}

	bankPaths := make(map[string]string)
	for _, path := range strings.Split(*bankStatementPaths, ",") {
		bankName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		bankPaths[bankName] = path
	}

	if *startDate == "" {
		log.Fatal("Start date is required")
	}

	startTime, err := time.Parse("2006-01-02", *startDate)
	if err != nil {
		log.Fatal("Start date format is invalid")
	}

	if *endDate == "" {
		log.Fatal("End date is required")
	}

	endTime, err := time.Parse("2006-01-02", *endDate)
	if err != nil {
		log.Fatal("End date format is invalid")
	}

	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())

	recon := reconciliation.NewReconciliation()
	result, err := recon.Reconcile(context.Background(),
		reconciliation.ReconciliationArgs{
			File: reconciliation.File{
				SystemPath: *systemTxPath,
				BankPath:   bankPaths,
			},
			Range: reconciliation.Range{
				Start: startTime,
				End:   endTime,
			},
			Tolerance: reconciliation.Tolerance{
				Days:   *daysTolerance,
				Amount: *amountTolerance,
			},
		},
	)
	if err != nil {
		log.Fatal("Reconciliation failed")
	}

	if err := writeResult(result, *outputPath); err != nil {
		log.Fatal(err)
	}
}

func writeResult(result reconciliation.ReconciliationResult, outputPath string) error {
	rendered, err := renderResult(result)
	if err != nil {
		return err
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}

	fmt.Print(rendered)
	return nil
}

func renderResult(result reconciliation.ReconciliationResult) (string, error) {
	var output strings.Builder
	tw := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)

	writeSummary(tw, result)
	writeUnmatchedRows(tw, result)

	if err := tw.Flush(); err != nil {
		return "", fmt.Errorf("failed to render output: %w", err)
	}

	return output.String(), nil
}

func writeSummary(tw *tabwriter.Writer, result reconciliation.ReconciliationResult) {
	fmt.Fprintln(tw, "Metric\tValue")
	fmt.Fprintf(tw, "Total Transactions\t%d\n", result.TotalTx)
	fmt.Fprintf(tw, "Total Matched\t%d\n", result.TotalMatched)
	fmt.Fprintf(tw, "Total Unmatched\t%d\n", result.TotalUnmatched)
	fmt.Fprintf(tw, "Total Discrepancy\t%s\n", result.TotalDiscrepancy.String())
	fmt.Fprintln(tw)
}

func writeUnmatchedRows(tw *tabwriter.Writer, result reconciliation.ReconciliationResult) {
	fmt.Fprintln(tw, "Status\tSystem ID\tBank\tTransaction Date\tAmount\tDetails")

	for _, tx := range result.UnmatchedSystem {
		fmt.Fprintf(tw, "UNMATCHED\t%s\tN/A\t%s\t%s\tNo matching bank transaction found\n",
			tx.TrxID,
			tx.TransactionTime.Format("2006-01-02"),
			tx.Amount.String(),
		)
	}

	for _, bankName := range sortedBankNames(result.UnmatchedBank) {
		for _, tx := range result.UnmatchedBank[bankName] {
			fmt.Fprintf(tw, "ORPHANED\tN/A\t%s\t%s\t%s\tNo matching system transaction found\n",
				tx.BankName,
				tx.Date.Format("2006-01-02"),
				tx.Amount.String(),
			)
		}
	}

	if result.TotalUnmatched == 0 {
		fmt.Fprintln(tw, "MATCHED\t-\t-\t-\t-\tAll transactions reconciled")
	}
}

func sortedBankNames(bankData map[string][]reconciliation.BankData) []string {
	names := make([]string, 0, len(bankData))
	for name := range bankData {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
