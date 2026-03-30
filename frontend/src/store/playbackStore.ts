import { create } from 'zustand';

// ---------------------------------------------------------------------------
// Playback Store
//
// Playback state is NOT persisted – it is always cleared on page reload.
//
// Scoping:
//   • Project-level: isActive, stepSize, playInterval
//     These are shared across all pairs within a project so that switching
//     pairs keeps the settings intact ("handle pair changes properly").
//   • Pair-level: currentIndex, isPlaying
//     Each pair remembers its own position and play state independently.
//     Key format: `${projectId}-${source}-${pair}`
// ---------------------------------------------------------------------------

interface PlaybackProjectState {
  isActive: boolean;
  /** N – candles to skip per step */
  stepSize: number;
  /** M – seconds between auto-play steps */
  playInterval: number;
}

interface PlaybackPairState {
  /** 0-based index into the loaded candle array */
  currentIndex: number;
  isPlaying: boolean;
}

interface PlaybackStoreState {
  // Per-project settings
  projects: Record<number, PlaybackProjectState>;
  // Per-pair state, key = `${projectId}-${source}-${pair}`
  pairs: Record<string, PlaybackPairState>;

  // Project-level actions
  activateProject: (projectId: number) => void;
  deactivateProject: (projectId: number) => void;
  setStepSize: (projectId: number, size: number) => void;
  setPlayInterval: (projectId: number, interval: number) => void;

  // Pair-level actions
  setCurrentIndex: (pairKey: string, index: number) => void;
  setIsPlaying: (pairKey: string, playing: boolean) => void;
}

const DEFAULT_PROJECT_STATE: PlaybackProjectState = {
  isActive: false,
  stepSize: 1,
  playInterval: 1,
};

const DEFAULT_PAIR_STATE: PlaybackPairState = {
  currentIndex: 0,
  isPlaying: false,
};

export function buildPairKey(projectId: number, source: string, pair: string): string {
  return `${projectId}-${source}-${pair}`;
}

export const usePlaybackStore = create<PlaybackStoreState>()((set, get) => ({
  projects: {},
  pairs: {},

  activateProject: (projectId) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: { ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE), isActive: true },
      },
    })),

  deactivateProject: (projectId) =>
    set((s) => {
      // Remove the project entry
      const { [projectId]: _proj, ...restProjects } = s.projects;

      // Remove all pair entries that belong to this project
      const prefix = `${projectId}-`;
      const restPairs = Object.fromEntries(
        Object.entries(s.pairs).filter(([k]) => !k.startsWith(prefix))
      );

      return { projects: restProjects, pairs: restPairs };
    }),

  setStepSize: (projectId, size) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: {
          ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE),
          stepSize: size,
        },
      },
    })),

  setPlayInterval: (projectId, interval) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: {
          ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE),
          playInterval: interval,
        },
      },
    })),

  setCurrentIndex: (pairKey, index) =>
    set((s) => ({
      pairs: {
        ...s.pairs,
        [pairKey]: { ...(s.pairs[pairKey] ?? DEFAULT_PAIR_STATE), currentIndex: index },
      },
    })),

  setIsPlaying: (pairKey, playing) => {
    // Stop playing for all pairs in the same project when a new one starts,
    // so only one pair auto-plays at a time.
    const existing = get().pairs;
    set((s) => {
      const updatedPairs = { ...existing };
      if (playing) {
        // Pause any currently playing pair
        for (const key of Object.keys(updatedPairs)) {
          if (updatedPairs[key].isPlaying) {
            updatedPairs[key] = { ...updatedPairs[key], isPlaying: false };
          }
        }
      }
      updatedPairs[pairKey] = {
        ...(s.pairs[pairKey] ?? DEFAULT_PAIR_STATE),
        isPlaying: playing,
      };
      return { pairs: updatedPairs };
    });
  },
}));
