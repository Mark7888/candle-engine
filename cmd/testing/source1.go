package main

import (
	"fmt"
	"testcharting/pkg/charting/sources"
	"time"
)

type MySource struct {
	Name string
}

func (s *MySource) GetName() string {
	return s.Name
}

func (s *MySource) GetTradingPairs() []sources.TradingPair {
	return []sources.TradingPair{}
}

func (s *MySource) GetValidIntervals() []string {
	return []string{"1m", "5m", "15m", "1h", "1d"}
}

func (s *MySource) GetIntervalSeconds(interval string) []int64 {
	switch interval {
	case "1m":
		return []int64{60}
	case "5m":
		return []int64{300}
	case "15m":
		return []int64{900}
	case "1h":
		return []int64{3600}
	case "1d":
		return []int64{86400}
	default:
		return []int64{}
	}
}

func (s *MySource) GetData(pairId string, interval string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	fmt.Printf("🔥 API CALL - GetData(%s, %s, %s, %s)\n",
		pairId, interval, timeRange.Start.Format("2006-01-02 15:04"), timeRange.End.Format("2006-01-02 15:04"))

	// Generate data for each hour between start and end
	var entries []sources.ChartEntry
	current := timeRange.Start

	intervalSecondsSlice := s.GetIntervalSeconds(interval)
	if len(intervalSecondsSlice) == 0 {
		fmt.Printf("   → Invalid interval, returning empty data\n")
		return entries, nil
	}
	intervalSeconds := intervalSecondsSlice[0]

	for !current.After(timeRange.End) {
		entries = append(entries, sources.ChartEntry{
			Timestamp: current,
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		})
		current = current.Add(time.Duration(intervalSeconds) * time.Second)
	}

	fmt.Printf("   → Returned %d entries\n", len(entries))
	return entries, nil
}

func CreateMySource() sources.Source {
	return &MySource{Name: "MySource1"}
}
