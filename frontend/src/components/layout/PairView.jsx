import { useCallback } from 'react';
import { Flex, Box } from '@chakra-ui/react';
import ChartContainer from '../chart/ChartContainer';
import LogPanel from '../logs/LogPanel';
import { useUIStore } from '../../store/uiStore';

/**
 * PairView – the main content area for a selected trading pair.
 *
 * Layout: [2/3 chart | 1/3 log panel]
 * When the log panel is closed, the chart expands to full width.
 *
 * Scrolling behaviour:
 *  - The whole page scrolls when the chart exceeds the viewport.
 *  - The log panel scrolls its content internally when taller than the viewport.
 */
export default function PairView({ renderData, loading, error, pairKey }) {
  // Log panel open/closed state is persisted in uiStore and synced across tabs.
  const logsOpen = useUIStore((s) => s.logsPanelOpen[pairKey] ?? true);
  const setLogsPanelOpen = useUIStore((s) => s.setLogsPanelOpen);

  const closePanel = useCallback(
    () => setLogsPanelOpen(pairKey, false),
    [pairKey, setLogsPanelOpen]
  );
  const openPanel = useCallback(
    () => setLogsPanelOpen(pairKey, true),
    [pairKey, setLogsPanelOpen]
  );

  const logs = renderData?.logs ?? [];

  return (
    <Flex
      direction={{ base: 'column', lg: 'row' }}
      gap={3}
      alignItems="flex-start"
      w="100%"
    >
      {/* Chart area – grows to fill available space */}
      <Box
        flex={logsOpen ? '2' : '1'}
        minW={0}
        transition="flex 0.3s ease"
      >
        <ChartContainer renderData={renderData} loading={loading} error={error} />
      </Box>

      {/* Log panel */}
      <Box
        flex={logsOpen ? '1' : '0 0 auto'}
        transition="flex 0.3s ease"
        alignSelf="stretch"
        minH="calc(100vh - 160px)"
        display="flex"
      >
        <LogPanel
          logs={logs}
          isOpen={logsOpen}
          onClose={closePanel}
          onOpen={openPanel}
        />
      </Box>
    </Flex>
  );
}
