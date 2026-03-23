import {
  Flex,
  HStack,
  Text,
  Button,
  Spinner,
  useColorModeValue,
} from '@chakra-ui/react';
import ThemeToggle from './ThemeToggle';

/**
 * Top navigation bar.
 * Left: project tab buttons (horizontal).
 * Right: dark-mode toggle.
 */
export default function Navbar({
  projects,
  loading,
  activeProjectId,
  onSelectProject,
}) {
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const activeColor = useColorModeValue('brand.600', 'brand.400');
  const inactiveColor = useColorModeValue('gray.500', 'gray.400');
  const navBg = useColorModeValue('white', 'gray.900');
  const hoverColor = useColorModeValue('gray.700', 'gray.200');
  const hoverBorderColor = useColorModeValue('gray.300', 'gray.600');

  return (
    <Flex
      as="nav"
      position="sticky"
      top={0}
      zIndex={100}
      px={6}
      borderBottom="1px solid"
      borderColor={borderColor}
      bg={navBg}
      h="52px"
      alignItems="center"
      justify="space-between"
    >
      {/* Left: project tabs */}
      <HStack spacing={0} flex={1} minW={0} overflowX="auto" className="hide-scrollbar">
        {loading ? (
          <HStack spacing={2} py={2}>
            <Spinner size="sm" color="brand.400" />
            <Text fontSize="sm" color="gray.500">
              Loading projects…
            </Text>
          </HStack>
        ) : (
          projects.map((project) => {
            const isActive = project.id === activeProjectId;
            return (
              <Button
                key={project.id}
                variant="unstyled"
                px={4}
                h="52px"
                borderBottom="2px solid"
                borderColor={isActive ? activeColor : 'transparent'}
                color={isActive ? activeColor : inactiveColor}
                fontWeight={isActive ? 'semibold' : 'medium'}
                fontSize="sm"
                borderRadius={0}
                flexShrink={0}
                onClick={() => onSelectProject(project.id)}
                _hover={{
                  color: hoverColor,
                  borderColor: hoverBorderColor,
                }}
                transition="border-color 0.15s, color 0.15s"
              >
                {project.title}
              </Button>
            );
          })
        )}
      </HStack>

      {/* Right: theme toggle */}
      <HStack spacing={3} ml={4} flexShrink={0}>
        <ThemeToggle />
      </HStack>
    </Flex>
  );
}
