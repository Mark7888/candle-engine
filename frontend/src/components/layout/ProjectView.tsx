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
 * All mutable settings (active pair, time frame, date range, candle colours)
 * are owned by the backend.  The frontend reads them from the project detail
 * response and calls PATCH /settings whenever the user changes them.
 */
export default function ProjectView({ projectId }: ProjectViewProps) {
  const loadProject = useProjectStore((s) => s.loadProject);
  const loadRender = useProjectStore((s) => s.loadRender);
  const updateSettings = useProjectStore((s) => s.updateSettings);
  const projectDetail = useProjectStore((s) => s.projectDetails[projectId]);

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
  // re-renders when loadRender() updates them.
  const renderData = useProjectStore((s) => s.renderData[projectId] ?? null);
  const loading = useProjectStore((s) => s.renderLoading[projectId] ?? false);
  const error = useProjectStore((s) => s.renderErrors[projectId] ?? null);

  // Load project detail on mount
  useEffect(() => {
    loadProject(projectId);
  }, [projectId, loadProject]);

  // Load render data once the project detail is available
  useEffect(() => {
    if (projectDetail) {
      loadRender(projectId);
    }
  }, [projectId, projectDetail, loadRender]);

  const handleSelectPair = useCallback(
    async (sourceName: string, pairId: string) => {
      try {
        await updateSettings(projectId, { activePair: { sourceName, pairId } });
      } catch {
        // error already logged in store
      }
    },
    [projectId, updateSettings]
  );

  // When a time-frame or range setting changes: update backend, then re-render
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

  // When candle colors change (from ProjectSettingsModal via the store)
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

  // Derive display values from the backend-owned project detail.
  const activePairEntry = projectDetail.activePair;

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
              activePairEntry?.sourceName === pair.sourceName &&
              activePairEntry?.pairId === pair.pairId;
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
