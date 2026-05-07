import axios from 'axios';
import type { Project, ProjectDetail, RenderResponse } from '../types/api';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

export const fetchProjects = (): Promise<Project[]> =>
  api.get<{ projects: Project[] }>('/api/v1/projects').then((r) => r.data.projects);

export const fetchProject = (id: number): Promise<ProjectDetail> =>
  api.get<ProjectDetail>(`/api/v1/projects/${id}`).then((r) => r.data);

interface RenderParams {
  source: string;
  pair: string;
  timeFrame?: string;
  rangeFrom?: string;
  rangeTo?: string;
  candleBullishColor?: string;
  candleBearishColor?: string;
}

/**
 * Per-request render configuration sent to the backend.
 *
 * Required fields:
 *   source    – data source name (e.g. 'Binance')
 *   pair      – trading pair ID (e.g. 'BTCUSDC')
 *
 * Optional overrides (omit to use the project's configured defaults):
 *   timeFrame         – time frame string (e.g. '4h')
 *   rangeFrom         – range start as ISO 8601 / RFC 3339 string
 *   rangeTo           – range end as ISO 8601 / RFC 3339 string
 *   candleBullishColor – bullish (up) candle color override (CSS hex, e.g. '#26a69a')
 *   candleBearishColor – bearish (down) candle color override (CSS hex, e.g. '#ef5350')
 */
export interface RenderConfig {
  source: string;
  pair: string;
  timeFrame?: string | null;
  rangeFrom?: string | null;
  rangeTo?: string | null;
  candleBullishColor?: string | null;
  candleBearishColor?: string | null;
}

/**
 * Fetch rendered chart data for a pair.
 *
 * @param id     - Project ID
 * @param config - Render configuration (source, pair, and optional overrides)
 */
export const fetchRender = (
  id: number,
  config: RenderConfig
): Promise<RenderResponse> => {
  const params: RenderParams = { source: config.source, pair: config.pair };
  if (config.timeFrame) params.timeFrame = config.timeFrame;
  if (config.rangeFrom) params.rangeFrom = config.rangeFrom;
  if (config.rangeTo) params.rangeTo = config.rangeTo;
  if (config.candleBullishColor) params.candleBullishColor = config.candleBullishColor;
  if (config.candleBearishColor) params.candleBearishColor = config.candleBearishColor;
  return api
    .get<RenderResponse>(`/api/v1/projects/${id}/render`, { params })
    .then((r) => r.data);
};

export default api;
