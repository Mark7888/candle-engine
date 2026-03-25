import axios from 'axios';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

export const fetchProjects = () =>
  api.get('/api/v1/projects').then((r) => r.data.projects);

export const fetchProject = (id) =>
  api.get(`/api/v1/projects/${id}`).then((r) => r.data);

export const fetchRender = (id, source, pair, settings = {}) => {
  const { timeFrame, timeRangeFrom, timeRangeTo } = settings;
  const params = { source, pair };
  if (timeFrame) params.timeFrame = timeFrame;
  if (timeRangeFrom) params.timeRangeFrom = timeRangeFrom;
  if (timeRangeTo) params.timeRangeTo = timeRangeTo;
  return api.get(`/api/v1/projects/${id}/render`, { params }).then((r) => r.data);
};

export default api;
