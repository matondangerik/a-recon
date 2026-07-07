package csv

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Read(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		assert.Error(t, ReadFile("random", struct{}{}))
	})

	t.Run("ok", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "csv-*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		records := [][]string{
			{"ID", "Name"},
			{"1", "Erik"},
		}

		writer := csv.NewWriter(tmpFile)
		err = writer.WriteAll(records)
		if err != nil {
			t.Fatalf("fatal to write CSV records: %v", err)
		}

		var temp []struct {
			ID   int64  `csv:"ID"`
			Name string `csv:"Name"`
		}

		err = ReadFile(tmpFile.Name(), &temp)
		assert.NoError(t, err)
		assert.NotEmpty(t, temp)
		assert.Equal(t, "Erik", temp[0].Name)
		assert.Equal(t, int64(1), temp[0].ID)
	})
}
