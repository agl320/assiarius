package screener

import (
	"strings"
	"time"
)

var finvizNewsLayouts = []string{
	"Jan-02-06 03:04PM",
	"Jan-2-06 03:04PM",
	"Jan-02-06 3:04PM",
	"Jan-2-06 3:04PM",
}

var finvizNewsTimeOnlyLayouts = []string{
	"03:04PM",
	"3:04PM",
}

// parseFinvizNewsTimestamp parses the Finviz quote-page news time cell.
//
// The first row for a date typically includes a date+time, while subsequent rows
// only include the time. For time-only rows, lastDate is used.
func parseFinvizNewsTimestamp(s string, lastDate time.Time, loc *time.Location) (ts time.Time, newLastDate time.Time, ok bool) {
	s = strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	if s == "" {
		return time.Time{}, lastDate, false
	}
	if loc == nil {
		loc = time.Local
	}

	// Date + time
	for _, layout := range finvizNewsLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
			return t, midnight, true
		}
	}

	// Time only (use last known date)
	if lastDate.IsZero() {
		return time.Time{}, lastDate, false
	}
	for _, layout := range finvizNewsTimeOnlyLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			combined := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), t.Hour(), t.Minute(), 0, 0, loc)
			return combined, lastDate, true
		}
	}

	return time.Time{}, lastDate, false
}
