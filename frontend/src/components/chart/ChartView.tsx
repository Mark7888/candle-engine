import { useMemo } from 'react';
import { Spinner, Alert } from '@chakra-ui/react';
import ChartContainer from './ChartContainer';
import { buildEChartsOption } from './chartUtils';
import type { ProjectRenderResponse } from '../../types/api';

interface ChartViewProps {
  data: ProjectRenderResponse | null;
  loading: boolean;
  error: string | null;
}

export default function ChartView({ data, loading, error }: ChartViewProps) {
  const result = useMemo(() => {
    if (!data || data.candles.length === 0) return null;
    return buildEChartsOption(data.candles, data.candleColors, data.indicators);
  }, [data]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="xl" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Title>{error}</Alert.Title>
      </Alert.Root>
    );
  }

  if (!result) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-400">
        No chart data available.
      </div>
    );
  }

  return <ChartContainer result={result} />;
}
