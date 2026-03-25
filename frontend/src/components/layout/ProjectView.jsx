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
import SettingsBar from './SettingsBar';

/**
 * ProjectView – renders the inner pair tabs (vertical sidebar) and the
 * PairView for the currently selected pair.
 */
export default function ProjectView({ projectId }) {
  const loadProject = useProjectStore((s) => s.loadProject);
  const loadRender = useProjectStore((s) => s.loadRender);
  const projectDetail = useProjectStore((s) => s.projectDetails[projectId]);

  const activePairs = useUIStore((s) => s.activePairs);
  const setActivePair = useUIStore((s) => s.setActivePair);

  const setProjectSettings = useUIStore((s) => s.setProjectSettings);
  const timeFrame = useUIStore((s) => s.projectSettings[projectId]?.timeFrame ?? '');
  const timeRangeFrom = useUIStore((s) => s.projectSettings[projectId]?.timeRangeFrom ?? '');
  const timeRangeTo = useUIStore((s) => s.projectSettings[projectId]?.timeRangeTo ?? '');

  const activePairEntry = activePairs[projectId];

  // Derive the cache key outside hooks so it can be used in selectors below.
  const renderKey = activePairEntry
    ? `${projectId}-${activePairEntry.source}-${activePairEntry.pair}-${timeFrame}-${timeRangeFrom}-${timeRangeTo}`
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

  // Set default pair once project detail arrives
  useEffect(() => {
    if (projectDetail && !activePairEntry && projectDetail.pairs?.length > 0) {
      const first = projectDetail.pairs[0];
      setActivePair(projectId, first.sourceName, first.pairId);
    }
  }, [projectDetail, activePairEntry, projectId, setActivePair]);

  // Load render data whenever active pair or settings change
  useEffect(() => {
    if (activePairEntry) {
      loadRender(projectId, activePairEntry.source, activePairEntry.pair, {
        timeFrame,
        timeRangeFrom,
        timeRangeTo,
      });
    }
  }, [projectId, activePairEntry, timeFrame, timeRangeFrom, timeRangeTo, loadRender]);

  const handleSelectPair = useCallback(
    (sourceName, pairId) => {
      setActivePair(projectId, sourceName, pairId);
    },
    [projectId, setActivePair]
  );

  const handleApplySettings = useCallback(
    (settings) => {
      setProjectSettings(projectId, settings);
    },
    [projectId, setProjectSettings]
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
    <Flex direction="column" gap={0} h="100%">
      {/* Settings bar – above pair selector and chart */}
      <SettingsBar
        settings={{ timeFrame, timeRangeFrom, timeRangeTo }}
        onApply={handleApplySettings}
      />

      {/* Pair sidebar + main content */}
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
