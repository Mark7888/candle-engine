import { Box, Flex, Text, IconButton, useColorMode } from '@chakra-ui/react';
import { MdOpenInNew, MdCallReceived } from 'react-icons/md';

/**
 * WidgetCard – shared wrapper for all drawer widget types.
 *
 * Props:
 *   title      – display name shown in the card header
 *   floating   – whether this card is currently floating (affects styling + icon)
 *   onFloat    – callback to pop out the card as a floating window
 *   onSnapBack – callback to return the card to the drawer
 *   children   – widget body content
 */
export default function WidgetCard({ title, floating = false, onFloat, onSnapBack, children, contentHeight }) {
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
        className={floating ? 'drag-handle' : undefined}
        cursor={floating ? 'grab' : 'default'}
        userSelect="none"
      >
        <Text fontSize="xs" fontWeight="semibold" noOfLines={1}>
          {title}
        </Text>
        {floating ? (
          <IconButton
            icon={<MdCallReceived size={12} />}
            size="xs"
            variant="ghost"
            aria-label="Return to drawer"
            title="Return to drawer"
            onClick={onSnapBack}
          />
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
