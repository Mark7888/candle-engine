import { useState, useMemo } from 'react';
import {
  Box,
  Flex,
  HStack,
  Text,
  Select,
  IconButton,
  VStack,
  Spinner,
  useColorMode,
} from '@chakra-ui/react';
import { MdChevronLeft, MdClose } from 'react-icons/md';
import LogEntry from './LogEntry';
import LogLevelFilters from './LogLevelFilters';
import { useUIStore } from '../../store/uiStore';
import type { LogEntry as LogEntryType } from '../../types/api';

interface LogPanelProps {
  logs?: LogEntryType[];
  isOpen: boolean;
  isLoading?: boolean;
  onClose: () => void;
  onOpen: () => void;
}

export default function LogPanel({ logs = [], isOpen, isLoading = false, onClose, onOpen }: LogPanelProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const [selectedChannel, setSelectedChannel] = useState('');

  const hiddenLevels = useUIStore((s) => s.hiddenLogLevels);
  const toggleLevel = useUIStore((s) => s.toggleLogLevel);

  const channels = useMemo(() => {
    const names = new Set(logs.map((l) => l.indicator).filter((x): x is string => Boolean(x)));
    return Array.from(names).sort();
  }, [logs]);

  const filteredLogs = useMemo(
    () =>
      logs.filter((l) => {
        if (selectedChannel && l.indicator !== selectedChannel) return false;
        if (hiddenLevels.includes(l.level)) return false;
        return true;
      }),
    [logs, selectedChannel, hiddenLevels]
  );

  const borderColor = isDark ? 'gray.700' : 'gray.200';
  const panelBg = isDark ? 'gray.800' : 'white';

  if (!isOpen) {
    return (
      <Flex
        w="40px"
        minH="100%"
        bg={isDark ? 'gray.800' : 'gray.50'}
        border="1px solid"
        borderColor={borderColor}
        borderRadius="md"
        cursor="pointer"
        onClick={onOpen}
        alignItems="center"
        justifyContent="center"
        flexShrink={0}
        title="Open logs"
        _hover={{ bg: isDark ? 'gray.700' : 'gray.100' }}
        transition="background 0.2s"
      >
        <MdChevronLeft size={20} color={isDark ? '#a0aec0' : '#718096'} />
      </Flex>
    );
  }

  return (
    <Flex
      direction="column"
      w={{ base: '100%', lg: '100%' }}
      border="1px solid"
      borderColor={borderColor}
      borderRadius="md"
      bg={panelBg}
      overflow="hidden"
      h="100%"
    >
      {/* Header */}
      <HStack
        px={3}
        py={2}
        borderBottom="1px solid"
        borderColor={borderColor}
        justify="space-between"
        flexShrink={0}
      >
        <HStack spacing={2}>
          <Text fontWeight="semibold" fontSize="sm">
            Logs
          </Text>
          <Select
            size="xs"
            value={selectedChannel}
            onChange={(e) => setSelectedChannel(e.target.value)}
            w="auto"
            minW="100px"
            borderColor={borderColor}
          >
            <option value="">All channels</option>
            {channels.map((ch) => (
              <option key={ch} value={ch}>
                {ch}
              </option>
            ))}
          </Select>
        </HStack>
        <IconButton
          icon={<MdClose size={14} />}
          size="xs"
          variant="ghost"
          colorScheme="gray"
          onClick={onClose}
          aria-label="Close logs"
          title="Close logs"
        />
      </HStack>

      {/* Level filters */}
      <Box px={3} py={2} borderBottom="1px solid" borderColor={borderColor} flexShrink={0}>
        <LogLevelFilters hiddenLevels={hiddenLevels} onToggle={toggleLevel} />
      </Box>

      {/* Log list – each row skips layout/paint when off-screen via content-visibility */}
      <Box flex={1} overflowY="auto" minH={0} px={2} py={1}>
        {isLoading ? (
          <Flex h="80px" alignItems="center" justifyContent="center">
            <Spinner size="sm" color="brand.400" />
          </Flex>
        ) : filteredLogs.length === 0 ? (
          <Flex h="80px" alignItems="center" justifyContent="center">
            <Text fontSize="sm" color="gray.500">
              No logs to display
            </Text>
          </Flex>
        ) : (
          <VStack spacing={0.5} alignItems="stretch">
            {filteredLogs.map((log, i) => (
              <LogEntry
                key={`${i}-${log.time ?? ''}-${log.level}-${log.indicator ?? ''}`}
                log={log}
                isDark={isDark}
              />
            ))}
          </VStack>
        )}
      </Box>
    </Flex>
  );
}

