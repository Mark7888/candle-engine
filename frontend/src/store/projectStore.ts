import { create } from 'zustand';
import { fetchProjects, fetchProject, fetchRender } from '../api/client';
import type { Project, ProjectDetail, RenderResponse } from '../types/api';

interface RenderSettings {
  timeFrame?: string | null;
  rangeFrom?: string | null;
  rangeTo?: string | null;
}

interface ProjectStoreState {
  // List of { id, title }
  projects: Project[];
  projectsLoading: boolean;
  projectsError: string | null;

  // Per-project detail cache
  projectDetails: Record<number, ProjectDetail>;

  // Per-pair render cache key: `${projectId}-${source}-${pair}`
  renderData: Record<string, RenderResponse>;
  renderLoading: Record<string, boolean>;
  renderErrors: Record<string, string | null>;

  loadProjects: () => Promise<void>;
  loadProject: (id: number) => Promise<void>;
  clearProjectRenderCache: (projectId: number) => void;
  loadRender: (
    projectId: number,
    source: string,
    pair: string,
    settings?: RenderSettings
  ) => Promise<void>;
  getRender: (
    projectId: number,
    source: string,
    pair: string
  ) => { data: RenderResponse | null; loading: boolean; error: string | null };
}

export const useProjectStore = create<ProjectStoreState>()((set, get) => ({
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
      set({ projectsError: (err as Error).message, projectsLoading: false });
    }
  },

  loadProject: async (id: number) => {
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
  clearProjectRenderCache: (projectId: number) => {
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
   * @param projectId
   * @param source
   * @param pair
   * @param settings  - Optional overrides forwarded to the API
   *   { timeFrame, rangeFrom, rangeTo }
   */
  loadRender: async (
    projectId: number,
    source: string,
    pair: string,
    settings: RenderSettings = {}
  ) => {
    const key = `${projectId}-${source}-${pair}`;
    // Skip if already cached. Cache invalidation when settings change is
    // handled by clearProjectRenderCache, so the cache is always safe to use.
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
        renderErrors: { ...s.renderErrors, [key]: (err as Error).message },
        renderLoading: { ...s.renderLoading, [key]: false },
      }));
    }
  },

  getRender: (projectId: number, source: string, pair: string) => {
    const key = `${projectId}-${source}-${pair}`;
    return {
      data: get().renderData[key] ?? null,
      loading: get().renderLoading[key] ?? false,
      error: get().renderErrors[key] ?? null,
    };
  },
}));
