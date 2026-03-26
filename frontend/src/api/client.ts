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

interface RenderSettings {
  timeFrame?: string | null;
  rangeFrom?: string | null;
  rangeTo?: string | null;
}

interface RenderParams {
  source: string;
  pair: string;
  timeFrame?: string;
  rangeFrom?: string;
  rangeTo?: string;
}

/**
 * Fetch rendered chart data for a pair.
 *
 * @param id        - Project ID
 * @param source    - Data source name (e.g. 'Binance')
 * @param pair      - Trading pair ID (e.g. 'BTCUSDC')
 * @param settings  - Optional overrides: { timeFrame, rangeFrom, rangeTo }
 *   timeFrame  – override time frame string (e.g. '4h')
 *   rangeFrom  – override range start as ISO 8601 string
 *   rangeTo    – override range end as ISO 8601 string
 */
export const fetchRender = (
  id: number,
  source: string,
  pair: string,
  settings: RenderSettings = {}
): Promise<RenderResponse> => {
  const params: RenderParams = { source, pair };
  if (settings.timeFrame) params.timeFrame = settings.timeFrame;
  if (settings.rangeFrom) params.rangeFrom = settings.rangeFrom;
  if (settings.rangeTo) params.rangeTo = settings.rangeTo;
  return api
    .get<RenderResponse>(`/api/v1/projects/${id}/render`, { params })
    .then((r) => r.data);
};

export default api;
