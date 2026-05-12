import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender, updateProjectSettings } from '../api/client';
import type { UpdateSettingsRequest } from '../types/api';
import type { Project, ProjectDetail, RenderResponse } from '../types/api';

interface ProjectStoreState {
  // List of { id, title }
  projects: Project[];
  projectsLoading: boolean;
  projectsError: string | null;

  // Per-project detail cache
  projectDetails: Record<string, ProjectDetail>;

  // Per-project render cache key: projectId
  renderData: Record<string, RenderResponse>;
  renderLoading: Record<string, boolean>;
  renderErrors: Record<string, string | null>;

  loadProjects: () => Promise<void>;
  loadProject: (id: string) => Promise<void>;
  reloadProject: (id: string) => Promise<void>;
  clearProjectRenderCache: (projectId: string) => void;
  loadRender: (projectId: string) => Promise<void>;
  getRender: (projectId: string) => { data: RenderResponse | null; loading: boolean; error: string | null };
  updateSettings: (projectId: string, settings: UpdateSettingsRequest) => Promise<void>;
}

export const useProjectStore = create<ProjectStoreState>()((set, get) => ({
  // List of { id, title }
  projects: [],
  projectsLoading: false,
  projectsError: null,

  // Per-project detail cache
  projectDetails: {},

  // Per-project render cache key: projectId
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

  /** Re-fetch project detail unconditionally (e.g. after settings change). */
  reloadProject: async (id: string) => {
    try {
      const detail = await fetchProject(id);
      set((s) => ({
        projectDetails: { ...s.projectDetails, [id]: detail },
      }));
    } catch (err) {
      console.error('Failed to reload project', id, err);
    }
  },

  /**
   * Clear cached render data for a project.
   * Called when settings change so the next loadRender() fetches fresh data.
   */
  clearProjectRenderCache: (projectId: string) => {
    set((s) => {
      const { [projectId]: _data, ...newRenderData } = s.renderData;
      const { [projectId]: _loading, ...newRenderLoading } = s.renderLoading;
      const { [projectId]: _error, ...newRenderErrors } = s.renderErrors;
      return {
        renderData: newRenderData,
        renderLoading: newRenderLoading,
        renderErrors: newRenderErrors,
      };
    });
  },

  /**
   * Load (or re-fetch) render data for a project using its stored backend settings.
   */
  loadRender: async (projectId: string) => {
    if (get().renderData[projectId]) return;

    set((s) => ({
      renderLoading: { ...s.renderLoading, [projectId]: true },
      renderErrors: { ...s.renderErrors, [projectId]: null },
    }));
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

  getRender: (projectId: string) => ({
    data: get().renderData[projectId] ?? null,
    loading: get().renderLoading[projectId] ?? false,
    error: get().renderErrors[projectId] ?? null,
  }),

  /**
   * Update backend settings for a project, then clear the render cache so the
   * next loadRender() fetches fresh data.  The updated project detail is stored
   * so the UI reflects the new state immediately.
   */
  updateSettings: async (projectId: string, settings: UpdateSettingsRequest) => {
    try {
      const updatedDetail = await updateProjectSettings(projectId, settings);
      set((s) => ({
        projectDetails: { ...s.projectDetails, [projectId]: updatedDetail },
      }));
      // Invalidate render cache so the next render reflects the new settings.
      get().clearProjectRenderCache(projectId);
    } catch (err) {
      console.error('Failed to update project settings', projectId, err);
      throw err;
    }
  },
}));
