import type { RenderResponse } from '../types/api';

/**
 * Returns a new RenderResponse sliced so that only data up to (and including)
 * `upToIndex` is present in the candle series and indicator plotPoints.
 *
 * Logs are filtered to those whose optional `time` field is ≤ the timestamp
 * of the candle at `upToIndex`.  Logs without a `time` value are always
 * included (they are not tied to a specific candle).
 *
 * No API call is made – this is a pure data transformation on already-loaded
 * render data, intended for the Playback mode feature.
 */
export function sliceRenderData(renderData: RenderResponse, upToIndex: number): RenderResponse {
  const slicedCandles = renderData.candles.map((series) => ({
    ...series,
    entries: series.entries.slice(0, upToIndex + 1),
  }));

  const slicedIndicators = renderData.indicators.map((indicator) => ({
    ...indicator,
    plotPoints: indicator.plotPoints.slice(0, upToIndex + 1),
  }));

  // Determine the timestamp cut-off from the active candle series.
  const { activePair } = renderData;
  const activeSeries = activePair
    ? renderData.candles.find(
        (c) => c.sourceName === activePair.sourceName && c.pairId === activePair.pairId
      )
    : renderData.candles[0];

  const cutoffTimestamp = activeSeries?.entries[upToIndex]?.timestamp;

  // Keep logs whose `time` is ≤ the cutoff, or logs with no `time` at all.
  const slicedLogs = cutoffTimestamp
    ? renderData.logs.filter((log) => !log.time || log.time <= cutoffTimestamp)
    : renderData.logs;

  return {
    ...renderData,
    candles: slicedCandles,
    indicators: slicedIndicators,
    logs: slicedLogs,
  };
}
