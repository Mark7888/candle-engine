import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FloatingWidget {
  id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

interface UIStoreState {
  // Logs panel open/closed state – global, synced across all pair tabs
  logsPanelOpen: boolean;
  setLogsPanelOpen: (value: boolean) => void;

  // Log level filter: levels in this array are HIDDEN (persisted globally)
  hiddenLogLevels: string[];
  toggleLogLevel: (level: string) => void;

  // Active project tab
  activeProjectId: string | null;
  setActiveProjectId: (id: string) => void;

  // Indicator panel open/closed — global, same behaviour as logsPanelOpen
  indicatorPanelOpen: boolean;
  setIndicatorPanelOpen: (value: boolean) => void;

  // Indicator panel height in pixels — remembered across open/close and page reloads
  indicatorPanelHeight: number;
  setIndicatorPanelHeight: (value: number) => void;

  // Floating widgets per project: { [projectId]: [{ id, x, y }] }
  // id is the indicator name. NOT persisted — session-only.
  floatingWidgets: Record<string, FloatingWidget[]>;
  setFloatingWidgets: (projectId: string, widgets: FloatingWidget[]) => void;
}

export const useUIStore = create<UIStoreState>()(
  persist(
    (set) => ({
      // Logs panel open/closed state – global, synced across all pair tabs
      logsPanelOpen: true,
      setLogsPanelOpen: (value: boolean) => set({ logsPanelOpen: value }),

      // Log level filter: levels in this array are HIDDEN (persisted globally)
      hiddenLogLevels: [],
      toggleLogLevel: (level: string) =>
        set((s) => ({
          hiddenLogLevels: s.hiddenLogLevels.includes(level)
            ? s.hiddenLogLevels.filter((l) => l !== level)
            : [...s.hiddenLogLevels, level],
        })),

      // Active project tab
      activeProjectId: null,
      setActiveProjectId: (id: string) => set({ activeProjectId: id }),

      // Indicator panel open/closed — global, same behaviour as logsPanelOpen
      indicatorPanelOpen: true,
      setIndicatorPanelOpen: (value: boolean) => set({ indicatorPanelOpen: value }),

      // Indicator panel height in pixels — remembered across open/close and page reloads
      indicatorPanelHeight: 280,
      setIndicatorPanelHeight: (value: number) => set({ indicatorPanelHeight: value }),

      // Floating widgets per project: { [projectId]: [{ id, x, y }] }
      // id is the indicator name. NOT persisted — session-only.
      floatingWidgets: {},
      setFloatingWidgets: (projectId: string, widgets: FloatingWidget[]) =>
        set((s) => ({
          floatingWidgets: { ...s.floatingWidgets, [projectId]: widgets },
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
