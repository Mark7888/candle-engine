import { useState, useMemo, useRef, useLayoutEffect, useCallback } from 'react';
import {
  Box,
  Flex,
  HStack,
  Text,
  Select,
  IconButton,
  Spinner,
  useColorMode,
} from '@chakra-ui/react';
import { MdChevronLeft, MdClose } from 'react-icons/md';
import LogEntry from './LogEntry';
import LogLevelFilters from './LogLevelFilters';
import { useUIStore } from '../../store/uiStore';
import type { LogEntry as LogEntryType } from '../../types/api';

const ITEM_HEIGHT = 24;
const BUFFER = 5;

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
  const [scrollTop, setScrollTop] = useState(0);
  const [containerHeight, setContainerHeight] = useState(400);
  const containerRef = useRef<HTMLDivElement>(null);

  const hiddenLevels = useUIStore((s) => s.hiddenLogLevels);
  const toggleLevel = useUIStore((s) => s.toggleLogLevel);

  useLayoutEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const observer = new ResizeObserver(() => {
      const h = el.clientHeight;
      setContainerHeight((prev) => (prev !== h ? h : prev));
    });
    observer.observe(el);
    setContainerHeight(el.clientHeight);
    return () => observer.disconnect();
  }, []);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  }, []);

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

  const startIdx = Math.max(0, Math.floor(scrollTop / ITEM_HEIGHT) - BUFFER);
  const endIdx = Math.min(
    filteredLogs.length,
    Math.ceil((scrollTop + containerHeight) / ITEM_HEIGHT) + BUFFER
  );
  const visibleItems = filteredLogs.slice(startIdx, endIdx);
  const totalHeight = filteredLogs.length * ITEM_HEIGHT;

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

      {/* Log list – virtual scrolling renders only visible rows */}
      <Box
        ref={containerRef}
        flex={1}
        overflowY="auto"
        minH={0}
        px={2}
        py={1}
        onScroll={handleScroll}
      >
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
          <Box position="relative" height={`${totalHeight}px`}>
            <Box position="absolute" top={`${startIdx * ITEM_HEIGHT}px`} left={0} right={0}>
              {visibleItems.map((log, i) => (
                <LogEntry key={`${startIdx + i}-${log.time ?? ''}-${log.level}-${log.indicator ?? ''}`} log={log} isDark={isDark} />
              ))}
            </Box>
          </Box>
        )}
      </Box>
    </Flex>
  );
}
