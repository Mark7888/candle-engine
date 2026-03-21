import axios from 'axios';
import type {
  ProjectListResponse,
  ProjectDetailResponse,
  ProjectRenderResponse,
  HealthResponse,
  RenderParams,
} from '../types/api';

const client = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
});

export const api = {
  health: (): Promise<HealthResponse> =>
    client.get<HealthResponse>('/health').then((r) => r.data),

  listProjects: (): Promise<ProjectListResponse> =>
    client.get<ProjectListResponse>('/projects').then((r) => r.data),

  getProject: (id: number): Promise<ProjectDetailResponse> =>
    client.get<ProjectDetailResponse>(`/projects/${id}`).then((r) => r.data),

  renderProject: (id: number, params?: RenderParams): Promise<ProjectRenderResponse> =>
    client
      .get<ProjectRenderResponse>(`/projects/${id}/render`, { params })
      .then((r) => r.data),
};
