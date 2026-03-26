import { useCallback } from 'react';
import { Flex, Box } from '@chakra-ui/react';
import ChartContainer from '../chart/ChartContainer';
import LogPanel from '../logs/LogPanel';
import IndicatorPanel from '../indicators/IndicatorPanel';
import { useUIStore } from '../../store/uiStore';
import type { RenderResponse } from '../../types/api';

interface PairViewProps {
  renderData: RenderResponse | null;
  loading: boolean;
  error: string | null;
  projectId: number;
}

/**
 * PairView – the main content area for a selected trading pair.
 *
 * Layout (column):
 *   Top row: [2/3 chart | 1/3 log panel]
 *   Bottom:  IndicatorPanel (collapsible drawer)
 *
 * Scrolling behaviour:
 *  - The chart area scrolls internally when the chart is taller than the
 *    available viewport height (multiple sub-grids).
 *  - The log panel is bounded to the same height as the chart area; long log
 *    lists scroll inside the panel independently of the page.
 */
export default function PairView({ renderData, loading, error, projectId }: PairViewProps) {
  // Log panel open/closed state is a single global flag synced across all pair tabs.
  const logsOpen = useUIStore((s) => s.logsPanelOpen);
  const setLogsPanelOpen = useUIStore((s) => s.setLogsPanelOpen);

  const closePanel = useCallback(
    () => setLogsPanelOpen(false),
    [setLogsPanelOpen]
  );
  const openPanel = useCallback(
    () => setLogsPanelOpen(true),
    [setLogsPanelOpen]
  );

  const logs = renderData?.logs ?? [];

  return (
    <Flex direction="column" h="100%" w="100%" gap={0}>
      {/* Top row: chart + log panel */}
      <Flex
        direction={{ base: 'column', lg: 'row' }}
        flex={1}
        minH={0}
        gap={3}
        alignItems="stretch"
      >
        {/* Chart area – scrolls internally when chart is taller than viewport */}
        <Box
          flex={logsOpen ? '2' : '1'}
          minW={0}
          overflowY="auto"
          sx={{ scrollbarGutter: 'stable' }}
          transition="flex 0.3s ease"
        >
          <ChartContainer renderData={renderData} loading={loading} error={error} />
        </Box>

        {/* Log panel – bounded to the same height; logs scroll inside the panel */}
        <Box
          flex={logsOpen ? '1' : '0 0 auto'}
          transition="flex 0.3s ease"
          display="flex"
          minW={0}
        >
          <LogPanel
            logs={logs}
            isOpen={logsOpen}
            onClose={closePanel}
            onOpen={openPanel}
          />
        </Box>
      </Flex>

      {/* Bottom: Indicator Drawer */}
      <IndicatorPanel renderData={renderData} projectId={projectId} />
    </Flex>
  );
}
