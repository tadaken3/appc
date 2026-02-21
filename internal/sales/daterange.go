package sales

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// DateRange returns a slice of date strings (YYYY-MM-DD) from start to end inclusive.
func DateRange(from, to string) ([]string, error) {
	start, err := time.Parse(dateLayout, from)
	if err != nil {
		return nil, fmt.Errorf("invalid --from date %q: %w: use YYYY-MM-DD", from, err)
	}
	end, err := time.Parse(dateLayout, to)
	if err != nil {
		return nil, fmt.Errorf("invalid --to date %q: %w: use YYYY-MM-DD", to, err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("--to must be after --from")
	}
	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateLayout))
	}
	return dates, nil
}
