package chartproject

import (
	"fmt"
	"sort"
	"time"

	"candle-engine/pkg/charting/chartproject/echarts"
	"candle-engine/pkg/charting/sources"
	"candle-engine/pkg/indicators"
)

// ============================================
// HELPER UTILITIES
// ============================================

// formatTimestamp converts a timestamp to a date string for Echarts x-axis
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02")
}

// convertToSeriesData converts PlotPointCollection array to series data
// Invalid data is represented as "-" string
func convertToSeriesData(collections []indicators.PlotPointCollection, pointIndex int) []interface{} {
	result := make([]interface{}, len(collections))
	for i, collection := range collections {
		if collection.IsInvalid() || pointIndex >= len(collection.Points) {
			result[i] = "-"
			continue
		}

		point := collection.Points[pointIndex]
		if point.Invalid {
			result[i] = "-"
		} else {
			result[i] = point.Value
		}
	}
	return result
}

// getSeriesColor extracts color from the first valid PlotPoint in collections
// Returns empty string if no valid color found
func getSeriesColor(collections []indicators.PlotPointCollection) string {
	for _, collection := range collections {
		if !collection.IsInvalid() && len(collection.Points) > 0 {
			for _, point := range collection.Points {
				if !point.Invalid && point.Color != "" {
					return normalizeColor(point.Color)
				}
			}
		}
	}
	return ""
}

// getSeriesColorForLine extracts color from a specific line index across all collections
// Returns empty string if no valid color found
func getSeriesColorForLine(collections []indicators.PlotPointCollection, lineIndex int) string {
	for _, collection := range collections {
		if !collection.IsInvalid() && lineIndex < len(collection.Points) {
			point := collection.Points[lineIndex]
			if !point.Invalid && point.Color != "" {
				return normalizeColor(point.Color)
			}
		}
	}
	return ""
}

// normalizeColor converts color strings to Echarts-compatible format
// Handles 8-digit hex colors (with alpha) by converting to rgba()
func normalizeColor(color string) string {
	// If color starts with # and has 8 hex digits (RRGGBBAA format)
	if len(color) == 9 && color[0] == '#' {
		// Convert #RRGGBBAA to rgba(r, g, b, a)
		r := hexToInt(color[1:3])
		g := hexToInt(color[3:5])
		b := hexToInt(color[5:7])
		a := float64(hexToInt(color[7:9])) / 255.0
		return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, a)
	}
	// Return as-is for other formats (#RGB, #RRGGBB, rgb(), rgba(), etc.)
	return color
}

// hexToInt converts a 2-character hex string to int
func hexToInt(hex string) int {
	var result int
	fmt.Sscanf(hex, "%x", &result)
	return result
}

// getSeriesTitle generates a title for a series, handling unnamed cases
func getSeriesTitle(titles []string, index int, indicatorName string) string {
	if index < len(titles) && titles[index] != "" {
		return titles[index]
	}
	return fmt.Sprintf("%s-UNNAMED-%d", indicatorName, index+1)
}

// ============================================
// SERIES BUILDERS FOR EACH INDICATOR TYPE
// ============================================

// buildLineIndicatorSeries creates Echarts line series from a line indicator
func buildLineIndicatorSeries(indicator RenderableIndicator, gridIndex int, xAxisIndex int, yAxisIndex int) []map[string]interface{} {
	var series []map[string]interface{}

	if len(indicator.PlotPoints) == 0 {
		return series
	}

	// Determine how many lines we need (based on first collection's point count)
	numLines := 0
	for _, collection := range indicator.PlotPoints {
		if !collection.IsInvalid() && collection.Length() > numLines {
			numLines = collection.Length()
		}
	}

	config := indicator.IndicatorConfig.LineIndicator

	// Create a separate series for each line
	for i := 0; i < numLines; i++ {
		seriesData := convertToSeriesData(indicator.PlotPoints, i)
		color := getSeriesColorForLine(indicator.PlotPoints, i)
		// TODO: Implement dynamic color changes within a single indicator (different color per data point)

		lineSeries := map[string]interface{}{
			"name":       getSeriesTitle(config.LineTitles, i, indicator.Name),
			"type":       "line",
			"xAxisIndex": xAxisIndex,
			"yAxisIndex": yAxisIndex,
			"smooth":     config.IsSmooth,
			"showSymbol": config.ShowSymbols,
			"data":       seriesData,
		}

		// Add item style for legend/tooltip colors
		if color != "" {
			lineSeries["itemStyle"] = map[string]interface{}{
				"color": color,
			}
		}

		// Add line style configuration
		lineStyle := make(map[string]interface{})
		if config.LineThickness > 0 {
			lineStyle["width"] = config.LineThickness
		}
		if color != "" {
			lineStyle["color"] = color
		}
		if config.Opacity > 0 {
			lineStyle["opacity"] = config.Opacity
		}
		if len(lineStyle) > 0 {
			lineSeries["lineStyle"] = lineStyle
		}

		series = append(series, lineSeries)
	}

	return series
}

// buildStepLineIndicatorSeries creates Echarts step line series from a step line indicator
func buildStepLineIndicatorSeries(indicator RenderableIndicator, gridIndex int, xAxisIndex int, yAxisIndex int) []map[string]interface{} {
	var series []map[string]interface{}

	if len(indicator.PlotPoints) == 0 {
		return series
	}

	// Determine how many lines we need
	numLines := 0
	for _, collection := range indicator.PlotPoints {
		if !collection.IsInvalid() && collection.Length() > numLines {
			numLines = collection.Length()
		}
	}

	config := indicator.IndicatorConfig.StepLineIndicator

	// Map step alignment to Echarts step value
	stepValue := "end"
	switch config.StepAlignment {
	case indicators.StepAlignStart:
		stepValue = "start"
	case indicators.StepAlignMidpoint:
		stepValue = "middle"
	case indicators.StepAlignEnd:
		stepValue = "end"
	}

	// Create a separate series for each line
	for i := 0; i < numLines; i++ {
		seriesData := convertToSeriesData(indicator.PlotPoints, i)
		color := getSeriesColorForLine(indicator.PlotPoints, i)
		// TODO: Implement dynamic color changes within a single indicator (different color per data point)

		stepSeries := map[string]interface{}{
			"name":       getSeriesTitle(config.LineTitles, i, indicator.Name),
			"type":       "line",
			"step":       stepValue,
			"xAxisIndex": xAxisIndex,
			"yAxisIndex": yAxisIndex,
			"data":       seriesData,
		}

		// Add item style for legend/tooltip colors
		if color != "" {
			stepSeries["itemStyle"] = map[string]interface{}{
				"color": color,
			}
		}

		// Add line style configuration
		lineStyle := make(map[string]interface{})
		if config.LineThickness > 0 {
			lineStyle["width"] = config.LineThickness
		}
		if color != "" {
			lineStyle["color"] = color
		}
		if config.Opacity > 0 {
			lineStyle["opacity"] = config.Opacity
		}
		if len(lineStyle) > 0 {
			stepSeries["lineStyle"] = lineStyle
		}

		series = append(series, stepSeries)
	}

	return series
}

// buildBarChartIndicatorSeries creates Echarts bar series from a bar chart indicator
func buildBarChartIndicatorSeries(indicator RenderableIndicator, gridIndex int, xAxisIndex int, yAxisIndex int) []map[string]interface{} {
	var series []map[string]interface{}

	if len(indicator.PlotPoints) == 0 {
		return series
	}

	// Determine how many bar groups we need
	numBars := 0
	for _, collection := range indicator.PlotPoints {
		if !collection.IsInvalid() && collection.Length() > numBars {
			numBars = collection.Length()
		}
	}

	config := indicator.IndicatorConfig.BarChartIndicator

	// Create a separate series for each bar group
	for i := 0; i < numBars; i++ {
		seriesData := convertToSeriesData(indicator.PlotPoints, i)
		color := getSeriesColorForLine(indicator.PlotPoints, i)
		// TODO: Implement dynamic color changes within a single indicator (different color per data point)

		barSeries := map[string]interface{}{
			"name":       getSeriesTitle(config.BarTitles, i, indicator.Name),
			"type":       "bar",
			"xAxisIndex": xAxisIndex,
			"yAxisIndex": yAxisIndex,
			"data":       seriesData,
		}

		// Add bar configuration
		if config.BarWidth != "" {
			barSeries["barWidth"] = config.BarWidth
		}

		// Add item style configuration
		itemStyle := make(map[string]interface{})
		if color != "" {
			itemStyle["color"] = color
		}
		if config.Opacity > 0 {
			itemStyle["opacity"] = config.Opacity
		}
		if len(itemStyle) > 0 {
			barSeries["itemStyle"] = itemStyle
		}

		// Add stacking if configured
		if config.Stacked {
			barSeries["stack"] = indicator.Name
		}

		series = append(series, barSeries)
	}

	return series
}

// buildAreaChartIndicatorSeries creates Echarts area chart series (line with fill)
func buildAreaChartIndicatorSeries(indicator RenderableIndicator, gridIndex int, xAxisIndex int, yAxisIndex int) []map[string]interface{} {
	var series []map[string]interface{}

	if len(indicator.PlotPoints) == 0 {
		return series
	}

	// Determine how many area series we need
	numAreas := 0
	for _, collection := range indicator.PlotPoints {
		if !collection.IsInvalid() && collection.Length() > numAreas {
			numAreas = collection.Length()
		}
	}

	config := indicator.IndicatorConfig.AreaChartIndicator

	// Create a separate series for each area
	for i := 0; i < numAreas; i++ {
		seriesData := convertToSeriesData(indicator.PlotPoints, i)
		color := getSeriesColorForLine(indicator.PlotPoints, i)
		// TODO: Implement dynamic color changes within a single indicator (different color per data point)

		areaSeries := map[string]interface{}{
			"name":       getSeriesTitle(config.LineTitles, i, indicator.Name),
			"type":       "line",
			"xAxisIndex": xAxisIndex,
			"yAxisIndex": yAxisIndex,
			"smooth":     config.IsSmooth,
			"showSymbol": config.ShowSymbols,
			"data":       seriesData,
		}

		// Add item style for legend/tooltip colors
		if color != "" {
			areaSeries["itemStyle"] = map[string]interface{}{
				"color": color,
			}
		}

		// Add area style (filled area under the line)
		areaStyle := make(map[string]interface{})
		if config.Opacity > 0 {
			areaStyle["opacity"] = config.Opacity
		} else {
			areaStyle["opacity"] = 0.3 // Default area opacity
		}
		if color != "" {
			areaStyle["color"] = color
		}
		areaSeries["areaStyle"] = areaStyle

		// Add line style configuration
		lineStyle := make(map[string]interface{})
		if config.LineThickness > 0 {
			lineStyle["width"] = config.LineThickness
		}
		if color != "" {
			lineStyle["color"] = color
		}
		if len(lineStyle) > 0 {
			areaSeries["lineStyle"] = lineStyle
		}

		series = append(series, areaSeries)
	}

	return series
}

// buildCandlestickSeries creates Echarts candlestick series from candle data
func buildCandlestickSeries(candleData []sources.ChartEntry, colors sources.CandleColorConfig, xAxisIndex int, yAxisIndex int) map[string]interface{} {
	data := make([][]float64, len(candleData))

	for i, entry := range candleData {
		if entry.IsValid() {
			// Echarts candlestick format: [open, close, low, high]
			data[i] = []float64{entry.Open, entry.Close, entry.Low, entry.High}
		} else {
			// Invalid data - use placeholder
			data[i] = []float64{0, 0, 0, 0}
		}
	}

	candleSeries := map[string]interface{}{
		"name":       "Candlestick",
		"type":       "candlestick",
		"xAxisIndex": xAxisIndex,
		"yAxisIndex": yAxisIndex,
		"data":       data,
		"itemStyle": map[string]interface{}{
			"color":        normalizeColor(colors.BullishColor),
			"color0":       normalizeColor(colors.BearishColor),
			"borderColor":  normalizeColor(colors.BullishColor),
			"borderColor0": normalizeColor(colors.BearishColor),
		},
	}

	return candleSeries
}

// ============================================
// GRID AND AXIS LAYOUT SYSTEM
// ============================================

// ChartLayoutInfo holds information about chart layout
type ChartLayoutInfo struct {
	ChartID     string
	GridIndex   int
	XAxisIndex  int
	YAxisIndex  int
	IsMainChart bool
}

// calculateChartLayout determines grid positions and creates axes
func calculateChartLayout(indicatorCharts map[string][]RenderableIndicator) ([]ChartLayoutInfo, []echarts.Grid, []echarts.XAxis, []echarts.YAxis) {
	// Collect and sort chart IDs (empty string first for main chart)
	chartIDs := make([]string, 0, len(indicatorCharts)+1)

	// Always ensure main chart (empty string) exists
	hasMainChart := false
	for chartID := range indicatorCharts {
		chartIDs = append(chartIDs, chartID)
		if chartID == "" {
			hasMainChart = true
		}
	}
	// Add main chart if it doesn't exist
	if !hasMainChart {
		chartIDs = append(chartIDs, "")
	}

	sort.Slice(chartIDs, func(i, j int) bool {
		// Empty string (main chart) comes first
		if chartIDs[i] == "" {
			return true
		}
		if chartIDs[j] == "" {
			return false
		}
		// Otherwise sort alphabetically
		return chartIDs[i] < chartIDs[j]
	})

	// Layout configuration
	const leftMargin = 20
	const rightMargin = 20
	const topStart = 110
	const mainChartHeight = 240 // 3x the subchart height
	const subChartHeight = 80
	const chartGap = 20

	var grids []echarts.Grid
	var xAxes []echarts.XAxis
	var yAxes []echarts.YAxis
	var layouts []ChartLayoutInfo

	currentTop := topStart

	for index, chartID := range chartIDs {
		isMainChart := (chartID == "")
		height := subChartHeight
		if isMainChart {
			height = mainChartHeight
		}

		// Create grid
		grid := echarts.Grid{
			Left:   leftMargin,
			Right:  rightMargin,
			Top:    currentTop,
			Height: height,
		}
		grids = append(grids, grid)

		// Create x-axis (will be populated with data later)
		xAxis := echarts.XAxis{
			Type:        "category",
			GridIndex:   index,
			Data:        []string{}, // Will be filled with dates
			BoundaryGap: false,
			AxisLine: echarts.AxisLine{
				LineStyle: echarts.LineStyle{
					Color: "#777",
				},
			},
			Min: "dataMin",
			Max: "dataMax",
		}

		// Show axis pointer on main chart
		if isMainChart {
			xAxis.AxisPointer = &echarts.XAxisPointer{
				Show: true,
			}
		} else {
			// Hide labels on subcharts
			xAxis.AxisLabel = echarts.AxisLabel{
				Show: false,
			}
			xAxis.AxisTick = &echarts.AxisTick{
				Show: false,
			}
			xAxis.SplitLine = &echarts.SplitLine{
				Show: false,
			}
		}

		xAxes = append(xAxes, xAxis)

		// Create y-axis
		yAxis := echarts.YAxis{
			Scale:       true,
			GridIndex:   index,
			SplitNumber: 2,
		}

		if isMainChart {
			yAxis.AxisLine = &echarts.AxisLine{
				LineStyle: echarts.LineStyle{
					Color: "#777",
				},
			}
			yAxis.SplitLine = &echarts.SplitLine{
				Show: true,
			}
			yAxis.AxisTick = &echarts.AxisTick{
				Show: false,
			}
			yAxis.AxisLabel = &echarts.AxisLabel{
				Inside:    true,
				Formatter: "{value}\n",
			}
		} else {
			yAxis.AxisLabel = &echarts.AxisLabel{
				Show: false,
			}
			yAxis.AxisLine = &echarts.AxisLine{
				Show: false,
			}
			yAxis.AxisTick = &echarts.AxisTick{
				Show: false,
			}
			yAxis.SplitLine = &echarts.SplitLine{
				Show: false,
			}
		}

		yAxes = append(yAxes, yAxis)

		// Store layout info
		layout := ChartLayoutInfo{
			ChartID:     chartID,
			GridIndex:   index,
			XAxisIndex:  index,
			YAxisIndex:  index,
			IsMainChart: isMainChart,
		}
		layouts = append(layouts, layout)

		currentTop += height + chartGap
	}

	return layouts, grids, xAxes, yAxes
}
