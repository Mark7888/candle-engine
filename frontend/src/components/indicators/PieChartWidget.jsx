import { Box, Center, Text, useColorMode } from '@chakra-ui/react';
import ReactECharts from 'echarts-for-react';
import { normalizeColor } from '../chart/chartEngine';

/**
 * PieChartWidget – renders a pie/donut chart for a pieChart indicator.
 *
 * Config fields (camelCase from Go PieChartIndicatorConfig):
 *   displayName, sliceTitles
 *
 * Each PlotPoint in the last valid collection is one slice.
 */
export default function PieChartWidget({ collection, config, height = 180 }) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const sliceTitles = config?.sliceTitles ?? [];
  const points = collection?.points ?? [];

  const validPoints = points.filter((pt) => pt && !pt.invalid);

  if (validPoints.length === 0) {
    return (
      <Center h={`${height}px`}>
        <Text fontSize="xs" color="gray.500">
          No data
        </Text>
      </Center>
    );
  }

  const sliceData = validPoints.map((pt, i) => ({
    name: sliceTitles[i] ?? `Slice ${i + 1}`,
    value: pt.value,
    itemStyle: { color: pt.color ? normalizeColor(pt.color) : undefined },
  }));

  const option = {
    animation: false,
    backgroundColor: isDark ? '#1a202c' : '#fff',
    tooltip: { trigger: 'item', appendToBody: true },
    series: [
      {
        type: 'pie',
        radius: ['35%', '65%'],
        data: sliceData,
        label: { color: isDark ? '#aaa' : '#555' },
      },
    ],
  };

  return (
    <Box>
      <ReactECharts
        option={option}
        style={{ width: '100%', height: `${height}px` }}
        notMerge
      />
    </Box>
  );
}
