# Complete Guide: Passing Go Data to JavaScript in a-h/Templ

## Overview
This guide covers all methods for passing large amounts of data from Go to JavaScript in templ script tags without using HTTP requests. For your use case (populating ECharts options with Go data), multiple approaches are available.

---

## Method 1: Script Templates with Parameters (⭐ RECOMMENDED)

### Description
The `script` keyword in templ creates reusable JavaScript functions that automatically JSON-encode Go data passed as parameters.

### How It Works
```templ
script chartRenderScript(chartId string, chartData ChartOptions) {
    var targetChart = echarts.init(document.getElementById(chartId));
    // chartData is automatically JSON-encoded
    targetChart.setOption(chartData);
}

templ ChartRender(id string, options ChartOptions) {
    <div data-chart-render id={ id }></div>
    @chartRenderScript(id, options)
}
```

### Generated Output
```html
<script>function __templ_chartRenderScript_abc123(chartId, chartData) {
    var targetChart = echarts.init(document.getElementById(chartId));
    targetChart.setOption(chartData);
}</script>
<script>__templ_chartRenderScript_abc123("chart-1", {"title":{"text":"My Chart"},"series":[...]})</script>
```

### Pros
- ✅ **Zero serialization overhead** - templ handles JSON encoding efficiently
- ✅ Clean separation between function definition and invocation
- ✅ Type-safe in Go code
- ✅ Reusable across multiple chart instances
- ✅ Handles complex nested structs automatically

### Cons
- ⚠️ Legacy feature (though still fully supported)
- ⚠️ Function names are auto-generated (includes hash suffix)

### Best For
- **Large chart configurations with complex nested data**
- Multiple chart instances on the same page
- Reusable chart components

---

## Method 2: Direct Interpolation in Script Tags

### Description
Use `{{ value }}` syntax within script tags to interpolate Go data directly into JavaScript code.

### How It Works
```templ
templ ChartRender(id string, options ChartOptions) {
    <div id={ id }></div>
    <script>
        const chartOptions = {{ options }};
        const chart = echarts.init(document.getElementById({{ id }}));
        chart.setOption(chartOptions);
    </script>
}
```

### String Interpolation vs Object Interpolation
```templ
script example(msg string, data MyStruct) {
    // Inside strings - plain interpolation
    const message = "Your message: {{ msg }}";  // Result: "Your message: Hello"
    
    // Outside strings - JSON encoded
    const obj = {{ data }};  // Result: {"field1":"value1","field2":123}
}
```

### Generated Output
```html
<div id="chart-1"></div>
<script>
    const chartOptions = {"title":{"text":"My Chart"},"series":[...]};
    const chart = echarts.init(document.getElementById("chart-1"));
    chart.setOption(chartOptions);
</script>
```

### Pros
- ✅ Very straightforward and readable
- ✅ Automatic JSON encoding
- ✅ XSS protection via automatic escaping
- ✅ No function wrapper overhead

### Cons
- ⚠️ Mixes data with logic (harder to maintain per docs)
- ⚠️ Not reusable across components
- ⚠️ Script runs inline with each render

### Best For
- **Single-use chart instances**
- Simpler configurations
- Quick prototyping

---

## Method 3: templ.JSONScript() - Separate Data from Logic

### Description
Embeds Go data as a JSON script tag that JavaScript can parse separately.

### How It Works
```templ
templ ChartRender(id string, options ChartOptions) {
    <div id={ id }></div>
    @templ.JSONScript(id + "-data", options)
    <script>
        const options = JSON.parse(document.getElementById('{{ id }}-data').textContent);
        const chart = echarts.init(document.getElementById('{{ id }}'));
        chart.setOption(options);
    </script>
}
```

### Generated Output
```html
<div id="chart-1"></div>
<script id="chart-1-data" type="application/json">
{"title":{"text":"My Chart"},"series":[...]}
</script>
<script>
    const options = JSON.parse(document.getElementById('chart-1-data').textContent);
    const chart = echarts.init(document.getElementById('chart-1'));
    chart.setOption(options);
</script>
```

### Pros
- ✅ **Best practice** - separates data from logic
- ✅ Easier to debug (data is visible in DOM)
- ✅ Can cache and reuse data across multiple scripts
- ✅ Clean separation of concerns

### Cons
- ⚠️ Requires additional DOM lookup (`getElementById`)
- ⚠️ Slightly more verbose

### Best For
- **Large datasets that might be reused**
- Cleaner code architecture
- Debugging complex data structures

---

## Method 4: templ.JSONString() in HTML Attributes

### Description
Pass Go data as JSON-encoded HTML attributes (commonly used with Alpine.js, HTMX, etc.)

### How It Works
```templ
templ ChartRender(id string, options ChartOptions) {
    <div 
        id={ id } 
        data-chart-options={ templ.JSONString(options) }
        data-chart-ready="false">
    </div>
    <script>
        const element = document.getElementById('{{ id }}');
        const options = JSON.parse(element.getAttribute('data-chart-options'));
        const chart = echarts.init(element);
        chart.setOption(options);
        element.setAttribute('data-chart-ready', 'true');
    </script>
}
```

### Generated Output
```html
<div id="chart-1" 
     data-chart-options="{&quot;title&quot;:{&quot;text&quot;:&quot;My Chart&quot;},&quot;series&quot;:[...]}">
</div>
<script>
    const element = document.getElementById('chart-1');
    const options = JSON.parse(element.getAttribute('data-chart-options'));
    const chart = echarts.init(element);
    chart.setOption(options);
</script>
```

### Pros
- ✅ Data is attached directly to the element
- ✅ Works well with progressive enhancement
- ✅ Compatible with frameworks like Alpine.js
- ✅ Self-contained component

### Cons
- ⚠️ HTML attribute encoding (uses `&quot;` etc.) - looks messy
- ⚠️ Browser limits on attribute size (though quite large)
- ⚠️ Not ideal for very large datasets

### Best For
- **Medium-sized chart configurations**
- Components that need to be self-contained
- Integration with attribute-based frameworks

---

## Method 5: templ.JSFuncCall() - Call JS Functions with Go Data

### Description
Explicitly call a JavaScript function with JSON-encoded Go parameters.

### How It Works
```templ
templ ChartRender(id string, options ChartOptions) {
    <div id={ id }></div>
    <script>
        function initChart(chartId, options) {
            const chart = echarts.init(document.getElementById(chartId));
            chart.setOption(options);
        }
    </script>
    @templ.JSFuncCall("initChart", id, options)
}
```

### Generated Output
```html
<div id="chart-1"></div>
<script>
    function initChart(chartId, options) {
        const chart = echarts.init(document.getElementById(chartId));
        chart.setOption(options);
    }
</script>
<script>
    initChart("chart-1", {"title":{"text":"My Chart"},"series":[...]});
</script>
```

### Pros
- ✅ Clear and explicit function calls
- ✅ Good for event handlers
- ✅ Automatic JSON encoding of parameters
- ✅ Security: function names are sanitized

### Cons
- ⚠️ Requires separate function definition
- ⚠️ More verbose than script templates

### Best For
- Event handlers with server data
- Explicit initialization patterns
- When you need clear function boundaries

---

## Method 6: Raw Go Variables in Script Tags

### Description
Use `{{ variableName := value }}` to declare scoped variables within templates.

### How It Works
```templ
templ ChartRender(items []DataPoint, chartType string) {
    {{ firstItem := items[0] }}
    {{ chartConfig := buildChartConfig(items, chartType) }}
    
    <div id="chart"></div>
    <script>
        const config = {{ chartConfig }};
        const chart = echarts.init(document.getElementById('chart'));
        chart.setOption(config);
        console.log('First data point:', {{ firstItem }});
    </script>
}
```

### Pros
- ✅ Reduces function calls in template
- ✅ Can pre-process data before rendering
- ✅ Variables scoped to template

### Cons
- ⚠️ Limited to template scope
- ⚠️ Can't be used for complex transformations

### Best For
- Simple data transformations
- Avoiding repeated function calls
- Accessing nested struct fields

---

## Comparison Table

| Method | Data Size | Performance | Maintainability | Reusability | Use Case |
|--------|-----------|-------------|-----------------|-------------|----------|
| Script Templates | ⭐⭐⭐⭐⭐ Large | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ High | Multiple charts, complex configs |
| Direct Interpolation | ⭐⭐⭐⭐ Medium | ⭐⭐⭐⭐ Very Good | ⭐⭐⭐ Fair | ⭐⭐ Low | Single-use, simple cases |
| JSONScript | ⭐⭐⭐⭐⭐ Large | ⭐⭐⭐⭐ Very Good | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good | Best practices, debugging |
| JSONString Attr | ⭐⭐⭐ Medium | ⭐⭐⭐ Good | ⭐⭐⭐⭐ Good | ⭐⭐⭐ Medium | Self-contained components |
| JSFuncCall | ⭐⭐⭐⭐ Medium-Large | ⭐⭐⭐⭐ Very Good | ⭐⭐⭐⭐ Good | ⭐⭐⭐ Medium | Event handlers, explicit calls |
| Raw Variables | ⭐⭐ Small | ⭐⭐⭐⭐ Very Good | ⭐⭐⭐ Fair | ⭐⭐ Low | Pre-processing, simple transforms |

---

## Recommended Approach for ECharts with Large Data

### Strategy: Hybrid Approach

```templ
package templates_charts

import "encoding/json"

// Define your Go structs for type safety
type EChartsOption struct {
    Title   TitleOption   `json:"title"`
    Tooltip TooltipOption `json:"tooltip"`
    Legend  LegendOption  `json:"legend"`
    XAxis   XAxisOption   `json:"xAxis"`
    YAxis   YAxisOption   `json:"yAxis"`
    Series  []SeriesOption `json:"series"`
}

type SeriesOption struct {
    Name string        `json:"name"`
    Type string        `json:"type"`
    Data []interface{} `json:"data"`
}

// Method 1: Using Script Template (Best for reusable components)
script initEChart(chartId string, options EChartsOption) {
    const element = document.getElementById(chartId);
    if (!element) {
        console.error("Element with ID " + chartId + " not found.");
        return;
    }
    const chart = echarts.init(element);
    chart.setOption(options);
    
    // Handle window resize
    window.addEventListener('resize', () => {
        chart.resize();
    });
}

templ ChartRender(id string, options EChartsOption) {
    <div data-chart-render id={ id }></div>
    @initEChart(id, options)
    @chartDimensionFixScript(id)
}

// Method 2: Using JSONScript (Best for debugging/architecture)
templ ChartRenderWithDataSeparation(id string, options EChartsOption) {
    <div data-chart-render id={ id }></div>
    @templ.JSONScript(id + "-options", options)
    <script>
        (function() {
            const chartId = {{ id }};
            const options = JSON.parse(
                document.getElementById(chartId + '-options').textContent
            );
            const chart = echarts.init(document.getElementById(chartId));
            chart.setOption(options);
            
            window.addEventListener('resize', () => chart.resize());
        })();
    </script>
    @chartDimensionFixScript(id)
}

// Method 3: For very large data - progressive loading pattern
script initChartWithLargeData(chartId string, initialOptions EChartsOption) {
    const element = document.getElementById(chartId);
    if (!element) {
        console.error("Element with ID " + chartId + " not found.");
        return;
    }
    
    const chart = echarts.init(element);
    
    // Set initial options immediately
    chart.setOption(initialOptions);
    
    // Mark as loading if needed
    if (initialOptions.series && initialOptions.series.some(s => s.data.length === 0)) {
        chart.showLoading();
    }
}

script appendChartData(chartId string, seriesIndex int, newData []interface{}) {
    const element = document.getElementById(chartId);
    if (!element) return;
    
    const chart = echarts.getInstanceByDom(element);
    if (!chart) return;
    
    const currentOption = chart.getOption();
    currentOption.series[seriesIndex].data = 
        currentOption.series[seriesIndex].data.concat(newData);
    
    chart.setOption(currentOption);
    chart.hideLoading();
}
```

---

## Performance Considerations

### 1. JSON Encoding is Fast
Templ uses Go's `encoding/json` package, which is highly optimized. For most use cases, this is not a bottleneck.

### 2. Data Structure Optimization
```go
// Good - flat structure
type ChartData struct {
    Labels []string  `json:"labels"`
    Values []float64 `json:"values"`
}

// Less optimal - deeply nested
type ChartData struct {
    Series []struct {
        Points []struct {
            Label string
            Value float64
        }
    }
}
```

### 3. Use JSON Tags
Always use json tags for predictable output:
```go
type SeriesData struct {
    Name      string    `json:"name"`
    Type      string    `json:"type"`
    DataSet   []float64 `json:"data"`      // Renamed in JSON
    ShowLabel bool      `json:"showLabel"` // camelCase for JS
}
```

### 4. Handle Time/Date Properly
```go
type TimeSeriesPoint struct {
    Time  string  `json:"time"`  // Format as string in Go
    Value float64 `json:"value"`
}

// In your handler
point := TimeSeriesPoint{
    Time: t.Format("2006-01-02"), // ISO format
    Value: value,
}
```

---

## Security Considerations

### 1. XSS Protection
All methods automatically escape/encode data:
- String interpolation: HTML escaped
- JSON encoding: Automatically escaped
- Attributes: HTML attribute encoded

### 2. CSP (Content Security Policy)
If using strict CSP, prefer:
- Script templates over inline scripts
- Separate script files when possible
- Use nonces with `templ.WithNonce()`

### 3. User-Generated Content
```go
// Always sanitize user input before rendering
type SafeChartData struct {
    Title string `json:"title"` // Will be JSON-encoded, then HTML-escaped
}

// For completely unsafe content (avoid if possible)
templ UnsafeExample(unsafeJS string) {
    @templ.JSUnsafeFuncCall(unsafeJS) // ⚠️ Dangerous - only use with trusted input
}
```

---

## Complete Working Example

```templ
package templates_charts

type CandlestickData struct {
    Timestamps []string      `json:"timestamps"`
    Data       [][]float64   `json:"data"` // [open, close, low, high]
    Volumes    []float64     `json:"volumes"`
}

type EChartsConfig struct {
    Title      string          `json:"-"` // Don't serialize, use separately
    ChartData  CandlestickData `json:"candlestick"`
    Indicators []IndicatorData `json:"indicators"`
}

script renderCandlestickChart(chartId string, config EChartsConfig, title string) {
    const chart = echarts.init(document.getElementById(chartId));
    
    const option = {
        title: {
            text: title
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross'
            }
        },
        xAxis: {
            data: config.candlestick.timestamps
        },
        yAxis: [
            {
                scale: true,
                splitArea: {
                    show: true
                }
            },
            {
                scale: true,
                gridIndex: 0,
                splitNumber: 2,
                axisLabel: { show: false },
                axisLine: { show: false },
                axisTick: { show: false },
                splitLine: { show: false }
            }
        ],
        series: [
            {
                name: 'Candlestick',
                type: 'candlestick',
                data: config.candlestick.data,
            },
            {
                name: 'Volume',
                type: 'bar',
                xAxisIndex: 0,
                yAxisIndex: 1,
                data: config.candlestick.volumes
            }
        ].concat(config.indicators.map(ind => ({
            name: ind.name,
            type: 'line',
            data: ind.values,
            smooth: true
        })))
    };
    
    chart.setOption(option);
    window.addEventListener('resize', () => chart.resize());
}

templ CandlestickChart(id string, config EChartsConfig) {
    <div 
        id={ id } 
        style="width: 100%; height: 500px;"
        data-chart-type="candlestick">
    </div>
    @renderCandlestickChart(id, config, config.Title)
}
```

### Usage in Go Handler:
```go
func ChartHandler(w http.ResponseWriter, r *http.Request) {
    data := templates_charts.EChartsConfig{
        Title: "BTC/USD",
        ChartData: templates_charts.CandlestickData{
            Timestamps: []string{"2024-01-01", "2024-01-02", ...},
            Data: [][]float64{
                {40000, 41000, 39500, 40500}, // open, close, low, high
                {41000, 42000, 40800, 41500},
                // ... thousands more
            },
            Volumes: []float64{1000000, 1500000, ...},
        },
        Indicators: []templates_charts.IndicatorData{
            {Name: "MA20", Values: []float64{40100, 40200, ...}},
            {Name: "MA50", Values: []float64{39900, 40000, ...}},
        },
    }
    
    component := templates_charts.CandlestickChart("btc-chart", data)
    component.Render(r.Context(), w)
}
```

---

## Conclusion

For your specific use case (ECharts with large datasets):

### ✅ **Primary Recommendation: Script Templates (Method 1)**
- Handles large data efficiently
- Clean, reusable code
- Type-safe in Go
- No serialization overhead concerns

### ✅ **Alternative: JSONScript (Method 3)**
- Better code organization
- Easier debugging
- Follows best practices
- Marginal performance difference

### ❌ **Avoid: HTTP Requests**
- As requested, not using AJAX/Fetch
- All data embedded in initial render

The JSON encoding performance of Go's standard library is excellent and should handle even large datasets (10,000+ data points) without issues. The actual bottleneck is more likely to be ECharts rendering time rather than templ's data passing.
