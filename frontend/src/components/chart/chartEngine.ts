/**
 * Chart Engine – builds ECharts option objects from raw API data.
 *
 * Key design decisions:
 *  - One ECharts instance with multiple grids stacked vertically.
 *  - Main chart (targetChart === "") sits at the top and contains the
 *    candlestick series plus any overlay indicators.
 *  - Sub-charts (non-empty targetChart) appear below, sorted alphabetically.
 *  - All X-axes share the same category data and are linked via axisPointer,
 *    so hovering highlights the same timestamp across every panel.
 *  - A single DataZoom (inside + slider) controls all X-axes simultaneously.
 */

import type {
  CandleEntry,
  CandleColors,
  Indicator,
  PlotPointCollection,
  RenderResponse,
  LineIndicatorConfig,
  BarIndicatorConfig,
  StepLineIndicatorConfig,
  CandleStickIndicatorConfig,
} from '../../types/api';

// Use a flexible type for ECharts option objects since ECharts types
// are complex and change across versions; we build them incrementally.
type EchartsObj = Record<string, unknown>;

export interface EchartsOptionWithHeight extends EchartsObj {
  _totalHeight?: number;
}

// ---------------------------------------------------------------------------
// Color utilities
// ---------------------------------------------------------------------------

/**
 * Converts #RRGGBBAA (8-char hex) to rgba().
 * Other color formats are returned as-is.
 */
export function normalizeColor(color: string | undefined | null): string | undefined {
  if (!color) return undefined;
  if (color.length === 9 && color[0] === '#') {
    const r = parseInt(color.slice(1, 3), 16);
    const g = parseInt(color.slice(3, 5), 16);
    const b = parseInt(color.slice(5, 7), 16);
    const a = (parseInt(color.slice(7, 9), 16) / 255).toFixed(2);
    return `rgba(${r}, ${g}, ${b}, ${a})`;
  }
  return color;
}

// ---------------------------------------------------------------------------
// Data extraction helpers
// ---------------------------------------------------------------------------

/**
 * Returns true when the timeFrame string represents a sub-day interval
 * (minutes or hours), e.g. "1m", "5m", "1h", "4h".
 */
function isSubDayTimeFrame(timeFrame: string | undefined): boolean {
  return typeof timeFrame === 'string' && /^\d+[mh]$/.test(timeFrame);
}

export function buildTimestamps(
  candleEntries: CandleEntry[],
  timeFrame: string | undefined
): string[] {
  const includeTime = isSubDayTimeFrame(timeFrame);
  return candleEntries.map((e) => {
    const d = new Date(e.timestamp);
    const yyyy = d.getUTCFullYear();
    const mo = String(d.getUTCMonth() + 1).padStart(2, '0');
    const dd = String(d.getUTCDate()).padStart(2, '0');
    if (!includeTime) return `${yyyy}-${mo}-${dd}`;
    const hh = String(d.getUTCHours()).padStart(2, '0');
    const mm = String(d.getUTCMinutes()).padStart(2, '0');
    return `${yyyy}-${mo}-${dd} ${hh}:${mm}`;
  });
}



function getPointValue(
  plotPoints: PlotPointCollection[],
  timeIdx: number,
  lineIdx: number
): number | string {
  const collection = plotPoints[timeIdx];
  if (!collection || collection.invalid) return '-';
  const point = collection.points[lineIdx];
  if (!point || point.invalid) return '-';
  return point.value;
}

function extractSeriesData(
  plotPoints: PlotPointCollection[],
  lineIdx: number
): (number | string)[] {
  return plotPoints.map((_, i) => getPointValue(plotPoints, i, lineIdx));
}

function findLineColor(
  plotPoints: PlotPointCollection[],
  lineIdx: number
): string | undefined {
  for (const collection of plotPoints) {
    if (collection.invalid) continue;
    const point = collection.points[lineIdx];
    if (point && !point.invalid && point.color) {
      return normalizeColor(point.color);
    }
  }
  return undefined;
}

function countLines(plotPoints: PlotPointCollection[]): number {
  for (const collection of plotPoints) {
    if (!collection.invalid && collection.points.length > 0) {
      return collection.points.length;
    }
  }
  return 0;
}

// ---------------------------------------------------------------------------
// Series builders per indicator type
// ---------------------------------------------------------------------------

function buildLineSeries(
  indicator: Indicator,
  xAxisIndex: number,
  yAxisIndex: number
): EchartsObj[] {
  const cfg = (indicator.config ?? {}) as LineIndicatorConfig;
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series: EchartsObj = {
      name,
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: cfg.isSmooth ?? false,
      showSymbol: cfg.showSymbols ?? false,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    const lineStyle: EchartsObj = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (cfg.opacity) lineStyle.opacity = cfg.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildStepLineSeries(
  indicator: Indicator,
  xAxisIndex: number,
  yAxisIndex: number
): EchartsObj[] {
  const cfg = (indicator.config ?? {}) as StepLineIndicatorConfig;
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  const stepMap: Record<number, string> = { 0: 'start', 1: 'middle', 2: 'end' };
  const step = (cfg.stepAlignment !== undefined ? stepMap[cfg.stepAlignment] : undefined) ?? 'end';
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series: EchartsObj = {
      name,
      type: 'line',
      step,
      xAxisIndex,
      yAxisIndex,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    const lineStyle: EchartsObj = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (cfg.opacity) lineStyle.opacity = cfg.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildBarSeries(
  indicator: Indicator,
  xAxisIndex: number,
  yAxisIndex: number
): EchartsObj[] {
  const cfg = (indicator.config ?? {}) as BarIndicatorConfig;
  const titles = cfg.barTitles ?? [];
  const numBars = countLines(indicator.plotPoints);
  return Array.from({ length: numBars }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series: EchartsObj = {
      name,
      type: 'bar',
      xAxisIndex,
      yAxisIndex,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    if (cfg.barWidth) series.barWidth = cfg.barWidth;
    if (cfg.stacked) series.stack = indicator.name;
    const itemStyle: EchartsObj = {};
    if (color) itemStyle.color = color;
    if (cfg.opacity) itemStyle.opacity = cfg.opacity;
    if (Object.keys(itemStyle).length) series.itemStyle = itemStyle;
    return series;
  });
}

function buildAreaSeries(
  indicator: Indicator,
  xAxisIndex: number,
  yAxisIndex: number
): EchartsObj[] {
  const cfg = (indicator.config ?? {}) as LineIndicatorConfig;
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series: EchartsObj = {
      name,
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: cfg.isSmooth ?? false,
      showSymbol: cfg.showSymbols ?? false,
      data: extractSeriesData(indicator.plotPoints, i),
      areaStyle: { opacity: cfg.opacity ?? 0.3, ...(color ? { color } : {}) },
    };
    const lineStyle: EchartsObj = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildCandlestickSeries(
  candleEntries: CandleEntry[],
  candleColors: CandleColors,
  xAxisIndex: number,
  yAxisIndex: number
): EchartsObj {
  const data = candleEntries.map((e) => [e.open, e.close, e.low, e.high]);
  return {
    name: 'Candlestick',
    type: 'candlestick',
    xAxisIndex,
    yAxisIndex,
    data,
    itemStyle: {
      color: normalizeColor(candleColors.bullishColor),
      color0: normalizeColor(candleColors.bearishColor),
      borderColor: normalizeColor(candleColors.bullishColor),
      borderColor0: normalizeColor(candleColors.bearishColor),
    },
  };
}

function buildCandleStickSubSeries(
  indicator: Indicator,
  xAxisIndex: number,
  yAxisIndex: number,
  fallbackColors: CandleColors
): EchartsObj[] {
  const cfg = (indicator.config ?? {}) as CandleStickIndicatorConfig;
  const bullish = cfg.bullishColor || fallbackColors.bullishColor;
  const bearish = cfg.bearishColor || fallbackColors.bearishColor;

  const data = indicator.plotPoints.map((coll) => {
    if (coll.invalid || coll.points.length < 4) return null;
    const [o, h, l, c] = coll.points;
    if (o.invalid || h.invalid || l.invalid || c.invalid) return null;
    // ECharts candlestick data format: [open, close, low, high]
    return [o.value, c.value, l.value, h.value];
  });

  return [{
    name: indicator.name,
    type: 'candlestick',
    xAxisIndex,
    yAxisIndex,
    data,
    itemStyle: {
      color:        normalizeColor(bullish),
      color0:       normalizeColor(bearish),
      borderColor:  normalizeColor(bullish),
      borderColor0: normalizeColor(bearish),
    },
  }];
}

// ---------------------------------------------------------------------------
// Layout calculator
// ---------------------------------------------------------------------------

const LAYOUT = {
  leftMargin: 60,
  rightMargin: 20,
  topStart: 50,
  mainChartHeight: 300,
  subChartHeight: 100,
  chartGap: 30,
  dataZoomHeight: 30,
  dataZoomGap: 10,
};

function groupIndicatorsByChart(indicators: Indicator[]): Record<string, Indicator[]> {
  const groups: Record<string, Indicator[]> = {};
  for (const ind of indicators) {
    const key = ind.targetChart ?? '';
    if (!groups[key]) groups[key] = [];
    groups[key].push(ind);
  }
  if (!groups['']) groups[''] = [];
  return groups;
}

function sortedChartIds(groups: Record<string, Indicator[]>): string[] {
  return Object.keys(groups).sort((a, b) => {
    if (a === '') return -1;
    if (b === '') return 1;
    return a.localeCompare(b);
  });
}

// ---------------------------------------------------------------------------
// Main option builder
// ---------------------------------------------------------------------------

/**
 * Builds a complete ECharts option from API render data.
 *
 * @param renderData - ProjectRenderResponse from the API
 * @param isDark - current color mode
 * @returns ECharts option object (includes _totalHeight for container sizing)
 */
export function buildEchartsOption(
  renderData: RenderResponse | null | undefined,
  isDark = true
): EchartsOptionWithHeight {
  if (!renderData) return {};

  const { candles, candleColors, indicators, activePair, activeTimeFrame } = renderData;

  // Use activePair from the response to find the correct candle series.
  // The backend may return data for multiple pairs; activePair identifies which
  // one the indicators are aligned with.
  const activeSeries = activePair
    ? candles?.find(
        (c) => c.sourceName === activePair.sourceName && c.pairId === activePair.pairId
      )
    : candles?.[0];
  if (!activeSeries) return {};

  const candleEntries = activeSeries.entries ?? [];
  const timestamps = buildTimestamps(candleEntries, activeTimeFrame);

  const textColor = isDark ? '#aaa' : '#555';
  const lineColor = isDark ? '#444' : '#ddd';
  const bgColor = isDark ? '#1a202c' : '#fff';

  // Filter out drawer-only types — they are rendered in IndicatorPanel, not ECharts.
  const echartsIndicators = (indicators ?? []).filter(
    (ind) => ind.type !== 'table' && ind.type !== 'pieChart'
  );

  const groups = groupIndicatorsByChart(echartsIndicators);
  const chartIds = sortedChartIds(groups);
  const numCharts = chartIds.length;

  let currentTop = LAYOUT.topStart;
  const grids: EchartsObj[] = [];
  const xAxes: EchartsObj[] = [];
  const yAxes: EchartsObj[] = [];
  const allXAxisIndices: number[] = [];

  for (let idx = 0; idx < numCharts; idx++) {
    const chartId = chartIds[idx];
    const isMain = chartId === '';
    const height = isMain ? LAYOUT.mainChartHeight : LAYOUT.subChartHeight;

    grids.push({
      left: LAYOUT.leftMargin,
      right: LAYOUT.rightMargin,
      top: currentTop,
      height,
    });

    xAxes.push({
      type: 'category',
      gridIndex: idx,
      data: timestamps,
      boundaryGap: false,
      axisLine: { lineStyle: { color: lineColor } },
      axisLabel: { show: isMain, color: textColor, fontSize: 11 },
      axisTick: { show: isMain },
      splitLine: { show: false },
      axisPointer: { show: true, type: 'line', label: { show: isMain } },
    });

    yAxes.push({
      scale: true,
      gridIndex: idx,
      splitNumber: isMain ? 4 : 2,
      axisLine: { show: isMain, lineStyle: { color: lineColor } },
      axisTick: { show: false },
      splitLine: {
        show: isMain,
        lineStyle: { color: lineColor, type: 'dashed' },
      },
      axisLabel: {
        show: true,
        inside: isMain,
        color: textColor,
        fontSize: isMain ? 11 : 10,
        formatter: isMain ? '{value}\n' : '{value}',
      },
    });

    allXAxisIndices.push(idx);
    currentTop += height + LAYOUT.chartGap;
  }

  // Build series
  const allSeries: EchartsObj[] = [];
  const legendData: EchartsObj[] = [];

  for (let idx = 0; idx < numCharts; idx++) {
    const chartId = chartIds[idx];
    const isMain = chartId === '';

    if (isMain) {
      const cs = buildCandlestickSeries(candleEntries, candleColors, idx, idx);
      allSeries.push(cs);
      legendData.push({ name: cs.name });
    }

    for (const ind of groups[chartId] ?? []) {
      let built: EchartsObj[] = [];
      switch (ind.type) {
        case 'line':
          built = buildLineSeries(ind, idx, idx);
          break;
        case 'stepLine':
          built = buildStepLineSeries(ind, idx, idx);
          break;
        case 'bar':
          built = buildBarSeries(ind, idx, idx);
          break;
        case 'area':
          built = buildAreaSeries(ind, idx, idx);
          break;
        case 'candleStick':
          built = buildCandleStickSubSeries(ind, idx, idx, candleColors);
          break;
        default:
          break;
      }
      for (const s of built) {
        legendData.push({ name: s.name });
        allSeries.push(s);
      }
    }
  }

  const legend = {
    show: true,
    top: 10,
    right: LAYOUT.rightMargin,
    textStyle: { color: textColor, fontSize: 11 },
    data: legendData,
  };

  const DEFAULT_CANDLES = 90;
  const total = candleEntries.length;
  const zoomStartValue = total > DEFAULT_CANDLES ? total - DEFAULT_CANDLES : 0;
  const zoomEndValue = Math.max(0, total - 1);

  const dataZoom = [
    {
      type: 'inside',
      xAxisIndex: allXAxisIndices,
      startValue: zoomStartValue,
      endValue: zoomEndValue,
    },
    {
      type: 'slider',
      xAxisIndex: allXAxisIndices,
      bottom: 10,
      height: LAYOUT.dataZoomHeight,
      startValue: zoomStartValue,
      endValue: zoomEndValue,
      handleStyle: { color: isDark ? '#6366f1' : '#4f46e5' },
      fillerColor: isDark ? 'rgba(99,102,241,0.15)' : 'rgba(79,70,229,0.1)',
      borderColor: isDark ? '#444' : '#ddd',
      textStyle: { color: textColor },
    },
  ];

  const tooltip = {
    trigger: 'axis',
    triggerOn: 'mousemove|click',
    transitionDuration: 0,
    confine: true,
    borderRadius: 6,
    borderWidth: 1,
    borderColor: isDark ? '#444' : '#ddd',
    backgroundColor: isDark ? '#1e2432' : '#fff',
    textStyle: { fontSize: 12, color: textColor },
    axisPointer: {
      type: 'cross',
      link: [{ xAxisIndex: allXAxisIndices }],
      lineStyle: { color: isDark ? '#6366f1' : '#4f46e5', type: 'dashed' },
      label: { show: true, backgroundColor: isDark ? '#1e2432' : '#eee' },
    },
  };

  const totalHeight =
    currentTop + LAYOUT.dataZoomHeight + LAYOUT.dataZoomGap + 20;

  return {
    _totalHeight: totalHeight,
    animation: false,
    backgroundColor: bgColor,
    legend,
    tooltip,
    axisPointer: {
      link: [{ xAxisIndex: allXAxisIndices }],
    },
    dataZoom,
    grid: grids,
    xAxis: xAxes,
    yAxis: yAxes,
    series: allSeries,
  };
}
