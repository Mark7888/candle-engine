import { memo, useState } from 'react';
import { HStack, Box, Text } from '@chakra-ui/react';
import { MdError, MdWarning, MdInfo, MdBugReport, MdChevronRight, MdExpandMore } from 'react-icons/md';
import type { IconType } from 'react-icons';
import type { LogEntry as LogEntryType } from '../../types/api';

interface LevelStyle {
  bg: string;
  text: string;
  Icon: IconType;
}

const LEVEL_STYLES_DARK: Record<string, LevelStyle> = {
  ERROR: { bg: 'red.900', text: 'red.300', Icon: MdError },
  WARN: { bg: 'yellow.900', text: 'yellow.300', Icon: MdWarning },
  INFO: { bg: 'green.900', text: 'green.300', Icon: MdInfo },
  DEBUG: { bg: 'gray.800', text: 'gray.400', Icon: MdBugReport },
};

const LEVEL_STYLES_LIGHT: Record<string, LevelStyle> = {
  ERROR: { bg: 'red.50', text: 'red.600', Icon: MdError },
  WARN: { bg: 'yellow.50', text: 'yellow.700', Icon: MdWarning },
  INFO: { bg: 'green.50', text: 'green.700', Icon: MdInfo },
  DEBUG: { bg: 'gray.50', text: 'gray.500', Icon: MdBugReport },
};

function formatTime(timeStr: string | undefined): string {
  if (!timeStr) return '';
  const d = new Date(timeStr);
  if (isNaN(d.getTime())) return timeStr;
  return d.toLocaleTimeString(undefined, { hour12: false });
}

interface LogEntryProps {
  log: LogEntryType;
  isDark: boolean;
}

const LogEntry = memo(function LogEntry({ log, isDark }: LogEntryProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const palette = isDark ? LEVEL_STYLES_DARK : LEVEL_STYLES_LIGHT;
  const { bg, text, Icon } = palette[log.level] ?? palette['DEBUG'];

  return (
    <Box
      borderRadius="sm"
      bg={bg}
      fontFamily="mono"
      fontSize="xs"
      cursor="pointer"
      onClick={() => setIsExpanded((e) => !e)}
      sx={{ contentVisibility: 'auto', containIntrinsicSize: '0 28px' }}
    >
      {/* Header row: icon · time · indicator · (truncated message when collapsed) · triangle */}
      <HStack spacing={2} px={2} py={0.5} alignItems="center">
        <Box color={text} flexShrink={0}>
          <Icon size={12} />
        </Box>
        {log.time && (
          <Text color={isDark ? 'gray.500' : 'gray.400'} flexShrink={0} whiteSpace="nowrap">
            [{formatTime(log.time)}]
          </Text>
        )}
        {log.indicator && (
          <Text color={isDark ? 'purple.400' : 'purple.600'} flexShrink={0} whiteSpace="nowrap">
            {log.indicator}:
          </Text>
        )}
        {/* Truncated message (collapsed) or flex spacer (expanded) to keep triangle on the right */}
        {!isExpanded ? (
          <Box color={text} flex={1} minW={0} overflow="hidden" whiteSpace="nowrap" textOverflow="ellipsis">
            {log.message}
          </Box>
        ) : (
          <Box flex={1} />
        )}
        {/* Expand / collapse triangle */}
        <Box color={isDark ? 'gray.400' : 'gray.500'} flexShrink={0}>
          {isExpanded ? <MdExpandMore size={14} /> : <MdChevronRight size={14} />}
        </Box>
      </HStack>

      {/* Full message shown below header when expanded */}
      {isExpanded && (
        <Box px={2} pb={1} pl={8} color={text} wordBreak="break-word">
          {log.message}
        </Box>
      )}
    </Box>
  );
});

export default LogEntry;
