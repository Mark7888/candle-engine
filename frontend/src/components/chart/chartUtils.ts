/**
 * Builds Apache ECharts options from the raw API render response.
 *
 * Layout model (mirrors Go echarts_builder.go):
 *  - One ECharts instance with multiple linked grids.
 *  - Main chart (targetChart === "") is always first and 3× taller.
 *  - Sub-charts are sorted alphabetically by targetChart ID.
 *  - All X-axes share the same category data and are linked for cross-chart highlighting.
 *  - DataZoom controls all grids simultaneously.
 */

import type {
  CandleSeries,
  CandleColors,
  IndicatorData,
  PlotPointCollection,
  LineIndicatorConfig,
  StepLineIndicatorConfig,
  BarChartIndicatorConfig,
  AreaChartIndicatorConfig,
} from '../../types/api';

// Layout constants matching Go renderer
const LEFT_MARGIN = '60px';
const RIGHT_MARGIN = '20px';
const TOP_START = 110;
const MAIN_CHART_HEIGHT = 240;
const SUB_CHART_HEIGHT = 80;
const CHART_GAP = 20;

// ─── Color helpers ─────────────────────────────────────────────────────────────

/**
 * Converts #RRGGBBAA → rgba(r,g,b,a) or passes through other formats unchanged.
 */
function normalizeColor(color: string): string {
  if (color.length === 9 && color[0] === '#') {
    const r = parseInt(color.slice(1, 3), 16);
    const g = parseInt(color.slice(3, 5), 16);
    const b = parseInt(color.slice(5, 7), 16);
    const a = parseInt(color.slice(7, 9), 16) / 255;
    return `rgba(${r},${g},${b},${a.toFixed(2)})`;
  }
  return color;
}

// ─── Data helpers ──────────────────────────────────────────────────────────────

function getSeriesColorForLine(collections: PlotPointCollection[], lineIndex: number): string {
  for (const col of collections) {
    if (!col.invalid && lineIndex < col.points.length) {
      const pt = col.points[lineIndex];
      if (!pt.invalid && pt.color) return normalizeColor(pt.color);
    }
  }
  return '';
}

function convertToSeriesData(collections: PlotPointCollection[], pointIndex: number): (number | '-')[] {
  return collections.map((col) => {
    if (col.invalid || pointIndex >= col.points.length) return '-';
    const pt = col.points[pointIndex];
    return pt.invalid ? '-' : pt.value;
  });
}

function getSeriesTitle(titles: string[], index: number, indicatorName: string): string {
  if (index < titles.length && titles[index]) return titles[index];
  return `${indicatorName}-${index + 1}`;
}

function numLines(collections: PlotPointCollection[]): number {
  let max = 0;
  for (const col of collections) {
    if (!col.invalid && col.points.length > max) max = col.points.length;
  }
  return max;
}

// ─── Series builders ───────────────────────────────────────────────────────────

function buildCandlestickSeries(
  candles: CandleSeries,
  colors: CandleColors,
  xAxisIndex: number,
  yAxisIndex: number,
) {
  const data = candles.entries.map((e) => [e.open, e.close, e.low, e.high]);
  return {
    name: 'Candlestick',
    type: 'candlestick',
    xAxisIndex,
    yAxisIndex,
    data,
    itemStyle: {
      color: normalizeColor(colors.bullishColor),
      color0: normalizeColor(colors.bearishColor),
      borderColor: normalizeColor(colors.bullishColor),
      borderColor0: normalizeColor(colors.bearishColor),
    },
  };
}

function buildLineSeries(
  indicator: IndicatorData,
  xAxisIndex: number,
  yAxisIndex: number,
) {
  const config = indicator.config as LineIndicatorConfig;
  const n = numLines(indicator.plotPoints);
  return Array.from({ length: n }, (_, i) => {
    const color = getSeriesColorForLine(indicator.plotPoints, i);
    const series: Record<string, unknown> = {
      name: getSeriesTitle(config.lineTitles ?? [], i, indicator.name),
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: config.isSmooth ?? false,
      showSymbol: config.showSymbols ?? false,
      data: convertToSeriesData(indicator.plotPoints, i),
    };
    if (color) series.itemStyle = { color };
    const lineStyle: Record<string, unknown> = {};
    if ((config.lineThickness ?? 0) > 0) lineStyle.width = config.lineThickness;
    if (color) lineStyle.color = color;
    if ((config.opacity ?? 0) > 0) lineStyle.opacity = config.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    return series;
  });
}

function buildStepLineSeries(
  indicator: IndicatorData,
  xAxisIndex: number,
  yAxisIndex: number,
) {
  const config = indicator.config as StepLineIndicatorConfig;
  const stepMap: Record<number, string> = { 0: 'start', 1: 'middle', 2: 'end' };
  const step = stepMap[config.stepAlignment ?? 2] ?? 'end';
  const n = numLines(indicator.plotPoints);
  return Array.from({ length: n }, (_, i) => {
    const color = getSeriesColorForLine(indicator.plotPoints, i);
    const series: Record<string, unknown> = {
      name: getSeriesTitle(config.lineTitles ?? [], i, indicator.name),
      type: 'line',
      step,
      xAxisIndex,
      yAxisIndex,
      data: convertToSeriesData(indicator.plotPoints, i),
    };
    if (color) series.itemStyle = { color };
    const lineStyle: Record<string, unknown> = {};
    if ((config.lineThickness ?? 0) > 0) lineStyle.width = config.lineThickness;
    if (color) lineStyle.color = color;
    if ((config.opacity ?? 0) > 0) lineStyle.opacity = config.opacity;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    return series;
  });
}

function buildBarSeries(
  indicator: IndicatorData,
  xAxisIndex: number,
  yAxisIndex: number,
) {
  const config = indicator.config as BarChartIndicatorConfig;
  const n = numLines(indicator.plotPoints);
  return Array.from({ length: n }, (_, i) => {
    const color = getSeriesColorForLine(indicator.plotPoints, i);
    const series: Record<string, unknown> = {
      name: getSeriesTitle(config.barTitles ?? [], i, indicator.name),
      type: 'bar',
      xAxisIndex,
      yAxisIndex,
      data: convertToSeriesData(indicator.plotPoints, i),
    };
    if (config.barWidth) series.barWidth = config.barWidth;
    const itemStyle: Record<string, unknown> = {};
    if (color) itemStyle.color = color;
    if ((config.opacity ?? 0) > 0) itemStyle.opacity = config.opacity;
    if (Object.keys(itemStyle).length) series.itemStyle = itemStyle;
    if (config.stacked) series.stack = indicator.name;
    return series;
  });
}

function buildAreaSeries(
  indicator: IndicatorData,
  xAxisIndex: number,
  yAxisIndex: number,
) {
  const config = indicator.config as AreaChartIndicatorConfig;
  const n = numLines(indicator.plotPoints);
  return Array.from({ length: n }, (_, i) => {
    const color = getSeriesColorForLine(indicator.plotPoints, i);
    const series: Record<string, unknown> = {
      name: getSeriesTitle(config.lineTitles ?? [], i, indicator.name),
      type: 'line',
      xAxisIndex,
      yAxisIndex,
      smooth: config.isSmooth ?? false,
      showSymbol: config.showSymbols ?? false,
      data: convertToSeriesData(indicator.plotPoints, i),
    };
    if (color) series.itemStyle = { color };
    const opacity = (config.opacity ?? 0) > 0 ? config.opacity : 0.3;
    const areaStyle: Record<string, unknown> = { opacity };
    if (color) areaStyle.color = color;
    series.areaStyle = areaStyle;
    const lineStyle: Record<string, unknown> = {};
    if ((config.lineThickness ?? 0) > 0) lineStyle.width = config.lineThickness;
    if (color) lineStyle.color = color;
    if (Object.keys(lineStyle).length) series.lineStyle = lineStyle;
    return series;
  });
}

// ─── Main builder ──────────────────────────────────────────────────────────────

export interface ChartLayoutMeta {
  chartId: string;
  gridIndex: number;
  isMainChart: boolean;
  label: string;
}

export interface EChartsResult {
  option: Record<string, unknown>;
  layouts: ChartLayoutMeta[];
  totalHeight: number;
}

export function buildEChartsOption(
  candles: CandleSeries[],
  colors: CandleColors,
  indicators: IndicatorData[],
): EChartsResult {
  // 1. Collect unique chart IDs, sort: "" first then alphabetical
  const idSet = new Set<string>(['']);
  for (const ind of indicators) {
    if (ind.type !== 'table') idSet.add(ind.targetChart);
  }
  const sortedIds = Array.from(idSet).sort((a, b) => {
    if (a === '') return -1;
    if (b === '') return 1;
    return a.localeCompare(b);
  });

  // 2. Build grids, axes, and layout meta
  const grids: Record<string, unknown>[] = [];
  const xAxes: Record<string, unknown>[] = [];
  const yAxes: Record<string, unknown>[] = [];
  const layouts: ChartLayoutMeta[] = [];
  const axisPointerLinks: { xAxisIndex: number[] }[] = [{ xAxisIndex: [] }];

  let currentTop = TOP_START;

  for (let idx = 0; idx < sortedIds.length; idx++) {
    const chartId = sortedIds[idx];
    const isMain = chartId === '';
    const height = isMain ? MAIN_CHART_HEIGHT : SUB_CHART_HEIGHT;

    grids.push({
      left: LEFT_MARGIN,
      right: RIGHT_MARGIN,
      top: currentTop,
      height,
    });

    const xAxis: Record<string, unknown> = {
      type: 'category',
      gridIndex: idx,
      boundaryGap: false,
      data: [], // filled below
      min: 'dataMin',
      max: 'dataMax',
      axisLine: { lineStyle: { color: '#777' } },
    };

    if (isMain) {
      xAxis.axisPointer = { show: true };
    } else {
      xAxis.axisLabel = { show: false };
      xAxis.axisTick = { show: false };
      xAxis.splitLine = { show: false };
    }
    xAxes.push(xAxis);

    const yAxis: Record<string, unknown> = {
      scale: true,
      gridIndex: idx,
      splitNumber: 2,
    };
    if (isMain) {
      yAxis.axisLine = { lineStyle: { color: '#777' } };
      yAxis.splitLine = { show: true };
      yAxis.axisTick = { show: false };
      yAxis.axisLabel = { inside: true, formatter: '{value}\n' };
    } else {
      yAxis.axisLabel = { show: false };
      yAxis.axisLine = { show: false };
      yAxis.axisTick = { show: false };
      yAxis.splitLine = { show: false };
    }
    yAxes.push(yAxis);

    layouts.push({ chartId, gridIndex: idx, isMainChart: isMain, label: isMain ? 'Main' : chartId });
    (axisPointerLinks[0].xAxisIndex as number[]).push(idx);

    currentTop += height + CHART_GAP;
  }

  // 3. Build timestamp category data from the active candle series
  //    All axes share the same timestamps (index-aligned).
  const activeSeries = candles[0];
  const timestamps = activeSeries?.entries.map((e) => {
    const d = new Date(e.timestamp);
    return d.toISOString().slice(0, 10); // YYYY-MM-DD
  }) ?? [];
  for (const xAxis of xAxes) {
    xAxis.data = timestamps;
  }

  // 4. Build series
  const allSeries: Record<string, unknown>[] = [];
  const legendData: string[] = [];

  // Group indicators by chartId
  const byChart = new Map<string, IndicatorData[]>();
  for (const id of sortedIds) byChart.set(id, []);
  for (const ind of indicators) {
    if (ind.type === 'table') continue;
    byChart.get(ind.targetChart)?.push(ind);
  }

  for (const layout of layouts) {
    const { chartId, gridIndex } = layout;
    const xAxisIndex = gridIndex;
    const yAxisIndex = gridIndex;

    // Candlestick goes in the main chart
    if (layout.isMainChart && activeSeries) {
      allSeries.push(buildCandlestickSeries(activeSeries, colors, xAxisIndex, yAxisIndex));
      legendData.push('Candlestick');
    }

    for (const ind of byChart.get(chartId) ?? []) {
      let built: Record<string, unknown>[] = [];
      switch (ind.type) {
        case 'line':
          built = buildLineSeries(ind, xAxisIndex, yAxisIndex);
          break;
        case 'stepLine':
          built = buildStepLineSeries(ind, xAxisIndex, yAxisIndex);
          break;
        case 'bar':
          built = buildBarSeries(ind, xAxisIndex, yAxisIndex);
          break;
        case 'area':
          built = buildAreaSeries(ind, xAxisIndex, yAxisIndex);
          break;
      }
      for (const s of built) {
        allSeries.push(s);
        legendData.push(s.name as string);
      }
    }
  }

  // 5. DataZoom — covers all grids
  const dataZoom = [
    {
      type: 'slider',
      xAxisIndex: sortedIds.map((_, i) => i),
      start: 20,
      end: 70,
      bottom: 0,
    },
    {
      type: 'inside',
      xAxisIndex: sortedIds.map((_, i) => i),
      start: 40,
      end: 70,
    },
  ];

  // 6. Total container height (room for title + charts + datazoom slider)
  const totalHeight = currentTop + 50; // 50px extra for datazoom slider

  // 7. Assemble full option
  const option: Record<string, unknown> = {
    animation: false,
    backgroundColor: 'transparent',
    color: [
      '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
      '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc', '#4ecdc4', '#45b7d1',
    ],
    legend: {
      top: 60,
      data: legendData,
      textStyle: { color: '#ccc' },
      inactiveColor: '#555',
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      transitionDuration: 0,
      confine: true,
      borderRadius: 4,
      backgroundColor: 'rgba(20,20,30,0.9)',
      borderColor: '#444',
      textStyle: { color: '#ddd', fontSize: 12 },
    },
    axisPointer: {
      link: axisPointerLinks,
      label: { backgroundColor: '#1e1e2e' },
    },
    dataZoom,
    grid: grids,
    xAxis: xAxes,
    yAxis: yAxes,
    series: allSeries,
  };

  return { option, layouts, totalHeight };
}
