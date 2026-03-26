import { Box } from '@chakra-ui/react';
import { keyframes } from '@emotion/react';
import { useColorMode } from '@chakra-ui/react';
import { MdLightMode, MdDarkMode } from 'react-icons/md';

const spinIn = keyframes`
  from { opacity: 0; transform: scale(0.5) rotate(-180deg); }
  to   { opacity: 1; transform: scale(1)   rotate(0deg); }
`;

const spinOut = keyframes`
  from { opacity: 1; transform: scale(1)   rotate(0deg); }
  to   { opacity: 0; transform: scale(0.5) rotate(180deg); }
`;

export default function ThemeToggle() {
  const { colorMode, toggleColorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Box
      as="button"
      onClick={toggleColorMode}
      position="relative"
      w={6}
      h={6}
      flexShrink={0}
      aria-label="Toggle dark mode"
      title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {/* Moon – visible in light mode */}
      <Box
        position="absolute"
        inset={0}
        display="flex"
        alignItems="center"
        justifyContent="center"
        animation={
          isDark
            ? `${spinOut} 0.4s ease forwards`
            : `${spinIn} 0.4s ease forwards`
        }
        color={isDark ? 'transparent' : 'gray.600'}
      >
        <MdDarkMode size={22} />
      </Box>

      {/* Sun – visible in dark mode */}
      <Box
        position="absolute"
        inset={0}
        display="flex"
        alignItems="center"
        justifyContent="center"
        animation={
          isDark
            ? `${spinIn} 0.4s ease forwards`
            : `${spinOut} 0.4s ease forwards`
        }
        color={isDark ? 'yellow.300' : 'transparent'}
      >
        <MdLightMode size={22} />
      </Box>
    </Box>
  );
}
