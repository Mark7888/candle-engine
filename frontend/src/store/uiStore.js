import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export const useUIStore = create(
  persist(
    (set) => ({
      // Logs panel open/closed state per pair key "projectId-source-pair"
      logsPanelOpen: {},
      setLogsPanelOpen: (key, value) =>
        set((s) => ({ logsPanelOpen: { ...s.logsPanelOpen, [key]: value } })),

      // Active project tab
      activeProjectId: null,
      setActiveProjectId: (id) => set({ activeProjectId: id }),

      // Active pair per project: { [projectId]: { source, pair } }
      activePairs: {},
      setActivePair: (projectId, source, pair) =>
        set((s) => ({
          activePairs: {
            ...s.activePairs,
            [projectId]: { source, pair },
          },
        })),
    }),
    {
      name: 'candle-engine-ui',
      // Only persist the logs-panel state to localStorage
      partialize: (s) => ({ logsPanelOpen: s.logsPanelOpen }),
    }
  )
);
