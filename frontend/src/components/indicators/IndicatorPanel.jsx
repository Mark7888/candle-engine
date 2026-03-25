import { useRef, useCallback, useMemo } from 'react';
import ReactDOM from 'react-dom';
import Draggable from 'react-draggable';
import {
  Box,
  Flex,
  HStack,
  Text,
  IconButton,
  Divider,
  useColorMode,
} from '@chakra-ui/react';
import { MdExpandLess, MdExpandMore } from 'react-icons/md';
import { useUIStore } from '../../store/uiStore';
import WidgetCard from './WidgetCard';
import TableWidget from './TableWidget';
import PieChartWidget from './PieChartWidget';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Return the last valid PlotPointCollection from a plotPoints array, or null. */
function getLastValidCollection(plotPoints) {
  if (!plotPoints) return null;
  for (let i = plotPoints.length - 1; i >= 0; i--) {
    if (!plotPoints[i].invalid) return plotPoints[i];
  }
  return null;
}

/** Render the correct widget body based on indicator type. */
function WidgetBody({ indicator, collection }) {
  switch (indicator.type) {
    case 'table':
      return <TableWidget collection={collection} config={indicator.config} />;
    case 'pieChart':
      return <PieChartWidget collection={collection} config={indicator.config} />;
    default:
      return null;
  }
}

// ---------------------------------------------------------------------------
// Floating widget portal
// ---------------------------------------------------------------------------

function FloatingWidget({ indicator, collection, position, onSnapBack, onStop }) {
  const nodeRef = useRef(null);
  return ReactDOM.createPortal(
    <Draggable
      handle=".drag-handle"
      defaultPosition={{ x: position.x, y: position.y }}
      nodeRef={nodeRef}
      onStop={(_, data) => onStop(data.x, data.y)}
    >
      <Box
        ref={nodeRef}
        position="fixed"
        top={0}
        left={0}
        zIndex={9999}
        minW="220px"
        maxW="400px"
      >
        <WidgetCard
          title={indicator.config?.displayName || indicator.name}
          floating
          onSnapBack={onSnapBack}
        >
          <WidgetBody indicator={indicator} collection={collection} />
        </WidgetCard>
      </Box>
    </Draggable>,
    document.body
  );
}

// ---------------------------------------------------------------------------
// Main IndicatorPanel
// ---------------------------------------------------------------------------

const MIN_HEIGHT = 120;
const MAX_HEIGHT = 600;
const HEADER_HEIGHT = 36;

/**
 * IndicatorPanel – collapsible, resizable bottom drawer that shows table and
 * pieChart indicator widgets. Other chart-rendered types are excluded.
 *
 * Props:
 *   renderData – ProjectRenderResponse from the API
 *   projectId  – used to scope floating widget state
 */
export default function IndicatorPanel({ renderData, projectId }) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const isOpen = useUIStore((s) => s.indicatorPanelOpen);
  const setOpen = useUIStore((s) => s.setIndicatorPanelOpen);
  const panelHeight = useUIStore((s) => s.indicatorPanelHeight);
  const setPanelHeight = useUIStore((s) => s.setIndicatorPanelHeight);
  const floatingWidgets = useUIStore((s) => s.floatingWidgets);
  const setFloatingWidgets = useUIStore((s) => s.setFloatingWidgets);

  const panelRef = useRef(null);

  // Drawer-only indicators
  const drawerIndicators = (renderData?.indicators ?? []).filter(
    (ind) => ind.type === 'table' || ind.type === 'pieChart'
  );

  const floatingForProject = useMemo(
    () => floatingWidgets[projectId] ?? [],
    [floatingWidgets, projectId]
  );

  // ---------------------------------------------------------------------------
  // Resize handle
  // ---------------------------------------------------------------------------
  const handleResizeMouseDown = useCallback(
    (e) => {
      e.preventDefault();
      const startY = e.clientY;
      const startH = panelHeight;

      const onMove = (moveEvt) => {
        const delta = startY - moveEvt.clientY;
        const clamped = Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, startH + delta));
        setPanelHeight(clamped);
      };

      const onUp = () => {
        window.removeEventListener('mousemove', onMove);
        window.removeEventListener('mouseup', onUp);
      };

      window.addEventListener('mousemove', onMove);
      window.addEventListener('mouseup', onUp);
    },
    [panelHeight, setPanelHeight]
  );

  // ---------------------------------------------------------------------------
  // Float / snap-back
  // ---------------------------------------------------------------------------
  const floatWidget = useCallback(
    (indicatorName) => {
      const existing = floatingForProject.find((w) => w.id === indicatorName);
      if (existing) return;
      setFloatingWidgets(projectId, [
        ...floatingForProject,
        { id: indicatorName, x: 100, y: 100 },
      ]);
    },
    [floatingForProject, projectId, setFloatingWidgets]
  );

  const snapBackWidget = useCallback(
    (indicatorName) => {
      setFloatingWidgets(
        projectId,
        floatingForProject.filter((w) => w.id !== indicatorName)
      );
    },
    [floatingForProject, projectId, setFloatingWidgets]
  );

  const updateWidgetPosition = useCallback(
    (indicatorName, x, y) => {
      setFloatingWidgets(
        projectId,
        floatingForProject.map((w) => (w.id === indicatorName ? { ...w, x, y } : w))
      );
    },
    [floatingForProject, projectId, setFloatingWidgets]
  );

  // ---------------------------------------------------------------------------
  // Group widgets by targetChart
  // ---------------------------------------------------------------------------
  const groups = {};
  for (const ind of drawerIndicators) {
    const key = ind.targetChart ?? '';
    if (!groups[key]) groups[key] = [];
    groups[key].push(ind);
  }
  const groupKeys = Object.keys(groups).sort((a, b) => {
    if (a === '') return -1;
    if (b === '') return 1;
    return a.localeCompare(b);
  });

  // ---------------------------------------------------------------------------
  // Styling
  // ---------------------------------------------------------------------------
  const borderColor = isDark ? 'gray.700' : 'gray.200';
  const headerBg = isDark ? 'gray.800' : 'white';
  const contentBg = isDark ? 'gray.900' : 'gray.50';

  // ---------------------------------------------------------------------------
  // Floating portals (always rendered regardless of panel open state)
  // ---------------------------------------------------------------------------
  const floatingPortals = floatingForProject.map((fw) => {
    const ind = drawerIndicators.find((i) => i.name === fw.id);
    if (!ind) return null;
    const collection = getLastValidCollection(ind.plotPoints);
    return (
      <FloatingWidget
        key={fw.id}
        indicator={ind}
        collection={collection}
        position={{ x: fw.x, y: fw.y }}
        onSnapBack={() => snapBackWidget(fw.id)}
        onStop={(x, y) => updateWidgetPosition(fw.id, x, y)}
      />
    );
  });

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------
  return (
    <Box
      ref={panelRef}
      flexShrink={0}
      borderTop="1px solid"
      borderColor={borderColor}
      bg={headerBg}
    >
      {/* Floating windows — always visible */}
      {floatingPortals}

      {/* Resize handle — only when open */}
      {isOpen && (
        <Box
          h="6px"
          cursor="row-resize"
          bg={isDark ? 'gray.700' : 'gray.200'}
          _hover={{ bg: isDark ? 'gray.600' : 'gray.300' }}
          onMouseDown={handleResizeMouseDown}
        />
      )}

      {/* Header bar — always visible */}
      <HStack
        h={`${HEADER_HEIGHT}px`}
        px={3}
        justify="space-between"
        borderBottom={isOpen ? '1px solid' : 'none'}
        borderColor={borderColor}
        flexShrink={0}
      >
        <Text fontSize="sm" fontWeight="semibold">
          Indicators
        </Text>
        <IconButton
          icon={isOpen ? <MdExpandMore size={18} /> : <MdExpandLess size={18} />}
          size="xs"
          variant="ghost"
          aria-label={isOpen ? 'Collapse indicator panel' : 'Expand indicator panel'}
          onClick={() => setOpen(!isOpen)}
        />
      </HStack>

      {/* Content — only when open */}
      {isOpen && (
        <Box
          h={`${panelHeight}px`}
          overflowX="auto"
          overflowY="auto"
          bg={contentBg}
          p={2}
        >
          {drawerIndicators.length === 0 ? (
            <Flex h="100%" alignItems="center" justifyContent="center">
              <Text fontSize="sm" color="gray.500">
                No drawer indicators for this project
              </Text>
            </Flex>
          ) : (
            groupKeys.map((key, gi) => {
              const groupIndicators = groups[key];
              const groupLabel = key === '' ? 'Main' : key;
              // Filter out floating indicators from the drawer
              const dockedIndicators = groupIndicators.filter(
                (ind) => !floatingForProject.find((fw) => fw.id === ind.name)
              );
              return (
                <Box key={key} mb={gi < groupKeys.length - 1 ? 3 : 0}>
                  <HStack mb={2} spacing={2}>
                    <Divider borderColor={borderColor} />
                    <Text
                      fontSize="xs"
                      color="gray.500"
                      whiteSpace="nowrap"
                      flexShrink={0}
                    >
                      {groupLabel}
                    </Text>
                    <Divider borderColor={borderColor} />
                  </HStack>
                  <Flex flexWrap="wrap" gap={2}>
                    {dockedIndicators.map((ind) => {
                      const collection = getLastValidCollection(ind.plotPoints);
                      const title = ind.config?.displayName || ind.name;
                      return (
                        <WidgetCard
                          key={ind.name}
                          title={title}
                          floating={false}
                          onFloat={() => floatWidget(ind.name)}
                        >
                          <WidgetBody indicator={ind} collection={collection} />
                        </WidgetCard>
                      );
                    })}
                  </Flex>
                </Box>
              );
            })
          )}
        </Box>
      )}
    </Box>
  );
}
