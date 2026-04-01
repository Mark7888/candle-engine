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
  projectId: number;
}

/**
 * ProjectView – renders the settings bar, the inner pair tabs (vertical
 * sidebar) and the PairView for the currently selected pair.
 */
export default function ProjectView({ projectId }: ProjectViewProps) {
  const loadProject = useProjectStore((s) => s.loadProject);
  const loadRender = useProjectStore((s) => s.loadRender);
  const clearProjectRenderCache = useProjectStore((s) => s.clearProjectRenderCache);
  const projectDetail = useProjectStore((s) => s.projectDetails[projectId]);

  const activePairs = useUIStore((s) => s.activePairs);
  const setActivePair = useUIStore((s) => s.setActivePair);

  // Select each setting individually so selectors always return a stable
  // primitive (string | null), avoiding the infinite-loop that `?? {}` causes
  // when the project key is not yet present in the store.
  const settingsTimeFrame = useUIStore((s) => s.projectSettings[projectId]?.timeFrame ?? null);
  const settingsRangeFrom = useUIStore((s) => s.projectSettings[projectId]?.rangeFrom ?? null);
  const settingsRangeTo = useUIStore((s) => s.projectSettings[projectId]?.rangeTo ?? null);
  const setProjectSetting = useUIStore((s) => s.setProjectSetting);

  const activePairEntry = activePairs[projectId];

  // Derive the cache key outside hooks so it can be used in selectors below.
  const renderKey = activePairEntry
    ? `${projectId}-${activePairEntry.source}-${activePairEntry.pair}`
    : null;

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

  // Subscribe directly to the data/loading/error slices so the component
  // re-renders when loadRender() updates them (subscribing to getRender — a
  // stable function reference — would never trigger a re-render).
  const renderData = useProjectStore((s) => (renderKey ? (s.renderData[renderKey] ?? null) : null));
  const loading = useProjectStore((s) => (renderKey ? (s.renderLoading[renderKey] ?? false) : false));
  const error = useProjectStore((s) => (renderKey ? (s.renderErrors[renderKey] ?? null) : null));

  // Load project detail on mount
  useEffect(() => {
    loadProject(projectId);
  }, [projectId, loadProject]);

  // Set default pair once project detail arrives
  useEffect(() => {
    if (projectDetail && !activePairEntry && projectDetail.pairs?.length > 0) {
      const first = projectDetail.pairs[0];
      setActivePair(projectId, first.sourceName, first.pairId);
    }
  }, [projectDetail, activePairEntry, projectId, setActivePair]);

  // Load render data whenever active pair or any setting value changes
  useEffect(() => {
    if (activePairEntry) {
      loadRender(projectId, activePairEntry.source, activePairEntry.pair, {
        timeFrame: settingsTimeFrame,
        rangeFrom: settingsRangeFrom,
        rangeTo: settingsRangeTo,
      });
    }
  }, [projectId, activePairEntry, settingsTimeFrame, settingsRangeFrom, settingsRangeTo, loadRender]);

  const handleSelectPair = useCallback(
    (sourceName: string, pairId: string) => {
      setActivePair(projectId, sourceName, pairId);
    },
    [projectId, setActivePair]
  );

  // When a setting changes: clear cached renders so the new request goes out
  const handleSettingChange = useCallback(
    (patch: { timeFrame?: string; rangeFrom?: string; rangeTo?: string }) => {
      clearProjectRenderCache(projectId);
      setProjectSetting(projectId, patch);
    },
    [projectId, clearProjectRenderCache, setProjectSetting]
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
    const { activePair } = renderData;
    return activePair
      ? renderData.candles.find(
          (c) => c.sourceName === activePair.sourceName && c.pairId === activePair.pairId
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

  // Use the first available time frame as the default when none is set yet
  const defaultTimeFrame = projectDetail.timeFrame ?? '';
  const defaultRangeFrom = projectDetail.timeRange?.start ?? '';
  const defaultRangeTo = projectDetail.timeRange?.end ?? '';

  return (
    <Flex direction="column" gap={0} h="100%">
      {/* Settings bar */}
      <SettingsBar
        projectId={projectId}
        availableTimeFrames={projectDetail.availableTimeFrames ?? []}
        defaultTimeFrame={defaultTimeFrame}
        defaultRangeFrom={defaultRangeFrom}
        defaultRangeTo={defaultRangeTo}
        timeFrame={settingsTimeFrame}
        rangeFrom={settingsRangeFrom}
        rangeTo={settingsRangeTo}
        onTimeFrameChange={(tf) => handleSettingChange({ timeFrame: tf })}
        onRangeFromChange={(v) => handleSettingChange({ rangeFrom: v })}
        onRangeToChange={(v) => handleSettingChange({ rangeTo: v })}
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
              activePairEntry?.source === pair.sourceName &&
              activePairEntry?.pair === pair.pairId;
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
