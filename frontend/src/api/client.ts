import axios from 'axios';
import type { Project, ProjectDetail, RenderResponse, UpdateSettingsRequest } from '../types/api';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

export const fetchProjects = (): Promise<Project[]> =>
  api.get<{ projects: Project[] }>('/api/v1/projects').then((r) => r.data.projects);

export const fetchProject = (id: string): Promise<ProjectDetail> =>
  api.get<ProjectDetail>(`/api/v1/projects/${id}`).then((r) => r.data);

/**
 * Update session settings for a project (active pair, time frame, date range, candle colours).
 * Returns the updated project detail so the caller can refresh its state in one round trip.
 */
export const updateProjectSettings = (
  id: string,
  settings: UpdateSettingsRequest
): Promise<ProjectDetail> =>
  api
    .patch<ProjectDetail>(`/api/v1/projects/${id}/settings`, settings)
    .then((r) => r.data);

/**
 * Fetch rendered chart data for a specific pair within a project.
 * sourceName and pairId identify the trading pair to render; both are required.
 */
export const fetchRender = (id: string, sourceName: string, pairId: string): Promise<RenderResponse> =>
  api
    .get<RenderResponse>(`/api/v1/projects/${id}/render`, {
      params: { sourceName, pairId },
    })
    .then((r) => r.data);

export default api;
