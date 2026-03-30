import { create } from 'zustand';

// ---------------------------------------------------------------------------
// Playback Store
//
// Playback state is NOT persisted – it is always cleared on page reload.
//
// Scoping:
//   • All state is project-level, including currentIndex and isPlaying.
//     Switching pairs within a project seamlessly shares the same position
//     and play state, so the playback continues without interruption.
//     Key format: projectId (number)
// ---------------------------------------------------------------------------

interface PlaybackProjectState {
  isActive: boolean;
  /** N – candles to skip per step */
  stepSize: number;
  /** M – seconds between auto-play steps */
  playInterval: number;
  /** 0-based index into the loaded candle array, shared across all pairs */
  currentIndex: number;
  /** Whether auto-play is running */
  isPlaying: boolean;
}

interface PlaybackStoreState {
  // Per-project settings
  projects: Record<number, PlaybackProjectState>;

  // Project-level actions
  activateProject: (projectId: number) => void;
  deactivateProject: (projectId: number) => void;
  setStepSize: (projectId: number, size: number) => void;
  setPlayInterval: (projectId: number, interval: number) => void;
  setCurrentIndex: (projectId: number, index: number) => void;
  setIsPlaying: (projectId: number, playing: boolean) => void;
}

const DEFAULT_PROJECT_STATE: PlaybackProjectState = {
  isActive: false,
  stepSize: 1,
  playInterval: 1,
  currentIndex: 0,
  isPlaying: false,
};

export const usePlaybackStore = create<PlaybackStoreState>()((set) => ({
  projects: {},

  activateProject: (projectId) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: { ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE), isActive: true },
      },
    })),

  deactivateProject: (projectId) =>
    set((s) => {
      const { [projectId]: _proj, ...restProjects } = s.projects;
      return { projects: restProjects };
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

  setCurrentIndex: (projectId, index) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: {
          ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE),
          currentIndex: index,
        },
      },
    })),

  setIsPlaying: (projectId, playing) =>
    set((s) => ({
      projects: {
        ...s.projects,
        [projectId]: {
          ...(s.projects[projectId] ?? DEFAULT_PROJECT_STATE),
          isPlaying: playing,
        },
      },
    })),
}));

