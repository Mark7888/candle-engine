package echarts

// EchartsOptions represents the complete configuration for an Apache ECharts instance
type EchartsOptions struct {
	Animation   bool                     `json:"animation"`
	Color       []string                 `json:"color,omitempty"`
	Title       Title                    `json:"title"`
	Legend      Legend                   `json:"legend"`
	Tooltip     Tooltip                  `json:"tooltip"`
	AxisPointer AxisPointer              `json:"axisPointer"`
	DataZoom    []DataZoom               `json:"dataZoom"`
	XAxis       []XAxis                  `json:"xAxis"`
	YAxis       []YAxis                  `json:"yAxis"`
	Grid        []Grid                   `json:"grid"`
	Series      []map[string]interface{} `json:"series"`
}

// Title configuration for the chart
type Title struct {
	Left string `json:"left"`
	Text string `json:"text"`
}

// Legend configuration
type Legend struct {
	Top  int      `json:"top"`
	Data []string `json:"data"`
}

// Tooltip configuration
type Tooltip struct {
	Trigger            string           `json:"trigger,omitempty"`
	TriggerOn          string           `json:"triggerOn,omitempty"`
	TransitionDuration int              `json:"transitionDuration"`
	Confine            bool             `json:"confine"`
	BorderRadius       int              `json:"borderRadius"`
	BorderWidth        int              `json:"borderWidth"`
	BorderColor        string           `json:"borderColor"`
	BackgroundColor    string           `json:"backgroundColor"`
	TextStyle          TooltipTextStyle `json:"textStyle"`
}

type TooltipTextStyle struct {
	FontSize int    `json:"fontSize"`
	Color    string `json:"color"`
}

// AxisPointer configuration for linking multiple axes
type AxisPointer struct {
	Link []AxisPointerLink `json:"link"`
}

type AxisPointerLink struct {
	XAxisIndex []int `json:"xAxisIndex"`
}

// DataZoom configuration for zooming and panning
type DataZoom struct {
	Type       string  `json:"type"`
	XAxisIndex []int   `json:"xAxisIndex"`
	Realtime   bool    `json:"realtime"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Top        int     `json:"top,omitempty"`
	Height     int     `json:"height,omitempty"`
}

// XAxis configuration
type XAxis struct {
	Type        string        `json:"type"`
	GridIndex   int           `json:"gridIndex,omitempty"`
	Data        []string      `json:"data"`
	BoundaryGap interface{}   `json:"boundaryGap"` // can be bool or array
	AxisLine    AxisLine      `json:"axisLine"`
	AxisLabel   AxisLabel     `json:"axisLabel,omitempty"`
	SplitLine   *SplitLine    `json:"splitLine,omitempty"`
	AxisTick    *AxisTick     `json:"axisTick,omitempty"`
	Min         interface{}   `json:"min,omitempty"` // can be string "dataMin" or number
	Max         interface{}   `json:"max,omitempty"` // can be string "dataMax" or number
	AxisPointer *XAxisPointer `json:"axisPointer,omitempty"`
}

type AxisLine struct {
	Show      bool      `json:"show,omitempty"`
	LineStyle LineStyle `json:"lineStyle"`
}

type LineStyle struct {
	Color string `json:"color"`
	Width int    `json:"width,omitempty"`
}

type AxisLabel struct {
	Show      bool        `json:"show,omitempty"`
	Inside    bool        `json:"inside,omitempty"`
	Formatter interface{} `json:"formatter,omitempty"` // can be string or function (as string in JSON)
}

type SplitLine struct {
	Show bool `json:"show"`
}

type AxisTick struct {
	Show bool `json:"show"`
}

type XAxisPointer struct {
	Show           bool               `json:"show,omitempty"`
	Type           string             `json:"type,omitempty"`
	Label          *AxisPointerLabel  `json:"label,omitempty"`
	TriggerTooltip bool               `json:"triggerTooltip,omitempty"`
	Handle         *AxisPointerHandle `json:"handle,omitempty"`
}

type AxisPointerLabel struct {
	Show bool `json:"show"`
}

type AxisPointerHandle struct {
	Show   bool   `json:"show"`
	Margin int    `json:"margin"`
	Color  string `json:"color"`
}

// YAxis configuration
type YAxis struct {
	Scale       bool       `json:"scale"`
	GridIndex   int        `json:"gridIndex,omitempty"`
	SplitNumber int        `json:"splitNumber"`
	AxisLine    *AxisLine  `json:"axisLine,omitempty"`
	SplitLine   *SplitLine `json:"splitLine,omitempty"`
	AxisTick    *AxisTick  `json:"axisTick,omitempty"`
	AxisLabel   *AxisLabel `json:"axisLabel,omitempty"`
}

// Grid configuration for chart layout
type Grid struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Height int `json:"height"`
}
