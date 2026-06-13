import { useRef, useCallback, useMemo, useEffect, useState } from 'react';
import type { MouseEvent as ReactMouseEvent } from 'react';
import ReactDOM from 'react-dom';
import Draggable from 'react-draggable';
import type { DraggableData, DraggableEvent } from 'react-draggable';
import {
  Box,
  Flex,
  HStack,
  Text,
  IconButton,
  Divider,
  useColorMode,
} from '@chakra-ui/react';
import { MdExpandLess, MdExpandMore, MdCallReceived } from 'react-icons/md';
import { useUIStore } from '../../store/uiStore';
import WidgetCard from './WidgetCard';
import TableWidget from './TableWidget';
import PieChartWidget from './PieChartWidget';
import type {
  Indicator,
  PlotPointCollection,
  RenderResponse,
  TableIndicatorConfig,
  PieChartIndicatorConfig,
} from '../../types/api';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Return the last valid PlotPointCollection from a plotPoints array, or null. */
function getLastValidCollection(plotPoints: PlotPointCollection[]): PlotPointCollection | null {
  if (!plotPoints) return null;
  for (let i = plotPoints.length - 1; i >= 0; i--) {
    if (!plotPoints[i].invalid) return plotPoints[i];
  }
  return null;
}

const FLOATING_HEADER_H = 36; // card header height in px
const FLOATING_MIN_W = 220;
const FLOATING_MIN_H = 160;
const FLOATING_MIN_CONTENT_H = 60;
const FLOATING_DEFAULT_W = 300;
const FLOATING_DEFAULT_H = 280;
const FLOATING_DEFAULT_X = 100;
const FLOATING_DEFAULT_Y = 100;

interface WidgetBodyProps {
  indicator: Indicator;
  collection: PlotPointCollection | null;
  height?: number;
}

/** Render the correct widget body based on indicator type. */
function WidgetBody({ indicator, collection, height }: WidgetBodyProps) {
  switch (indicator.type) {
    case 'table':
      return (
        <TableWidget
          collection={collection}
          config={indicator.config as TableIndicatorConfig | undefined}
        />
      );
    case 'pieChart':
      return (
        <PieChartWidget
          collection={collection}
          config={indicator.config as PieChartIndicatorConfig | undefined}
          height={height}
        />
      );
    default:
      return null;
  }
}

// ---------------------------------------------------------------------------
// Floating widget portal
// ---------------------------------------------------------------------------

interface FloatingWidgetProps {
  indicator: Indicator;
  collection: PlotPointCollection | null;
  position: { x: number; y: number };
  size: { w: number; h: number };
  onSnapBack: () => void;
  onStop: (x: number, y: number) => void;
  onResize: (w: number, h: number) => void;
}

function FloatingWidget({
  indicator,
  collection,
  position,
  size,
  onSnapBack,
  onStop,
  onResize,
}: FloatingWidgetProps) {
  const nodeRef = useRef<HTMLDivElement>(null);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const { w, h } = size;
  const contentH = Math.max(FLOATING_MIN_CONTENT_H, h - FLOATING_HEADER_H);

  // Controlled drag position — stays in sync with the store (e.g. after resize clamping)
  const [dragPos, setDragPos] = useState({ x: position.x, y: position.y });
  useEffect(() => {
    setDragPos({ x: position.x, y: position.y });
  }, [position.x, position.y]);

  // Track live viewport size so drag bounds update when the window is resized
  const [viewport, setViewport] = useState(() => ({ w: window.innerWidth, h: window.innerHeight }));
  useEffect(() => {
    const onResize = () => setViewport({ w: window.innerWidth, h: window.innerHeight });
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  // Constrain dragging so the widget (at minimum its header) stays fully on-screen
  const bounds = {
    left: 0,
    top: 0,
    right: Math.max(0, viewport.w - w),
    bottom: Math.max(0, viewport.h - FLOATING_HEADER_H),
  };

  // Use refs to avoid stale closure in resize handler
  const wRef = useRef(w);
  const hRef = useRef(h);
  useEffect(() => { wRef.current = w; hRef.current = h; }, [w, h]);

  const handleResizeMouseDown = useCallback(
    (e: ReactMouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      const startX = e.clientX;
      const startY = e.clientY;
      const startW = wRef.current;
      const startH = hRef.current;

      const onMove = (moveEvt: MouseEvent) => {
        const newW = Math.max(FLOATING_MIN_W, startW + moveEvt.clientX - startX);
        const newH = Math.max(FLOATING_MIN_H, startH + moveEvt.clientY - startY);
        onResize(newW, newH);
      };
      const onUp = () => {
        window.removeEventListener('mousemove', onMove);
        window.removeEventListener('mouseup', onUp);
      };
      window.addEventListener('mousemove', onMove);
      window.addEventListener('mouseup', onUp);
    },
    [onResize]
  );

  const handleColor = isDark ? '#718096' : '#A0AEC0';

  return ReactDOM.createPortal(
    <Draggable
      handle=".drag-handle"
      position={dragPos}
      bounds={bounds}
      nodeRef={nodeRef}
      onDrag={(_e: DraggableEvent, data: DraggableData) => setDragPos({ x: data.x, y: data.y })}
      onStop={(_e: DraggableEvent, data: DraggableData) => {
        setDragPos({ x: data.x, y: data.y });
        onStop(data.x, data.y);
      }}
    >
      <Box
        ref={nodeRef}
        position="fixed"
        top={0}
        left={0}
        zIndex={9999}
        w={`${w}px`}
      >
        <WidgetCard
          title={
            (indicator.config as TableIndicatorConfig | PieChartIndicatorConfig | undefined)
              ?.displayName || indicator.name
          }
          floating
          onSnapBack={onSnapBack}
          contentHeight={contentH}
        >
          <WidgetBody indicator={indicator} collection={collection} height={contentH} />
        </WidgetCard>

        {/* Resize handle — bottom-right corner */}
        <Box
          as="svg"
          viewBox="0 0 16 16"
          position="absolute"
          bottom="4px"
          right="4px"
          w="16px"
          h="16px"
          cursor="se-resize"
          zIndex={10001}
          onMouseDown={handleResizeMouseDown}
          opacity={0.6}
          _hover={{ opacity: 1 }}
          sx={{ color: handleColor, pointerEvents: 'all' }}
        >
          <line x1="4" y1="16" x2="16" y2="4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          <line x1="9" y1="16" x2="16" y2="9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          <line x1="14" y1="16" x2="16" y2="14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </Box>
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

interface IndicatorPanelProps {
  renderData: RenderResponse | null;
  projectId: string;
}

/**
 * IndicatorPanel – collapsible, resizable bottom drawer that shows table and
 * pieChart indicator widgets. Other chart-rendered types are excluded.
 *
 * Props:
 *   renderData – ProjectRenderResponse from the API
 *   projectId  – used to scope floating widget state
 */
export default function IndicatorPanel({ renderData, projectId }: IndicatorPanelProps) {
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

  // Track viewport so percentage-based positions convert to correct pixels on resize
  const [viewport, setViewport] = useState(() => ({ w: window.innerWidth, h: window.innerHeight }));
  useEffect(() => {
    const onResize = () => setViewport({ w: window.innerWidth, h: window.innerHeight });
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  // ---------------------------------------------------------------------------
  // Resize handle
  // ---------------------------------------------------------------------------
  const handleResizeMouseDown = useCallback(
    (e: ReactMouseEvent) => {
      e.preventDefault();
      const startY = e.clientY;
      const startH = panelHeight;

      const onMove = (moveEvt: MouseEvent) => {
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
    (indicatorName: string) => {
      if (floatingForProject.find((w) => w.id === indicatorName)) return;
      setFloatingWidgets(projectId, [
        ...floatingForProject,
        {
          id: indicatorName,
          xPct: FLOATING_DEFAULT_X / viewport.w,
          yPct: FLOATING_DEFAULT_Y / viewport.h,
          w: FLOATING_DEFAULT_W,
          h: FLOATING_DEFAULT_H,
        },
      ]);
    },
    [floatingForProject, projectId, setFloatingWidgets, viewport]
  );

  const snapBackWidget = useCallback(
    (indicatorName: string) => {
      setFloatingWidgets(
        projectId,
        floatingForProject.filter((w) => w.id !== indicatorName)
      );
    },
    [floatingForProject, projectId, setFloatingWidgets]
  );

  // Pixels come in from the drag; convert to viewport fractions before storing
  const updateWidgetPosition = useCallback(
    (indicatorName: string, x: number, y: number) => {
      setFloatingWidgets(
        projectId,
        floatingForProject.map((w) =>
          w.id === indicatorName
            ? { ...w, xPct: x / viewport.w, yPct: y / viewport.h }
            : w
        )
      );
    },
    [floatingForProject, projectId, setFloatingWidgets, viewport]
  );

  const updateWidgetSize = useCallback(
    (indicatorName: string, w: number, h: number) => {
      setFloatingWidgets(
        projectId,
        floatingForProject.map((fw) => (fw.id === indicatorName ? { ...fw, w, h } : fw))
      );
    },
    [floatingForProject, projectId, setFloatingWidgets]
  );

  const snapBackAllWidgets = useCallback(() => {
    setFloatingWidgets(projectId, []);
  }, [projectId, setFloatingWidgets]);

  // ---------------------------------------------------------------------------
  // Group widgets by targetChart
  // ---------------------------------------------------------------------------
  const groups: Record<string, Indicator[]> = {};
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
    const w = fw.w ?? FLOATING_DEFAULT_W;
    const h = fw.h ?? FLOATING_DEFAULT_H;
    // Convert stored fractions to pixels, clamping so the widget stays on-screen
    // even if it is wider/taller than the current viewport.
    const x = Math.min(Math.round(fw.xPct * viewport.w), Math.max(0, viewport.w - w));
    const y = Math.min(Math.round(fw.yPct * viewport.h), Math.max(0, viewport.h - FLOATING_HEADER_H));
    return (
      <FloatingWidget
        key={fw.id}
        indicator={ind}
        collection={collection}
        position={{ x, y }}
        size={{ w, h }}
        onSnapBack={() => snapBackWidget(fw.id)}
        onStop={(px, py) => updateWidgetPosition(fw.id, px, py)}
        onResize={(nw, nh) => updateWidgetSize(fw.id, nw, nh)}
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
        <HStack spacing={1}>
          {floatingForProject.length > 0 && (
            <IconButton
              icon={<MdCallReceived size={14} />}
              size="xs"
              variant="ghost"
              aria-label="Recover all floating widgets"
              title="Recover all floating widgets"
              onClick={snapBackAllWidgets}
            />
          )}
          <IconButton
            icon={isOpen ? <MdExpandMore size={18} /> : <MdExpandLess size={18} />}
            size="xs"
            variant="ghost"
            aria-label={isOpen ? 'Collapse indicator panel' : 'Expand indicator panel'}
            onClick={() => setOpen(!isOpen)}
          />
        </HStack>
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
                      const indConfig = ind.config as
                        | TableIndicatorConfig
                        | PieChartIndicatorConfig
                        | undefined;
                      const title = indConfig?.displayName || ind.name;
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
