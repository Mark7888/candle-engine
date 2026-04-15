package sources

// ForwardFillInvalidEntries fills OHLCV values of invalid entries from the last valid entry.
func ForwardFillInvalidEntries(entries []ChartEntry) []ChartEntry {
	result := make([]ChartEntry, len(entries))
	copy(result, entries)

	lastValidIdx := -1
	for i := range result {
		if !result[i].Invalid {
			lastValidIdx = i
		} else if lastValidIdx >= 0 {
			result[i].Open = result[lastValidIdx].Open
			result[i].High = result[lastValidIdx].High
			result[i].Low = result[lastValidIdx].Low
			result[i].Close = result[lastValidIdx].Close
			result[i].Volume = result[lastValidIdx].Volume
		}
	}
	return result
}
