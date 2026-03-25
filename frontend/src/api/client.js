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

/**
 * Fetch rendered chart data for a pair.
 *
 * @param {number} id        - Project ID
 * @param {string} source    - Data source name (e.g. 'Binance')
 * @param {string} pair      - Trading pair ID (e.g. 'BTCUSDC')
 * @param {object} [settings] - Optional overrides: { timeFrame, rangeFrom, rangeTo }
 *   timeFrame  – override time frame string (e.g. '4h')
 *   rangeFrom  – override range start as ISO 8601 string
 *   rangeTo    – override range end as ISO 8601 string
 */
export const fetchRender = (id, source, pair, settings = {}) => {
  const params = { source, pair };
  if (settings.timeFrame) params.timeFrame = settings.timeFrame;
  if (settings.rangeFrom) params.rangeFrom = settings.rangeFrom;
  if (settings.rangeTo) params.rangeTo = settings.rangeTo;
  return api
    .get(`/api/v1/projects/${id}/render`, { params })
    .then((r) => r.data);
};

export default api;
