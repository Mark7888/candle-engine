import { useState } from 'react';
import {
  Flex,
  HStack,
  Text,
  Select,
  Input,
  Button,
  useColorModeValue,
} from '@chakra-ui/react';

const DEFAULT_SETTINGS = {
  timeFrame: '',
  timeRangeFrom: '',
  timeRangeTo: '',
};

const TIME_FRAMES = [
  { value: '', label: 'Default' },
  { value: '1m', label: '1m' },
  { value: '5m', label: '5m' },
  { value: '15m', label: '15m' },
  { value: '30m', label: '30m' },
  { value: '1h', label: '1h' },
  { value: '4h', label: '4h' },
  { value: '1d', label: '1d' },
  { value: '1w', label: '1w' },
  { value: '1M', label: '1M' },
];

/**
 * SettingsBar – horizontal bar displayed per-project between the project tabs
 * and the pair selector / chart area.
 *
 * Props:
 *   settings  – committed settings object { timeFrame, timeRangeFrom, timeRangeTo }
 *   onApply   – called with the new settings object when the user clicks Apply
 */
export default function SettingsBar({ settings, onApply }) {
  // Initialize local draft from committed settings. Re-sync is not needed
  // because ProjectView (and therefore SettingsBar) is remounted whenever the
  // active project changes (via key={activeProjectId} in App.jsx), and
  // settings can only be committed via the Apply button inside this component.
  const [local, setLocal] = useState(() => ({ ...DEFAULT_SETTINGS, ...settings }));

  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const bg = useColorModeValue('white', 'gray.800');
  const labelColor = useColorModeValue('gray.600', 'gray.400');

  const set = (key, value) => setLocal((s) => ({ ...s, [key]: value }));

  const handleApply = () => onApply({ ...local });

  const handleReset = () => {
    setLocal({ ...DEFAULT_SETTINGS });
    onApply({ ...DEFAULT_SETTINGS });
  };

  return (
    <Flex
      align="center"
      gap={4}
      px={3}
      py={2}
      borderBottom="1px solid"
      borderColor={borderColor}
      bg={bg}
      flexShrink={0}
      flexWrap="wrap"
    >
      {/* Time Frame */}
      <HStack spacing={2}>
        <Text fontSize="sm" fontWeight="medium" color={labelColor} whiteSpace="nowrap">
          Time Frame
        </Text>
        <Select
          size="sm"
          value={local.timeFrame}
          onChange={(e) => set('timeFrame', e.target.value)}
          w="80px"
          borderRadius="md"
        >
          {TIME_FRAMES.map((tf) => (
            <option key={tf.value} value={tf.value}>
              {tf.label}
            </option>
          ))}
        </Select>
      </HStack>

      {/* Time Range From */}
      <HStack spacing={2}>
        <Text fontSize="sm" fontWeight="medium" color={labelColor} whiteSpace="nowrap">
          From
        </Text>
        <Input
          type="datetime-local"
          size="sm"
          value={local.timeRangeFrom}
          onChange={(e) => set('timeRangeFrom', e.target.value)}
          w="200px"
          borderRadius="md"
        />
      </HStack>

      {/* Time Range To */}
      <HStack spacing={2}>
        <Text fontSize="sm" fontWeight="medium" color={labelColor} whiteSpace="nowrap">
          To
        </Text>
        <Input
          type="datetime-local"
          size="sm"
          value={local.timeRangeTo}
          onChange={(e) => set('timeRangeTo', e.target.value)}
          w="200px"
          borderRadius="md"
        />
      </HStack>

      {/* Actions */}
      <HStack spacing={2}>
        <Button size="sm" colorScheme="blue" onClick={handleApply}>
          Apply
        </Button>
        <Button size="sm" variant="ghost" onClick={handleReset}>
          Reset
        </Button>
      </HStack>
    </Flex>
  );
}
