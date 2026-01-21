import React, { useState } from 'react';
import { useKeyboard, useTerminalDimensions } from '@opentui/react';
import { Box as StyledBox } from './shared/Box';
import { StatusBar } from './shared/StatusBar';
import { SearchBar } from './shared/SearchBar';
import type { DateRange } from '../types';
import { rangeFilteredTotalCAD, rangeFilteredTotalIDR, formatRangeDisplay } from '../types/dateRange';
import { formatCAD, formatIDR, formatDateShort, truncate } from '../utils/format';

export enum RangeViewAction {
  None = 'None',
  Back = 'Back',
  SelectDay = 'SelectDay',
}

interface RangeViewProps {
  dateRange: DateRange;
  onAction: (action: RangeViewAction, selectedDate?: Date) => void;
}

interface DisplayRow {
  id: string;
  date: Date;
  description: string;
  idr: number;
  cad: number;
  isJournal: boolean;
}

export function RangeView({ dateRange, onAction }: RangeViewProps) {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchMode, setSearchMode] = useState(false);

  const { width: termWidth, height: terminalRows } = useTerminalDimensions();

  // Build display rows: entries + journal entries
  const allRows: DisplayRow[] = [];
  for (const day of dateRange.days) {
    // Add regular entries
    for (const entry of day.entries) {
      if (!searchQuery || entry.description.toLowerCase().includes(searchQuery.toLowerCase())) {
        allRows.push({
          id: entry.id,
          date: entry.date,
          description: entry.description,
          idr: entry.idr,
          cad: entry.cad,
          isJournal: false,
        });
      }
    }
    // Add journal entry if exists
    if (day.journal && day.journal.trim()) {
      const journalPreview = day.journal.replace(/\n/g, ' ').trim();
      if (!searchQuery || journalPreview.toLowerCase().includes(searchQuery.toLowerCase())) {
        allRows.push({
          id: `journal-${day.date.toISOString()}`,
          date: day.date,
          description: `[Journal] ${journalPreview}`,
          idr: 0,
          cad: 0,
          isJournal: true,
        });
      }
    }
  }

  // Sort by date descending (most recent first)
  allRows.sort((a, b) => b.date.getTime() - a.date.getTime());

  const safeSelectedIndex = Math.min(selectedIndex, Math.max(0, allRows.length - 1));

  useKeyboard((event) => {
    const key = event.name;
    const char = event.raw || '';

    if (searchMode) {
      if (key === 'escape' || key === 'return') {
        setSearchMode(false);
      } else if (key === 'backspace') {
        setSearchQuery(prev => prev.slice(0, -1));
      } else if (char && char.length === 1) {
        setSearchQuery(prev => prev + char);
      }
      return;
    }

    if (key === 'escape' || char === 'q') {
      if (searchQuery) {
        setSearchQuery('');
      } else {
        onAction(RangeViewAction.Back);
      }
    } else if (key === 'return' && allRows.length > 0) {
      const row = allRows[safeSelectedIndex];
      onAction(RangeViewAction.SelectDay, row.date);
    } else if (key === 'up') {
      setSelectedIndex(prev => Math.max(0, prev - 1));
    } else if (key === 'down') {
      setSelectedIndex(prev => Math.min(allRows.length - 1, prev + 1));
    } else if (char === '/') {
      setSearchMode(true);
    }
  });

  // Calculate column widths based on terminal width
  // Account for: StyledBox border (2), safety margin (4), table borders (5 │ chars)
  const availableWidth = termWidth - 6;
  const totalContentWidth = Math.max(availableWidth - 5, 50); // min 50 chars for content
  const col1 = 12; // Date - fixed
  const col2 = Math.floor((totalContentWidth - col1) * 0.55); // Description
  const col3 = Math.floor((totalContentWidth - col1) * 0.25); // IDR
  const col4 = totalContentWidth - col1 - col2 - col3; // CAD gets remainder

  const pad = (s: string, w: number) => s.length >= w ? s.slice(0, w) : s + ' '.repeat(w - s.length);

  const makeBorder = (l: string, m: string, r: string) =>
    l + '─'.repeat(col1) + m + '─'.repeat(col2) + m + '─'.repeat(col3) + m + '─'.repeat(col4) + r;

  const topBorder = makeBorder('┌', '┬', '┐');
  const headerSep = makeBorder('├', '┼', '┤');
  const bottomBorder = makeBorder('└', '┴', '┘');

  const TableRow = ({ c1, c2, c3, c4, contentColor, bold, isJournal }: {
    c1: string; c2: string; c3: string; c4: string;
    contentColor: string; bold?: boolean; isJournal?: boolean;
  }) => {
    const cell1 = pad(' ' + c1, col1);
    const cell2 = pad(' ' + c2, col2);
    const cell3 = pad(' ' + c3, col3);
    const cell4 = pad(' ' + c4, col4);

    const journalColor = isJournal ? '#C9637D' : contentColor;

    const content = bold ? (
      <text>
        <span fg="#888888">│</span><span fg={contentColor}><b>{cell1}</b></span>
        <span fg="#888888">│</span><span fg={journalColor}><b>{cell2}</b></span>
        <span fg="#888888">│</span><span fg={contentColor}><b>{cell3}</b></span>
        <span fg="#888888">│</span><span fg={contentColor}><b>{cell4}</b></span>
        <span fg="#888888">│</span>
      </text>
    ) : (
      <text>
        <span fg="#888888">│</span><span fg={contentColor}>{cell1}</span>
        <span fg="#888888">│</span><span fg={journalColor}>{cell2}</span>
        <span fg="#888888">│</span><span fg={contentColor}>{cell3}</span>
        <span fg="#888888">│</span><span fg={contentColor}>{cell4}</span>
        <span fg="#888888">│</span>
      </text>
    );
    return <box>{content}</box>;
  };

  const cadTotal = rangeFilteredTotalCAD(dateRange, searchQuery);
  const idrTotal = rangeFilteredTotalIDR(dateRange, searchQuery);

  const helpItems = searchMode
    ? [{ key: 'Enter', description: 'confirm' }, { key: 'Esc', description: 'exit' }]
    : [{ key: 'Enter', description: 'view day' }, { key: '/', description: 'search' }, { key: 'q', description: 'back' }];

  // Calculate filler rows - use terminal height as upper bound, accounting for UI chrome
  const maxFillerRows = Math.max(0, terminalRows - 15 - allRows.length);

  return (
    <StyledBox
      title={formatRangeDisplay(dateRange)}
      footer={<StatusBar notification={`${allRows.length} items`} helpItems={helpItems} />}
    >
      <box flexDirection="column" height="100%">
        {/* Search bar */}
        {(searchMode || searchQuery) && (
          <box marginBottom={1}>
            <SearchBar
              value={searchQuery}
              matchCount={allRows.length}
              totalCount={dateRange.days.reduce((sum, d) => sum + d.entries.length + (d.journal ? 1 : 0), 0)}
            />
          </box>
        )}

        {/* Table with flex layout */}
        <box flexDirection="column" height="100%">
          {/* Header section */}
          <box><text fg="#888888">{topBorder}</text></box>
          <TableRow c1="Date" c2="Description" c3="IDR" c4="CAD" contentColor="#FFFFFF" bold />
          <box><text fg="#888888">{headerSep}</text></box>

          {/* Content section - flexGrow to fill available space */}
          <box flexDirection="column" flexGrow={1} overflow="hidden">
            {/* Data rows */}
            {allRows.length === 0 ? (
              <TableRow c1="" c2={searchQuery ? 'No matches' : 'No entries'} c3="" c4="" contentColor="#888888" />
            ) : (
              allRows.map((row, idx) => {
                const isSelected = idx === safeSelectedIndex;
                return (
                  <TableRow
                    key={row.id}
                    c1={formatDateShort(row.date)}
                    c2={truncate(row.description, col2 - 2)}
                    c3={row.isJournal ? '' : formatIDR(row.idr)}
                    c4={row.isJournal ? '' : formatCAD(row.cad)}
                    contentColor={isSelected ? '#FFFFFF' : '#888888'}
                    bold={isSelected}
                    isJournal={row.isJournal}
                  />
                );
              })
            )}
            {/* Filler rows with side borders - render enough to fill, overflow hidden */}
            {Array.from({ length: maxFillerRows }).map((_, i) => (
              <TableRow key={`filler-${i}`} c1="" c2="" c3="" c4="" contentColor="#888888" />
            ))}
          </box>

          {/* Total section */}
          <box><text fg="#888888">{headerSep}</text></box>
          <TableRow c1="" c2="Total:" c3={formatIDR(idrTotal)} c4={formatCAD(cadTotal)} contentColor="#FFFFFF" bold />
          <box><text fg="#888888">{bottomBorder}</text></box>
        </box>
      </box>
    </StyledBox>
  );
}
