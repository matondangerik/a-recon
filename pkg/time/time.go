package time

import "time"

type DateOnly struct {
	time.Time
}

func (d *DateOnly) UnmarshalCSV(value string) error {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	d.Time = t.UTC()
	return nil
}
