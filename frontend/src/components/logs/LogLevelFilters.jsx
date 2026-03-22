import { HStack, Button } from '@chakra-ui/react';
import { MdBugReport, MdInfo, MdWarning, MdError } from 'react-icons/md';

const LEVELS = [
  { key: 'DEBUG', colorScheme: 'gray', Icon: MdBugReport },
  { key: 'INFO', colorScheme: 'green', Icon: MdInfo },
  { key: 'WARN', colorScheme: 'yellow', Icon: MdWarning },
  { key: 'ERROR', colorScheme: 'red', Icon: MdError },
];

/**
 * Renders four toggle buttons (DEBUG / INFO / WARN / ERROR).
 * When a level is in `hiddenLevels`, its button appears dimmed/outlined.
 */
export default function LogLevelFilters({ hiddenLevels, onToggle }) {
  return (
    <HStack spacing={1} flexWrap="wrap">
      {LEVELS.map(({ key, colorScheme, Icon }) => {
        const isHidden = hiddenLevels.includes(key);
        return (
          <Button
            key={key}
            size="xs"
            variant={isHidden ? 'ghost' : 'solid'}
            colorScheme={isHidden ? 'gray' : colorScheme}
            opacity={isHidden ? 0.45 : 1}
            onClick={() => onToggle(key)}
            borderRadius="full"
            fontWeight="semibold"
            leftIcon={<Icon size={11} />}
          >
            {key}
          </Button>
        );
      })}
    </HStack>
  );
}
