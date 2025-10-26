package main

import (
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

func (s *MySource) GetData(pairId string, interval string, start time.Time, end time.Time) ([]sources.ChartEntry, error) {
	return []sources.ChartEntry{
		{
			Open:   1.00,
			High:   1.05,
			Low:    0.95,
			Close:  1.00,
			Volume: 1000,
		},
		{
			Open:   1.00,
			High:   1.10,
			Low:    0.90,
			Close:  1.05,
			Volume: 1500,
		},
		{
			Open:   2.00,
			High:   2.05,
			Low:    1.95,
			Close:  2.00,
			Volume: 2000,
		},
	}, nil
}

func CreateMySource() sources.Source {
	return &MySource{Name: "MySource1"}
}
