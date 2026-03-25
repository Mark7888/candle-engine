import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender, patchProjectSettings } from '../api/client';

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

  loadRender: async (projectId, source, pair) => {
    const key = `${projectId}-${source}-${pair}`;
    // Skip if data is already cached or a fetch is already in progress.
    if (get().renderData[key] || get().renderLoading[key]) return;
    set((s) => ({
      renderLoading: { ...s.renderLoading, [key]: true },
      renderErrors: { ...s.renderErrors, [key]: null },
    }));
    try {
      const data = await fetchRender(projectId, source, pair);
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

  getRender: (projectId, source, pair) => {
    const key = `${projectId}-${source}-${pair}`;
    return {
      data: get().renderData[key] ?? null,
      loading: get().renderLoading[key] ?? false,
      error: get().renderErrors[key] ?? null,
    };
  },

  // Clear all render cache entries for a given project so the next loadRender
  // call re-fetches with any updated settings.
  clearProjectRender: (projectId) => {
    set((s) => {
      const prefix = `${projectId}-`;
      const filterKeys = (obj) =>
        Object.fromEntries(Object.entries(obj).filter(([k]) => !k.startsWith(prefix)));
      return {
        renderData: filterKeys(s.renderData),
        renderLoading: filterKeys(s.renderLoading),
        renderErrors: filterKeys(s.renderErrors),
      };
    });
  },

  // PATCH project settings on the backend, update the cached project detail,
  // and clear the render cache so the next render re-fetches with new settings.
  updateProjectSettings: async (projectId, settings) => {
    try {
      const updated = await patchProjectSettings(projectId, settings);
      set((s) => ({
        projectDetails: { ...s.projectDetails, [projectId]: updated },
      }));
      get().clearProjectRender(projectId);
    } catch (err) {
      console.error('Failed to update project settings', projectId, err);
      throw err;
    }
  },
}));
