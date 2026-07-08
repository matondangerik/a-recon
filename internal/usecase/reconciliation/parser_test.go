package reconciliation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSystemData(t *testing.T) {
	r := &recon{}
	dir := t.TempDir()
	path := writeTempCSV(t, dir, "system.csv", ""+
		"trx_id,amount,type,transaction_time,bank_name\n"+
		"trx-1,100,CREDIT,2026-01-10T10:00:00Z,bank-a\n"+
		"trx-2,50,DEBIT,2026-01-10T11:00:00Z,bank-b\n"+
		"trx-3,70,CREDIT,2026-01-15T11:00:00Z,bank-a\n"+
		"trx-4,80,CREDIT,2026-01-10T12:00:00Z,bank-c\n",
	)

	result, err := r.parseSystemData(context.Background(), Range{
		Start: at(2026, time.January, 10, 0, 0),
		End:   at(2026, time.January, 12, 0, 0),
	}, path, map[string]string{
		"bank-a": "bank-a.csv",
		"bank-b": "bank-b.csv",
	})

	assert.NoError(t, err)
	if !assert.Len(t, result, 2) {
		return
	}
	if assert.Len(t, result["bank-a"], 1) {
		assert.Equal(t, "trx-1", result["bank-a"][0].TrxID)
	}
	if assert.Len(t, result["bank-b"], 1) {
		assert.Equal(t, "trx-2", result["bank-b"][0].TrxID)
	}
}

func TestParseBankData(t *testing.T) {
	r := &recon{}
	dir := t.TempDir()
	path := writeTempCSV(t, dir, "bank.csv", ""+
		"unique_identifier,amount,date\n"+
		"bank-1,100,2026-01-10T00:00:00Z\n"+
		"bank-2,120,2026-01-12T00:00:00Z\n"+
		"bank-3,130,2026-01-15T00:00:00Z\n",
	)

	result, err := r.parseBankData(context.Background(), Range{
		Start: at(2026, time.January, 10, 0, 0),
		End:   at(2026, time.January, 12, 0, 0),
	}, "bank-a", path)

	assert.NoError(t, err)
	if !assert.Len(t, result, 2) {
		return
	}
	assert.Equal(t, "bank-1", result[0].Unique_Identifier)
	assert.Equal(t, "bank-2", result[1].Unique_Identifier)
	assert.Equal(t, "bank-a", result[0].BankName)
	assert.Equal(t, "bank-a", result[1].BankName)
}

func TestParseReconciliationData(t *testing.T) {
	r := &recon{}
	dir := t.TempDir()

	systemPath := writeTempCSV(t, dir, "system.csv", ""+
		"trx_id,amount,type,transaction_time,bank_name\n"+
		"trx-1,100,CREDIT,2026-01-10T10:00:00Z,bank-a\n"+
		"trx-2,50,DEBIT,2026-01-10T11:00:00Z,bank-b\n",
	)
	bankAPath := writeTempCSV(t, dir, "bank-a.csv", ""+
		"unique_identifier,amount,date\n"+
		"bank-a-1,100,2026-01-10T00:00:00Z\n",
	)
	bankBPath := writeTempCSV(t, dir, "bank-b.csv", ""+
		"unique_identifier,amount,date\n"+
		"bank-b-1,-50,2026-01-10T00:00:00Z\n",
	)

	systemData, bankData, err := r.parseReconciliationData(context.Background(), ReconciliationArgs{
		File: File{
			SystemPath: systemPath,
			BankPath: map[string]string{
				"bank-a": bankAPath,
				"bank-b": bankBPath,
			},
		},
		Range: Range{
			Start: at(2026, time.January, 10, 0, 0),
			End:   at(2026, time.January, 11, 0, 0),
		},
	})

	assert.NoError(t, err)
	if !assert.Len(t, systemData, 2) {
		return
	}
	if assert.Len(t, bankData, 2) {
		if assert.Len(t, bankData["bank-a"], 1) {
			assert.Equal(t, "bank-a", bankData["bank-a"][0].BankName)
		}
		if assert.Len(t, bankData["bank-b"], 1) {
			assert.Equal(t, "bank-b", bankData["bank-b"][0].BankName)
		}
	}
}

func writeTempCSV(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}
	return path
}
