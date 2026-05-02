// Package validation provides reusable input-validation helpers for the API
// layer.  Every rule that can be expressed without hitting external services
// (e.g. date-range semantics) belongs here so that it is easy to extend and
// to test in isolation.
package validation

import (
	"fmt"
	"time"

	"candle-engine/pkg/charting/sources"
)

// MinDate is the earliest date accepted as a valid range boundary.
var MinDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

// ValidationError describes a single failed validation rule.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateTimeRange checks the semantic validity of a TimeRange:
//   - Both Start and End must be on or after MinDate.
//   - Start must not be after End.
//
// It returns a slice of ValidationErrors; a nil slice means the range is valid.
func ValidateTimeRange(tr sources.TimeRange) []error {
	var errs []error

	if tr.Start.Before(MinDate) {
		errs = append(errs, &ValidationError{
			Field:   "rangeFrom",
			Message: fmt.Sprintf("must be on or after %s", MinDate.Format("2006-01-02")),
		})
	}
	if tr.End.Before(MinDate) {
		errs = append(errs, &ValidationError{
			Field:   "rangeTo",
			Message: fmt.Sprintf("must be on or after %s", MinDate.Format("2006-01-02")),
		})
	}
	if tr.Start.After(tr.End) {
		errs = append(errs, &ValidationError{
			Field:   "rangeFrom",
			Message: "must not be later than rangeTo",
		})
	}

	return errs
}
