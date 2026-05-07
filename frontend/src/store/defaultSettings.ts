/**
 * Default values for per-project settings.
 *
 * All new setting fields should be added here so that projects without an
 * existing saved record receive a sensible initial value.
 */

export interface ProjectSettingsConfig {
  /** Candle bullish (up) color in CSS hex format (e.g. '#26a69a'). */
  candleBullishColor: string;
  /** Candle bearish (down) color in CSS hex format (e.g. '#ef5350'). */
  candleBearishColor: string;
}

export const DEFAULT_PROJECT_SETTINGS: ProjectSettingsConfig = {
  candleBullishColor: '#26a69a',
  candleBearishColor: '#ef5350',
};
