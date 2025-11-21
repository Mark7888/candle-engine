package main

import (
	"fmt"
	"testcharting/pkg/charting/cachedsource"
	"testcharting/pkg/charting/sources"
	"time"
)

func testCaching() {
	fmt.Println("=== Testing Cached Source ===")
	fmt.Println()

	// Create a basic source and wrap it with caching
	basicSource := CreateMySource()
	cachedSource := cachedsource.NewCachedSource(basicSource, "./cache.json")

	// Define time ranges
	start1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	timeRange1 := sources.TimeRange{Start: start1, End: end1}

	start2 := time.Date(2025, 1, 1, 5, 0, 0, 0, time.UTC)
	end2 := time.Date(2025, 1, 1, 6, 0, 0, 0, time.UTC)
	timeRange2 := sources.TimeRange{Start: start2, End: end2}

	start3 := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	end3 := time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC)
	timeRange3 := sources.TimeRange{Start: start3, End: end3}

	start4 := time.Date(2025, 1, 1, 3, 0, 0, 0, time.UTC)
	end4 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	timeRange4 := sources.TimeRange{Start: start4, End: end4}

	// Test 1: First request - should make API call
	fmt.Println("📋 Test 1: Request 2025-01-01 00:00 to 10:00 (should make API call)")
	data1, err := cachedSource.GetData("BTC-USD", "1h", timeRange1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data1))

	// Test 2: Request subset of cached data - should NOT make API call
	fmt.Println("📋 Test 2: Request 2025-01-01 05:00 to 06:00 (fully cached, no API call)")
	data2, err := cachedSource.GetData("BTC-USD", "1h", timeRange2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data2))

	// Test 3: Request partial overlap - should make API call for missing part
	fmt.Println("📋 Test 3: Request 2025-01-01 08:00 to 15:00 (partial cache, API call for 10:00-15:00)")
	data3, err := cachedSource.GetData("BTC-USD", "1h", timeRange3)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data3))

	// Test 4: Request that spans cached ranges - should fill in the gaps
	fmt.Println("📋 Test 4: Request 2025-01-01 03:00 to 12:00 (needs 03:00-00:00 only)")
	data4, err := cachedSource.GetData("BTC-USD", "1h", timeRange4)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data4))

	// Test 5: Same request again - should be fully cached
	fmt.Println("📋 Test 5: Request 2025-01-01 00:00 to 10:00 again (fully cached)")
	data5, err := cachedSource.GetData("BTC-USD", "1h", timeRange1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data5))

	// Test 6: Different interval - should make API call
	fmt.Println("📋 Test 6: Request with different interval 5m (new cache key)")
	data6, err := cachedSource.GetData("BTC-USD", "5m", timeRange1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✅ Received %d entries\n\n", len(data6))

	fmt.Println("=== Caching Test Complete ===")
}
