package time

import (
	stdtime "time"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateOnlyUnmarshalCSV(t *testing.T) {
	var d DateOnly

	err := d.UnmarshalCSV("2026-01-10")

	assert.NoError(t, err)
	assert.True(t, d.Time.Equal(stdtime.Date(2026, stdtime.January, 10, 0, 0, 0, 0, stdtime.UTC)))
	assert.Equal(t, stdtime.UTC, d.Location())
}

func TestDateOnlyUnmarshalCSVRejectsTimestamp(t *testing.T) {
	var d DateOnly

	err := d.UnmarshalCSV("2026-01-10T00:00:00Z")

	assert.Error(t, err)
}
