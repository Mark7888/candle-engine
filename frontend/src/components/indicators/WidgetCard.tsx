import { Box, Flex, HStack, Text, IconButton, useColorMode } from '@chakra-ui/react';
import { MdOpenInNew, MdCallReceived, MdPushPin } from 'react-icons/md';
import type { ReactNode } from 'react';

interface WidgetCardProps {
  title: string;
  floating?: boolean;
  pinned?: boolean;
  onFloat?: () => void;
  onSnapBack?: () => void;
  onTogglePin?: () => void;
  children?: ReactNode;
  contentHeight?: number;
}

export default function WidgetCard({
  title,
  floating = false,
  pinned = false,
  onFloat,
  onSnapBack,
  onTogglePin,
  children,
  contentHeight,
}: WidgetCardProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const bg = isDark ? 'gray.800' : 'white';
  const borderColor = isDark ? 'gray.700' : 'gray.200';

  return (
    <Box
      bg={bg}
      border="1px solid"
      borderColor={borderColor}
      borderRadius="md"
      minW="220px"
      boxShadow={floating ? 'lg' : 'none'}
      overflow="hidden"
      display="flex"
      flexDirection="column"
    >
      {/* Header */}
      <Flex
        px={3}
        py={1}
        borderBottom="1px solid"
        borderColor={borderColor}
        align="center"
        justify="space-between"
        flexShrink={0}
        className={floating && !pinned ? 'drag-handle' : undefined}
        cursor={floating && !pinned ? 'grab' : 'default'}
        userSelect="none"
      >
        <Text fontSize="xs" fontWeight="semibold" noOfLines={1}>
          {title}
        </Text>
        {floating ? (
          <HStack spacing={0}>
            <IconButton
              icon={<MdPushPin size={12} />}
              size="xs"
              variant="ghost"
              aria-label={pinned ? 'Unpin widget' : 'Pin widget'}
              title={pinned ? 'Unpin widget' : 'Pin widget'}
              color={pinned ? 'blue.400' : undefined}
              onClick={onTogglePin}
            />
            <IconButton
              icon={<MdCallReceived size={12} />}
              size="xs"
              variant="ghost"
              aria-label="Return to drawer"
              title="Return to drawer"
              onClick={onSnapBack}
            />
          </HStack>
        ) : (
          <IconButton
            icon={<MdOpenInNew size={12} />}
            size="xs"
            variant="ghost"
            aria-label="Float widget"
            title="Float widget"
            onClick={onFloat}
          />
        )}
      </Flex>

      {/* Body */}
      <Box
        overflowY="auto"
        flex={contentHeight !== undefined ? undefined : 1}
        h={contentHeight !== undefined ? `${contentHeight}px` : undefined}
      >
        {children}
      </Box>
    </Box>
  );
}
