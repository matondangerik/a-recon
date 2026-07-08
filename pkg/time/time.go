package time

import "time"

// Date represents a date.
type Date struct {
	time.Time
}

func (d *Date) UnmarshalCSV(value string) error {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	d.Time = t.UTC()
	return nil
}
