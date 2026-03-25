import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender } from '../api/client';

export const useProjectStore = create((set, get) => ({
  // List of { id, title }
  projects: [],
  projectsLoading: false,
  projectsError: null,

  // Per-project detail cache
  projectDetails: {},

  // Per-pair render cache key: `${projectId}-${source}-${pair}`
  renderData: {},
  renderLoading: {},
  renderErrors: {},

  loadProjects: async () => {
    set({ projectsLoading: true, projectsError: null });
    try {
      const projects = await fetchProjects();
      set({ projects, projectsLoading: false });
    } catch (err) {
      set({ projectsError: err.message, projectsLoading: false });
    }
  },

  loadProject: async (id) => {
    if (get().projectDetails[id]) return;
    try {
      const detail = await fetchProject(id);
      set((s) => ({
        projectDetails: { ...s.projectDetails, [id]: detail },
      }));
    } catch (err) {
      console.error('Failed to load project', id, err);
    }
  },

  loadRender: async (projectId, source, pair, settings = {}) => {
    const { timeFrame = '', timeRangeFrom = '', timeRangeTo = '' } = settings;
    const key = `${projectId}-${source}-${pair}-${timeFrame}-${timeRangeFrom}-${timeRangeTo}`;
    if (get().renderData[key]) return;
    set((s) => ({
      renderLoading: { ...s.renderLoading, [key]: true },
      renderErrors: { ...s.renderErrors, [key]: null },
    }));
    try {
      const data = await fetchRender(projectId, source, pair, settings);
      set((s) => ({
        renderData: { ...s.renderData, [key]: data },
        renderLoading: { ...s.renderLoading, [key]: false },
      }));
    } catch (err) {
      set((s) => ({
        renderErrors: { ...s.renderErrors, [key]: err.message },
        renderLoading: { ...s.renderLoading, [key]: false },
      }));
    }
  },

  getRender: (projectId, source, pair, settings = {}) => {
    const { timeFrame = '', timeRangeFrom = '', timeRangeTo = '' } = settings;
    const key = `${projectId}-${source}-${pair}-${timeFrame}-${timeRangeFrom}-${timeRangeTo}`;
    return {
      data: get().renderData[key] ?? null,
      loading: get().renderLoading[key] ?? false,
      error: get().renderErrors[key] ?? null,
    };
  },
}));
