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

func (s *MySource) GetValidTimeFrames() []string {
	return []string{"1m", "5m", "15m", "1h", "1d"}
}

func (s *MySource) GetTimeFrameSeconds(timeFrame string) int64 {
	switch timeFrame {
	case "1m":
		return 60
	case "5m":
		return 300
	case "15m":
		return 900
	case "1h":
		return 3600
	case "1d":
		return 86400
	default:
		return 0
	}
}

func (s *MySource) GetTimeFrameString(timeFrameSeconds int64) string {
	switch timeFrameSeconds {
	case 60:
		return "1m"
	case 300:
		return "5m"
	case 900:
		return "15m"
	case 3600:
		return "1h"
	case 86400:
		return "1d"
	default:
		return ""
	}
}

func (s *MySource) GetData(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	fmt.Printf("🔥 API CALL - GetData(%s, %s, %s, %s)\n",
		pairId, timeFrame, timeRange.Start.Format("2006-01-02 15:04"), timeRange.End.Format("2006-01-02 15:04"))

	// Generate data for each hour between start and end
	var entries []sources.ChartEntry
	current := timeRange.Start

	timeFrameSeconds := s.GetTimeFrameSeconds(timeFrame)
	if timeFrameSeconds == 0 {
		fmt.Printf("   → Invalid timeFrame, returning empty data\n")
		return entries, nil
	}

	for !current.After(timeRange.End) {
		entries = append(entries, sources.ChartEntry{
			Timestamp: current,
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		})
		current = current.Add(time.Duration(timeFrameSeconds) * time.Second)
	}

	fmt.Printf("   → Returned %d entries\n", len(entries))
	return entries, nil
}

func CreateMySource() sources.Source {
	return &MySource{Name: "MySource1"}
}
