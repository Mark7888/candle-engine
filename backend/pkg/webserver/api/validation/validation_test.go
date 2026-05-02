package validation

import (
	"strings"
	"testing"
	"time"

	"candle-engine/pkg/charting/sources"
)

func TestValidateTimeRange_Valid(t *testing.T) {
	tr := sources.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateTimeRange_EqualStartEnd(t *testing.T) {
	ts := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	tr := sources.TimeRange{Start: ts, End: ts}
	errs := ValidateTimeRange(tr)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for equal start/end, got %v", errs)
	}
}

func TestValidateTimeRange_StartAfterEnd(t *testing.T) {
	tr := sources.TimeRange{
		Start: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	if len(errs) == 0 {
		t.Fatal("expected error for start after end, got none")
	}
	if !containsField(errs, "rangeFrom") {
		t.Errorf("expected rangeFrom error, got %v", errs)
	}
}

func TestValidateTimeRange_StartBeforeMinDate(t *testing.T) {
	tr := sources.TimeRange{
		Start: time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	if len(errs) == 0 {
		t.Fatal("expected error for start before MinDate, got none")
	}
	if !containsField(errs, "rangeFrom") {
		t.Errorf("expected rangeFrom error, got %v", errs)
	}
}

func TestValidateTimeRange_EndBeforeMinDate(t *testing.T) {
	tr := sources.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	if len(errs) == 0 {
		t.Fatal("expected error for end before MinDate, got none")
	}
	if !containsField(errs, "rangeTo") {
		t.Errorf("expected rangeTo error, got %v", errs)
	}
}

func TestValidateTimeRange_BothBeforeMinDate(t *testing.T) {
	tr := sources.TimeRange{
		Start: time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(1850, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	// Expect errors for both rangeFrom and rangeTo (start < end so no order error).
	if !containsField(errs, "rangeFrom") {
		t.Errorf("expected rangeFrom error, got %v", errs)
	}
	if !containsField(errs, "rangeTo") {
		t.Errorf("expected rangeTo error, got %v", errs)
	}
}

func TestValidateTimeRange_ExactlyMinDate(t *testing.T) {
	tr := sources.TimeRange{
		Start: MinDate,
		End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	errs := ValidateTimeRange(tr)
	if len(errs) != 0 {
		t.Fatalf("expected no errors when start equals MinDate, got %v", errs)
	}
}

// containsField returns true when any error in errs has the given field prefix.
func containsField(errs []error, field string) bool {
	for _, e := range errs {
		if strings.HasPrefix(e.Error(), field+":") {
			return true
		}
	}
	return false
}
