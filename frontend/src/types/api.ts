// TypeScript types mirroring the Go API models in backend/internal/webserver/api/models/responses.go

export interface ProjectSummary {
  id: number;
  title: string;
}

export interface ProjectListResponse {
  projects: ProjectSummary[];
}

export interface SourcePair {
  sourceName: string;
  pairId: string;
}

export interface TimeRange {
  start: string;
  end: string;
}

export interface ProjectDetailResponse {
  id: number;
  title: string;
  timeFrame: string;
  availableTimeFrames: string[];
  pairs: SourcePair[];
  activePair: SourcePair;
  timeRange: TimeRange;
}

export interface CandleEntry {
  timestamp: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
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
  color: string;
  invalid: boolean;
}

export interface PlotPointCollection {
  points: PlotPoint[];
  invalid: boolean;
}

// Config types per indicator type
export interface LineIndicatorConfig {
  lineThickness: number;
  isSmooth: boolean;
  lineTitles: string[];
  showSymbols: boolean;
  opacity: number;
}

export interface StepLineIndicatorConfig {
  stepAlignment: number; // 0=start, 1=midpoint, 2=end
  lineTitles: string[];
  lineThickness: number;
  opacity: number;
}

export interface BarChartIndicatorConfig {
  barTitles: string[];
  barWidth: string;
  opacity: number;
  stacked: boolean;
}

export interface AreaChartIndicatorConfig {
  lineThickness: number;
  isSmooth: boolean;
  lineTitles: string[];
  showSymbols: boolean;
  opacity: number;
}

export type IndicatorConfig =
  | LineIndicatorConfig
  | StepLineIndicatorConfig
  | BarChartIndicatorConfig
  | AreaChartIndicatorConfig
  | Record<string, unknown>;

export type IndicatorType = 'line' | 'stepLine' | 'bar' | 'area' | 'table';

export interface IndicatorData {
  name: string;
  type: IndicatorType;
  targetChart: string;
  config: IndicatorConfig;
  plotPoints: PlotPointCollection[];
  dataError: boolean;
  errors?: string[];
}

export interface LogEntry {
  indicator: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  time: string;
  message: string;
}

export interface ProjectRenderResponse {
  title: string;
  timeRange: TimeRange;
  availableTimeFrames: string[];
  activeTimeFrame: string;
  activePair: SourcePair;
  candleColors: CandleColors;
  candles: CandleSeries[];
  indicators: IndicatorData[];
  logs: LogEntry[];
}

export interface HealthResponse {
  status: string;
  serverTime: string;
}

export interface ErrorResponse {
  error: string;
  details?: string;
}

export interface RenderParams {
  timeFrame?: string;
  source?: string;
  pair?: string;
}
