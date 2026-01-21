import React, { ReactNode } from 'react';

type Alignment = 'left' | 'right' | 'center';

export interface Column {
  header: string;
  minWidth: number;
  align?: Alignment;
}

interface TableProps {
  columns: Column[];
  rows: ReactNode[][];
  selectedIndex?: number;
  totalsRow?: ReactNode[];
}

const BORDER_COLOR = '#888888';
const HEADER_COLOR = '#FFFFFF';
const TEXT_COLOR = '#CCCCCC';
const SELECTED_BG = '#333333';

const justify = (align?: Alignment) => {
  if (align === 'right') return 'flex-end';
  if (align === 'center') return 'center';
  return 'flex-start';
};

function buildSeparator(columns: Column[], left: string, mid: string, right: string) {
  const parts = columns.map(col => '─'.repeat(col.minWidth + 2));
  return left + parts.join(mid) + right;
}

function renderCellContent(cell: ReactNode, isSelected: boolean) {
  if (typeof cell === 'string' || typeof cell === 'number') {
    return <text fg={isSelected ? HEADER_COLOR : TEXT_COLOR}>{String(cell)}</text>;
  }
  return cell;
}

export function Table({ columns, rows, selectedIndex = -1, totalsRow }: TableProps) {
  const topBorder = buildSeparator(columns, '┌', '┬', '┐');
  const headerSeparator = buildSeparator(columns, '├', '┼', '┤');
  const bottomBorder = buildSeparator(columns, '└', '┴', '┘');

  const renderRow = (cells: ReactNode[], opts: { isSelected?: boolean; isHeader?: boolean } = {}) => {
    const isSelected = !!opts.isSelected;
    const isHeader = !!opts.isHeader;

    return (
      <box flexDirection="row" backgroundColor={isSelected ? SELECTED_BG : undefined}>
        <text fg={BORDER_COLOR}>│</text>
        {columns.map((col, colIdx) => (
          <React.Fragment key={colIdx}>
            <box
              flexDirection="row"
              width={col.minWidth + 2}
              paddingLeft={1}
              paddingRight={1}
              justifyContent={justify(col.align)}
            >
              {isHeader ? (
                <text fg={HEADER_COLOR}>
                  <b>{col.header}</b>
                </text>
              ) : (
                renderCellContent(cells[colIdx] ?? '', isSelected)
              )}
            </box>
            <text fg={BORDER_COLOR}>│</text>
          </React.Fragment>
        ))}
      </box>
    );
  };

  return (
    <box flexDirection="column">
      {/* Top border */}
      <text fg={BORDER_COLOR}>{topBorder}</text>

      {/* Header row */}
      {renderRow([], { isHeader: true })}

      {/* Header separator */}
      <text fg={BORDER_COLOR}>{headerSeparator}</text>

      {/* Data rows */}
      {rows.map((row, rowIdx) => renderRow(row, { isSelected: rowIdx === selectedIndex }))}

      {/* Totals separator + row (keeps vertical lines aligned) */}
      {totalsRow && (
        <>
          <text fg={BORDER_COLOR}>{headerSeparator}</text>
          {renderRow(totalsRow)}
        </>
      )}

      {/* Bottom border */}
      <text fg={BORDER_COLOR}>{bottomBorder}</text>
    </box>
  );
}
