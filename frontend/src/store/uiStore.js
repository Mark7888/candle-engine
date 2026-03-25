import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export const useUIStore = create(
  persist(
    (set) => ({
      // Logs panel open/closed state – global, synced across all pair tabs
      logsPanelOpen: true,
      setLogsPanelOpen: (value) => set({ logsPanelOpen: value }),

      // Log level filter: levels in this array are HIDDEN (persisted globally)
      hiddenLogLevels: [],
      toggleLogLevel: (level) =>
        set((s) => ({
          hiddenLogLevels: s.hiddenLogLevels.includes(level)
            ? s.hiddenLogLevels.filter((l) => l !== level)
            : [...s.hiddenLogLevels, level],
        })),

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

      // Indicator panel open/closed — global, same behaviour as logsPanelOpen
      indicatorPanelOpen: true,
      setIndicatorPanelOpen: (value) => set({ indicatorPanelOpen: value }),

      // Indicator panel height in pixels — remembered across open/close and page reloads
      indicatorPanelHeight: 280,
      setIndicatorPanelHeight: (value) => set({ indicatorPanelHeight: value }),

      // Floating widgets per project: { [projectId]: [{ id, x, y }] }
      // id is the indicator name. NOT persisted — session-only.
      floatingWidgets: {},
      setFloatingWidgets: (projectId, widgets) =>
        set((s) => ({
          floatingWidgets: { ...s.floatingWidgets, [projectId]: widgets },
        })),

      // Chart settings per project: { [projectId]: { timeFrame, timeRangeFrom, timeRangeTo } }
      // NOT persisted — session-only.
      projectSettings: {},
      setProjectSettings: (projectId, settings) =>
        set((s) => ({
          projectSettings: { ...s.projectSettings, [projectId]: settings },
        })),
    }),
    {
      name: 'candle-engine-ui',
      // Persist log panel open state and log level filters across page reloads.
      partialize: (s) => ({
        logsPanelOpen: s.logsPanelOpen,
        hiddenLogLevels: s.hiddenLogLevels,
        indicatorPanelOpen: s.indicatorPanelOpen,
        indicatorPanelHeight: s.indicatorPanelHeight,
      }),
    }
  )
);

// Sync persisted state across browser tabs: when another tab writes to
// localStorage, rehydrate this store so both tabs stay in sync immediately.
if (typeof window !== 'undefined') {
  window.addEventListener('storage', (e) => {
    if (e.key === 'candle-engine-ui') {
      useUIStore.persist.rehydrate();
    }
  });
}
