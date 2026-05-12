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
import type { CandleColors } from '../../types/api';

interface ProjectSettingsModalProps {
  projectId: string;
  isOpen: boolean;
  onClose: () => void;
  /** Current candle colors read from the backend-owned project detail. */
  candleColors: CandleColors;
  /** Called with the new colors when the user saves. */
  onColorsChange: (bullish: string, bearish: string) => void;
}

/**
 * ProjectSettingsModal – a project-scoped settings dialog.
 *
 * Visual settings (candle colors) are owned by the backend.  Clicking Save
 * calls onColorsChange, which the parent uses to PATCH the backend and
 * re-fetch the render.  Clicking Cancel discards any unsaved edits.
 */
export default function ProjectSettingsModal({
  isOpen,
  onClose,
  candleColors,
  onColorsChange,
}: ProjectSettingsModalProps) {
  // Draft is initialised from the current backend-owned colors each time the
  // modal opens (SettingsBar passes a new key every time it opens).
  const [draft, setDraft] = useState<CandleColors>(() => ({ ...candleColors }));

  const labelColor = useColorModeValue('gray.600', 'gray.400');
  const dividerColor = useColorModeValue('gray.200', 'gray.600');
  const sectionBg = useColorModeValue('gray.50', 'gray.750');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  const handleSave = useCallback(() => {
    onColorsChange(draft.bullishColor, draft.bearishColor);
    onClose();
  }, [draft, onColorsChange, onClose]);

  const handleCancel = useCallback(() => {
    setDraft({ ...candleColors });
    onClose();
  }, [candleColors, onClose]);

  const handleBullishColorChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      setDraft((d) => ({ ...d, bullishColor: e.target.value }));
    },
    []
  );

  const handleBearishColorChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      setDraft((d) => ({ ...d, bearishColor: e.target.value }));
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
          {/* ── Visual Settings section ─────────────────────────────── */}
          <Text
            fontSize="xs"
            fontWeight="semibold"
            textTransform="uppercase"
            letterSpacing="wider"
            color={labelColor}
            mb={3}
          >
            Visual Settings
          </Text>

          <Box
            bg={sectionBg}
            border="1px solid"
            borderColor={borderColor}
            borderRadius="md"
            px={5}
            py={4}
          >
            {/* Bullish candle color */}
            <FormControl mb={4}>
              <FormLabel fontSize="sm" fontWeight="medium" mb={2}>
                Bullish Candle Color
              </FormLabel>
              <HStack spacing={3}>
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
                    value={draft.bullishColor}
                    onChange={handleBullishColorChange}
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
                <Input
                  value={draft.bullishColor}
                  onChange={handleBullishColorChange}
                  size="sm"
                  maxW="120px"
                  fontFamily="mono"
                  placeholder="#26a69a"
                />
              </HStack>
            </FormControl>

            {/* Bearish candle color */}
            <FormControl>
              <FormLabel fontSize="sm" fontWeight="medium" mb={2}>
                Bearish Candle Color
              </FormLabel>
              <HStack spacing={3}>
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
                    value={draft.bearishColor}
                    onChange={handleBearishColorChange}
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
                <Input
                  value={draft.bearishColor}
                  onChange={handleBearishColorChange}
                  size="sm"
                  maxW="120px"
                  fontFamily="mono"
                  placeholder="#ef5350"
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
