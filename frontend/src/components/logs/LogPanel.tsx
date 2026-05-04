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

const ROW_HEIGHT = 24;
const EXPANDED_LINE_HEIGHT = 18;
// ~characters that fit on one line of the expanded message area
const CHARS_PER_LINE = 44;
const EXPANDED_PADDING = 8;
const BUFFER = 8;

function rowKey(log: LogEntryType): string {
  return `${log.time ?? ''}\x00${log.level}\x00${log.indicator ?? ''}\x00${(log.message ?? '').slice(0, 32)}`;
}

function getRowHeight(log: LogEntryType, expanded: boolean): number {
  if (!expanded) return ROW_HEIGHT;
  const lines = Math.max(1, Math.ceil((log.message?.length ?? 0) / CHARS_PER_LINE));
  return ROW_HEIGHT + lines * EXPANDED_LINE_HEIGHT + EXPANDED_PADDING;
}

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
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());
  const containerRef = useRef<HTMLDivElement>(null);

  const hiddenLevels = useUIStore((s) => s.hiddenLogLevels);
  const toggleLevel = useUIStore((s) => s.toggleLogLevel);

  useLayoutEffect(() => {
    if (!isOpen) {
      // Reset before next open so the first render already has scrollTop = 0.
      setScrollTop(0);
      return;
    }
    const el = containerRef.current;
    if (!el) return;
    el.scrollTop = 0;
    const observer = new ResizeObserver(() => {
      const h = el.clientHeight;
      setContainerHeight((prev) => (prev !== h ? h : prev));
    });
    observer.observe(el);
    setContainerHeight(el.clientHeight);
    return () => observer.disconnect();
  }, [isOpen]);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  }, []);

  const toggleRow = useCallback((key: string) => {
    setExpandedKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
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

  const offsets = useMemo(() => {
    const arr = new Array<number>(filteredLogs.length + 1);
    arr[0] = 0;
    for (let i = 0; i < filteredLogs.length; i++) {
      arr[i + 1] = arr[i] + getRowHeight(filteredLogs[i], expandedKeys.has(rowKey(filteredLogs[i])));
    }
    return arr;
  }, [filteredLogs, expandedKeys]);

  const totalHeight = offsets[filteredLogs.length] ?? 0;

  const rawStart = useMemo(() => {
    let lo = 0;
    let hi = filteredLogs.length;
    while (lo < hi) {
      const mid = (lo + hi) >> 1;
      if (offsets[mid + 1] <= scrollTop) lo = mid + 1;
      else hi = mid;
    }
    return lo;
  }, [offsets, scrollTop, filteredLogs.length]);

  const startIdx = Math.max(0, rawStart - BUFFER);

  const endIdx = useMemo(() => {
    const visBottom = scrollTop + containerHeight;
    let i = rawStart;
    while (i < filteredLogs.length && offsets[i] < visBottom) i++;
    return Math.min(filteredLogs.length, i + BUFFER);
  }, [offsets, rawStart, scrollTop, containerHeight, filteredLogs.length]);

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

      <Box px={3} py={2} borderBottom="1px solid" borderColor={borderColor} flexShrink={0}>
        <LogLevelFilters hiddenLevels={hiddenLevels} onToggle={toggleLevel} />
      </Box>

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
            <Box position="absolute" top={`${offsets[startIdx]}px`} left={0} right={0}>
              {filteredLogs.slice(startIdx, endIdx).map((log, i) => {
                const absIdx = startIdx + i;
                const key = rowKey(log);
                return (
                  <LogEntry
                    key={String(absIdx)}
                    log={log}
                    isDark={isDark}
                    isExpanded={expandedKeys.has(key)}
                    onToggle={() => toggleRow(key)}
                  />
                );
              })}
            </Box>
          </Box>
        )}
      </Box>
    </Flex>
  );
}

