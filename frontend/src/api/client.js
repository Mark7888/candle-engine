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

export const fetchRender = (id, source, pair) =>
  api
    .get(`/api/v1/projects/${id}/render`, { params: { source, pair } })
    .then((r) => r.data);

export default api;
