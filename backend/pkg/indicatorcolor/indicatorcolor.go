package indicatorcolor

import (
	"candle-engine/pkg/charting/sources"
	"candle-engine/pkg/indicators"
	"fmt"
)

// TODO - this only works for single indicators
func GetNewColoredIndicator(i indicators.BaseIndicator, numberOfColors int) indicators.BaseIndicator {
	return &ColoredIndicator{
		Indicator:      i,
		NumberOfColors: numberOfColors,
	}
}

type ColoredIndicator struct {
	Indicator      indicators.BaseIndicator
	NumberOfColors int
	KnownColors    []string
}

func (r *ColoredIndicator) GetType() indicators.IndicatorType {
	return r.Indicator.GetType()
}

func (r *ColoredIndicator) GetName() string {
	return r.Indicator.GetName()
}

func (r *ColoredIndicator) GetConfig() indicators.IndicatorConfig {
	// We have to update config with the separate lines for coloring
	config := r.Indicator.GetConfig()
	// Update config with the number of colors
	config.LineIndicator.LineTitles = make([]string, r.NumberOfColors)
	for i := 0; i < r.NumberOfColors; i++ {
		config.LineIndicator.LineTitles[i] = "Color " + fmt.Sprintf("%d", i+1)
	}
	// TODO - this is temporary, will have to set each to the same title, but match the number ot titles.
	return config
}

func (r *ColoredIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	outputHandler, err := r.Indicator.HandleEntry(points, chartsData, config, handler)
	if err != nil {
		return outputHandler, err
	}

	// Apply coloring logic here, by rendering it as multiple lines, each with a different color, hiding the lines other than the actual color
	// We have to keep track of how many colors there are?

	value := outputHandler.GetOutput()
	point := value.Points[0]

	if point.Invalid {
		// Still have to fill in the number of points, but all invalid
		value.Points = make([]indicators.PlotPoint, r.NumberOfColors)
		for i := 0; i < r.NumberOfColors; i++ {
			value.Points[i] = indicators.NewInvalidPlotPoint()
		}
		outputHandler.SetOutput(value)
		return outputHandler, nil
	}

	color := point.Color
	lineIndex, ok := r.isColorWithinLimitAndTrack(color)
	if !ok {
		// This means we have more colors than we can handle,
		// which means we have a problem with the coloring logic in the indicator, as it should only be outputting a single color per point,
		// and that color should be consistent for the same value (so that it doesn't randomly change colors on the chart)
		return outputHandler, fmt.Errorf("ColoredIndicator: More colors found while rendering than specified in constructor. Color: %s, KnownColors: %v, NumberOfColors: %d", color, r.KnownColors, r.NumberOfColors)
	}

	value.Points = make([]indicators.PlotPoint, r.NumberOfColors)
	for i := 0; i < r.NumberOfColors; i++ {
		if i == lineIndex {
			value.Points[i] = point
			outputHandler.LogDebug(fmt.Sprintf("Color %s assigned to line index %d", color, lineIndex))
		} else {
			value.Points[i] = indicators.NewInvalidPlotPoint()
		}
	}

	outputHandler.SetOutput(value)

	return outputHandler, nil
}

func (r *ColoredIndicator) isColorWithinLimitAndTrack(color string) (int, bool) {
	for i, c := range r.KnownColors {
		if c == color {
			return i, true
		}
	}
	r.KnownColors = append(r.KnownColors, color)
	return len(r.KnownColors) - 1, len(r.KnownColors) <= r.NumberOfColors
}
