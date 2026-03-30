import { useEffect, useRef, useCallback } from 'react';
import {
  Flex,
  Box,
  Text,
  IconButton,
  Input,
  Tooltip,
  Slider,
  SliderTrack,
  SliderFilledTrack,
  SliderThumb,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  BsSkipBackward,
  BsSkipForward,
  BsPlayFill,
  BsPauseFill,
  BsX,
} from 'react-icons/bs';

interface PlaybackBarProps {
  /** Total number of candles in the loaded dataset. */
  totalCandles: number;
  /** 0-based index of the currently displayed last candle. */
  currentIndex: number;
  /** N – candles to skip per step. */
  stepSize: number;
  /** M – seconds between auto-play steps. */
  playInterval: number;
  /** Whether auto-play is currently running. */
  isPlaying: boolean;
  /** ISO timestamp of the candle at currentIndex (used for the date display). */
  currentTimestamp: string;
  onDeactivate: () => void;
  onIndexChange: (index: number) => void;
  onStepSizeChange: (size: number) => void;
  onPlayIntervalChange: (interval: number) => void;
  onPlayingChange: (playing: boolean) => void;
}

/**
 * PlaybackBar – the control bar that appears below the SettingsBar when
 * Playback mode is active.
 *
 * Contains:
 *  • X button to exit Playback mode
 *  • Current end-date and remaining-candles display
 *  • Skip-back / Play-pause / Skip-forward buttons
 *  • Progress slider (drag to any position)
 *  • Step size (N) and play interval (M seconds) inputs
 */
export default function PlaybackBar({
  totalCandles,
  currentIndex,
  stepSize,
  playInterval,
  isPlaying,
  currentTimestamp,
  onDeactivate,
  onIndexChange,
  onStepSizeChange,
  onPlayIntervalChange,
  onPlayingChange,
}: PlaybackBarProps) {
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const barBg = useColorModeValue('brand.50', 'brand.950');
  const labelColor = useColorModeValue('gray.600', 'gray.400');
  const inputBg = useColorModeValue('white', 'gray.800');

  const atStart = currentIndex === 0;
  const atEnd = currentIndex >= totalCandles - 1;
  const remaining = Math.max(0, totalCandles - 1 - currentIndex);

  // Keep mutable values in a ref so the interval callback never goes stale.
  const stateRef = useRef({ currentIndex, stepSize, totalCandles });
  useEffect(() => {
    stateRef.current = { currentIndex, stepSize, totalCandles };
  });

  // Auto-play: only recreate the interval when isPlaying or playInterval changes.
  useEffect(() => {
    if (!isPlaying) return;

    const id = setInterval(() => {
      const { currentIndex: idx, stepSize: step, totalCandles: total } = stateRef.current;
      const next = Math.min(idx + step, total - 1);
      onIndexChange(next);
      if (next >= total - 1) {
        onPlayingChange(false);
      }
    }, playInterval * 1000);

    return () => clearInterval(id);
  }, [isPlaying, playInterval, onIndexChange, onPlayingChange]);

  const handleBack = useCallback(() => {
    onPlayingChange(false);
    onIndexChange(Math.max(0, currentIndex - stepSize));
  }, [currentIndex, stepSize, onIndexChange, onPlayingChange]);

  const handleForward = useCallback(() => {
    onPlayingChange(false);
    onIndexChange(Math.min(totalCandles - 1, currentIndex + stepSize));
  }, [currentIndex, stepSize, totalCandles, onIndexChange, onPlayingChange]);

  const handlePlayPause = useCallback(() => {
    onPlayingChange(!isPlaying);
  }, [isPlaying, onPlayingChange]);

  const handleSliderChange = useCallback(
    (value: number) => {
      onPlayingChange(false);
      onIndexChange(value);
    },
    [onIndexChange, onPlayingChange]
  );

  // Format timestamp for display: "YYYY-MM-DD HH:mm UTC"
  const displayDate = currentTimestamp
    ? `${new Date(currentTimestamp).toISOString().replace('T', ' ').slice(0, 16)} UTC`
    : '—';

  return (
    <Flex
      align="center"
      gap={3}
      px={3}
      py={1.5}
      borderBottom="1px solid"
      borderColor={borderColor}
      bg={barBg}
      flexShrink={0}
      flexWrap="wrap"
    >
      {/* Current position info */}
      <Box minW="120px">
        <Text fontSize="xs" color={labelColor} lineHeight={1.3}>
          Playback
        </Text>
        <Text fontSize="xs" fontWeight="semibold" lineHeight={1.3} whiteSpace="nowrap">
          {displayDate}
        </Text>
      </Box>

      {/* Remaining / end-of-dataset notice */}
      <Box minW="90px">
        {atEnd ? (
          <Text fontSize="xs" color="orange.400" fontWeight="medium" whiteSpace="nowrap">
            End of dataset
          </Text>
        ) : (
          <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
            {remaining} candle{remaining !== 1 ? 's' : ''} left
          </Text>
        )}
      </Box>

      {/* Transport controls */}
      <Flex align="center" gap={0}>
        <Tooltip label={`Back ${stepSize} candle${stepSize !== 1 ? 's' : ''}`} placement="bottom">
          <IconButton
            aria-label="Skip backward"
            icon={<BsSkipBackward />}
            size="sm"
            variant="ghost"
            isDisabled={atStart}
            onClick={handleBack}
          />
        </Tooltip>

        <Tooltip label={isPlaying ? 'Pause' : 'Play'} placement="bottom">
          <IconButton
            aria-label={isPlaying ? 'Pause' : 'Play'}
            icon={isPlaying ? <BsPauseFill /> : <BsPlayFill />}
            size="sm"
            variant="ghost"
            colorScheme="brand"
            isDisabled={atEnd && !isPlaying}
            onClick={handlePlayPause}
          />
        </Tooltip>

        <Tooltip
          label={`Forward ${stepSize} candle${stepSize !== 1 ? 's' : ''}`}
          placement="bottom"
        >
          <IconButton
            aria-label="Skip forward"
            icon={<BsSkipForward />}
            size="sm"
            variant="ghost"
            isDisabled={atEnd}
            onClick={handleForward}
          />
        </Tooltip>
      </Flex>

      {/* Progress slider */}
      <Flex flex={1} minW="120px" align="center" gap={2}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap" flexShrink={0}>
          {currentIndex + 1} / {totalCandles}
        </Text>
        <Slider
          flex={1}
          min={0}
          max={Math.max(0, totalCandles - 1)}
          value={currentIndex}
          onChange={handleSliderChange}
          colorScheme="brand"
          size="sm"
          focusThumbOnChange={false}
        >
          <SliderTrack>
            <SliderFilledTrack />
          </SliderTrack>
          <SliderThumb />
        </Slider>
      </Flex>

      {/* Step size */}
      <Flex align="center" gap={1} flexShrink={0}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          Step
        </Text>
        <Input
          type="number"
          size="xs"
          value={stepSize}
          min={1}
          onChange={(e) => {
            const v = parseInt(e.target.value, 10);
            if (v > 0) onStepSizeChange(v);
          }}
          w="52px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      {/* Play interval (seconds) */}
      <Flex align="center" gap={1} flexShrink={0}>
        <Text fontSize="xs" color={labelColor} whiteSpace="nowrap">
          Interval (s)
        </Text>
        <Input
          type="number"
          size="xs"
          value={playInterval}
          min={0.1}
          step={0.1}
          onChange={(e) => {
            const v = parseFloat(e.target.value);
            if (v >= 0.1) onPlayIntervalChange(v);
          }}
          w="60px"
          bg={inputBg}
          borderRadius="md"
        />
      </Flex>

      {/* Exit button – placed on the right */}
      <Tooltip label="Exit Playback mode" placement="bottom">
        <IconButton
          aria-label="Exit Playback mode"
          icon={<BsX size={18} />}
          size="sm"
          variant="ghost"
          colorScheme="red"
          onClick={onDeactivate}
        />
      </Tooltip>
    </Flex>
  );
}
