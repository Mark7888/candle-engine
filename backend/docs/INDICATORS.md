# Custom Indicator Development Guide

> **IMPORTANT**: This indicator system is currently in active development and APIs may change in future versions. Please refer to this documentation for the latest information.

This guide provides comprehensive documentation for creating custom indicators in the charting system. It covers all interfaces, functions, configuration options, and best practices. It is written to be both human-readable and usable as an instruction set for an LLM to generate custom indicator implementations.

---

## Table of Contents

1. [Overview](#overview)
2. [Core Interface](#core-interface)
3. [Indicator Types](#indicator-types)
4. [Configuration](#configuration)
   - [ChartConfig (input)](#chartconfig)
   - [IndicatorConfig (output)](#indicatorconfig)
   - [LineIndicatorConfig](#lineindicatorconfig)
   - [StepLineIndicatorConfig](#steplineindicatorconfig)
   - [BarChartIndicatorConfig](#barchartindicatorconfig)
   - [TableIndicatorConfig](#tableindicatorconfig)
   - [AreaChartIndicatorConfig](#areachartindicatorconfig)
   - [PieChartIndicatorConfig](#piechartindicatorconfig)
   - [CandleStickIndicatorConfig](#candlestickindicatorconfig)
5. [Data Access](#data-access)
6. [Output Handling](#output-handling)
   - [PlotPoint](#plotpoint)
   - [PlotPointCollection](#plotpointcollection)
   - [Setting Output](#setting-output)
   - [Accessing Previous Values (PlotPointCollectionHistory)](#accessing-previous-values)
7. [Logging](#logging)
8. [Error Handling](#error-handling)
9. [Simple Examples](#simple-examples)
10. [Complete Example](#complete-example)
11. [Type-Specific Examples](#type-specific-examples)
    - [Line Chart](#example-line-chart)
    - [Step Line Chart](#example-step-line-chart)
    - [Bar Chart](#example-bar-chart)
    - [Area Chart](#example-area-chart)
    - [CandleStick Chart](#example-candlestick-chart)
    - [Table](#example-table)
    - [Pie Chart](#example-pie-chart)
12. [Advanced Topics](#advanced-topics)
13. [Best Practices](#best-practices)
14. [Troubleshooting](#troubleshooting)
15. [API Reference Summary](#api-reference-summary)
16. [Indicator Development Checklist](#indicator-development-checklist)
17. [Project Setup (Quick Reference)](#project-setup-quick-reference)

---

## Overview

Custom indicators are implemented by satisfying the `BaseIndicator` interface. Indicators process chart data iteratively from **oldest to newest** candles, allowing them to build state and calculate values progressively.

### Key Concepts

- **Iterative Processing**: `HandleEntry` is called once for each candle, from oldest to newest
- **Historical Access**: Access current and previous candle data using indexed lookups (index 0 = most recent)
- **Multi-Source Support**: Work with data from multiple trading pairs simultaneously
- **Flexible Output**: Support for single values, multiple plots, and seven visualization types
- **Previous Output Access**: Read your own previous indicator outputs via `PlotPointCollectionHistory`
- **Real-time Logging**: Debug and monitor indicator behavior with built-in logging (Debug, Info, Warning, Error)
- **Pair Reactivity**: Use `config.ActivePair` to make your indicator respond to the currently displayed pair

---

## Core Interface

### BaseIndicator

All custom indicators must implement this interface (defined in `pkg/indicators/indicator_base.go`):

```go
type BaseIndicator interface {
    GetName() string
    GetType() IndicatorType
    GetConfig() IndicatorConfig
    HandleEntry(
        points PlotPointCollectionHistory,
        chartsData sources.MultiSourceEntries,
        config ChartConfig,
        handler IndicatorOutputHandler,
    ) (IndicatorOutputHandler, error)
}
```

> **Note on Struct Implementation**: Your indicator structs can contain additional fields and methods beyond the interface requirements. However, **avoid using mutable state fields** as they can lead to repainting issues and false data across different chart contexts. Only use struct fields for immutable configuration (like period settings). State should be maintained through the `points` parameter instead.

#### Methods

**`GetName() string`**
- Returns the display name of your indicator
- This name is used in the UI and as a key for data storage
- Must be **unique** among all indicators in the same project
- Example: `"Simple Moving Average (SMA)"`

**`GetType() IndicatorType`**
- Returns the visualization type for your indicator
- See [Indicator Types](#indicator-types) for all seven available options

**`GetConfig() IndicatorConfig`**
- Returns configuration for rendering and display
- Only fill in the sub-config that matches your `GetType()` return value
- See [Configuration](#configuration) for all options

**`HandleEntry(...) (IndicatorOutputHandler, error)`**
- Core calculation logic, executed **once for each candle** from oldest to newest
- **Parameters**:
  - `points PlotPointCollectionHistory` — History of your indicator's own previous outputs. Use `points.GetCollection(0)` for the most recent previous output, `points.GetCollection(1)` for the one before that, etc.
  - `chartsData sources.MultiSourceEntries` — All available OHLCV chart data from loaded sources. Use `chartsData.GetEntry(pair, index)` to access candles (index 0 = most recent).
  - `config ChartConfig` — Chart configuration providing the active pair, loaded pairs, and timeframe.
  - `handler IndicatorOutputHandler` — Output handler; call `handler.SetOutput()` or `handler.SetSingleOutput()` to set your result, and `handler.LogDebug/Info/Warning/Error()` to log.
- **Returns**:
  - `IndicatorOutputHandler` — The updated handler with your calculated values and any logs
  - `error` — Return a non-nil error **only** for fatal failures (this stops indicator processing entirely). For recoverable issues, use logging + invalid output instead.

---

## Indicator Types

Indicators can be rendered in seven different visualization styles. Choose the appropriate type for your data.

### IndicatorType Enum

Defined in `pkg/indicators/indicator_config.go`:

```go
type IndicatorType int

const (
    IndicatorTypeLine        IndicatorType = iota // 0
    IndicatorTypeStepLine                         // 1
    IndicatorTypeBarChart                         // 2
    IndicatorTypeTable                            // 3
    IndicatorTypeAreaChart                        // 4
    IndicatorTypePieChart                         // 5
    IndicatorTypeCandleStick                      // 6
)
```

### Type Descriptions

| Type | Value | Description | Use Case | Points per Candle |
|------|-------|-------------|----------|-------------------|
| `IndicatorTypeLine` | 0 | Smooth or straight continuous line plot | Moving averages, oscillators, price-based indicators | 1+ (one per line) |
| `IndicatorTypeStepLine` | 1 | Step-wise line with configurable alignment | Discrete signals, levels, binary indicators | 1+ (one per line) |
| `IndicatorTypeBarChart` | 2 | Vertical bar chart | Volume, comparative analysis, multi-value comparisons | 1+ (one per bar series) |
| `IndicatorTypeTable` | 3 | Tabular data display (not time-series) | Statistical summaries, OHLC snapshots, multi-metric displays | rows × columns |
| `IndicatorTypeAreaChart` | 4 | Filled area chart with line border | Cumulative data, range displays, filled oscillators | 1+ (one per area) |
| `IndicatorTypePieChart` | 5 | Pie/donut chart (not time-series) | Portfolio allocation, percentage breakdowns, proportional data | 1 per slice |
| `IndicatorTypeCandleStick` | 6 | Custom candlestick sub-chart | Alternative price representations, synthetic instruments | Exactly 4: [Open, High, Low, Close] |

---

## Configuration

### ChartConfig

Contains information about the chart the indicator is running on. This is **provided to your indicator** in `HandleEntry`. Defined in `pkg/indicators/indicator_config.go`:

```go
type ChartConfig struct {
    TimeFrameSeconds int64                  // Candle timeframe in seconds (e.g. 86400 for 1d)
    ActivePair       sources.SourcePairId   // The currently active/displayed trading pair
    LoadedPairs      []sources.SourcePairId // All trading pairs loaded in the project
}
```

**Usage**:
- Use `config.ActivePair` to make your indicator **reactive** to the current chart pair
- Use `config.LoadedPairs` to work with multiple pairs for cross-pair analysis
- Use `config.TimeFrameSeconds` for time-based calculations (e.g. converting periods to real time)

### IndicatorConfig

Defines how your indicator should be rendered. This is **returned by your indicator** in `GetConfig()`. Defined in `pkg/indicators/indicator_config.go`:

```go
type IndicatorConfig struct {
    // Empty string for main chart overlay, any other string for a custom sub-chart panel
    TargetChart string

    LineIndicator        LineIndicatorConfig
    StepLineIndicator    StepLineIndicatorConfig
    BarChartIndicator    BarChartIndicatorConfig
    CandleStickIndicator CandleStickIndicatorConfig
    TableIndicator       TableIndicatorConfig
    AreaChartIndicator   AreaChartIndicatorConfig
    PieChartIndicator    PieChartIndicatorConfig
}
```

> **Rule**: Only populate the sub-config field that corresponds to your `GetType()` return value. For example, if `GetType()` returns `IndicatorTypeLine`, only fill in `LineIndicator`. The other fields are ignored.

#### TargetChart

- **Type**: `string`
- **Purpose**: Groups indicators into separate chart panels
- **Values**:
  - `""` (empty string): Display on the **main** candlestick chart (overlay)
  - `"volume"`, `"rsi"`, `"my-panel"`, etc.: Create or use a **separate sub-chart panel** with this ID
- **Behavior**: Multiple indicators that share the same `TargetChart` string appear in the same panel
- **Example**: Oscillators (RSI, MACD) typically use their own panel; overlays (SMA, EMA) use `""`

---

### LineIndicatorConfig

For `IndicatorTypeLine`. Renders one or more continuous lines on a chart.

```go
type LineIndicatorConfig struct {
    LineThickness int      `json:"lineThickness"` // Width of the line in pixels
    IsSmooth      bool     `json:"isSmooth"`      // Whether the line uses smooth curve interpolation
    LineTitles    []string `json:"lineTitles"`    // Names for each line (for multi-line indicators)
    ShowSymbols   bool     `json:"showSymbols"`   // Show circle markers at data points
    Opacity       float64  `json:"opacity"`       // Line opacity (0.0 to 1.0)
}
```

- `LineTitles` length defines how many lines your indicator produces — your `PlotPointCollection` must output the same number of `PlotPoint` values each iteration.
- `IsSmooth: true` produces bezier-curved lines; `false` produces straight segments.

**Example**:
```go
LineIndicator: indicators.LineIndicatorConfig{
    LineThickness: 2,
    IsSmooth:      true,
    LineTitles:    []string{"Fast MA", "Slow MA"},
    ShowSymbols:   false,
    Opacity:       0.8,
}
```

---

### StepLineIndicatorConfig

For `IndicatorTypeStepLine`. Renders step-wise lines with configurable alignment.

```go
type StepLineIndicatorConfig struct {
    StepAlignment StepAlignment `json:"stepAlignment"` // How steps align with candles
    LineTitles    []string      `json:"lineTitles"`    // Names for each line
    LineThickness int           `json:"lineThickness"` // Width of the line in pixels
    Opacity       float64       `json:"opacity"`       // Line opacity (0.0 to 1.0)
}
```

**StepAlignment Values**:

```go
type StepAlignment int

const (
    StepAlignStart    StepAlignment = iota // 0 — Step occurs at candle start
    StepAlignMidpoint                      // 1 — Step occurs at candle midpoint
    StepAlignEnd                           // 2 — Step occurs at candle end
)
```

**Example**:
```go
StepLineIndicator: indicators.StepLineIndicatorConfig{
    StepAlignment: indicators.StepAlignMidpoint,
    LineTitles:    []string{"Close"},
    LineThickness: 1,
    Opacity:       1.0,
}
```

---

### BarChartIndicatorConfig

For `IndicatorTypeBarChart`. Renders vertical bars, optionally stacked.

```go
type BarChartIndicatorConfig struct {
    BarTitles []string `json:"barTitles"` // Names for each bar series
    BarWidth  string   `json:"barWidth"`  // Bar width (e.g., "60%", "20", "auto")
    Opacity   float64  `json:"opacity"`   // Bar opacity (0.0 to 1.0)
    Stacked   bool     `json:"stacked"`   // Whether bars should be stacked
}
```

- `BarTitles` length defines the number of bar series — match it with the number of `PlotPoint` values you output each iteration.
- `BarWidth` accepts CSS-style values: `"60%"` (of available space), `"20"` (pixels), or `"auto"`.

**Example**:
```go
BarChartIndicator: indicators.BarChartIndicatorConfig{
    BarTitles: []string{"Volume"},
    BarWidth:  "60%",
    Opacity:   0.8,
    Stacked:   false,
}
```

---

### TableIndicatorConfig

For `IndicatorTypeTable`. Renders a static table (not time-series). Shows the **last** (most recent) `PlotPointCollection` as a table.

```go
type TableIndicatorConfig struct {
    DisplayName        string   `json:"displayName"`        // Title of the table
    RowTitles          []string `json:"rowTitles"`          // Row labels (also defines the number of rows)
    ColumnTitles       []string `json:"columnTitles"`       // Column labels (also defines the number of columns)
    ValueDisplayFormat string   `json:"valueDisplayFormat"` // Printf-style format string (e.g., "%.2f")
}
```

- The total number of `PlotPoint` values per collection must equal `len(RowTitles) × len(ColumnTitles)`.
- Points are laid out **row-major**: the first `len(ColumnTitles)` points fill row 0, the next fill row 1, and so on.

**Example**:
```go
TableIndicator: indicators.TableIndicatorConfig{
    DisplayName:        "OHLC Snapshot",
    RowTitles:          []string{"Current", "Previous"},
    ColumnTitles:       []string{"Open", "Close"},
    ValueDisplayFormat: "%.2f",
}
// PlotPoints: [CurrentOpen, CurrentClose, PreviousOpen, PreviousClose]
```

---

### AreaChartIndicatorConfig

For `IndicatorTypeAreaChart`. Renders filled area charts with a line border.

```go
type AreaChartIndicatorConfig struct {
    LineTitles    []string `json:"lineTitles"`    // Names for each area series
    LineThickness int      `json:"lineThickness"` // Width of the line border in pixels
    IsSmooth      bool     `json:"isSmooth"`      // Enable smooth curve interpolation
    ShowSymbols   bool     `json:"showSymbols"`   // Show circle markers at data points
    Opacity       float64  `json:"opacity"`       // Area fill opacity (0.0 to 1.0)
}
```

- Functions identically to `LineIndicatorConfig` but with a filled area below the line.
- `Opacity` controls the area fill transparency; the line border is always fully opaque.

**Example**:
```go
AreaChartIndicator: indicators.AreaChartIndicatorConfig{
    LineTitles:    []string{"RSI"},
    LineThickness: 1,
    IsSmooth:      true,
    ShowSymbols:   false,
    Opacity:       0.4,
}
```

---

### PieChartIndicatorConfig

For `IndicatorTypePieChart`. Renders a pie chart (not time-series). Shows the **last** (most recent) `PlotPointCollection` as proportional slices.

```go
type PieChartIndicatorConfig struct {
    DisplayName string   `json:"displayName"` // Title of the pie chart
    SliceTitles []string `json:"sliceTitles"` // Label for each slice; length defines number of slices
}
```

- The number of `PlotPoint` values per collection must equal `len(SliceTitles)`.
- Each `PlotPoint.Value` represents the **magnitude** of that slice (values are automatically normalized to percentages).
- Each `PlotPoint.Color` sets the color for that slice.
- If all slice values are zero or negative, return an invalid collection.

**Example**:
```go
PieChartIndicator: indicators.PieChartIndicatorConfig{
    DisplayName: "Market Snapshot",
    SliceTitles: []string{"Open", "Hi-Close Range", "Volume (k)"},
}
// PlotPoints: [OpenValue, RangeValue, VolumeValue] with colors
```

---

### CandleStickIndicatorConfig

For `IndicatorTypeCandleStick`. Renders a custom candlestick sub-chart — separate from the main price candles.

```go
type CandleStickIndicatorConfig struct {
    // Custom sub-chart candlestick (not the main price candles — a separate indicator sub-chart).
    // PlotPoints per timestamp: exactly 4 points in order [Open, High, Low, Close].
    // Falls back to project CandleColors if these are empty strings.
    BullishColor string `json:"bullishColor"` // Hex color for bullish candles (e.g. "#26a69a")
    BearishColor string `json:"bearishColor"` // Hex color for bearish candles (e.g. "#ef5350")
}
```

- Each `PlotPointCollection` must contain **exactly 4** `PlotPoint` values in the order: **Open, High, Low, Close**.
- The `Color` field on individual `PlotPoint` values is ignored; `BullishColor` and `BearishColor` from the config are used instead.
- If `BullishColor` or `BearishColor` are empty strings, the project's default candle colors are used as fallback.

**Example**:
```go
CandleStickIndicator: indicators.CandleStickIndicatorConfig{
    BullishColor: "#26a69a",
    BearishColor: "#ef5350",
}
// PlotPoints: [Open, High, Low, Close] — 4 points per candle
```

---

## Data Access

### MultiSourceEntries

The `chartsData` parameter provides access to all chart data from loaded sources. Defined in `pkg/sources/multisource.go`.

#### Key Methods

**`GetEntry(pairId SourcePairId, index int) *ChartEntry`**
- Retrieves a single candle by pair and index
- **Index 0** = Most recent (current) candle being processed
- **Index 1** = Previous candle
- **Index n** = Candle n positions back in time
- Returns `nil` if pair doesn't exist or index is out of bounds

**`GetEntryCount(pairId SourcePairId) int`**
- Returns the number of available candles for a pair at the current iteration
- Returns `0` if pair doesn't exist
- Useful for validating sufficient data before calculations

**`GetEntries(pairId SourcePairId) []ChartEntry`**
- Returns all available candles for a pair as a slice
- Returns `nil` if pair doesn't exist

### ChartEntry Structure

Defined in `pkg/sources/models.go`:

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
- `IsValid() bool`: Returns `false` if entry contains invalid data (-999 values). Always check this before using entry values.

### SourcePairId

Identifies a trading pair from a specific data source:

```go
type SourcePairId struct {
    SourceName string // e.g. "Binance", "Coinbase"
    PairId     string // e.g. "BTCUSDC", "ETH-USDT"
}
```

### Making Indicators Pair-Specific

To make an indicator react to the currently displayed chart, use `config.ActivePair`:

```go
func (i *MyIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Use config.ActivePair to get data for the current chart
    current := chartsData.GetEntry(config.ActivePair, 0)
    
    if current == nil || !current.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    // Your calculation using current pair's data
    // ...
    
    return handler, nil
}
```

Without using `config.ActivePair`, the indicator will show the same values across all pairs.

---

## Output Handling

### PlotPoint

Represents a single data point with optional styling. Defined in `pkg/indicators/indicator_models.go`:

```go
type PlotPoint struct {
    Value   float64 // The numeric value to plot
    Color   string  // CSS color string (e.g., "#00ff00", "red", "#ff000080")
    Invalid bool    // If true, this point won't be displayed (creates a gap)
}
```

**Important**: If a `PlotPoint` has `Invalid` set to `true`, that specific point will not be displayed in the chart. This allows for gaps in your indicator line at specific positions where data is insufficient or invalid.

**Helper**:
```go
// Creates an invalid point (Value: -999, Color: "red", Invalid: true)
invalidPoint := indicators.NewInvalidPlotPoint()
```

### PlotPointCollection

Container for one or more plot points for a single candle/iteration:

```go
type PlotPointCollection struct {
    Points  []PlotPoint
    Invalid bool // If true, no data will be rendered for this iteration
}
```

**Important Behavior**:
- If the **entire collection** has `Invalid` set to `true`, **none of the points** from that collection will be displayed for that candle/iteration
- If **individual points** within a collection have their `Invalid` field set to `true`, only those specific points won't be displayed, while other valid points in the same collection will still render

**Creating Collections**:
```go
// Single point
collection := indicators.NewPlotPointCollection(
    indicators.PlotPoint{Value: 50.0, Color: "#00ff00"},
)

// Multiple points (variadic arguments)
collection := indicators.NewPlotPointCollection(
    indicators.PlotPoint{Value: 20.0, Color: "#ff0000"},
    indicators.PlotPoint{Value: 80.0, Color: "#00ff00"},
)

// Invalid/no data — entire collection is invalid
collection := indicators.NewInvalidPlotPointCollection()

// Create a single invalid point (for use within a collection)
invalidPoint := indicators.NewInvalidPlotPoint()

// Mixed valid and invalid points in one collection
collection := indicators.NewPlotPointCollection(
    indicators.PlotPoint{Value: 20.0, Color: "#ff0000"}, // Valid — will be displayed
    indicators.NewInvalidPlotPoint(),                     // Invalid — creates a gap
    indicators.PlotPoint{Value: 80.0, Color: "#00ff00"}, // Valid — will be displayed
)
```

**Collection Methods**:
- `AddPoints(points ...PlotPoint)` — Append additional points to the collection
- `IsInvalid() bool` — Check if the entire collection is invalid
- `Length() int` — Get the number of points in the collection

### Setting Output

Use the `handler` parameter to set your indicator's output for the current candle.

**Single Value**:
```go
handler.SetSingleOutput(indicators.PlotPoint{
    Value: calculatedValue,
    Color: "#00ffff",
})
```

**No Output** (insufficient data — mark the entire collection as invalid):
```go
handler.SetOutput(indicators.NewInvalidPlotPointCollection())
```

**Partial Output** (some valid, some invalid points in a multi-line indicator):
```go
handler.SetOutput(indicators.NewPlotPointCollection(
    indicators.PlotPoint{Value: fastMA, Color: "#00ff00"}, // Valid point — displays
    indicators.NewInvalidPlotPoint(),                       // Invalid point — creates gap
    indicators.PlotPoint{Value: slowMA, Color: "#ff0000"}, // Valid point — displays
))
```

**Multiple Values** (e.g. multi-line indicator):
```go
handler.SetOutput(indicators.NewPlotPointCollection(
    indicators.PlotPoint{Value: fastMA, Color: "#00ff00"},
    indicators.PlotPoint{Value: slowMA, Color: "#ff0000"},
))
```

### Accessing Previous Values

The `points` parameter is a `PlotPointCollectionHistory` that provides access to your indicator's own previous outputs with convenient helper methods.

#### PlotPointCollectionHistory

```go
type PlotPointCollectionHistory struct {
    Collections []PlotPointCollection
}
```

**Methods**:
- `GetCollection(index int) *PlotPointCollection` — Get collection at reverse index (0 = most recent, 1 = previous, etc.). Returns `nil` if out of bounds.
- `Length() int` — Returns the number of historical collections available
- `GetAllCollections() []PlotPointCollection` — Returns all collections as a slice (in chronological order)
- `AddCollection(collection PlotPointCollection)` — Append a collection (used internally by the engine)

**Index Convention**: Like chart data, index 0 is the most recent (previous) output, index 1 is the one before that, and so on.

**Example Usage**:

```go
func (i *MyIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    // Check if we have previous indicator output
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() {
            previousValue := prevCollection.Points[0].Value
            // Use previousValue in calculation (e.g. for EMA, KAMA)
        }
    }
    
    // Access multiple historical values
    if points.Length() >= 3 {
        recent   := points.GetCollection(0) // Most recent previous output
        oneBack  := points.GetCollection(1) // Two iterations ago
        twoBack  := points.GetCollection(2) // Three iterations ago
        // Use for trend analysis across your own outputs
    }
    
    // ...
    return handler, nil
}
```

---

## Logging

Indicators can log messages that appear in the UI for debugging and monitoring. Logs are attached to the current candle's timestamp automatically.

### Log Levels

| Function | Level | Use Case |
|----------|-------|----------|
| `handler.LogDebug(message)` | `DEBUG` | Detailed diagnostic information |
| `handler.LogInfo(message)` | `INFO` | General informational messages, significant events |
| `handler.LogWarning(message)` | `WARNING` | Potential issues or unusual conditions |
| `handler.LogError(message)` | `ERROR` | Non-fatal errors (calculation continues) |

### IndicatorLog Structure

```go
type IndicatorLog struct {
    Level   IndicatorLogLevel // "DEBUG", "INFO", "WARNING", "ERROR"
    Time    time.Time         // Automatically set to current candle's timestamp
    Message string
}
```

### Example

```go
func (i *MyIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
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

Return a non-nil error to stop indicator processing completely. Use this **only** for truly unrecoverable conditions:

```go
if critical_condition {
    return handler, fmt.Errorf("critical failure: reason")
}
```

When a fatal error is returned, the engine marks the current output as invalid and records the error for display.

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

A simple line indicator that displays the closing price of the current candle on the main chart.

```go
package sample

import (
    "candle-engine/pkg/sources"
    "candle-engine/pkg/indicators"
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
    current := chartsData.GetEntry(config.ActivePair, 0)
    
    if current == nil || !current.IsValid() {
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

Shows the difference between current and previous close prices on a separate panel.

```go
package sample

import (
    "candle-engine/pkg/sources"
    "candle-engine/pkg/indicators"
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

### Example 3: High/Low Range (Bar Chart)

Displays the difference between high and low for each candle as a bar chart.

```go
package sample

import (
    "candle-engine/pkg/sources"
    "candle-engine/pkg/indicators"
)

type RangeIndicator struct{}

func (r *RangeIndicator) GetName() string {
    return "Candle Range"
}

func (r *RangeIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeBarChart
}

func (r *RangeIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "range",
        BarChartIndicator: indicators.BarChartIndicatorConfig{
            BarTitles: []string{"Range"},
            BarWidth:  "60%",
            Opacity:   0.8,
        },
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
- Multi-value output (two lines)
- Conditional coloring based on crossover detection
- State tracking using previous indicator values via `PlotPointCollectionHistory`
- Logging at multiple levels
- Error handling

```go
package sample

import (
    "fmt"
    "candle-engine/pkg/sources"
    "candle-engine/pkg/indicators"
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
    // Validate input parameters (fatal error for broken config)
    if s.FastPeriod <= 0 || s.SlowPeriod <= 0 {
        return handler, fmt.Errorf("invalid period configuration: fast=%d, slow=%d", s.FastPeriod, s.SlowPeriod)
    }
    
    if s.FastPeriod >= s.SlowPeriod {
        handler.LogWarning("Fast period should be less than slow period")
    }
    
    // Check if we have enough data
    availableData := chartsData.GetEntryCount(config.ActivePair)
    if availableData < s.SlowPeriod {
        handler.LogDebug(fmt.Sprintf("Insufficient data: need %d, have %d", s.SlowPeriod, availableData))
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
    fastColor := "#ffaa00" // Default orange for fast
    slowColor := "#00aaff" // Default blue for slow
    
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() && len(prevCollection.Points) >= 2 {
            prevFastSMA := prevCollection.Points[0].Value
            prevSlowSMA := prevCollection.Points[1].Value
            
            if prevFastSMA <= prevSlowSMA && fastSMA > slowSMA {
                handler.LogInfo("Bullish crossover detected")
                fastColor = "#00ff00"
                slowColor = "#00ff00"
            } else if prevFastSMA >= prevSlowSMA && fastSMA < slowSMA {
                handler.LogInfo("Bearish crossover detected")
                fastColor = "#ff0000"
                slowColor = "#ff0000"
            }
        } else {
            handler.LogDebug("First calculation, no crossover detection available")
        }
    }
    
    // Output both moving averages
    handler.SetOutput(indicators.NewPlotPointCollection(
        indicators.PlotPoint{Value: fastSMA, Color: fastColor},
        indicators.PlotPoint{Value: slowSMA, Color: slowColor},
    ))
    
    return handler, nil
}
```

### Registering Indicators

Indicators are registered by adding them to a `ChartProject`:

```go
// Create your indicators
smaIndicator := NewSMAIndicator()
rsiIndicator := NewRSIIndicator()

// Add them to your chart project (variadic — add as many as needed)
chartProject.AddIndicators(smaIndicator, rsiIndicator)
```

> **Note**: For a minimal example of setting up a full project with sources, pairs, and indicators, see [Project Setup (Quick Reference)](#project-setup-quick-reference) at the end of this document.

---

## Type-Specific Examples

These examples demonstrate how to correctly implement each of the seven indicator types.

---

### Example: Line Chart

A multi-line moving average indicator (SMA-5, SMA-10, SMA-20) on the main chart.

```go
type MAIndicator struct{}

func (r *MAIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeLine
}

func (r *MAIndicator) GetName() string {
    return "Moving Average (MA)"
}

func (r *MAIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "",
        LineIndicator: indicators.LineIndicatorConfig{
            LineThickness: 1,
            LineTitles:    []string{"MA5", "MA10", "MA20"},
        },
    }
}

func (r *MAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    availableData := chartsData.GetEntryCount(config.ActivePair)

    calcMA := func(period int) indicators.PlotPoint {
        if availableData < period {
            return indicators.NewInvalidPlotPoint()
        }
        var sum float64
        for i := 0; i < period; i++ {
            entry := chartsData.GetEntry(config.ActivePair, i)
            if entry == nil || !entry.IsValid() {
                return indicators.NewInvalidPlotPoint()
            }
            sum += entry.Close
        }
        return indicators.PlotPoint{Value: sum / float64(period), Color: "#00ff00"}
    }

    handler.SetOutput(indicators.NewPlotPointCollection(
        calcMA(5),  // MA5 — matches LineTitles[0]
        calcMA(10), // MA10 — matches LineTitles[1]
        calcMA(20), // MA20 — matches LineTitles[2]
    ))
    return handler, nil
}
```

**Key pattern**: The number of `PlotPoint` values (3) matches `len(LineTitles)` (3).

---

### Example: Step Line Chart

A step line showing the closing price.

```go
type StepPriceIndicator struct{}

func (d *StepPriceIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeStepLine
}

func (d *StepPriceIndicator) GetName() string {
    return "Price Step"
}

func (d *StepPriceIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "step-panel",
        StepLineIndicator: indicators.StepLineIndicatorConfig{
            StepAlignment: indicators.StepAlignMidpoint,
            LineTitles:    []string{"Close"},
            LineThickness: 1,
        },
    }
}

func (d *StepPriceIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    entry := chartsData.GetEntry(config.ActivePair, 0)
    if entry == nil || !entry.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    handler.SetSingleOutput(indicators.PlotPoint{
        Value: entry.Close,
        Color: "#00e5ff",
    })
    return handler, nil
}
```

---

### Example: Bar Chart

A volume bar chart with logging for high-volume candles.

```go
type VolumeIndicator struct{}

func (d *VolumeIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeBarChart
}

func (d *VolumeIndicator) GetName() string {
    return "Volume"
}

func (d *VolumeIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "volume",
        BarChartIndicator: indicators.BarChartIndicatorConfig{
            BarTitles: []string{"Volume"},
            BarWidth:  "60%",
            Opacity:   0.8,
        },
    }
}

func (d *VolumeIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    entry := chartsData.GetEntry(config.ActivePair, 0)
    if entry == nil || !entry.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    vol := entry.Volume
    if vol >= 50000 {
        handler.LogInfo(fmt.Sprintf("High volume: %.2f", vol))
    }

    handler.SetSingleOutput(indicators.PlotPoint{
        Value: vol,
        Color: "#ab47bc",
    })
    return handler, nil
}
```

---

### Example: Area Chart

An RSI indicator rendered as a filled area chart with conditional coloring.

```go
type AreaRSIIndicator struct{}

func (d *AreaRSIIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeAreaChart
}

func (d *AreaRSIIndicator) GetName() string {
    return "Area RSI"
}

func (d *AreaRSIIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "rsi-area",
        AreaChartIndicator: indicators.AreaChartIndicatorConfig{
            LineTitles:    []string{"RSI"},
            LineThickness: 1,
            IsSmooth:      true,
            Opacity:       0.4,
        },
    }
}

func (d *AreaRSIIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    period := 14
    if chartsData.GetEntryCount(config.ActivePair) < period+1 {
        handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
        return handler, nil
    }

    var gains, losses float64
    for i := 0; i < period; i++ {
        current := chartsData.GetEntry(config.ActivePair, i)
        previous := chartsData.GetEntry(config.ActivePair, i+1)
        change := current.Close - previous.Close
        if change > 0 {
            gains += change
        } else {
            losses += -change
        }
    }

    avgGain := gains / float64(period)
    avgLoss := losses / float64(period)

    var rsi float64
    if avgLoss == 0 {
        rsi = 100
    } else {
        rs := avgGain / avgLoss
        rsi = 100 - (100 / (1 + rs))
    }

    // Conditional coloring: overbought (>70) red, oversold (<30) green, neutral blue
    color := "#29b6f6"
    if rsi > 70 {
        color = "#ff7043"
    } else if rsi < 30 {
        color = "#66bb6a"
    }

    handler.SetSingleOutput(indicators.PlotPoint{Value: rsi, Color: color})
    return handler, nil
}
```

---

### Example: CandleStick Chart

A custom candlestick indicator that mirrors the main price candles on a sub-chart. Each iteration must output **exactly 4 points**: Open, High, Low, Close.

```go
type CustomCandlesIndicator struct{}

func (d *CustomCandlesIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeCandleStick
}

func (d *CustomCandlesIndicator) GetName() string {
    return "Custom Candles"
}

func (d *CustomCandlesIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "custom-candles",
        CandleStickIndicator: indicators.CandleStickIndicatorConfig{
            BullishColor: "#26a69a",
            BearishColor: "#ef5350",
        },
    }
}

func (d *CustomCandlesIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    entry := chartsData.GetEntry(config.ActivePair, 0)
    if entry == nil || !entry.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    // Exactly 4 points in order: Open, High, Low, Close
    handler.SetOutput(indicators.NewPlotPointCollection(
        indicators.PlotPoint{Value: entry.Open, Color: ""},
        indicators.PlotPoint{Value: entry.High, Color: ""},
        indicators.PlotPoint{Value: entry.Low, Color: ""},
        indicators.PlotPoint{Value: entry.Close, Color: ""},
    ))
    return handler, nil
}
```

**Key pattern**: Always 4 points per collection in [Open, High, Low, Close] order. Color on individual points is ignored — use `BullishColor`/`BearishColor` in the config instead.

---

### Example: Table

An OHLC snapshot table with 2 rows × 2 columns.

```go
type OHLCSnapshotIndicator struct{}

func (d *OHLCSnapshotIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypeTable
}

func (d *OHLCSnapshotIndicator) GetName() string {
    return "OHLC Snapshot"
}

func (d *OHLCSnapshotIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "",
        TableIndicator: indicators.TableIndicatorConfig{
            DisplayName:        "OHLC Snapshot",
            RowTitles:          []string{"Current", "Previous"},
            ColumnTitles:       []string{"Open", "Close"},
            ValueDisplayFormat: "%.2f",
        },
    }
}

func (d *OHLCSnapshotIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    if chartsData.GetEntryCount(config.ActivePair) < 2 {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    current := chartsData.GetEntry(config.ActivePair, 0)
    previous := chartsData.GetEntry(config.ActivePair, 1)

    closeColor := func(e *sources.ChartEntry) string {
        if e.Close >= e.Open {
            return "#26a69a"
        }
        return "#ef5350"
    }

    // Points laid out row-major: Row0Col0, Row0Col1, Row1Col0, Row1Col1
    handler.SetOutput(indicators.NewPlotPointCollection(
        indicators.PlotPoint{Value: current.Open, Color: "#ffffff"},       // Current Open
        indicators.PlotPoint{Value: current.Close, Color: closeColor(current)},  // Current Close
        indicators.PlotPoint{Value: previous.Open, Color: "#ffffff"},      // Previous Open
        indicators.PlotPoint{Value: previous.Close, Color: closeColor(previous)}, // Previous Close
    ))
    return handler, nil
}
```

**Key pattern**: Total points = `len(RowTitles) × len(ColumnTitles)` = 2 × 2 = 4, laid out in row-major order.

---

### Example: Pie Chart

A market snapshot pie chart with 3 slices.

```go
type MarketSnapshotIndicator struct{}

func (d *MarketSnapshotIndicator) GetType() indicators.IndicatorType {
    return indicators.IndicatorTypePieChart
}

func (d *MarketSnapshotIndicator) GetName() string {
    return "Market Snapshot"
}

func (d *MarketSnapshotIndicator) GetConfig() indicators.IndicatorConfig {
    return indicators.IndicatorConfig{
        TargetChart: "",
        PieChartIndicator: indicators.PieChartIndicatorConfig{
            DisplayName: "Market Snapshot",
            SliceTitles: []string{"Open", "Hi-Close Range", "Volume (k)"},
        },
    }
}

func (d *MarketSnapshotIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    entry := chartsData.GetEntry(config.ActivePair, 0)
    if entry == nil || !entry.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    v0 := math.Max(0, entry.Open)
    v1 := math.Max(0, math.Abs(entry.High-entry.Close))
    v2 := math.Max(0, entry.Volume/1000)

    if v0 == 0 && v1 == 0 && v2 == 0 {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }

    // One point per slice, matching SliceTitles order
    handler.SetOutput(indicators.NewPlotPointCollection(
        indicators.PlotPoint{Value: v0, Color: "#42a5f5"},
        indicators.PlotPoint{Value: v1, Color: "#66bb6a"},
        indicators.PlotPoint{Value: v2, Color: "#ffa726"},
    ))
    return handler, nil
}
```

**Key pattern**: Number of points = `len(SliceTitles)`. Values represent magnitudes (auto-normalized to percentages).

---

## Advanced Topics

### Working with Multiple Pairs

Access data from multiple trading pairs simultaneously for cross-pair analysis:

```go
func (i *CrossPairIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    btcData := chartsData.GetEntry(sources.SourcePairId{
        SourceName: "Binance",
        PairId:     "BTCUSDC",
    }, 0)
    
    ethData := chartsData.GetEntry(sources.SourcePairId{
        SourceName: "Binance",
        PairId:     "ETHUSDC",
    }, 0)
    
    if btcData == nil || ethData == nil {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    ratio := ethData.Close / btcData.Close
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: ratio,
        Color: "#ff00ff",
    })
    
    return handler, nil
}
```

> **Tip**: Use `config.LoadedPairs` to discover available pairs dynamically instead of hardcoding them.

### Stateful Indicators (Using Previous Outputs)

Use the `points` parameter to maintain state across iterations. This is the **recommended** way to carry state — avoid mutable struct fields.

```go
func (i *EMAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
    current := chartsData.GetEntry(config.ActivePair, 0)
    if current == nil || !current.IsValid() {
        handler.SetOutput(indicators.NewInvalidPlotPointCollection())
        return handler, nil
    }
    
    alpha := 2.0 / float64(i.Period+1)
    
    var ema float64
    if points.Length() > 0 {
        prevCollection := points.GetCollection(0)
        if prevCollection != nil && !prevCollection.IsInvalid() {
            prevEMA := prevCollection.Points[0].Value
            ema = alpha*current.Close + (1-alpha)*prevEMA
        } else {
            ema = current.Close // First valid calculation
        }
    } else {
        ema = current.Close // Seed value
    }
    
    handler.SetSingleOutput(indicators.PlotPoint{
        Value: ema,
        Color: "#00ffff",
    })
    
    return handler, nil
}
```

---

## Best Practices

### Data Validation
- Always check if `GetEntry()` returns `nil` before using the result
- Use `GetEntryCount()` to verify sufficient data before starting calculations
- Validate indicator configuration parameters early (in `HandleEntry`)
- Use `IsValid()` to check `ChartEntry` integrity before accessing OHLCV fields

### State Management
- **Prefer** using `points` (PlotPointCollectionHistory) to access previous indicator outputs
- **Avoid** mutable struct fields for state — they can cause repainting issues across different chart contexts
- Struct fields are fine for **immutable** configuration (like period settings initialized at creation)

### Performance
- Avoid unnecessary loops over large datasets
- Return early with invalid output when data is insufficient
- Use `config.ActivePair` for pair-specific data unless cross-pair analysis is needed
- Reduce logging verbosity in production indicators

### Logging
- Use `LogDebug` for detailed calculation diagnostics
- Use `LogInfo` for significant events (crossovers, signals, threshold breaches)
- Use `LogWarning` for unexpected but recoverable conditions
- Use `LogError` for data integrity issues that don't warrant a fatal error
- Keep log messages concise and informative

### Output Consistency
- Use descriptive CSS colors that make sense for your data (hex preferred, e.g. `"#00ff00"`)
- Return invalid output (`NewInvalidPlotPointCollection()`) rather than zero values when data is insufficient
- For multi-line/multi-series indicators, ensure consistent point ordering across iterations
- **Match** the number of output points to your config titles:
  - `LineTitles` count = number of `PlotPoint` values for Line/StepLine/AreaChart
  - `BarTitles` count = number of `PlotPoint` values for BarChart
  - `SliceTitles` count = number of `PlotPoint` values for PieChart
  - `RowTitles × ColumnTitles` = total `PlotPoint` values for Table
  - CandleStick always requires exactly **4** points: [Open, High, Low, Close]

### Error Handling
- Return fatal errors (`return handler, fmt.Errorf(...)`) only for truly unrecoverable conditions
- Use invalid output + logging for all recoverable issues
- Provide descriptive error messages that help diagnose the problem

---

## Troubleshooting

### Common Issues

**Indicator shows no data:**
- Check if you're returning `NewInvalidPlotPointCollection()` when data is insufficient
- Verify `config.ActivePair` exists in `chartsData` (use `GetEntryCount()` to check)
- Use `handler.LogDebug()` to trace your calculation steps
- Ensure your `GetName()` returns a unique name among all project indicators

**Data consistency errors (DataError flag):**
- The engine checks that all `PlotPointCollection` outputs have the same number of points
- If some iterations return 2 points and others return 3, a `DataError` is flagged
- Ensure your point count is consistent across all iterations (use `NewInvalidPlotPoint()` as placeholder for missing individual values)

**Incorrect values:**
- Remember index 0 is most recent, not oldest
- Verify period calculations don't exceed available data (`GetEntryCount()`)
- Check for off-by-one errors in loops (e.g. `i < period` vs `i <= period`)
- Remember that `points.GetCollection(0)` is your **previous** output, not the current one

**Colors not appearing:**
- Ensure color strings are valid CSS colors (e.g. `"#ff0000"`, `"red"`, `"#ff000080"`)
- For CandleStick indicators, colors on individual `PlotPoint` values are ignored — set `BullishColor`/`BearishColor` in the config
- Verify your `PlotPoint` contains both `Value` and `Color`

**Performance issues:**
- Reduce logging in production
- Optimize loops over historical data
- Consider caching repeated calculations

---

## API Reference Summary

### Interfaces

| Interface | Description |
|-----------|-------------|
| `BaseIndicator` | Main indicator interface — all custom indicators must implement this |

### Types

| Type | Description |
|------|-------------|
| `IndicatorType` | Visualization type enum (0–6) |
| `ChartConfig` | Chart configuration (timeframe, active pair, loaded pairs) |
| `IndicatorConfig` | Indicator rendering configuration (target chart + type-specific config) |
| `PlotPoint` | Single data point with value, color, and invalid flag |
| `PlotPointCollection` | Collection of plot points for a single candle/iteration |
| `PlotPointCollectionHistory` | History container for previous indicator outputs |
| `IndicatorOutputHandler` | Output and logging handler for each iteration |
| `IndicatorLog` | A single log entry with level, time, and message |
| `IndicatorLogLevel` | Log level string enum: `"DEBUG"`, `"INFO"`, `"WARNING"`, `"ERROR"` |
| `StepAlignment` | Step alignment enum for StepLine: `StepAlignStart`, `StepAlignMidpoint`, `StepAlignEnd` |

### Config Types

| Config Type | For IndicatorType | Key Fields |
|-------------|-------------------|------------|
| `LineIndicatorConfig` | `IndicatorTypeLine` | `LineThickness`, `IsSmooth`, `LineTitles`, `ShowSymbols`, `Opacity` |
| `StepLineIndicatorConfig` | `IndicatorTypeStepLine` | `StepAlignment`, `LineTitles`, `LineThickness`, `Opacity` |
| `BarChartIndicatorConfig` | `IndicatorTypeBarChart` | `BarTitles`, `BarWidth`, `Opacity`, `Stacked` |
| `TableIndicatorConfig` | `IndicatorTypeTable` | `DisplayName`, `RowTitles`, `ColumnTitles`, `ValueDisplayFormat` |
| `AreaChartIndicatorConfig` | `IndicatorTypeAreaChart` | `LineTitles`, `LineThickness`, `IsSmooth`, `ShowSymbols`, `Opacity` |
| `PieChartIndicatorConfig` | `IndicatorTypePieChart` | `DisplayName`, `SliceTitles` |
| `CandleStickIndicatorConfig` | `IndicatorTypeCandleStick` | `BullishColor`, `BearishColor` |

### Key Functions

| Function | Description |
|----------|-------------|
| `chartsData.GetEntry(pair, index)` | Get candle by pair and index (0 = most recent) |
| `chartsData.GetEntryCount(pair)` | Get available candle count for a pair |
| `chartsData.GetEntries(pair)` | Get all candles for a pair |
| `handler.SetSingleOutput(point)` | Set single-value output |
| `handler.SetOutput(collection)` | Set multi-value output |
| `handler.GetOutput()` | Retrieve the current output |
| `handler.LogDebug(msg)` | Log debug message |
| `handler.LogInfo(msg)` | Log info message |
| `handler.LogWarning(msg)` | Log warning message |
| `handler.LogError(msg)` | Log error message |
| `handler.GetLogs()` | Retrieve all logs for current iteration |
| `NewPlotPointCollection(points ...PlotPoint)` | Create valid collection (variadic) |
| `NewInvalidPlotPointCollection()` | Create invalid collection (nothing displayed) |
| `NewInvalidPlotPoint()` | Create invalid point (gap in line) |
| `points.GetCollection(index)` | Get historical output (0 = most recent previous) |
| `points.Length()` | Get number of historical outputs available |
| `points.GetAllCollections()` | Get all historical outputs as slice |

---

## Indicator Development Checklist

Use this checklist when creating a new indicator:

- [ ] Create a struct and implement all four `BaseIndicator` interface methods
- [ ] Choose the correct `IndicatorType` (Line, StepLine, BarChart, Table, AreaChart, PieChart, CandleStick)
- [ ] Fill in the **matching** sub-config in `IndicatorConfig` (only the one for your type)
- [ ] Set `TargetChart` appropriately: `""` for main chart overlay, or a custom panel ID string
- [ ] Ensure `GetName()` returns a **unique** name among all indicators in the project
- [ ] In `HandleEntry`, validate data availability with `GetEntryCount()` before calculations
- [ ] Use `config.ActivePair` for pair-specific data access
- [ ] Return `NewInvalidPlotPointCollection()` when data is insufficient (not zero values)
- [ ] Ensure point count matches config titles (LineTitles, BarTitles, SliceTitles, RowTitles×ColumnTitles, or exactly 4 for CandleStick)
- [ ] Keep point count **consistent** across all iterations to avoid `DataError`
- [ ] Use `points.GetCollection()` for state across iterations instead of mutable struct fields
- [ ] Use logging for debugging (`LogDebug`) and significant events (`LogInfo`)
- [ ] Return fatal errors only for truly unrecoverable conditions
- [ ] Test with different timeframes and trading pairs
- [ ] Consider edge cases: missing data, invalid entries, first iteration, insufficient history

---

## Project Setup (Quick Reference)

This section shows the minimal steps to create a chart project, attach a data source, add trading pairs, and register your indicators. Detailed project configuration documentation is covered separately.

```go
package main

import (
    "candle-engine/pkg/chartproject"
    "candle-engine/pkg/sources"
    "candle-engine/pkg/sources/binance"
    "time"
)

func main() {
    // 1. Create a data source
    sourceBinance := binance.NewBinanceSource()

    // 2. Define the trading pairs to load
    pairs := []sources.SourcePairId{
        {SourceName: "Binance", PairId: "BTCUSDC"},
        {SourceName: "Binance", PairId: "ETHUSDC"},
    }

    // 3. Set the time range
    start, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
    end, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")

    // 4. Build the project
    project := chartproject.NewChartProject("My Project")
    project.AddSources(sourceBinance)
    project.AddTradingPairs(pairs...)
    project.SetTimeFrame("1d")
    project.SetTimeRange(sources.TimeRange{Start: start, End: end})

    // 5. Register your custom indicators
    project.AddIndicators(
        NewSMAIndicator(),
        NewRSIIndicator(),
        NewVolumeIndicator(),
    )

    // The project is now ready to be rendered or served via the API.
}
```

---

**Last Updated**: March 31, 2026
