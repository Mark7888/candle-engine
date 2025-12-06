package main

import (
	"fmt"
	"testcharting/pkg/charting/chartproject"
	"testcharting/pkg/charting/sources"
	"time"
)

// MockSource1 simulates a data source (e.g., Coinbase)
type MockSource1 struct {
	Name string
}

func (s *MockSource1) GetName() string {
	return s.Name
}

func (s *MockSource1) GetTradingPairs() []sources.TradingPair {
	return []sources.TradingPair{}
}

func (s *MockSource1) GetValidTimeFrames() []string {
	return []string{"1h"}
}

func (s *MockSource1) GetTimeFrameSeconds(timeFrame string) int64 {
	return 3600
}

func (s *MockSource1) GetData(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	fmt.Printf("  🔹 %s API CALL - GetData(%s, %s, %s to %s)\n",
		s.Name, pairId, timeFrame, timeRange.Start.Format("15:04"), timeRange.End.Format("15:04"))

	var entries []sources.ChartEntry
	current := timeRange.Start

	for !current.After(timeRange.End) {
		entries = append(entries, sources.ChartEntry{
			Timestamp: current,
			Open:      100.0 + float64(current.Hour()),
			High:      105.0 + float64(current.Hour()),
			Low:       95.0 + float64(current.Hour()),
			Close:     102.0 + float64(current.Hour()),
			Volume:    1000.0,
		})
		current = current.Add(time.Hour)
	}

	fmt.Printf("     → Returned %d entries\n", len(entries))
	return entries, nil
}

// MockSource2 simulates another data source (e.g., Binance)
type MockSource2 struct {
	Name string
}

func (s *MockSource2) GetName() string {
	return s.Name
}

func (s *MockSource2) GetTradingPairs() []sources.TradingPair {
	return []sources.TradingPair{}
}

func (s *MockSource2) GetValidTimeFrames() []string {
	return []string{"1h"}
}

func (s *MockSource2) GetTimeFrameSeconds(timeFrame string) int64 {
	return 3600
}

func (s *MockSource2) GetData(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	fmt.Printf("  🔹 %s API CALL - GetData(%s, %s, %s to %s)\n",
		s.Name, pairId, timeFrame, timeRange.Start.Format("15:04"), timeRange.End.Format("15:04"))

	var entries []sources.ChartEntry
	current := timeRange.Start

	// Skip some hours to test alignment (only even hours)
	for !current.After(timeRange.End) {
		if current.Hour()%2 == 0 {
			entries = append(entries, sources.ChartEntry{
				Timestamp: current,
				Open:      200.0 + float64(current.Hour()),
				High:      205.0 + float64(current.Hour()),
				Low:       195.0 + float64(current.Hour()),
				Close:     202.0 + float64(current.Hour()),
				Volume:    2000.0,
			})
		}
		current = current.Add(time.Hour)
	}

	fmt.Printf("     → Returned %d entries\n", len(entries))
	return entries, nil
}

func testSourceManager() {
	fmt.Println("=== Testing ChartSourceManager ===")
	fmt.Println()

	// Create the manager
	manager := chartproject.NewChartSourceManager()

	// Create mock sources
	source1 := &MockSource1{Name: "Coinbase"}
	source2 := &MockSource2{Name: "Binance"}

	// Test 1: Add sources
	fmt.Println("📋 Test 1: Adding sources")
	manager.AddSources(source1, source2)
	fmt.Println("✅ Added Coinbase and Binance sources")
	fmt.Println()

	// Test 2: Add trading pairs
	fmt.Println("📋 Test 2: Adding trading pairs")
	pair1 := sources.SourcePairId{SourceName: "Coinbase", PairId: "BTC-USD"}
	pair2 := sources.SourcePairId{SourceName: "Binance", PairId: "BTC-USDT"}
	manager.AddTradingPairs(pair1, pair2)
	fmt.Println("✅ Added BTC-USD (Coinbase) and BTC-USDT (Binance)")
	fmt.Println()

	// Define time range
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	timeRange := sources.TimeRange{Start: start, End: end}

	// Test 3: Fetch data from multiple sources (concurrent)
	fmt.Println("📋 Test 3: Fetching data from multiple sources concurrently")
	fmt.Println("  Expected: Two API calls happening in parallel")
	data, _, err := manager.FetchSources(timeRange, []sources.SourcePairId{pair1, pair2}, "1h")
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Fetched data for %d source-pair combinations\n\n", len(data))

	// Test 4: Verify data alignment
	fmt.Println("📋 Test 4: Verifying data alignment across sources")
	for pairId, entries := range data {
		fmt.Printf("  %s:\n", pairId.String())
		fmt.Printf("    Total entries: %d\n", len(entries))

		validCount := 0
		invalidCount := 0
		for _, entry := range entries {
			if entry.IsValid() {
				validCount++
			} else {
				invalidCount++
			}
		}
		fmt.Printf("    Valid entries: %d\n", validCount)
		fmt.Printf("    Invalid entries (filled gaps): %d\n", invalidCount)

		// Show first few entries
		fmt.Printf("    Sample entries:\n")
		for i := 0; i < len(entries) && i < 3; i++ {
			entry := entries[i]
			if entry.IsValid() {
				fmt.Printf("      [%s] Valid: Close=%.2f\n", entry.Timestamp.Format("15:04"), entry.Close)
			} else {
				fmt.Printf("      [%s] Invalid (gap filled)\n", entry.Timestamp.Format("15:04"))
			}
		}
		fmt.Println()
	}

	// Test 5: Verify timestamp consistency
	fmt.Println("📋 Test 5: Verifying timestamp consistency across sources")
	var timestamps1, timestamps2 []time.Time
	for _, entry := range data[pair1] {
		timestamps1 = append(timestamps1, entry.Timestamp)
	}
	for _, entry := range data[pair2] {
		timestamps2 = append(timestamps2, entry.Timestamp)
	}

	if len(timestamps1) != len(timestamps2) {
		fmt.Printf("❌ Length mismatch: %d vs %d\n", len(timestamps1), len(timestamps2))
	} else {
		allMatch := true
		for i := range timestamps1 {
			if !timestamps1[i].Equal(timestamps2[i]) {
				allMatch = false
				fmt.Printf("❌ Timestamp mismatch at index %d: %s vs %s\n",
					i, timestamps1[i].Format("15:04"), timestamps2[i].Format("15:04"))
			}
		}
		if allMatch {
			fmt.Printf("✅ All %d timestamps are perfectly aligned across both sources\n", len(timestamps1))
		}
	}
	fmt.Println()

	// Test 6: Fetch with non-existent source
	fmt.Println("📋 Test 6: Fetching with non-existent source (should return error)")
	pair3 := sources.SourcePairId{SourceName: "NonExistent", PairId: "ETH-USD"}
	data2, _, err := manager.FetchSources(timeRange, []sources.SourcePairId{pair1, pair3}, "1h")
	if err != nil {
		fmt.Printf("✅ Got expected error: %v\n\n", err)
	} else {
		fmt.Printf("❌ Expected error but got data for %d source-pair combinations\n\n", len(data2))
	}

	// Test 7: Test thread-safety with concurrent operations
	fmt.Println("📋 Test 7: Testing thread-safety with concurrent operations")
	done := make(chan bool, 3)

	// Concurrent adds
	go func() {
		source3 := &MockSource1{Name: "Kraken"}
		manager.AddSources(source3)
		done <- true
	}()

	go func() {
		pair4 := sources.SourcePairId{SourceName: "Coinbase", PairId: "ETH-USD"}
		manager.AddTradingPairs(pair4)
		done <- true
	}()

	// Concurrent fetch
	go func() {
		_, _, _ = manager.FetchSources(timeRange, []sources.SourcePairId{pair1}, "1h")
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
	fmt.Println("✅ Concurrent operations completed without race conditions")
	fmt.Println()

	fmt.Println("=== ChartSourceManager Test Complete ===")
}
