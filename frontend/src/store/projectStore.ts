import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender, updateProjectSettings } from '../api/client';
import type { UpdateSettingsRequest } from '../types/api';
import type { Project, ProjectDetail, RenderResponse } from '../types/api';

// Render cache key: identifies the exact data slice (project + pair).
const renderKey = (projectId: string, sourceName: string, pairId: string) =>
  `${projectId}:${sourceName}:${pairId}`;

interface ProjectStoreState {
  // List of { id, title }
  projects: Project[];
  projectsLoading: boolean;
  projectsError: string | null;

  // Per-project detail cache (authoritative state from backend)
  projectDetails: Record<string, ProjectDetail>;

  // Per-project-per-pair render data cache
  // Key: renderKey(projectId, sourceName, pairId)
  renderData: Record<string, RenderResponse>;
  renderLoading: Record<string, boolean>;
  renderErrors: Record<string, string | null>;

  loadProjects: () => Promise<void>;
  loadProject: (id: string) => Promise<void>;
  loadRender: (projectId: string, sourceName: string, pairId: string) => Promise<void>;
  updateSettings: (projectId: string, settings: UpdateSettingsRequest) => Promise<void>;
}

export const useProjectStore = create<ProjectStoreState>()((set, get) => ({
  // List of { id, title }
  projects: [],
  projectsLoading: false,
  projectsError: null,

  // Per-project detail cache
  projectDetails: {},

  // Per-project-per-pair render cache
  renderData: {},
  renderLoading: {},
  renderErrors: {},

  loadProjects: async () => {
    // Guard: skip if a fetch is already in flight or data is already loaded.
    if (get().projectsLoading || get().projects.length > 0) return;
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
   * Load render data for the given project + pair.
   * Skips if already cached or a fetch is already in progress.
   * The cache is cleared per-project by updateSettings, so calling this again
   * after a settings change will always fetch fresh data for that pair.
   */
  loadRender: async (projectId: string, sourceName: string, pairId: string) => {
    const key = renderKey(projectId, sourceName, pairId);

    // Atomically check-and-set the loading flag to prevent concurrent duplicate fetches.
    let shouldFetch = false;
    set((s) => {
      if (s.renderData[key] || s.renderLoading[key]) {
        return s; // already cached or in flight — no change
      }
      shouldFetch = true;
      return {
        ...s,
        renderLoading: { ...s.renderLoading, [key]: true },
        renderErrors: { ...s.renderErrors, [key]: null },
      };
    });
    if (!shouldFetch) return;

    try {
      const data = await fetchRender(projectId, sourceName, pairId);
      set((s) => ({
        renderData: { ...s.renderData, [key]: data },
        renderLoading: { ...s.renderLoading, [key]: false },
      }));
    } catch (err) {
      set((s) => ({
        renderErrors: { ...s.renderErrors, [key]: (err as Error).message },
        renderLoading: { ...s.renderLoading, [key]: false },
      }));
    }
  },

  /**
   * Update backend settings for a project (timeFrame, timeRange, candleColors).
   * Clears ALL render caches for this project since every pair's render will
   * be stale after a settings change.  The updated project detail is stored so
   * the UI reflects the new state immediately, and the effect in ProjectView
   * will re-trigger loadRender() for the currently active pair.
   */
  updateSettings: async (projectId: string, settings: UpdateSettingsRequest) => {
    try {
      const updatedDetail = await updateProjectSettings(projectId, settings);
      set((s) => {
        // Remove all render cache entries for this project (all pairs).
        const prefix = `${projectId}:`;
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
