import { useRef, useEffect, useCallback, useState } from 'react';
import ReactECharts from 'echarts-for-react';
import {
  Box,
  Flex,
  Text,
  Spinner,
  Center,
  useColorMode,
  Wrap,
  WrapItem,
  Button,
  Tooltip,
} from '@chakra-ui/react';
import { buildEchartsOption, extractSeriesNames } from './chartEngine';

/**
 * ChartContainer renders the full multi-grid ECharts chart from API render data.
 * Handles:
 *  - Building the ECharts option from raw API data via chartEngine
 *  - Legend toggle (series visibility)
 *  - Responsive resizing via ResizeObserver (React-idiomatic ref pattern)
 *  - Dark / light mode theming
 */
export default function ChartContainer({ renderData, loading, error }) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';
  const echartsRef = useRef(null);
  const containerRef = useRef(null);
  const [hiddenSeries, setHiddenSeries] = useState([]);

  const option = buildEchartsOption(renderData, hiddenSeries, isDark);
  const totalHeight = option._totalHeight ?? 600;
  const seriesNames = extractSeriesNames(option);

  const toggleSeries = useCallback((name) => {
    setHiddenSeries((prev) =>
      prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]
    );
  }, []);

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
      {/* Series toggle legend */}
      {seriesNames.length > 0 && (
        <Wrap spacing={1} px={1}>
          {seriesNames.map((name) => {
            const isHidden = hiddenSeries.includes(name);
            return (
              <WrapItem key={name}>
                <Tooltip label={isHidden ? 'Show' : 'Hide'} placement="top">
                  <Button
                    size="xs"
                    variant={isHidden ? 'outline' : 'solid'}
                    colorScheme={isHidden ? 'gray' : 'brand'}
                    opacity={isHidden ? 0.5 : 1}
                    onClick={() => toggleSeries(name)}
                    borderRadius="full"
                  >
                    {name}
                  </Button>
                </Tooltip>
              </WrapItem>
            );
          })}
        </Wrap>
      )}

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
