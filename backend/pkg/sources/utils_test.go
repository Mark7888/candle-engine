package sources

import (
	"testing"
	"time"
)

func ts(n int) time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Hour)
}

func validEntry(n int, price float64) ChartEntry {
	return ChartEntry{
		Timestamp: ts(n),
		Open:      price,
		High:      price + 1,
		Low:       price - 1,
		Close:     price,
		Volume:    price * 10,
	}
}

// TestFillInvalidEntries_LeadingGap covers the chart rendering candles as
// wick-thin lines: a pair whose history starts later than the chart's other
// pairs gets placeholder entries in front, and their -999 sentinel used to
// survive the fill and drag the shared y-axis down with it.
func TestFillInvalidEntries_LeadingGap(t *testing.T) {
	entries := []ChartEntry{
		GenerateInvalidEntry(ts(0)),
		GenerateInvalidEntry(ts(1)),
		validEntry(2, 100),
		validEntry(3, 110),
	}

	filled := FillInvalidEntries(entries)

	for i := 0; i < 2; i++ {
		if filled[i].Close != 100 {
			t.Errorf("leading entry %d: expected back-fill from the first valid close 100, got %v",
				i, filled[i].Close)
		}
		if filled[i].Open < 0 || filled[i].High < 0 || filled[i].Low < 0 || filled[i].Volume < 0 {
			t.Errorf("leading entry %d still carries sentinel values: %+v", i, filled[i])
		}
		if !filled[i].Invalid {
			t.Errorf("leading entry %d lost its Invalid flag; filled values must stay marked", i)
		}
	}
}

func TestFillInvalidEntries_MiddleGapFillsForward(t *testing.T) {
	entries := []ChartEntry{
		validEntry(0, 100),
		GenerateInvalidEntry(ts(1)),
		validEntry(2, 130),
	}

	filled := FillInvalidEntries(entries)

	if filled[1].Close != 100 {
		t.Errorf("gap should be filled forward from the previous valid close 100, got %v", filled[1].Close)
	}
	if filled[0].Close != 100 || filled[2].Close != 130 {
		t.Errorf("valid entries must not be modified: %v, %v", filled[0].Close, filled[2].Close)
	}
}

func TestFillInvalidEntries_AllInvalid(t *testing.T) {
	entries := []ChartEntry{GenerateInvalidEntry(ts(0)), GenerateInvalidEntry(ts(1))}

	filled := FillInvalidEntries(entries)

	if len(filled) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(filled))
	}
	for i, e := range filled {
		if !e.Invalid {
			t.Errorf("entry %d should still be invalid", i)
		}
	}
}

func TestFillInvalidEntries_DoesNotMutateInput(t *testing.T) {
	entries := []ChartEntry{GenerateInvalidEntry(ts(0)), validEntry(1, 100)}

	FillInvalidEntries(entries)

	if entries[0].Close != -999 {
		t.Errorf("input slice was mutated: Close=%v", entries[0].Close)
	}
}

// TestIsValid_HonoursInvalidFlag pins the contract indicators rely on: a filled
// placeholder holds plausible numbers but is still not real market data.
func TestIsValid_HonoursInvalidFlag(t *testing.T) {
	filled := FillInvalidEntries([]ChartEntry{
		GenerateInvalidEntry(ts(0)),
		validEntry(1, 100),
	})

	if filled[0].IsValid() {
		t.Error("a filled placeholder entry must not report itself as valid")
	}
	if !filled[1].IsValid() {
		t.Error("a real entry must report itself as valid")
	}
}
