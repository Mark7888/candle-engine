package sources

// FillInvalidEntries replaces the sentinel OHLCV values of invalid entries with
// the values of the nearest valid entry, so that the placeholder numbers from
// GenerateInvalidEntry never reach indicator math or the chart.
//
// Gaps are filled forward from the last valid entry. A leading run of invalid
// entries has no preceding valid entry, so it is filled backwards from the
// first valid one instead – otherwise those entries would keep their sentinel
// values and drag the chart's y-axis down to -999, squashing every real candle
// into a wick-thin line.
//
// Entries keep their Invalid flag, so callers can still tell filled placeholders
// apart from real data.
func FillInvalidEntries(entries []ChartEntry) []ChartEntry {
	result := make([]ChartEntry, len(entries))
	copy(result, entries)

	copyValues := func(dst *ChartEntry, src ChartEntry) {
		dst.Open = src.Open
		dst.High = src.High
		dst.Low = src.Low
		dst.Close = src.Close
		dst.Volume = src.Volume
	}

	lastValidIdx := -1
	for i := range result {
		if !result[i].Invalid {
			lastValidIdx = i
		} else if lastValidIdx >= 0 {
			copyValues(&result[i], result[lastValidIdx])
		}
	}

	// Back-fill the leading run of invalid entries (if any) from the first
	// valid entry, which is the one the forward pass started from.
	firstValidIdx := -1
	for i := range result {
		if !result[i].Invalid {
			firstValidIdx = i
			break
		}
	}
	if firstValidIdx > 0 {
		for i := 0; i < firstValidIdx; i++ {
			copyValues(&result[i], result[firstValidIdx])
		}
	}

	return result
}
