import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { DEFAULT_PROJECT_SETTINGS, type ProjectSettingsConfig } from './defaultSettings';

interface ProjectSettingsStoreState {
  /** Saved settings keyed by project ID. */
  settings: Record<number, ProjectSettingsConfig>;

  /**
   * Returns the saved settings for the given project, falling back to the
   * defaults when no record has been saved yet.
   */
  getSettings: (projectId: number) => ProjectSettingsConfig;

  /** Persists the complete settings object for the given project. */
  saveSettings: (projectId: number, config: ProjectSettingsConfig) => void;
}

export const useProjectSettingsStore = create<ProjectSettingsStoreState>()(
  persist(
    (set, get) => ({
      settings: {},

      getSettings: (projectId: number) =>
        get().settings[projectId] ?? DEFAULT_PROJECT_SETTINGS,

      saveSettings: (projectId: number, config: ProjectSettingsConfig) =>
        set((s) => ({
          settings: { ...s.settings, [projectId]: config },
        })),
    }),
    {
      name: 'candle-engine-project-settings',
    }
  )
);
