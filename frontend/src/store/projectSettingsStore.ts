import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { DEFAULT_PROJECT_SETTINGS, type ProjectSettingsConfig } from './defaultSettings';

interface ProjectSettingsStoreState {
  /** Saved settings keyed by project ID. */
  settings: Record<string, ProjectSettingsConfig>;

  /**
   * Returns the saved settings for the given project, falling back to the
   * defaults when no record has been saved yet.
   */
  getSettings: (projectId: string) => ProjectSettingsConfig;

  /** Persists the complete settings object for the given project. */
  saveSettings: (projectId: string, config: ProjectSettingsConfig) => void;
}

export const useProjectSettingsStore = create<ProjectSettingsStoreState>()(
  persist(
    (set, get) => ({
      settings: {},

      getSettings: (projectId: string) =>
        get().settings[projectId] ?? DEFAULT_PROJECT_SETTINGS,

      saveSettings: (projectId: string, config: ProjectSettingsConfig) =>
        set((s) => ({
          settings: { ...s.settings, [projectId]: config },
        })),
    }),
    {
      name: 'candle-engine-project-settings',
    }
  )
);
