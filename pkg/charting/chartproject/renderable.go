package chartproject

import (
	"testcharting/pkg/charting/chartproject/echarts"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
	templates_chartproject "testcharting/templates/chartproject"
)

type RenderableIndicator struct {
	Name            string
	Type            indicators.IndicatorType
	IndicatorConfig indicators.IndicatorConfig
	PlotPoints      []indicators.PlotPointCollection

	DataError       bool // if true, indicates that there was an error in data rendering (e.g., inconsistent data lengths)
	RenderingErrors []error
	Logs            []indicators.IndicatorLog
}

// TODO - add methods to RenderableIndicator struct

// This will contain all the data needed to render the chart and indicators in the frontend
type RenderableChartProject struct {
	Title               string
	TimeRange           sources.TimeRange
	AvailableTimeFrames []string
	ActiveTimeFrame     string
	LoadedPairs         []sources.SourcePairId
	ActivePair          sources.SourcePairId

	CandleData   map[sources.SourcePairId][]sources.ChartEntry
	CandleColors sources.CandleColorConfig

	IndicatorCharts map[string][]RenderableIndicator

	// Other metadata
}

func NewRenderableChartProject() RenderableChartProject {
	return RenderableChartProject{}
}

func (rcp *RenderableChartProject) GetLogs() map[string][]templates_chartproject.ProjectLogData {
	var logList map[string][]templates_chartproject.ProjectLogData
	logList = make(map[string][]templates_chartproject.ProjectLogData)

	for _, indicators := range rcp.IndicatorCharts {
		for _, indicator := range indicators {
			// First collect rendering errors
			if indicator.DataError {
				var renderErrorLogs []templates_chartproject.ProjectLogData
				for _, err := range indicator.RenderingErrors {
					renderErrorLogs = append(renderErrorLogs, templates_chartproject.ProjectLogData{
						LogLevel:      templates_chartproject.LogLevelError,
						TimeFormatted: "",
						Message:       "Rendering error for indicator " + indicator.Name + ": " + err.Error(),
					})
				}

				if logList["RenderingErrors"] == nil {
					logList["RenderingErrors"] = []templates_chartproject.ProjectLogData{}
				}
				logList["RenderingErrors"] = append(logList["RenderingErrors"], renderErrorLogs...)
			}

			var indicatorLogList []templates_chartproject.ProjectLogData
			for _, logEntry := range indicator.Logs {
				indicatorLogList = append(indicatorLogList, templates_chartproject.ProjectLogData{
					LogLevel:      templates_chartproject.LogLevelEnum(logEntry.Level),
					TimeFormatted: logEntry.Time.Format("2006:01:02 15:04:05"),
					Message:       logEntry.Message,
				})
			}

			if logList[indicator.Name] == nil {
				logList[indicator.Name] = []templates_chartproject.ProjectLogData{}
			}
			logList[indicator.Name] = append(logList[indicator.Name], indicatorLogList...)
		}
	}

	return logList
}

func (rcp *RenderableChartProject) GetEchartsOptions() echarts.EchartsOptions {
	// Calculate layout for all charts
	layouts, grids, xAxes, yAxes := calculateChartLayout(rcp.IndicatorCharts)

	// Extract timestamps from active pair's candle data
	timestamps := []string{}
	if candleData, ok := rcp.CandleData[rcp.ActivePair]; ok {
		for _, entry := range candleData {
			timestamps = append(timestamps, formatTimestamp(entry.Timestamp))
		}
	}

	// Populate all x-axes with timestamp data
	for i := range xAxes {
		xAxes[i].Data = timestamps
	}

	// Build all series
	var allSeries []map[string]interface{}
	var legendData []string

	// Process each chart layout
	for _, layout := range layouts {
		chartIndicators := rcp.IndicatorCharts[layout.ChartID]

		// Add candlestick data to main chart
		if layout.IsMainChart {
			if candleData, ok := rcp.CandleData[rcp.ActivePair]; ok {
				candleSeries := buildCandlestickSeries(candleData, rcp.CandleColors, layout.XAxisIndex, layout.YAxisIndex)
				allSeries = append(allSeries, candleSeries)
				legendData = append(legendData, "Candlestick")
			}
		}

		// Add indicator series
		for _, indicator := range chartIndicators {
			var indicatorSeries []map[string]interface{}

			switch indicator.Type {
			case indicators.IndicatorTypeLine:
				indicatorSeries = buildLineIndicatorSeries(indicator, layout.GridIndex, layout.XAxisIndex, layout.YAxisIndex)
			case indicators.IndicatorTypeStepLine:
				indicatorSeries = buildStepLineIndicatorSeries(indicator, layout.GridIndex, layout.XAxisIndex, layout.YAxisIndex)
			case indicators.IndicatorTypeBarChart:
				indicatorSeries = buildBarChartIndicatorSeries(indicator, layout.GridIndex, layout.XAxisIndex, layout.YAxisIndex)
			case indicators.IndicatorTypeAreaChart:
				indicatorSeries = buildAreaChartIndicatorSeries(indicator, layout.GridIndex, layout.XAxisIndex, layout.YAxisIndex)
			case indicators.IndicatorTypeTable:
				// Table indicators are handled differently (not rendered in Echarts)
				continue
			}

			allSeries = append(allSeries, indicatorSeries...)

			// Add series names to legend
			for _, series := range indicatorSeries {
				if name, ok := series["name"].(string); ok {
					legendData = append(legendData, name)
				}
			}
		}
	}

	// Create axis pointer links (link all x-axes together)
	xAxisIndices := make([]int, len(xAxes))
	for i := range xAxes {
		xAxisIndices[i] = i
	}

	// Create data zoom configuration (applies to all x-axes)
	dataZoom := []echarts.DataZoom{
		{
			Type:       "slider",
			XAxisIndex: xAxisIndices,
			Realtime:   true,
			Start:      20,
			End:        70,
			Top:        65,
			Height:     20,
		},
		{
			Type:       "inside",
			XAxisIndex: xAxisIndices,
			Start:      40,
			End:        70,
			Top:        30,
			Height:     20,
		},
	}

	// Assemble the complete EchartsOptions
	options := echarts.EchartsOptions{
		Animation: false, // Disable animations for better performance
		Color:     []string{"#c23531", "#2f4554", "#61a0a8", "#d48265", "#91c7ae", "#749f83", "#ca8622", "#bda29a", "#6e7074", "#546570", "#c4ccd3"},
		Title: echarts.Title{
			Left: "center",
			Text: rcp.Title,
		},
		Legend: echarts.Legend{
			Top:  30,
			Data: legendData,
		},
		Tooltip: echarts.Tooltip{
			Trigger:            "axis",
			TriggerOn:          "mousemove|click",
			TransitionDuration: 0,
			Confine:            true,
			BorderRadius:       4,
			BorderWidth:        1,
			BorderColor:        "#333",
			BackgroundColor:    "rgba(255,255,255,0.9)",
			TextStyle: echarts.TooltipTextStyle{
				FontSize: 12,
				Color:    "#333",
			},
		},
		AxisPointer: echarts.AxisPointer{
			Link: []echarts.AxisPointerLink{
				{
					XAxisIndex: xAxisIndices,
				},
			},
		},
		DataZoom: dataZoom,
		XAxis:    xAxes,
		YAxis:    yAxes,
		Grid:     grids,
		Series:   allSeries,
	}

	return options
}

// TODO - add methods to RenderableChartProject struct
