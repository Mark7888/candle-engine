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

  /**
   * Clear all cached render data for a project (all pairs).
   * Called when project-level settings (timeFrame/range) change so the next
   * loadRender() call fetches fresh data with the updated parameters.
   */
  clearProjectRenderCache: (projectId) => {
    set((s) => {
      const prefix = `${projectId}-`;
      const newRenderData = Object.fromEntries(
        Object.entries(s.renderData).filter(([k]) => !k.startsWith(prefix))
      );
      const newRenderLoading = Object.fromEntries(
        Object.entries(s.renderLoading).filter(([k]) => !k.startsWith(prefix))
      );
      const newRenderErrors = Object.fromEntries(
        Object.entries(s.renderErrors).filter(([k]) => !k.startsWith(prefix))
      );
      return {
        renderData: newRenderData,
        renderLoading: newRenderLoading,
        renderErrors: newRenderErrors,
      };
    });
  },

  /**
   * Load (or re-fetch) render data for a pair.
   *
   * @param {number} projectId
   * @param {string} source
   * @param {string} pair
   * @param {object} [settings]  - Optional overrides forwarded to the API
   *   { timeFrame, rangeFrom, rangeTo }
   */
  loadRender: async (projectId, source, pair, settings = {}) => {
    const key = `${projectId}-${source}-${pair}`;
    // Skip if already cached and no settings override is active
    const hasSettings = Object.values(settings).some((v) => v !== undefined && v !== null && v !== '');
    if (!hasSettings && get().renderData[key]) return;

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

  getRender: (projectId, source, pair) => {
    const key = `${projectId}-${source}-${pair}`;
    return {
      data: get().renderData[key] ?? null,
      loading: get().renderLoading[key] ?? false,
      error: get().renderErrors[key] ?? null,
    };
  },
}));
