import { useEffect, useCallback, useMemo } from 'react';
import {
  Flex,
  Box,
  VStack,
  Button,
  Text,
  Spinner,
  Center,
  useColorModeValue,
} from '@chakra-ui/react';
import { useProjectStore } from '../../store/projectStore';
import { useUIStore } from '../../store/uiStore';
import { usePlaybackStore } from '../../store/playbackStore';
import { sliceRenderData } from '../../utils/sliceRenderData';
import PairView from '../layout/PairView';
import SettingsBar from '../layout/SettingsBar';
import PlaybackBar from '../layout/PlaybackBar';

interface ProjectViewProps {
  projectId: string;
}

/**
 * ProjectView – renders the settings bar, the inner pair tabs (vertical
 * sidebar) and the PairView for the currently selected pair.
 *
 * Active pair selection is frontend-only state (uiStore), cached per pair so
 * switching between already-fetched pairs is instant.  Backend-owned settings
 * (timeFrame, timeRange, candleColors) are updated via PATCH /settings which
 * invalidates all per-pair render caches for the project and triggers a
 * fresh render for the current pair.
 */
export default function ProjectView({ projectId }: ProjectViewProps) {
  const loadProject = useProjectStore((s) => s.loadProject);
  const loadRender = useProjectStore((s) => s.loadRender);
  const updateSettings = useProjectStore((s) => s.updateSettings);
  const projectDetail = useProjectStore((s) => s.projectDetails[projectId]);

  // Active pair lives in uiStore (frontend session state, not persisted).
  const activePair = useUIStore((s) => s.activePairs[projectId]);
  const setActivePair = useUIStore((s) => s.setActivePair);

  // Render cache subscriptions are keyed by project + pair.
  const renderCacheKey = activePair
    ? `${projectId}:${activePair.sourceName}:${activePair.pairId}`
    : null;
  const renderData = useProjectStore((s) =>
    renderCacheKey ? (s.renderData[renderCacheKey] ?? null) : null
  );
  const loading = useProjectStore((s) =>
    renderCacheKey ? (s.renderLoading[renderCacheKey] ?? false) : false
  );
  const error = useProjectStore((s) =>
    renderCacheKey ? (s.renderErrors[renderCacheKey] ?? null) : null
  );

  // --- Playback store – all state is project-level so pair switches share position ---
  const playbackProjectState = usePlaybackStore((s) => s.projects[projectId]);
  const playbackIsActive = playbackProjectState?.isActive ?? false;
  const playbackStepSize = playbackProjectState?.stepSize ?? 1;
  const playbackPlayInterval = playbackProjectState?.playInterval ?? 1;
  const playbackCurrentIndex = playbackProjectState?.currentIndex ?? 0;
  const playbackIsPlaying = playbackProjectState?.isPlaying ?? false;

  const activateProject = usePlaybackStore((s) => s.activateProject);
  const deactivateProject = usePlaybackStore((s) => s.deactivateProject);
  const setPlaybackCurrentIndex = usePlaybackStore((s) => s.setCurrentIndex);
  const setPlaybackIsPlaying = usePlaybackStore((s) => s.setIsPlaying);
  const setPlaybackStepSize = usePlaybackStore((s) => s.setStepSize);
  const setPlaybackPlayInterval = usePlaybackStore((s) => s.setPlayInterval);

  // Load project detail on mount (guarded by the store).
  useEffect(() => {
    loadProject(projectId);
  }, [projectId, loadProject]);

  // Once we have the project detail and an active pair, load the render.
  // This effect also fires when:
  //  - activePair changes (user selected a different pair) → loadRender is
  //    guarded so it's a no-op if the pair is already cached.
  //  - projectDetail changes (after updateSettings cleared the cache) → loadRender
  //    fetches fresh data because the cache was invalidated.
  useEffect(() => {
    if (!projectDetail?.pairs?.length) return;

    if (!activePair) {
      // Default to the first pair on initial load.
      const first = projectDetail.pairs[0];
      setActivePair(projectId, first.sourceName, first.pairId);
      return; // effect will re-run once activePair is set
    }

    loadRender(projectId, activePair.sourceName, activePair.pairId);
  }, [projectId, projectDetail, activePair, loadRender, setActivePair]);

  // Pair selection is purely frontend – update uiStore, then the effect above
  // will call loadRender (cache hit = instant, cache miss = fetches from backend).
  const handleSelectPair = useCallback(
    (sourceName: string, pairId: string) => {
      setActivePair(projectId, sourceName, pairId);
    },
    [projectId, setActivePair]
  );

  // When a backend-owned setting changes: PATCH, then the store clears all
  // render caches for this project. The effect above re-triggers and fetches
  // fresh render data for the current pair.
  const handleSettingChange = useCallback(
    async (patch: { timeFrame?: string; rangeFrom?: string; rangeTo?: string }) => {
      try {
        await updateSettings(projectId, patch);
      } catch {
        // error already logged in store
      }
    },
    [projectId, updateSettings]
  );

  const handleColorsChange = useCallback(
    async (candleBullishColor: string, candleBearishColor: string) => {
      try {
        await updateSettings(projectId, { candleBullishColor, candleBearishColor });
      } catch {
        // error already logged in store
      }
    },
    [projectId, updateSettings]
  );

  // --- Playback callbacks (stable references to prevent interval restarts) ---
  const handlePlaybackToggle = useCallback(() => {
    activateProject(projectId);
  }, [projectId, activateProject]);

  const handlePlaybackDeactivate = useCallback(() => {
    deactivateProject(projectId);
  }, [projectId, deactivateProject]);

  const handlePlaybackIndexChange = useCallback(
    (index: number) => {
      setPlaybackCurrentIndex(projectId, index);
    },
    [projectId, setPlaybackCurrentIndex]
  );

  const handlePlaybackPlayingChange = useCallback(
    (playing: boolean) => {
      setPlaybackIsPlaying(projectId, playing);
    },
    [projectId, setPlaybackIsPlaying]
  );

  const handlePlaybackStepSizeChange = useCallback(
    (size: number) => {
      setPlaybackStepSize(projectId, size);
    },
    [projectId, setPlaybackStepSize]
  );

  const handlePlaybackPlayIntervalChange = useCallback(
    (interval: number) => {
      setPlaybackPlayInterval(projectId, interval);
    },
    [projectId, setPlaybackPlayInterval]
  );

  // --- Derive playback-sliced render data and candle metadata ---
  const activeSeries = useMemo(() => {
    if (!renderData) return null;
    const ap = renderData.activePair;
    return ap
      ? renderData.candles.find(
          (c) => c.sourceName === ap.sourceName && c.pairId === ap.pairId
        ) ?? null
      : renderData.candles[0] ?? null;
  }, [renderData]);

  const totalCandles = activeSeries?.entries.length ?? 0;

  // Clamp index to valid range (e.g. after a data reload with fewer candles)
  const safeIndex = Math.min(playbackCurrentIndex, Math.max(0, totalCandles - 1));
  const currentTimestamp = activeSeries?.entries[safeIndex]?.timestamp ?? '';

  const displayRenderData = useMemo(() => {
    if (!renderData || !playbackIsActive) return renderData;
    return sliceRenderData(renderData, safeIndex);
  }, [renderData, playbackIsActive, safeIndex]);

  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const activeBg = useColorModeValue('brand.50', 'brand.900');
  const activeColor = useColorModeValue('brand.700', 'brand.300');
  const hoverBg = useColorModeValue('gray.100', 'gray.700');

  if (!projectDetail) {
    return (
      <Center h="200px">
        <Spinner size="lg" color="brand.400" />
      </Center>
    );
  }

  return (
    <Flex direction="column" gap={0} h="100%">
      {/* Settings bar */}
      <SettingsBar
        projectId={projectId}
        availableTimeFrames={projectDetail.availableTimeFrames ?? []}
        timeFrame={projectDetail.timeFrame ?? ''}
        rangeFrom={projectDetail.timeRange?.start ?? ''}
        rangeTo={projectDetail.timeRange?.end ?? ''}
        candleColors={projectDetail.candleColors}
        onTimeFrameChange={(tf) => handleSettingChange({ timeFrame: tf })}
        onRangeFromChange={(v) => handleSettingChange({ rangeFrom: v })}
        onRangeToChange={(v) => handleSettingChange({ rangeTo: v })}
        onColorsChange={handleColorsChange}
        playbackActive={playbackIsActive}
        onPlaybackToggle={handlePlaybackToggle}
      />

      {/* Playback control bar – visible only when Playback mode is active */}
      {playbackIsActive && (
        <PlaybackBar
          totalCandles={totalCandles}
          currentIndex={safeIndex}
          stepSize={playbackStepSize}
          playInterval={playbackPlayInterval}
          isPlaying={playbackIsPlaying}
          currentTimestamp={currentTimestamp}
          onDeactivate={handlePlaybackDeactivate}
          onIndexChange={handlePlaybackIndexChange}
          onStepSizeChange={handlePlaybackStepSizeChange}
          onPlayIntervalChange={handlePlaybackPlayIntervalChange}
          onPlayingChange={handlePlaybackPlayingChange}
        />
      )}

      {/* Pair selector + chart */}
      <Flex gap={0} flex={1} minH={0}>
        {/* Left pair sidebar */}
        <VStack
          spacing={1}
          align="stretch"
          minW="160px"
          maxW="200px"
          pr={3}
          mr={3}
          borderRight="1px solid"
          borderColor={borderColor}
          py={2}
          flexShrink={0}
        >
          {projectDetail.pairs.map((pair) => {
            const isActive =
              activePair?.sourceName === pair.sourceName &&
              activePair?.pairId === pair.pairId;
            return (
              <Button
                key={`${pair.sourceName}-${pair.pairId}`}
                variant="ghost"
                textAlign="left"
                px={3}
                py={2}
                h="auto"
                borderRadius="md"
                bg={isActive ? activeBg : 'transparent'}
                color={isActive ? activeColor : undefined}
                _hover={{ bg: isActive ? activeBg : hoverBg }}
                onClick={() => handleSelectPair(pair.sourceName, pair.pairId)}
                display="block"
                whiteSpace="normal"
                fontWeight={isActive ? 'semibold' : 'normal'}
              >
                <Text fontSize="sm" fontWeight="inherit" lineHeight={1.3}>
                  {pair.pairId}
                </Text>
                <Text
                  fontSize="xs"
                  color={isActive ? activeColor : 'gray.500'}
                  lineHeight={1.3}
                >
                  {pair.sourceName}
                </Text>
              </Button>
            );
          })}
        </VStack>

        {/* Main content */}
        <Box flex={1} minW={0} py={2} h="100%" overflow="hidden">
          <PairView
            renderData={displayRenderData}
            loading={loading}
            error={error}
            projectId={projectId}
          />
        </Box>
      </Flex>
    </Flex>
  );
}
