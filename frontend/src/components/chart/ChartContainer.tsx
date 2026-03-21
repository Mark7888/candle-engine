import { useRef, useEffect, useCallback } from 'react';
import * as echarts from 'echarts';
import type { EChartsResult } from './chartUtils';

interface ChartContainerProps {
  result: EChartsResult;
}

/**
 * Renders a single ECharts instance that contains multiple linked grids
 * (main candlestick chart + sub-indicator charts).
 */
export default function ChartContainer({ result }: ChartContainerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const instanceRef = useRef<echarts.ECharts | null>(null);

  const initChart = useCallback(() => {
    if (!containerRef.current) return;

    // Dispose previous instance if switching projects/pairs
    if (instanceRef.current) {
      instanceRef.current.dispose();
      instanceRef.current = null;
    }

    const chart = echarts.init(containerRef.current, 'dark', { renderer: 'canvas' });
    instanceRef.current = chart;
    chart.setOption(result.option);
  }, [result]);

  useEffect(() => {
    initChart();
  }, [initChart]);

  // Resize observer
  useEffect(() => {
    if (!containerRef.current) return;
    const observer = new ResizeObserver(() => {
      instanceRef.current?.resize();
    });
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      instanceRef.current?.dispose();
    };
  }, []);

  return (
    <div
      ref={containerRef}
      style={{ width: '100%', height: result.totalHeight }}
    />
  );
}
