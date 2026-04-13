import { useState, useCallback, type ChangeEvent } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  FormControl,
  FormLabel,
  HStack,
  Input,
  Box,
  Divider,
  Text,
  useColorModeValue,
} from '@chakra-ui/react';
import { useProjectSettingsStore } from '../../store/projectSettingsStore';
import type { ProjectSettingsConfig } from '../../store/defaultSettings';

interface ProjectSettingsModalProps {
  projectId: number;
  isOpen: boolean;
  onClose: () => void;
}

/**
 * ProjectSettingsModal – a project-scoped settings dialog.
 *
 * The modal occupies 60 % of the viewport (20 % inset on each side) and
 * renders a semi-transparent overlay for the remaining area.  Clicking
 * Cancel discards any unsaved edits; Save persists them to localStorage via
 * the projectSettingsStore.
 */
export default function ProjectSettingsModal({
  projectId,
  isOpen,
  onClose,
}: ProjectSettingsModalProps) {
  const getSettings = useProjectSettingsStore((s) => s.getSettings);
  const saveSettings = useProjectSettingsStore((s) => s.saveSettings);

  // Draft is initialised from the persisted settings when the component mounts.
  // SettingsBar passes a new `key` prop every time the modal is opened, which
  // remounts this component and re-runs this initialiser – no effect needed.
  const [draft, setDraft] = useState<ProjectSettingsConfig>(() =>
    getSettings(projectId)
  );

  const labelColor = useColorModeValue('gray.600', 'gray.400');
  const dividerColor = useColorModeValue('gray.200', 'gray.600');
  const sectionBg = useColorModeValue('gray.50', 'gray.750');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  const handleSave = useCallback(() => {
    saveSettings(projectId, draft);
    onClose();
  }, [projectId, draft, saveSettings, onClose]);

  const handleCancel = useCallback(() => {
    setDraft(getSettings(projectId));
    onClose();
  }, [projectId, getSettings, onClose]);

  const handleColorInputChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      setDraft((d) => ({ ...d, testColor: e.target.value }));
    },
    []
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleCancel}
      isCentered
      scrollBehavior="inside"
    >
      <ModalOverlay bg="blackAlpha.600" />
      <ModalContent
        sx={{
          width: '60vw',
          maxWidth: '60vw',
          minHeight: '80vh',
          maxHeight: '80vh',
        }}
        display="flex"
        flexDirection="column"
      >
        <ModalHeader borderBottom="1px solid" borderColor={dividerColor} pb={3}>
          Project Settings
        </ModalHeader>

        <ModalBody flex={1} overflowY="auto" py={6} px={8}>
          {/* ── Appearance section ─────────────────────────────────── */}
          <Text
            fontSize="xs"
            fontWeight="semibold"
            textTransform="uppercase"
            letterSpacing="wider"
            color={labelColor}
            mb={3}
          >
            Appearance
          </Text>

          <Box
            bg={sectionBg}
            border="1px solid"
            borderColor={borderColor}
            borderRadius="md"
            px={5}
            py={4}
          >
            <FormControl>
              <FormLabel fontSize="sm" fontWeight="medium" mb={2}>
                Test Color Change
              </FormLabel>
              <HStack spacing={3}>
                {/* Native color swatch – wrapped in a Box so Chakra styles apply */}
                <Box
                  as="label"
                  w="36px"
                  h="36px"
                  borderRadius="md"
                  overflow="hidden"
                  border="1px solid"
                  borderColor={borderColor}
                  cursor="pointer"
                  flexShrink={0}
                >
                  <input
                    type="color"
                    value={draft.testColor}
                    onChange={handleColorInputChange}
                    style={{
                      width: '100%',
                      height: '100%',
                      border: 'none',
                      padding: 0,
                      cursor: 'pointer',
                      display: 'block',
                    }}
                  />
                </Box>

                {/* Hex value text input */}
                <Input
                  value={draft.testColor}
                  onChange={handleColorInputChange}
                  size="sm"
                  maxW="120px"
                  fontFamily="mono"
                  placeholder="#004040"
                />
              </HStack>
            </FormControl>
          </Box>
        </ModalBody>

        <Divider borderColor={dividerColor} />

        <ModalFooter gap={3}>
          <Button variant="ghost" size="sm" onClick={handleCancel}>
            Cancel
          </Button>
          <Button colorScheme="brand" size="sm" onClick={handleSave}>
            Save
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
}
