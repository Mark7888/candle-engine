import { useRef, useEffect } from 'react';
import ReactECharts from 'echarts-for-react';
import {
  Box,
  Flex,
  Text,
  Spinner,
  Center,
  useColorMode,
} from '@chakra-ui/react';
import { buildEchartsOption } from './chartEngine';

/**
 * ChartContainer renders the full multi-grid ECharts chart from API render data.
 * Handles:
 *  - Building the ECharts option from raw API data via chartEngine
 *  - Responsive resizing via ResizeObserver (React-idiomatic ref pattern)
 *  - Dark / light mode theming
 */
export default function ChartContainer({ renderData, loading, error }) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';
  const echartsRef = useRef(null);
  const containerRef = useRef(null);

  const option = buildEchartsOption(renderData, isDark);
  const totalHeight = option._totalHeight ?? 600;

  // Resize chart when container width changes
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new ResizeObserver(() => {
      echartsRef.current?.getEchartsInstance()?.resize();
    });
    observer.observe(container);
    return () => observer.disconnect();
  }, []);

  if (loading) {
    return (
      <Center h="400px">
        <Spinner size="xl" color="brand.400" />
      </Center>
    );
  }

  if (error) {
    return (
      <Center h="200px" flexDir="column" gap={2}>
        <Text color="red.400" fontWeight="semibold">
          Failed to load chart data
        </Text>
        <Text fontSize="sm" color="gray.500">
          {error}
        </Text>
      </Center>
    );
  }

  if (!renderData) {
    return (
      <Center h="200px">
        <Text color="gray.500">No chart data</Text>
      </Center>
    );
  }

  return (
    <Flex direction="column" gap={2} ref={containerRef} w="100%">
      {/* Chart */}
      <Box w="100%">
        <ReactECharts
          ref={echartsRef}
          option={option}
          style={{ width: '100%', height: `${totalHeight}px` }}
          notMerge
          lazyUpdate={false}
          theme={isDark ? 'dark' : undefined}
        />
      </Box>
    </Flex>
  );
}
