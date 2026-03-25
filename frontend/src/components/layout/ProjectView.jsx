import { useEffect, useCallback } from 'react';
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
import PairView from '../layout/PairView';
import ProjectSettingsBar from '../layout/ProjectSettingsBar';

/**
 * ProjectView – renders the inner pair tabs (vertical sidebar) and the
 * PairView for the currently selected pair.
 */
export default function ProjectView({ projectId }) {
  const loadProject = useProjectStore((s) => s.loadProject);
  const loadRender = useProjectStore((s) => s.loadRender);
  const updateProjectSettings = useProjectStore((s) => s.updateProjectSettings);
  const projectDetail = useProjectStore((s) => s.projectDetails[projectId]);

  const activePairs = useUIStore((s) => s.activePairs);
  const setActivePair = useUIStore((s) => s.setActivePair);
  const storedSettings = useUIStore((s) => s.projectSettings[projectId]);
  const setProjectSettings = useUIStore((s) => s.setProjectSettings);

  const activePairEntry = activePairs[projectId];

  // Derive the cache key outside hooks so it can be used in selectors below.
  const renderKey = activePairEntry
    ? `${projectId}-${activePairEntry.source}-${activePairEntry.pair}`
    : null;

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

  // Once project detail is loaded, apply any stored settings overrides to the
  // backend so they persist across page reloads.
  useEffect(() => {
    if (!projectDetail || !storedSettings) return;

    const patch = {};
    if (storedSettings.timeFrame && storedSettings.timeFrame !== projectDetail.timeFrame) {
      patch.timeFrame = storedSettings.timeFrame;
    }
    const storedStart = storedSettings.timeRangeStart;
    const storedEnd = storedSettings.timeRangeEnd;
    const defaultStart = projectDetail.timeRange?.start?.slice(0, 10) ?? '';
    const defaultEnd = projectDetail.timeRange?.end?.slice(0, 10) ?? '';
    if (
      (storedStart && storedStart !== defaultStart) ||
      (storedEnd && storedEnd !== defaultEnd)
    ) {
      const resolvedStart = storedStart ?? defaultStart;
      const resolvedEnd = storedEnd ?? defaultEnd;
      const startDate = resolvedStart ? new Date(resolvedStart) : null;
      const endDate = resolvedEnd ? new Date(resolvedEnd) : null;
      if (startDate && !isNaN(startDate.getTime()) && endDate && !isNaN(endDate.getTime())) {
        patch.timeRange = {
          start: startDate.toISOString(),
          end: endDate.toISOString(),
        };
      }
    }

    if (Object.keys(patch).length > 0) {
      updateProjectSettings(projectId, patch).catch(() => {});
    }
    // We intentionally run this only once after projectDetail first loads.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectDetail?.id]);

  // Set default pair once project detail arrives
  useEffect(() => {
    if (projectDetail && !activePairEntry && projectDetail.pairs?.length > 0) {
      const first = projectDetail.pairs[0];
      setActivePair(projectId, first.sourceName, first.pairId);
    }
  }, [projectDetail, activePairEntry, projectId, setActivePair]);

  // Load render data whenever active pair changes
  useEffect(() => {
    if (activePairEntry) {
      loadRender(projectId, activePairEntry.source, activePairEntry.pair);
    }
  }, [projectId, activePairEntry, loadRender]);

  const handleSelectPair = useCallback(
    (sourceName, pairId) => {
      setActivePair(projectId, sourceName, pairId);
    },
    [projectId, setActivePair]
  );

  // Called by ProjectSettingsBar when any setting changes immediately.
  const handleSettingChange = useCallback(
    async (patch) => {
      // Build the patch for uiStore (plain strings / date strings).
      const uiPatch = {};
      if (patch.timeFrame !== undefined) uiPatch.timeFrame = patch.timeFrame;
      if (patch.timeRangeStart !== undefined) uiPatch.timeRangeStart = patch.timeRangeStart;
      if (patch.timeRangeEnd !== undefined) uiPatch.timeRangeEnd = patch.timeRangeEnd;

      // Persist locally first for instant UI feedback.
      setProjectSettings(projectId, uiPatch);

      // Build the API patch (need full ISO timestamps for timeRange).
      const apiPatch = {};
      if (patch.timeFrame !== undefined) {
        apiPatch.timeFrame = patch.timeFrame;
      }
      if (patch.timeRangeStart !== undefined || patch.timeRangeEnd !== undefined) {
        const merged = { ...(storedSettings ?? {}) };
        if (patch.timeRangeStart !== undefined) merged.timeRangeStart = patch.timeRangeStart;
        if (patch.timeRangeEnd !== undefined) merged.timeRangeEnd = patch.timeRangeEnd;

        const startStr = merged.timeRangeStart ?? projectDetail?.timeRange?.start?.slice(0, 10) ?? '';
        const endStr = merged.timeRangeEnd ?? projectDetail?.timeRange?.end?.slice(0, 10) ?? '';

        if (startStr && endStr) {
          const startDate = new Date(startStr);
          const endDate = new Date(endStr);
          if (!isNaN(startDate.getTime()) && !isNaN(endDate.getTime())) {
            apiPatch.timeRange = {
              start: startDate.toISOString(),
              end: endDate.toISOString(),
            };
          }
        }
      }

      try {
        await updateProjectSettings(projectId, apiPatch);
        // Render cache has been cleared by updateProjectSettings; reload now.
        if (activePairEntry) {
          loadRender(projectId, activePairEntry.source, activePairEntry.pair);
        }
      } catch {
        // Errors are logged inside updateProjectSettings; nothing extra to do.
      }
    },
    [projectId, storedSettings, projectDetail, activePairEntry, setProjectSettings, updateProjectSettings, loadRender]
  );

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
    <Flex direction="column" h="100%" gap={0}>
      {/* Settings bar – full width, above pair list and chart */}
      <ProjectSettingsBar
        projectDetail={projectDetail}
        storedSettings={storedSettings}
        onSettingChange={handleSettingChange}
      />

      {/* Pair sidebar + main content */}
      <Flex flex={1} gap={0} minH={0}>
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
          overflowY="auto"
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
            renderData={renderData}
            loading={loading}
            error={error}
            projectId={projectId}
          />
        </Box>
      </Flex>
    </Flex>
  );
}
