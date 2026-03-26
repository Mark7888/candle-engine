import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Text,
  useColorMode,
} from '@chakra-ui/react';
import { normalizeColor } from '../chart/chartEngine';
import type { PlotPointCollection, TableIndicatorConfig } from '../../types/api';

interface TableWidgetProps {
  collection: PlotPointCollection | null;
  config?: TableIndicatorConfig;
}

/**
 * Parses a Go-style format string like "%.2f" and returns precision N.
 * Falls back to 4 if the string does not match.
 */
function parseFormatPrecision(fmt: string | undefined): number {
  if (!fmt) return 4;
  const match = fmt.match(/^%\.(\d+)f$/);
  return match ? parseInt(match[1], 10) : 4;
}

/**
 * TableWidget – renders a 2-D table indicator using the last valid PlotPointCollection.
 *
 * Config fields (camelCase from Go TableIndicatorConfig):
 *   displayName, rowTitles, columnTitles, valueDisplayFormat
 *
 * Layout: row-major — point at [row][col] = points[row * columnCount + col].
 */
export default function TableWidget({ collection, config }: TableWidgetProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const rowTitles = config?.rowTitles ?? [];
  const columnTitles = config?.columnTitles ?? [];
  const precision = parseFormatPrecision(config?.valueDisplayFormat);

  const columnCount = Math.max(columnTitles.length, 1);
  const hasRowHeaders = rowTitles.length > 0;
  const hasColHeaders = columnTitles.length > 0;
  const points = collection?.points ?? [];

  const headerBg = isDark ? 'gray.700' : 'gray.100';
  const headerColor = isDark ? 'gray.300' : 'gray.600';
  const borderColor = isDark ? 'gray.700' : 'gray.200';

  const renderValue = (pt: (typeof points)[number] | undefined) => {
    if (!pt || pt.invalid) return '—';
    if (typeof pt.value !== 'number' || !isFinite(pt.value)) return '—';
    const color = pt.color ? normalizeColor(pt.color) : undefined;
    return (
      <Text as="span" color={color}>
        {pt.value.toFixed(precision)}
      </Text>
    );
  };

  const rows = rowTitles.length > 0 ? rowTitles.length : Math.ceil(points.length / columnCount);

  return (
    <Table size="sm" variant="simple">
      {hasColHeaders && (
        <Thead>
          <Tr>
            {hasRowHeaders && <Th bg={headerBg} borderColor={borderColor} />}
            {columnTitles.map((col, ci) => (
              <Th key={ci} bg={headerBg} borderColor={borderColor} color={headerColor} fontSize="xs">
                {col}
              </Th>
            ))}
          </Tr>
        </Thead>
      )}
      <Tbody>
        {Array.from({ length: rows }, (_, ri) => (
          <Tr key={ri}>
            {hasRowHeaders && (
              <Td
                fontWeight="semibold"
                fontSize="xs"
                color={headerColor}
                bg={headerBg}
                borderColor={borderColor}
                whiteSpace="nowrap"
              >
                {rowTitles[ri] ?? ''}
              </Td>
            )}
            {Array.from({ length: columnCount }, (_, ci) => {
              const idx = ri * columnCount + ci;
              const pt = points[idx];
              return (
                <Td key={ci} fontSize="xs" borderColor={borderColor}>
                  {renderValue(pt)}
                </Td>
              );
            })}
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
}
