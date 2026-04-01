/**
 * Default values for per-project settings.
 *
 * All new setting fields should be added here so that projects without an
 * existing saved record receive a sensible initial value.
 */

export interface ProjectSettingsConfig {
  /** A dummy color setting used to demonstrate the settings persistence flow. */
  testColor: string;
}

export const DEFAULT_PROJECT_SETTINGS: ProjectSettingsConfig = {
  testColor: '#004040',
};
