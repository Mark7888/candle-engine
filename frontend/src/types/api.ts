// ---------------------------------------------------------------------------
// Shared API response types for the candle-engine backend
// ---------------------------------------------------------------------------

export interface Project {
  id: string;
  title: string;
}

export interface PairInfo {
  sourceName: string;
  pairId: string;
}

export interface TimeRange {
  start: string;
  end: string;
}

export interface ProjectDetail {
  id: string;
  title: string;
  timeFrame: string;
  timeRange: TimeRange;
  availableTimeFrames: string[];
  pairs: PairInfo[];
  candleColors: CandleColors;
}

export interface CandleEntry {
  timestamp: string;
  open: number;
  close: number;
  high: number;
  low: number;
  invalid: boolean;
}

export interface CandleSeries {
  sourceName: string;
  pairId: string;
  entries: CandleEntry[];
}

export interface CandleColors {
  bullishColor: string;
  bearishColor: string;
}

export interface PlotPoint {
  value: number;
  color?: string;
  invalid: boolean;
}

export interface PlotPointCollection {
  points: PlotPoint[];
  invalid: boolean;
}

// Indicator configs --------------------------------------------------------

export interface LineIndicatorConfig {
  lineTitles?: string[];
  isSmooth?: boolean;
  showSymbols?: boolean;
  lineThickness?: number;
  opacity?: number;
}

export interface BarIndicatorConfig {
  barTitles?: string[];
  barWidth?: number;
  stacked?: boolean;
  opacity?: number;
}

export interface StepLineIndicatorConfig extends LineIndicatorConfig {
  stepAlignment?: number;
}

export interface CandleStickIndicatorConfig {
  bullishColor?: string;
  bearishColor?: string;
}

export interface TableIndicatorConfig {
  displayName?: string;
  rowTitles?: string[];
  columnTitles?: string[];
  valueDisplayFormat?: string;
}

export interface PieChartIndicatorConfig {
  displayName?: string;
  sliceTitles?: string[];
}

export type IndicatorConfig =
  | LineIndicatorConfig
  | BarIndicatorConfig
  | StepLineIndicatorConfig
  | CandleStickIndicatorConfig
  | TableIndicatorConfig
  | PieChartIndicatorConfig
  | Record<string, unknown>;

export type IndicatorType =
  | 'line'
  | 'stepLine'
  | 'bar'
  | 'area'
  | 'candleStick'
  | 'table'
  | 'pieChart';

export interface Indicator {
  name: string;
  type: IndicatorType;
  targetChart?: string;
  plotPoints: PlotPointCollection[];
  config?: IndicatorConfig;
}

export interface LogEntry {
  time?: string;
  level: string;
  message: string;
  indicator?: string;
}

export interface ActivePair {
  sourceName: string;
  pairId: string;
}

export interface RenderResponse {
  candles: CandleSeries[];
  candleColors: CandleColors;
  indicators: Indicator[];
  logs: LogEntry[];
  activePair?: ActivePair;
  activeTimeFrame?: string;
}

export interface UpdateSettingsRequest {
  timeFrame?: string;
  rangeFrom?: string;
  rangeTo?: string;
  candleBullishColor?: string;
  candleBearishColor?: string;
}
