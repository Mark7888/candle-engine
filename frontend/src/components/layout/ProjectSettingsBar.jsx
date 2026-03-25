import {
  Flex,
  Box,
  Select,
  Input,
  Text,
  IconButton,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import { FaPlay } from 'react-icons/fa';

/**
 * Converts a timeframe string (e.g. "1m", "1h", "1d") to seconds so we can
 * sort the list from smallest to largest.  Falls back to 0 for unknown values.
 */
const TIMEFRAME_SECONDS = {
  '1m': 60,
  '3m': 180,
  '5m': 300,
  '15m': 900,
  '30m': 1800,
  '1h': 3600,
  '2h': 7200,
  '4h': 14400,
  '6h': 21600,
  '8h': 28800,
  '12h': 43200,
  '1d': 86400,
  '3d': 259200,
  '1w': 604800,
  '1M': 2592000,
};

const sortedTimeFrames = (tfs) =>
  [...tfs].sort(
    (a, b) => (TIMEFRAME_SECONDS[a] ?? 0) - (TIMEFRAME_SECONDS[b] ?? 0)
  );

/** Format an ISO timestamp string (or Date) to a YYYY-MM-DD string for <input type="date">. */
const toDateValue = (iso) => {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '';
  return d.toISOString().slice(0, 10);
};

/**
 * ProjectSettingsBar – horizontal bar displayed above the pair list and chart.
 *
 * Props:
 *   projectDetail  – the current project detail object from the store
 *   storedSettings – per-project settings from uiStore (may be undefined)
 *   onSettingChange(patch) – called immediately when a setting changes
 *                            patch is a partial object, e.g. { timeFrame: "1h" }
 *                            or { timeRange: { start, end } }
 */
export default function ProjectSettingsBar({ projectDetail, storedSettings, onSettingChange }) {
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const labelColor = useColorModeValue('gray.500', 'gray.400');
  const inputBg = useColorModeValue('white', 'gray.800');

  if (!projectDetail) return null;

  const availableTimeFrames = sortedTimeFrames(projectDetail.availableTimeFrames ?? []);

  // Active settings: prefer stored overrides, fall back to project defaults.
  const activeTimeFrame = storedSettings?.timeFrame ?? projectDetail.timeFrame;
  const activeStart = storedSettings?.timeRangeStart ?? toDateValue(projectDetail.timeRange?.start);
  const activeEnd = storedSettings?.timeRangeEnd ?? toDateValue(projectDetail.timeRange?.end);

  const handleTimeFrameChange = (e) => {
    const tf = e.target.value;
    if (tf && tf !== activeTimeFrame) {
      onSettingChange({ timeFrame: tf });
    }
  };

  const handleStartChange = (e) => {
    const val = e.target.value; // YYYY-MM-DD
    if (!val) return;
    onSettingChange({ timeRangeStart: val });
  };

  const handleEndChange = (e) => {
    const val = e.target.value; // YYYY-MM-DD
    if (!val) return;
    onSettingChange({ timeRangeEnd: val });
  };

  return (
    <Flex
      align="center"
      gap={4}
      px={3}
      py={2}
      borderBottom="1px solid"
      borderColor={borderColor}
      flexShrink={0}
      flexWrap="wrap"
    >
      {/* Time Frame */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap" fontWeight="medium">
          Time Frame
        </Text>
        <Select
          size="sm"
          value={activeTimeFrame}
          onChange={handleTimeFrameChange}
          w="90px"
          bg={inputBg}
          borderRadius="md"
        >
          {availableTimeFrames.map((tf) => (
            <option key={tf} value={tf}>
              {tf}
            </option>
          ))}
        </Select>
      </Flex>

      {/* Time Range */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap" fontWeight="medium">
          From
        </Text>
        <Input
          type="date"
          size="sm"
          value={activeStart}
          onChange={handleStartChange}
          max={activeEnd || undefined}
          w="140px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap" fontWeight="medium">
          To
        </Text>
        <Input
          type="date"
          size="sm"
          value={activeEnd}
          onChange={handleEndChange}
          min={activeStart || undefined}
          w="140px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      {/* Spacer pushes the PlayBack button to the right */}
      <Box flex={1} />

      {/* PlayBack mode button – disabled, no functionality yet */}
      <Tooltip label="PlayBack mode" placement="bottom">
        <IconButton
          aria-label="PlayBack mode"
          icon={<FaPlay />}
          size="sm"
          isDisabled
          variant="outline"
          colorScheme="brand"
        />
      </Tooltip>
    </Flex>
  );
}
