import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender, updateProjectSettings } from '../api/client';
import type { UpdateSettingsRequest } from '../types/api';
import type { Project, ProjectDetail, RenderResponse } from '../types/api';

interface ProjectStoreState {
  // List of { id, title }
  projects: Project[];
  projectsLoading: boolean;
  projectsError: string | null;

  // Per-project detail cache (authoritative state from backend)
  projectDetails: Record<string, ProjectDetail>;

  // Per-project render data cache
  renderData: Record<string, RenderResponse>;
  renderLoading: Record<string, boolean>;
  renderErrors: Record<string, string | null>;

  loadProjects: () => Promise<void>;
  loadProject: (id: string) => Promise<void>;
  loadRender: (projectId: string) => Promise<void>;
  updateSettings: (projectId: string, settings: UpdateSettingsRequest) => Promise<void>;
}

export const useProjectStore = create<ProjectStoreState>()((set, get) => ({
  // List of { id, title }
  projects: [],
  projectsLoading: false,
  projectsError: null,

  // Per-project detail cache
  projectDetails: {},

  // Per-project render data cache
  renderData: {},
  renderLoading: {},
  renderErrors: {},

  loadProjects: async () => {
    set({ projectsLoading: true, projectsError: null });
    try {
      const projects = await fetchProjects();
      set({ projects, projectsLoading: false });
    } catch (err) {
      set({ projectsError: (err as Error).message, projectsLoading: false });
    }
  },

  loadProject: async (id: string) => {
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
   * Load render data for a project using its current backend settings.
   * Skips if a fetch is already in progress or data is already cached.
   * The render cache is cleared by updateSettings, so calling this again
   * after a settings change will always fetch fresh data.
   */
  loadRender: async (projectId: string) => {
    // Atomically check-and-set the loading flag so concurrent callers
    // (e.g. two React effects firing in the same batch) cannot both pass
    // the guard before either has set loading = true.
    let shouldFetch = false;
    set((s) => {
      if (s.renderData[projectId] || s.renderLoading[projectId]) {
        return s; // already cached or in flight — no change
      }
      shouldFetch = true;
      return {
        ...s,
        renderLoading: { ...s.renderLoading, [projectId]: true },
        renderErrors: { ...s.renderErrors, [projectId]: null },
      };
    });
    if (!shouldFetch) return;
    try {
      const data = await fetchRender(projectId);
      set((s) => ({
        renderData: { ...s.renderData, [projectId]: data },
        renderLoading: { ...s.renderLoading, [projectId]: false },
      }));
    } catch (err) {
      set((s) => ({
        renderErrors: { ...s.renderErrors, [projectId]: (err as Error).message },
        renderLoading: { ...s.renderLoading, [projectId]: false },
      }));
    }
  },

  /**
   * Update backend settings for a project, then clear the render cache so the
   * next loadRender() fetches fresh data.  The updated project detail is stored
   * so the UI reflects the new state immediately.
   */
  updateSettings: async (projectId: string, settings: UpdateSettingsRequest) => {
    try {
      const updatedDetail = await updateProjectSettings(projectId, settings);
      // Clear render cache first, then update project detail.
      // Updating project detail triggers the useEffect in ProjectView which
      // calls loadRender() to fetch fresh render data.
      set((s) => {
        const { [projectId]: _data, ...newRenderData } = s.renderData;
        const { [projectId]: _loading, ...newRenderLoading } = s.renderLoading;
        const { [projectId]: _error, ...newRenderErrors } = s.renderErrors;
        return {
          projectDetails: { ...s.projectDetails, [projectId]: updatedDetail },
          renderData: newRenderData,
          renderLoading: newRenderLoading,
          renderErrors: newRenderErrors,
        };
      });
    } catch (err) {
      console.error('Failed to update project settings', projectId, err);
      throw err;
    }
  },
}));
