import { useState } from 'react';
import {
  Flex,
  Select,
  Input,
  Text,
  IconButton,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import { BsPlayCircle } from 'react-icons/bs';

interface SettingsBarProps {
  availableTimeFrames?: string[];
  defaultTimeFrame?: string;
  defaultRangeFrom?: string;
  defaultRangeTo?: string;
  timeFrame?: string | null;
  rangeFrom?: string | null;
  rangeTo?: string | null;
  onTimeFrameChange?: (value: string) => void;
  onRangeFromChange?: (value: string) => void;
  onRangeToChange?: (value: string) => void;
  /** Whether Playback mode is currently active for this project. */
  playbackActive?: boolean;
  /** Called when the user clicks the Playback mode toggle button. */
  onPlaybackToggle?: () => void;
}

/**
 * SettingsBar – a horizontal toolbar rendered inside a project, above the pair
 * selector and chart.  It lets the user override the project-level time frame
 * and date range without touching any backend state.
 *
 * Props:
 *   availableTimeFrames  string[]   – ordered list of valid time frames
 *   defaultTimeFrame     string     – project default (used as placeholder)
 *   defaultRangeFrom     string     – project default range start (ISO)
 *   defaultRangeTo       string     – project default range end (ISO)
 *   timeFrame            string|null  – current client-side override
 *   rangeFrom            string|null  – current client-side override
 *   rangeTo              string|null  – current client-side override
 *   onTimeFrameChange    (value: string) => void
 *   onRangeFromChange    (value: string) => void
 *   onRangeToChange      (value: string) => void
 *   playbackActive       boolean – whether Playback mode is currently active
 *   onPlaybackToggle     () => void – called when the Playback button is clicked
 */
export default function SettingsBar({
  availableTimeFrames = [],
  defaultTimeFrame = '',
  defaultRangeFrom = '',
  defaultRangeTo = '',
  timeFrame = null,
  rangeFrom = null,
  rangeTo = null,
  onTimeFrameChange,
  onRangeFromChange,
  onRangeToChange,
  playbackActive = false,
  onPlaybackToggle,
}: SettingsBarProps) {
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const labelColor = useColorModeValue('gray.600', 'gray.400');
  const inputBg = useColorModeValue('white', 'gray.800');

  // Convert ISO timestamp to the value expected by <input type="datetime-local">
  const toDatetimeLocal = (iso: string): string => {
    if (!iso) return '';
    try {
      const d = new Date(iso);
      // Format: YYYY-MM-DDTHH:mm
      return d.toISOString().slice(0, 16);
    } catch {
      return '';
    }
  };

  // Convert datetime-local value back to full ISO 8601 string for the API
  const fromDatetimeLocal = (value: string): string => {
    if (!value) return '';
    return new Date(value).toISOString();
  };

  const displayFrom = rangeFrom ?? defaultRangeFrom;
  const displayTo = rangeTo ?? defaultRangeTo;
  const displayTimeFrame = timeFrame ?? defaultTimeFrame;

  // Local draft state for the datetime-local inputs so the user can freely edit
  // them without the controlled value resetting mid-edit.
  const [fromDraft, setFromDraft] = useState(() => toDatetimeLocal(displayFrom));
  const [toDraft, setToDraft] = useState(() => toDatetimeLocal(displayTo));
  const [fromFocused, setFromFocused] = useState(false);
  const [toFocused, setToFocused] = useState(false);

  // Sync from external prop changes only when the input is not focused.
  // Using the "previous value in state" pattern (derived state) avoids calling
  // setState inside a useEffect, which the linter flags as a cascade risk.
  const [prevDisplayFrom, setPrevDisplayFrom] = useState(displayFrom);
  const [prevDisplayTo, setPrevDisplayTo] = useState(displayTo);

  if (displayFrom !== prevDisplayFrom) {
    setPrevDisplayFrom(displayFrom);
    if (!fromFocused) setFromDraft(toDatetimeLocal(displayFrom));
  }
  if (displayTo !== prevDisplayTo) {
    setPrevDisplayTo(displayTo);
    if (!toFocused) setToDraft(toDatetimeLocal(displayTo));
  }

  return (
    <Flex
      align="center"
      gap={4}
      px={3}
      py={1}
      borderBottom="1px solid"
      borderColor={borderColor}
      flexShrink={0}
      flexWrap="wrap"
    >
      {/* Time Frame */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          Time Frame
        </Text>
        <Select
          size="xs"
          value={displayTimeFrame}
          onChange={(e) => onTimeFrameChange?.(e.target.value)}
          w="80px"
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

      {/* Range From */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          From
        </Text>
        <Input
          type="datetime-local"
          size="xs"
          value={fromDraft}
          onChange={(e) => {
            setFromDraft(e.target.value);
            if (e.target.value) onRangeFromChange?.(fromDatetimeLocal(e.target.value));
          }}
          onFocus={() => setFromFocused(true)}
          onBlur={() => setFromFocused(false)}
          w="180px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      {/* Range To */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          To
        </Text>
        <Input
          type="datetime-local"
          size="xs"
          value={toDraft}
          onChange={(e) => {
            setToDraft(e.target.value);
            if (e.target.value) onRangeToChange?.(fromDatetimeLocal(e.target.value));
          }}
          onFocus={() => setToFocused(true)}
          onBlur={() => setToFocused(false)}
          w="180px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      {/* Spacer */}
      <Flex flex={1} />

      {/* PlayBack mode toggle button */}
      <Tooltip
        label={
          playbackActive
            ? 'Playback active – use × to exit'
            : 'Start Playback mode'
        }
        placement="bottom"
        shouldWrapChildren
      >
        <IconButton
          aria-label="PlayBack mode"
          icon={<BsPlayCircle />}
          size="sm"
          variant={playbackActive ? 'solid' : 'ghost'}
          colorScheme={playbackActive ? 'brand' : undefined}
          isDisabled={playbackActive}
          onClick={onPlaybackToggle}
        />
      </Tooltip>
    </Flex>
  );
}
