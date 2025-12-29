# Custom Indicator Development Guide

> **IMPORTANT**: This indicator system is currently in active development and APIs may change in future versions. Please refer to this documentation for the latest information.

This guide provides comprehensive documentation for creating custom indicators in the charting system. It covers all interfaces, functions, configuration options, and best practices.

---

## Table of Contents

1. [Overview](#overview)
2. [Core Interface](#core-interface)
3. [Indicator Types](#indicator-types)
4. [Configuration](#configuration)
5. [Data Access](#data-access)
6. [Output Handling](#output-handling)
7. [Logging](#logging)
8. [Error Handling](#error-handling)
9. [Simple Examples](#simple-examples)
10. [Complete Example](#complete-example)

---

## Overview

Custom indicators are implemented by satisfying the `BaseIndicator` interface. Indicators process chart data iteratively from **oldest to newest** candles, allowing them to build state and calculate values progressively.

### Key Concepts

- **Iterative Processing**: HandleEntry is called for each candle, from oldest to newest
- **Historical Access**: Access current and previous candle data using indexed lookups
- **Multi-Source Support**: Work with data from multiple trading pairs simultaneously
- **Flexible Output**: Support for single values, multiple plots, and various visualization types
- **Real-time Logging**: Debug and monitor indicator behavior with built-in logging

---

## Core Interface

### BaseIndicator

All custom indicators must implement this interface:

```go
type BaseIndicator interface {
    GetName() string
    GetType() IndicatorType
    GetConfig() IndicatorConfig
    HandleEntry(points PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config ChartConfig, handler IndicatorOutputHandler) (IndicatorOutputHandler, error)
}
```

> **Note on Struct Implementation**: Your indicator structs can contain additional fields and methods beyond the interface requirements. However, **avoid using mutable state fields** as they can lead to repainting issues and false data across different chart contexts. Only use struct fields for immutable configuration (like period settings). State should be maintained through the `points` parameter instead.

#### Methods

**`GetName() string`**
- Returns the display name of your indicator
- Example: `"Simple Moving Average (SMA)"`

**`GetType() IndicatorType`**
- Returns the visualization type for your indicator
- See [Indicator Types](#indicator-types) for available options

**`GetConfig() IndicatorConfig`**
- Returns configuration for rendering and display
- See [Configuration](#configuration) for details

**`HandleEntry(...) (IndicatorOutputHandler, error)`**
- Core calculation logic executed for each candle
- **Parameters**:
  - `points PlotPointCollectionHistory`: History of previous indicator outputs with helper methods
  - `chartsData sources.MultiSourceEntries`: All available chart data from loaded sources
  - `config ChartConfig`: Chart configuration (timeframe, active pair, etc.)
  - `handler IndicatorOutputHandler`: Output handler for this iteration
- **Returns**:
  - `IndicatorOutputHandler`: Updated handler with calculated values
  - `error`: Non-nil for fatal errors (stops indicator processing)

---

## Indicator Types

Indicators can be rendered in different visualization styles. Choose the appropriate type for your data.

### IndicatorType Enum

```go
type IndicatorType int

const (
    IndicatorTypeLine IndicatorType = iota
    IndicatorTypeStepLine
    IndicatorTypeTable
    IndicatorTypeHistogram
)
```

### Type Descriptions

| Type | Value | Description | Use Case |
|------|-------|-------------|----------|
| **IndicatorTypeLine** | 0 | Smooth continuous line plot | Moving averages, oscillators, price-based indicators |
| **IndicatorTypeStepLine** | 1 | Step-wise line with alignment options | Discrete signals, levels, binary indicators |
| **IndicatorTypeTable** | 2 | Tabular data display | Statistical summaries, multi-metric displays |
| **IndicatorTypeHistogram** | 3 | Vertical bar histogram | Volume analysis, distribution displays |

### Future Indicator Types (Coming Soon)

The following indicator types are planned for future releases:

| Type | Description | Use Case |
|------|-------------|----------|
| **IndicatorTypeBarChart** | Bar chart visualization | Comparative analysis, multi-value comparisons |
| **IndicatorTypePieChart** | Pie chart for proportional data | Portfolio allocation, percentage breakdowns |
| **IndicatorTypeAreaChart** | Filled area charts | Cumulative data, volume profiles |
| **IndicatorTypeCandleStick** | Custom candlestick overlays | Alternative price representations, synthetic instruments |

---

## Configuration

### ChartConfig

Contains information about the chart the indicator is running on. This is **provided to your indicator** in HandleEntry.

```go
type ChartConfig struct {
    TimeFrameSeconds int64                  // Candle timeframe in seconds
    ActivePair       sources.SourcePairId   // The currently active trading pair
    LoadedPairs      []sources.SourcePairId // All available trading pairs
}
```

**Usage**:
- Use `config.ActivePair` to make your indicator **reactive** to the current chart
- Use `config.LoadedPairs` to work with multiple pairs for cross-pair analysis
- Use `config.TimeFrameSeconds` for time-based calculations

### IndicatorConfig

Defines how your indicator should be rendered. This is **returned by your indicator** in GetConfig().

```go
type IndicatorConfig struct {
    TargetChart string
    
    LineIndicator      LineIndicatorConfig
    StepLineIndicator  StepLineIndicatorConfig
    TableIndicator     TableIndicatorConfig
    HistogramIndicator HistogramIndicatorConfig
}
```

#### TargetChart

- **Type**: `string`
- **Purpose**: Groups indicators into separate chart panels
- **Values**:
  - `""` (empty string): Display on the main candlestick chart
  - `"custom_panel"`: Create/use a custom panel with this ID
- **Example**: Oscillators typically use their own panel, overlays use `""`

#### LineIndicatorConfig

For `IndicatorTypeLine`:

```go
type LineIndicatorConfig struct {
    LineThickness int      // Width of the line in pixels
    LineTitles    []string // Names for each line (for multi-line indicators)
}
```

**Example**:
```go
LineIndicator: indicators.LineIndicatorConfig{
    LineThickness: 2,
    LineTitles:    []string{"Fast MA", "Slow MA"},
}
```

#### StepLineIndicatorConfig

For `IndicatorTypeStepLine`:

```go
type StepLineIndicatorConfig struct {
    StepAlignment StepAlignment // How steps align with candles
    LineTitles    []string      // Names for each line
}
```

**StepAlignment Values**:
- `StepAlignStart` (0): Step occurs at candle start
- `StepAlignMidpoint` (1): Step occurs at candle midpoint
- `StepAlignEnd` (2): Step occurs at candle end

#### TableIndicatorConfig

For `IndicatorTypeTable`:

```go
type TableIndicatorConfig struct {
    DisplayName        string   // Title of the table
    RowTitles          []string // Row labels (defines row count)
    ColumnTitles       []string // Column labels (defines column count)
    ValueDisplayFormat string   // Format string (e.g., "%.2f")
}
```

#### HistogramIndicatorConfig

For `IndicatorTypeHistogram`:

```go
type HistogramIndicatorConfig struct{}
```

Currently no configuration options for histograms.

---

## Data Access

### MultiSourceEntries

The `chartsData` parameter provides access to all chart data from loaded sources.

#### Key Methods

**`GetEntry(pairId SourcePairId, index int) *ChartEntry`**
- Retrieves a single candle by pair and index
- **Index 0** = Most recent (current) candle
- **Index 1** = Previous candle
- **Index n** = Candle n positions back in time
- Returns `nil` if pair doesn't exist or index out of bounds

**`GetEntryCount(pairId SourcePairId) int`**
- Returns the number of available candles for a pair
- Returns `0` if pair doesn't exist
- Useful for validating sufficient data before calculations

**`GetEntries(pairId SourcePairId) []ChartEntry`**
- Returns all candles for a pair
- Returns `nil` if pair doesn't exist

### ChartEntry Structure

```go
type ChartEntry struct {
    Timestamp time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    float64
}
```

**Methods**:
- `IsValid() bool`: Returns false if entry contains invalid data (-999 values)

### Making Indicators Pair-Specific

To make an indicator react to the currently displayed chart:

```go
func (i *MyIndicator) HandleEntry(points []PlotPointCollection, chartsData sources.MultiSourceEntries, config ChartConfig, handler IndicatorOutputHandler) (IndicatorOutputHandler, error) {
    // Use config.ActivePair to get data for the current chart
    current := chartsData.GetEntry(config.ActivePair, 0)
    
    // Your calculation using current pair's data
    // ...
}
```

Without using `config.ActivePair`, the indicator will show the same values across all pairs.

---

## Output Handling

### PlotPoint

Represents a single data point with optional styling:

```go
type PlotPoint struct {
    Value float64 // The numeric value to plot
    Color string  // CSS color string (e.g., "#00ff00", "red")
}
```

### PlotPointCollection

Container for one or more plot points:

```go
type PlotPointCollection struct {
    Points  []PlotPoint
    Invalid bool // If true, no data will be rendered for this iteration
}
```

**Creating Collections**:
```go
// Single point
collection := indicators.NewPlotPointCollection([]indicators.PlotPoint{
    {Value: 50.0, Color: "#00ff00"},
})

// Multiple points
collection := indicators.NewPlotPointCollection([]indicators.PlotPoint{
    {Value: 20.0, Color: "#ff0000"},
    {Value: 80.0, Color: "#00ff00"},
})

// Invalid/no data
collection := indicators.NewInvalidPlotPointCollection()
```

### Setting Output

**Single Value**:
```go
handler.SetSingleOutput(indicators.PlotPoint{
    Value: calculatedValue,
    Color: "#00ffff",
})
```

**Multiple Values**:
```go
handler.SetOutput(indicators.NewPlotPointCollection([]indicators.PlotPoint{
    {Value: fastMA, Color: "#00ff00"},
    {Value: slowMA, Color: "#ff0000"},
}))
```

**No Output** (insufficient data):
```go
handler.SetOutput(indicators.NewInvalidPlotPointCollection())
```

### Accessing Previous Values

The `points` parameter is a `PlotPointCollectionHistory` that provides access to previous indicator outputs with convenient helper methods.

#### PlotPointCollectionHistory

```go
type PlotPointCollectionHistory struct {
    Collections []PlotPointCollection
}
```

**Key Methods**:
- `GetCollection(index int) *PlotPointCollection`: Get collection at index (0 = most recent, 1 = previous, etc.)
- `Length() int`: Returns the number of historical collections available
- `GetAllCollections() []PlotPointCollection`: Returns all collections as a slice

**Index Convention**: Like chart data, index 0 is the most recent (current) output, index 1 is the previous, and so on.

**Example Usage**:

```go
func (i *MyIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Check if we have previous data
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() {
            previousValue := prevCollection.Points[0].Value
            // Use previousValue in calculation
        }
    }
    
    // Access multiple historical values
    if points.Length() >= 3 {
        current := points.GetCollection(0)   // Most recent
        previous := points.GetCollection(1)  // One back
        twoPrev := points.GetCollection(2)   // Two back
        // Use for trend analysis
    }
    
    // ...
}
```

---

## Logging

Indicators can log messages that appear in the UI for debugging and monitoring.

### Log Levels

| Function | Level | Use Case |
|----------|-------|----------|
| `handler.LogDebug(message)` | Debug | Detailed diagnostic information |
| `handler.LogInfo(message)` | Info | General informational messages |
| `handler.LogWarning(message)` | Warning | Potential issues or unusual conditions |
| `handler.LogError(message)` | Error | Non-fatal errors (calculation continues) |

### Example

```go
func (i *MyIndicator) HandleEntry(points []PlotPointCollection, chartsData sources.MultiSourceEntries, config ChartConfig, handler IndicatorOutputHandler) (IndicatorOutputHandler, error) {
    period := 20
    
    if chartsData.GetEntryCount(config.ActivePair) < period {
        handler.LogWarning("Insufficient data for calculation")
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    handler.LogDebug("Calculating with 20 periods of data")
    // Perform calculation
    
    return handler, nil
}
```

---

## Error Handling

### Fatal Errors

Return a non-nil error to stop indicator processing completely:

```go
if critical_condition {
    return handler, fmt.Errorf("critical failure: reason")
}
```

### Non-Fatal Issues

Use logging and invalid output for recoverable issues:

```go
if insufficient_data {
    handler.LogWarning("Not enough data, skipping")
    handler.SetOutput(indicators.NewInvalidPlotPointCollection())
    return handler, nil
}
```

---

## Simple Examples

### Example 1: Last Close Price Display

A simple indicator that displays the closing price of the current candle.

```go
package sample

import (
    "testcharting/pkg/charting/sources"
    "testcharting/pkg/indicators"
)

type LastPriceIndicator struct{}

func (l *LastPriceIndicator) GetName() string {
    return "Last Close Price"
}

func (l *LastPriceIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeLine
}

func (l *LastPriceIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "", // Display on main chart
        LineIndicator: indicators.LineIndicatorConfig{
            LineThickness: 2,
            LineTitles:    []string{"Close"},
        },
    }
}

func (l *LastPriceIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Get the current candle for the active pair
    current := chartsData.GetEntry(config.ActivePair, 0)
    
    if current == nil {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: current.Close,
        Color: "#ffffff",
    })
    
    return handler, nil
}
```

### Example 2: Simple Price Change

Shows the difference between current and previous close prices.

```go
package sample

import (
    "testcharting/pkg/charting/sources"
    "testcharting/pkg/indicators"
)

type PriceChangeIndicator struct{}

func (p *PriceChangeIndicator) GetName() string {
    return "Price Change"
}

func (p *PriceChangeIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeLine
}

func (p *PriceChangeIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "price_change", // Separate panel
        LineIndicator: indicators.LineIndicatorConfig{
            LineThickness: 1,
            LineTitles:    []string{"Change"},
        },
    }
}

func (p *PriceChangeIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Need at least 2 candles
    if chartsData.GetEntryCount(config.ActivePair) < 2 {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    current := chartsData.GetEntry(config.ActivePair, 0)
    previous := chartsData.GetEntry(config.ActivePair, 1)
    
    change := current.Close - previous.Close
    color := "#00ff00" // Green for positive
    if change < 0 {
        color = "#ff0000" // Red for negative
    }
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: change,
        Color: color,
    })
    
    return handler, nil
}
```

### Example 3: High/Low Range

Displays the difference between high and low for each candle.

```go
package sample

import (
    "testcharting/pkg/charting/sources"
    "testcharting/pkg/indicators"
)

type RangeIndicator struct{}

func (r *RangeIndicator) GetName() string {
    return "Candle Range"
}

func (r *RangeIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeHistogram
}

func (r *RangeIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "range",
        HistogramIndicator: indicators.HistogramIndicatorConfig{},
    }
}

func (r *RangeIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    current := chartsData.GetEntry(config.ActivePair, 0)
    
    if current == nil || !current.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    rangeValue := current.High - current.Low
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: rangeValue,
        Color: "#ffaa00",
    })
    
    return handler, nil
}
```

---

## Complete Example

### Simple Moving Average (SMA) with Crossover Detection

This comprehensive example demonstrates:
- Data validation
- Historical data access
- Multi-value output
- Conditional coloring
- State tracking using previous indicator values
- Logging
- Error handling

```go
package sample

import (
    "fmt"
    "testcharting/pkg/charting/sources"
    "testcharting/pkg/indicators"
)

type SimpleMovingAverageIndicator struct {
    FastPeriod int
    SlowPeriod int
}

func NewSMAIndicator() *SimpleMovingAverageIndicator {
    return &SimpleMovingAverageIndicator{
        FastPeriod: 10,
        SlowPeriod: 20,
    }
}

func (s *SimpleMovingAverageIndicator) GetName() string {
    return fmt.Sprintf("SMA (%d/%d)", s.FastPeriod, s.SlowPeriod)
}

func (s *SimpleMovingAverageIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeLine
}

func (s *SimpleMovingAverageIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "", // Display on main chart as overlay
        LineIndicator: indicators.LineIndicatorConfig{
            LineThickness: 2,
            LineTitles:    []string{"Fast SMA", "Slow SMA"},
        },
    }
}

func (s *SimpleMovingAverageIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Validate input parameters
    if s.FastPeriod <= 0 || s.SlowPeriod <= 0 {
        return handler, fmt.Errorf("invalid period configuration: fast=%d, slow=%d", s.FastPeriod, s.SlowPeriod)
    }
    
    if s.FastPeriod >= s.SlowPeriod {
        handler.LogWarning("Fast period should be less than slow period")
    }
    
    // Check if we have enough data
    requiredPeriod := s.SlowPeriod
    availableData := chartsData.GetEntryCount(config.ActivePair)
    
    if availableData < requiredPeriod {
        handler.LogDebug(fmt.Sprintf("Insufficient data: need %d, have %d", requiredPeriod, availableData))
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    // Calculate Fast SMA
    var fastSum float64
    for i := 0; i < s.FastPeriod; i++ {
        entry := chartsData.GetEntry(config.ActivePair, i)
        if entry == nil || !entry.IsValid() {
            handler.LogError(fmt.Sprintf("Invalid data at index %d", i))
            handler.SetOutput(indicators.NewInvalidPlotPointCollection())
            return handler, nil
        }
        fastSum += entry.Close
    }
    fastSMA := fastSum / float64(s.FastPeriod)
    
    // Calculate Slow SMA
    var slowSum float64
    for i := 0; i < s.SlowPeriod; i++ {
        entry := chartsData.GetEntry(config.ActivePair, i)
        slowSum += entry.Close
    }
    slowSMA := slowSum / float64(s.SlowPeriod)
    
    // Detect crossovers using previous indicator values
    var fastColor, slowColor string
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() && len(prevCollection.Points) >= 2 {
            prevFastSMA := prevCollection.Points[0].Value
            prevSlowSMA := prevCollection.Points[1].Value
        
        // Bullish crossover: fast crosses above slow
        if prevFastSMA <= prevSlowSMA && fastSMA > slowSMA {
            handler.LogInfo("Bullish crossover detected")
            fastColor = "#00ff00" // Bright green
            slowColor = "#00ff00"
        } else if prevFastSMA >= prevSlowSMA && fastSMA < slowSMA {
            // Bearish crossover: fast crosses below slow
            handler.LogInfo("Bearish crossover detected")
            fastColor = "#ff0000" // Bright red
            slowColor = "#ff0000"
        } else {
            // No crossover
            fastColor = "#ffaa00" // Orange for fast
            slowColor = "#00aaff" // Blue for slow
        }
    } else {
        // First iteration, no previous values
        fastColor = "#ffaa00"
        slowColor = "#00aaff"
        handler.LogDebug("First calculation, no crossover detection available")
    }
    
    // Output both moving averages
    handler.SetOutput(indicators.NewPlotPointCollection([]indicators.PlotPoint{
        {
            Value: fastSMA,
            Color: fastColor,
        },
        {
            Value: slowSMA,
            Color: slowColor,
        },
    }))
    
    return handler, nil
}
```

### Using the Complete Example

Indicators are registered by adding them to a `ChartProject`:

```go
// Create your indicators
smaIndicator := sample.NewSMAIndicator()
rsiIndicator := sample.GetSampleRSIIndicator()

// Add them to your chart project
chartProject.AddIndicators(smaIndicator, rsiIndicator)
```

> **Note**: Chart project creation, source management, and project configuration are covered in separate documentation. This guide focuses solely on indicator development.

---

## Best Practices

### Data Validation
- Always check if `GetEntry()` returns `nil` before using
- Use `GetEntryCount()` to verify sufficient data before calculations
- Validate indicator configuration in HandleEntry
- Use `IsValid()` to check ChartEntry integrity

### Performance
- Avoid unnecessary loops over large datasets
- Cache calculations when possible
- Return early with invalid output when data is insufficient
- Use `config.ActivePair` unless cross-pair analysis is needed

### Logging
- Use Debug for calculation details
- Use Info for significant events (crossovers, signals)
- Use Warning for unexpected but recoverable conditions
- Use Error for data integrity issues
- Keep log messages concise

### Output
- Use descriptive colors that make sense for your data
- Return invalid output rather than zero values when data is insufficient
- For multi-line indicators, ensure consistent point ordering
- Match the number of output points to your LineTitles configuration

### Error Handling
- Return errors only for truly unrecoverable conditions
- Use invalid output + logging for recoverable issues
- Provide descriptive error messages
- Validate configuration early

---

## Advanced Topics

### Working with Multiple Pairs

```go
func (i *CrossPairIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Access data from multiple pairs
    btcData := chartsData.GetEntry(sources.SourcePairId{
        SourceName: "binance",
        PairId:     "BTCUSDT",
    }, 0)
    
    ethData := chartsData.GetEntry(sources.SourcePairId{
        SourceName: "binance",
        PairId:     "ETHUSDT",
    }, 0)
    
    if btcData == nil || ethData == nil {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    // Calculate correlation or ratio
    ratio := ethData.Close / btcData.Close
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: ratio,
        Color: "#ff00ff",
    })
    
    return handler, nil
}
```

### Stateful Indicators

Use the `points` parameter to maintain state across iterations:

```go
func (i *EMAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    current := chartsData.GetEntry(config.ActivePair, 0)
    alpha := 2.0 / float64(i.Period+1)
    
    var ema float64
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() {
            // Use previous EMA value
            prevEMA := prevCollection.Points[0].Value
            ema = alpha*current.Close + (1-alpha)*prevEMA
        } else {
            // First calculation, use SMA
            ema = current.Close // Simplified
        }
    } else {
        // First calculation, use SMA
        ema = current.Close // Simplified
    }
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: ema,
        Color: "#00ffff",
    })
    
    return handler, nil
}
```

---

## Troubleshooting

### Common Issues

**Indicator shows no data:**
- Check if you're returning invalid output when data is insufficient
- Verify `config.ActivePair` exists in `chartsData`
- Use logging to debug calculation steps

**Incorrect values:**
- Remember index 0 is most recent, not oldest
- Verify period calculations don't exceed available data
- Check for off-by-one errors in loops

**Colors not appearing:**
- Ensure color strings are valid CSS colors
- Check if color logic is executing
- Verify PlotPoint contains both Value and Color

**Performance issues:**
- Reduce logging in production
- Optimize loops over historical data
- Consider caching repeated calculations

---

## API Reference Summary

### Interfaces
- `BaseIndicator`: Main indicator interface

### Types
- `IndicatorType`: Visualization type enum
- `ChartConfig`: Chart configuration
- `IndicatorConfig`: Indicator rendering configuration
- `PlotPoint`: Single data point with styling
- `PlotPointCollection`: Collection of plot points
- `PlotPointCollectionHistory`: History container for previous indicator outputs
- `IndicatorOutputHandler`: Output and logging handler

### Key Functions
- `chartsData.GetEntry(pair, index)`: Get candle by index
- `chartsData.GetEntryCount(pair)`: Get available candle count
- `handler.SetSingleOutput(point)`: Set single value output
- `handler.SetOutput(collection)`: Set multi-value output
- `handler.LogDebug/Info/Warning/Error(msg)`: Log messages
- `NewPlotPointCollection(points)`: Create valid collection
- `NewInvalidPlotPointCollection()`: Create invalid collection
- `points.GetCollection(index)`: Get historical indicator output by index
- `points.Length()`: Get number of historical outputs available
- `points.GetAllCollections()`: Get all historical outputs as slice

---

## Additional Resources

- Review existing indicators in the codebase for patterns
- Test indicators with different timeframes and pairs
- Use logging extensively during development
- Consider edge cases: missing data, invalid values, first iteration

---

**Last Updated**: December 29, 2025
