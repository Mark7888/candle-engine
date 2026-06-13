import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FloatingWidget {
  id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface ActivePairEntry {
  sourceName: string;
  pairId: string;
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

  // Active pair per project – frontend session state, NOT persisted.
  // Keyed by projectId; the render cache is also keyed per-pair so switching
  // between already-fetched pairs is instant.
  activePairs: Record<string, ActivePairEntry>;
  setActivePair: (projectId: string, sourceName: string, pairId: string) => void;

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

      // Active pair per project – session-only, not persisted
      activePairs: {},
      setActivePair: (projectId: string, sourceName: string, pairId: string) =>
        set((s) => ({
          activePairs: { ...s.activePairs, [projectId]: { sourceName, pairId } },
        })),

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
      // Persist only the frontend-specific UI state. Active pairs are session-only.
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
