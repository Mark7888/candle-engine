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

// ---------------------------------------------------------------------------
// Color utilities
// ---------------------------------------------------------------------------

/**
 * Converts #RRGGBBAA (8-char hex) to rgba().
 * Other color formats are returned as-is.
 */
export function normalizeColor(color) {
  if (!color) return color;
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

export function buildTimestamps(candleEntries) {
  return candleEntries.map((e) => {
    const d = new Date(e.timestamp);
    return d.toISOString().split('T')[0];
  });
}

/**
 * Returns the DataZoom start/end percentages so that roughly the last 3 months
 * of data are visible by default. Uses the last candle's timestamp as the
 * reference point so the window is always relative to the actual data, not
 * wall-clock time.
 */
function computeDefaultZoom(candleEntries) {
  const total = candleEntries.length;
  if (total === 0) return { start: 0, end: 100 };

  const lastDate = new Date(candleEntries[total - 1].timestamp);
  // Use 90 days in ms instead of setMonth() to avoid month-boundary rollover
  // issues (e.g. Jan 31 − 3 months ≠ Oct 31 on some engines).
  const THREE_MONTHS_MS = 90 * 24 * 60 * 60 * 1000;
  const windowStart = new Date(lastDate.getTime() - THREE_MONTHS_MS);

  let startIdx = 0;
  for (let i = 0; i < total; i++) {
    if (new Date(candleEntries[i].timestamp) >= windowStart) {
      startIdx = i;
      break;
    }
  }

  const startPercent = Math.round((startIdx / total) * 100);
  // Cap at 95 so at least 5 % of the data range is always visible when the
  // 3-month window is very close to the end of a short dataset.
  const MAX_START = 95;
  return { start: Math.min(startPercent, MAX_START), end: 100 };
}

function getPointValue(plotPoints, timeIdx, lineIdx) {
  const collection = plotPoints[timeIdx];
  if (!collection || collection.invalid) return '-';
  const point = collection.points[lineIdx];
  if (!point || point.invalid) return '-';
  return point.value;
}

function extractSeriesData(plotPoints, lineIdx) {
  return plotPoints.map((_, i) => getPointValue(plotPoints, i, lineIdx));
}

function findLineColor(plotPoints, lineIdx) {
  for (const collection of plotPoints) {
    if (collection.invalid) continue;
    const point = collection.points[lineIdx];
    if (point && !point.invalid && point.color) {
      return normalizeColor(point.color);
    }
  }
  return undefined;
}

function countLines(plotPoints) {
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

function buildLineSeries(indicator, xAxisIndex, yAxisIndex) {
  const cfg = indicator.config ?? {};
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series = {
      name,
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: cfg.isSmooth ?? false,
      showSymbol: cfg.showSymbols ?? false,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    const lineStyle = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (cfg.opacity) lineStyle.opacity = cfg.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildStepLineSeries(indicator, xAxisIndex, yAxisIndex) {
  const cfg = indicator.config ?? {};
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  const stepMap = { 0: 'start', 1: 'middle', 2: 'end' };
  const step = stepMap[cfg.stepAlignment] ?? 'end';
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series = {
      name,
      type: 'line',
      step,
      xAxisIndex,
      yAxisIndex,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    const lineStyle = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (cfg.opacity) lineStyle.opacity = cfg.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildBarSeries(indicator, xAxisIndex, yAxisIndex) {
  const cfg = indicator.config ?? {};
  const titles = cfg.barTitles ?? [];
  const numBars = countLines(indicator.plotPoints);
  return Array.from({ length: numBars }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series = {
      name,
      type: 'bar',
      xAxisIndex,
      yAxisIndex,
      data: extractSeriesData(indicator.plotPoints, i),
    };
    if (cfg.barWidth) series.barWidth = cfg.barWidth;
    if (cfg.stacked) series.stack = indicator.name;
    const itemStyle = {};
    if (color) itemStyle.color = color;
    if (cfg.opacity) itemStyle.opacity = cfg.opacity;
    if (Object.keys(itemStyle).length) series.itemStyle = itemStyle;
    return series;
  });
}

function buildAreaSeries(indicator, xAxisIndex, yAxisIndex) {
  const cfg = indicator.config ?? {};
  const titles = cfg.lineTitles ?? [];
  const numLines = countLines(indicator.plotPoints);
  return Array.from({ length: numLines }, (_, i) => {
    const color = findLineColor(indicator.plotPoints, i);
    const name = titles[i] || `${indicator.name}-${i + 1}`;
    const series = {
      name,
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: cfg.isSmooth ?? false,
      showSymbol: cfg.showSymbols ?? false,
      data: extractSeriesData(indicator.plotPoints, i),
      areaStyle: { opacity: cfg.opacity ?? 0.3, ...(color ? { color } : {}) },
    };
    const lineStyle = {};
    if (cfg.lineThickness) lineStyle.width = cfg.lineThickness;
    if (color) lineStyle.color = color;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    if (color) series.itemStyle = { color };
    return series;
  });
}

function buildCandlestickSeries(candleEntries, candleColors, xAxisIndex, yAxisIndex) {
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

function groupIndicatorsByChart(indicators) {
  const groups = {};
  for (const ind of indicators) {
    const key = ind.targetChart ?? '';
    if (!groups[key]) groups[key] = [];
    groups[key].push(ind);
  }
  if (!groups['']) groups[''] = [];
  return groups;
}

function sortedChartIds(groups) {
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
 * @param {object} renderData - ProjectRenderResponse from the API
 * @param {boolean} isDark - current color mode
 * @returns {object} ECharts option object (includes _totalHeight for container sizing)
 */
export function buildEchartsOption(renderData, isDark = true) {
  if (!renderData) return {};

  const { candles, candleColors, indicators, activePair } = renderData;

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
  const timestamps = buildTimestamps(candleEntries);

  const textColor = isDark ? '#aaa' : '#555';
  const lineColor = isDark ? '#444' : '#ddd';
  const bgColor = isDark ? '#1a202c' : '#fff';

  const groups = groupIndicatorsByChart(indicators ?? []);
  const chartIds = sortedChartIds(groups);
  const numCharts = chartIds.length;

  let currentTop = LAYOUT.topStart;
  const grids = [];
  const xAxes = [];
  const yAxes = [];
  const allXAxisIndices = [];

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
  const allSeries = [];
  const legendData = [];

  for (let idx = 0; idx < numCharts; idx++) {
    const chartId = chartIds[idx];
    const isMain = chartId === '';

    if (isMain) {
      const cs = buildCandlestickSeries(candleEntries, candleColors, idx, idx);
      allSeries.push(cs);
      legendData.push({ name: cs.name });
    }

    for (const ind of groups[chartId] ?? []) {
      let built = [];
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

  const { start: zoomStart, end: zoomEnd } = computeDefaultZoom(candleEntries);

  const dataZoom = [
    {
      type: 'inside',
      xAxisIndex: allXAxisIndices,
      start: zoomStart,
      end: zoomEnd,
    },
    {
      type: 'slider',
      xAxisIndex: allXAxisIndices,
      bottom: 10,
      height: LAYOUT.dataZoomHeight,
      start: zoomStart,
      end: zoomEnd,
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
