import { useState, useCallback } from 'react';
import {
  Flex,
  Select,
  Input,
  Text,
  IconButton,
  Tooltip,
  Box,
  useColorModeValue,
} from '@chakra-ui/react';
import { BsPlayCircle } from 'react-icons/bs';
import { LuSettings2 } from 'react-icons/lu';
import ProjectSettingsModal from '../settings/ProjectSettingsModal';
import { validateDateRange, isDateRangeValid } from '../../utils/validation';

interface SettingsBarProps {
  /** The ID of the project this bar belongs to (used for per-project settings). */
  projectId: number;
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
 *   projectId            number     – the owning project (required for per-project settings)
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
  projectId,
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

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [settingsKey, setSettingsKey] = useState(0);
  const handleSettingsOpen = useCallback(() => {
    setSettingsKey((k) => k + 1);
    setSettingsOpen(true);
  }, []);
  const handleSettingsClose = useCallback(() => setSettingsOpen(false), []);

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

  // Validate the current draft values on every render (cheap, pure computation).
  const rangeValidation = validateDateRange(fromDraft, toDraft);

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
        <Tooltip
          label={rangeValidation.fromError ?? ''}
          isDisabled={!rangeValidation.fromError}
          placement="bottom"
          hasArrow
        >
          <Input
            type="datetime-local"
            size="xs"
            value={fromDraft}
            onChange={(e) => {
              setFromDraft(e.target.value);
              const next = validateDateRange(e.target.value, toDraft);
              if (e.target.value && isDateRangeValid(next)) {
                onRangeFromChange?.(fromDatetimeLocal(e.target.value));
              }
            }}
            onFocus={() => setFromFocused(true)}
            onBlur={() => setFromFocused(false)}
            w="180px"
            bg={inputBg}
            borderRadius="md"
            isInvalid={!!rangeValidation.fromError}
            borderColor={rangeValidation.fromError ? 'red.400' : undefined}
            _invalid={{ borderColor: 'red.400', boxShadow: '0 0 0 1px var(--chakra-colors-red-400)' }}
          />
        </Tooltip>
      </Flex>

      {/* Range To */}
      <Flex align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          To
        </Text>
        <Tooltip
          label={rangeValidation.toError ?? ''}
          isDisabled={!rangeValidation.toError}
          placement="bottom"
          hasArrow
        >
          <Input
            type="datetime-local"
            size="xs"
            value={toDraft}
            onChange={(e) => {
              setToDraft(e.target.value);
              const next = validateDateRange(fromDraft, e.target.value);
              if (e.target.value && isDateRangeValid(next)) {
                onRangeToChange?.(fromDatetimeLocal(e.target.value));
              }
            }}
            onFocus={() => setToFocused(true)}
            onBlur={() => setToFocused(false)}
            w="180px"
            bg={inputBg}
            borderRadius="md"
            isInvalid={!!rangeValidation.toError}
            borderColor={rangeValidation.toError ? 'red.400' : undefined}
            _invalid={{ borderColor: 'red.400', boxShadow: '0 0 0 1px var(--chakra-colors-red-400)' }}
          />
        </Tooltip>
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

      {/* Vertical divider */}
      <Box w="1px" h="20px" bg={borderColor} mx={1} flexShrink={0} />

      {/* Project settings button */}
      <Tooltip label="Project settings" placement="bottom" shouldWrapChildren>
        <IconButton
          aria-label="Project settings"
          icon={<LuSettings2 />}
          size="sm"
          variant="ghost"
          onClick={handleSettingsOpen}
        />
      </Tooltip>

      <ProjectSettingsModal
        key={settingsKey}
        projectId={projectId}
        isOpen={settingsOpen}
        onClose={handleSettingsClose}
      />
    </Flex>
  );
}
