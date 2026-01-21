import React, { useState, useCallback } from 'react';
import { useKeyboard, useTerminalDimensions } from '@opentui/react';
import { Box as StyledBox } from './shared/Box';
import { StatusBar } from './shared/StatusBar';
import { SearchBar } from './shared/SearchBar';
import type { Day, Entry } from '../types';
import { createEntry, cloneEntry } from '../types/entry';
import { filterEntries, totalCAD, totalIDR, filteredTotalCAD, filteredTotalIDR } from '../types/day';
import { formatCAD, formatIDR, formatDateDisplay, truncate } from '../utils/format';
import { CurrencyConverter, currencyConverter } from '../services/currencyConverter';
import { UndoManager, undoManager } from '../services/undoManager';

export enum EditorMode {
  Normal = 'Normal',
  Search = 'Search',
  InlineEdit = 'InlineEdit',
  ScreenTime = 'ScreenTime',
  Journal = 'Journal',
}

export enum EditColumn {
  Description = 0,
  IDR = 1,
  CAD = 2,
}

export enum EditorAction {
  None = 'None',
  Back = 'Back',
  Saved = 'Saved',
  Reload = 'Reload',
}

interface EditorProps {
  day: Day;
  onAction: (action: EditorAction) => void;
  onSave: (day: Day) => void;
  converter?: CurrencyConverter;
  undo?: UndoManager;
}

export function Editor({
  day,
  onAction,
  onSave,
  converter = currencyConverter,
  undo = undoManager,
}: EditorProps) {
  // Mode state
  const [mode, setMode] = useState<EditorMode>(EditorMode.Normal);
  const [selectedRow, setSelectedRow] = useState(0);
  const [selectedCol, setSelectedCol] = useState<EditColumn>(EditColumn.Description);

  // Edit state
  const [editValue, setEditValue] = useState('');
  const [editOriginal, setEditOriginal] = useState<Entry | null>(null);
  const [isNewEntry, setIsNewEntry] = useState(false);
  const [hasTypedInCell, setHasTypedInCell] = useState(false);

  // Search state
  const [searchQuery, setSearchQuery] = useState('');

  // Delete confirmation
  const [pendingDelete, setPendingDelete] = useState(false);

  // Screen time edit
  const [screenTimeValue, setScreenTimeValue] = useState(day.screenTime);

  // Journal edit
  const [journalValue, setJournalValue] = useState(day.journal);

  // Notification
  const [notification, setNotification] = useState<string | null>(null);

  // Terminal dimensions for table layout
  const { height: terminalRows } = useTerminalDimensions();

  // Filtered entries
  const entries = filterEntries(day, searchQuery);

  // Ensure selected row is valid
  const safeSelectedRow = Math.min(selectedRow, Math.max(0, entries.length - 1));

  const showNotification = useCallback((msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 3000);
  }, []);

  const startInlineEdit = useCallback((entry: Entry, col: EditColumn) => {
    setEditOriginal(cloneEntry(entry));
    setSelectedCol(col);
    setHasTypedInCell(false);

    switch (col) {
      case EditColumn.Description:
        setEditValue(entry.description);
        break;
      case EditColumn.CAD:
        setEditValue(entry.cad.toFixed(2));
        break;
      case EditColumn.IDR:
        setEditValue(Math.round(entry.idr).toString());
        break;
    }

    setMode(EditorMode.InlineEdit);
  }, []);

  const saveCurrentCell = useCallback((entry: Entry) => {
    const val = editValue.trim().replace(/,/g, '');

    switch (selectedCol) {
      case EditColumn.Description:
        entry.description = editValue;
        break;
      case EditColumn.CAD:
        if (val === '') {
          entry.cad = 0;
          entry.idr = 0;
        } else {
          const cad = parseFloat(val);
          if (!isNaN(cad)) {
            entry.cad = cad;
            entry.idr = converter.cadToIDR(cad);
          }
        }
        break;
      case EditColumn.IDR:
        if (val === '') {
          entry.idr = 0;
          entry.cad = 0;
        } else {
          const idr = parseFloat(val);
          if (!isNaN(idr)) {
            entry.idr = idr;
            entry.cad = converter.idrToCAD(idr);
          }
        }
        break;
    }
  }, [editValue, selectedCol, converter]);

  const finishEdit = useCallback((entry: Entry, save: boolean) => {
    if (!save) {
      // Restore original
      if (editOriginal) {
        if (editOriginal.description === '' && isNewEntry) {
          // Cancel new entry
          const idx = day.entries.findIndex(e => e.id === entry.id);
          if (idx !== -1) day.entries.splice(idx, 1);
        } else {
          Object.assign(entry, editOriginal);
        }
      }
    } else {
      // Save changes
      if (entry.description === '') {
        // Remove empty entry
        const idx = day.entries.findIndex(e => e.id === entry.id);
        if (idx !== -1) day.entries.splice(idx, 1);
      } else {
        // Record for undo
        if (editOriginal) {
          if (editOriginal.description === '' && isNewEntry) {
            undo.recordAddEntry(day.date, entry);
            showNotification('Added');
          } else {
            undo.recordEditEntry(day.date, editOriginal, entry);
            showNotification('Saved');
          }
        }
        onSave(day);
      }
    }

    setMode(EditorMode.Normal);
    setEditOriginal(null);
    setIsNewEntry(false);
  }, [day, editOriginal, isNewEntry, undo, onSave, showNotification]);

  const addNewEntry = useCallback(() => {
    const entry = createEntry(day.date, '', 0, 0, day.screenTime);
    day.entries.push(entry);
    setSelectedRow(day.entries.length - 1);
    setIsNewEntry(true);
    startInlineEdit(entry, EditColumn.Description);
  }, [day, startInlineEdit]);

  const performUndo = useCallback(async () => {
    const msg = await undo.undo();
    if (msg) {
      showNotification(msg);
      onAction(EditorAction.Reload);
    } else {
      showNotification('No undo');
    }
  }, [undo, showNotification, onAction]);

  const handleNormalMode = (key: string, char: string) => {
    if (pendingDelete && key !== 'd') {
      setPendingDelete(false);
    }

    if (key === 'up') {
      setSelectedRow(prev => Math.max(0, prev - 1));
    } else if (key === 'down') {
      setSelectedRow(prev => Math.min(entries.length - 1, prev + 1));
    } else if (key === 'left') {
      setSelectedCol(prev => Math.max(EditColumn.Description, prev - 1) as EditColumn);
    } else if (key === 'right') {
      setSelectedCol(prev => Math.min(EditColumn.CAD, prev + 1) as EditColumn);
    } else if (key === '/' || char === '/') {
      setMode(EditorMode.Search);
    } else if (key === 'a' || char === 'a') {
      addNewEntry();
    } else if (key === 'return' && entries.length > 0) {
      const entry = entries[safeSelectedRow];
      startInlineEdit(entry, selectedCol);
    } else if (key === 'd' || char === 'd') {
      if (pendingDelete && entries.length > 0) {
        const entry = entries[safeSelectedRow];
        undo.recordDeleteEntry(day.date, entry);
        const idx = day.entries.findIndex(e => e.id === entry.id);
        if (idx !== -1) day.entries.splice(idx, 1);
        showNotification('Deleted');
        onSave(day);
        setPendingDelete(false);
      } else {
        setPendingDelete(true);
      }
    } else if (key === 's' || char === 's') {
      setScreenTimeValue(day.screenTime);
      setMode(EditorMode.ScreenTime);
    } else if (key === 'j' || char === 'j') {
      setJournalValue(day.journal);
      setMode(EditorMode.Journal);
    } else if (key === 'u' || char === 'u') {
      performUndo();
    } else if (key === 'escape') {
      if (searchQuery) {
        setSearchQuery('');
      } else {
        onAction(EditorAction.Back);
      }
    } else if (key === 'q' || char === 'q') {
      onAction(EditorAction.Back);
    }
  };

  const handleInlineEditMode = (key: string, char: string) => {
    const entry = entries[safeSelectedRow];
    if (!entry) return;

    if (key === 'tab') {
      saveCurrentCell(entry);
      if (selectedCol < EditColumn.CAD) {
        startInlineEdit(entry, (selectedCol + 1) as EditColumn);
      } else if (isNewEntry) {
        finishEdit(entry, true);
      } else {
        startInlineEdit(entry, EditColumn.Description);
      }
    } else if (key === 'return') {
      saveCurrentCell(entry);
      if (selectedCol === EditColumn.Description && isNewEntry) {
        if (entry.description === '') {
          finishEdit(entry, false);
        } else {
          startInlineEdit(entry, EditColumn.IDR);
        }
      } else {
        finishEdit(entry, true);
      }
    } else if (key === 'escape') {
      finishEdit(entry, false);
    } else if (key === 'backspace') {
      setEditValue(prev => prev.slice(0, -1));
      setHasTypedInCell(true);
    } else if (char && char.length === 1) {
      // First keypress in currency column clears the field
      if (!hasTypedInCell && (selectedCol === EditColumn.CAD || selectedCol === EditColumn.IDR)) {
        setEditValue(char);
      } else {
        setEditValue(prev => prev + char);
      }
      setHasTypedInCell(true);
    }
  };

  const handleSearchMode = (key: string, char: string) => {
    if (key === 'escape' || key === 'return') {
      setMode(EditorMode.Normal);
    } else if (key === 'backspace') {
      setSearchQuery(prev => prev.slice(0, -1));
    } else if (char && char.length === 1) {
      setSearchQuery(prev => prev + char);
    }
  };

  const handleScreenTimeMode = (key: string, char: string) => {
    if (key === 'return') {
      const oldScreenTime = day.screenTime;
      day.screenTime = screenTimeValue.trim();
      for (const entry of day.entries) {
        entry.screenTime = day.screenTime;
      }
      undo.recordSetScreenTime(day.date, oldScreenTime, day.screenTime);
      showNotification('Saved');
      onSave(day);
      setMode(EditorMode.Normal);
    } else if (key === 'escape') {
      setMode(EditorMode.Normal);
    } else if (key === 'backspace') {
      setScreenTimeValue(prev => prev.slice(0, -1));
    } else if (char && char.length === 1) {
      setScreenTimeValue(prev => prev + char);
    }
  };

  const handleJournalMode = (key: string, char: string) => {
    if (key === 'escape') {
      const oldJournal = day.journal;
      day.journal = journalValue.trim();
      if (day.journal !== oldJournal) {
        undo.recordSetJournal(day.date, oldJournal, day.journal);
        showNotification('Saved');
        onSave(day);
      }
      setMode(EditorMode.Normal);
    } else if (key === 'backspace') {
      setJournalValue(prev => prev.slice(0, -1));
    } else if (key === 'return') {
      setJournalValue(prev => prev + '\n');
    } else if (char && char.length === 1) {
      setJournalValue(prev => prev + char);
    }
  };

  useKeyboard((event) => {
    const key = event.name;
    const char = event.raw || '';

    // Handle different modes
    switch (mode) {
      case EditorMode.Normal:
        handleNormalMode(key, char);
        break;
      case EditorMode.InlineEdit:
        handleInlineEditMode(key, char);
        break;
      case EditorMode.Search:
        handleSearchMode(key, char);
        break;
      case EditorMode.ScreenTime:
        handleScreenTimeMode(key, char);
        break;
      case EditorMode.Journal:
        handleJournalMode(key, char);
        break;
    }
  });

  // Mode indicator
  const getModeText = () => {
    switch (mode) {
      case EditorMode.Search: return 'SEARCH';
      case EditorMode.InlineEdit: return isNewEntry ? 'ADD' : 'EDIT';
      case EditorMode.ScreenTime: return 'SCREEN TIME';
      case EditorMode.Journal: return 'JOURNAL';
      default: return pendingDelete ? 'd...' : 'NORMAL';
    }
  };

  // Help items based on mode
  const getHelpItems = () => {
    switch (mode) {
      case EditorMode.InlineEdit:
        return [
          { key: 'Tab', description: 'next' },
          { key: 'Enter', description: 'save' },
          { key: 'Esc', description: 'cancel' },
        ];
      case EditorMode.Search:
        return [
          { key: 'Enter', description: 'confirm' },
          { key: 'Esc', description: 'exit' },
        ];
      case EditorMode.ScreenTime:
      case EditorMode.Journal:
        return [
          { key: 'Enter/Esc', description: 'save' },
        ];
      default:
        return [
          { key: 'Enter', description: 'edit' },
          { key: 'a', description: 'add' },
          { key: 'dd', description: 'del' },
          { key: 's', description: 'screen' },
          { key: 'j', description: 'journal' },
          { key: '/', description: 'search' },
        ];
    }
  };

  const cadTotal = searchQuery ? filteredTotalCAD(day, searchQuery) : totalCAD(day);
  const idrTotal = searchQuery ? filteredTotalIDR(day, searchQuery) : totalIDR(day);

  return (
    <StyledBox
      footer={
        <StatusBar
          modeInfo={getModeText()}
          rateInfo={converter.formatRate()}
          notification={notification || undefined}
          helpItems={getHelpItems()}
        />
      }
    >
      <box flexDirection="column">
        {/* Date + Screen time (top, centered, minimal width, bordered) */}
        <box justifyContent="center" alignItems="center" width="100%">
          <box
            borderStyle="single"
            borderColor="#888888"
            width="auto"
            paddingLeft={1}
            paddingRight={1}
          >
            <box flexDirection="row">
              <text fg="#FFFFFF">
                <b>{formatDateDisplay(day.date)}</b>
              </text>
              {/* Separator line between date and screen time */}
              <text fg="#888888">{' | '}</text>
              {mode === EditorMode.ScreenTime ? (
                <text fg="#FFFFFF">
                  <b>Screen Time:</b>{' '}{screenTimeValue || '(type here)'}_
                </text>
              ) : (
                <text fg="#FFFFFF">
                  <b>Screen Time:</b>{' '}{day.screenTime || 'not set'}
                </text>
              )}
            </box>
          </box>
        </box>

        {/* Search bar */}
        {(mode === EditorMode.Search || searchQuery) && (
          <box marginTop={0}>
            <SearchBar
              value={searchQuery}
              matchCount={entries.length}
              totalCount={day.entries.length}
            />
          </box>
        )}

        {/* Main content row: Ledger (left) and Journal (right) */}
        <box
          flexDirection="row"
          gap={1}
          flexGrow={1}
          height="100%"
          alignItems="stretch"
					marginLeft={1}
					marginRight={1}
        >
          {/* Ledger panel (left) - manually drawn table */}
          {(() => {
            // Use terminal width to calculate panel width (2/3 minus margin for borders/safety)
            const { width: termWidth } = useTerminalDimensions();
            const panelWidth = Math.floor((termWidth * 2) / 3) - 4;

            // Distribute width: Description gets 45%, IDR gets 30%, CAD gets 25%
            const totalContentWidth = Math.max(panelWidth - 4, 30); // subtract 4 for │ chars, min 30
            const col1 = Math.floor(totalContentWidth * 0.45);
            const col2 = Math.floor(totalContentWidth * 0.30);
            const col3 = totalContentWidth - col1 - col2; // remainder to CAD

            // Helper to pad string to exact width
            const pad = (s: string, w: number) => s.length >= w ? s.slice(0, w) : s + ' '.repeat(w - s.length);

            // Build rows - consistent structure for all
            const makeBorder = (l: string, m: string, r: string) =>
              l + '─'.repeat(col1) + m + '─'.repeat(col2) + m + '─'.repeat(col3) + r;

            const topBorder = makeBorder('┌', '┬', '┐');
            const headerSep = makeBorder('├', '┼', '┤');
            const bottomBorder = makeBorder('└', '┴', '┘');

            // Row component with proper border coloring
            // selectedCell: 0=desc, 1=idr, 2=cad, undefined=none
            const TableRow = ({ c1, c2, c3, contentColor, bold, selectedCell }: {
              c1: string; c2: string; c3: string;
              contentColor: string; bold?: boolean; selectedCell?: number;
            }) => {
              const cell1 = pad(' ' + c1, col1);
              const cell2 = pad(' ' + c2, col2);
              const cell3 = pad(' ' + c3, col3);

              // Selected cell gets highlight color, others get contentColor
              const color1 = selectedCell === 0 ? '#C9637D' : contentColor;
              const color2 = selectedCell === 1 ? '#C9637D' : contentColor;
              const color3 = selectedCell === 2 ? '#C9637D' : contentColor;

              const content = bold ? (
                <text>
                  <span fg="#888888">│</span><span fg={color1}><b>{cell1}</b></span>
                  <span fg="#888888">│</span><span fg={color2}><b>{cell2}</b></span>
                  <span fg="#888888">│</span><span fg={color3}><b>{cell3}</b></span>
                  <span fg="#888888">│</span>
                </text>
              ) : (
                <text>
                  <span fg="#888888">│</span><span fg={color1}>{cell1}</span>
                  <span fg="#888888">│</span><span fg={color2}>{cell2}</span>
                  <span fg="#888888">│</span><span fg={color3}>{cell3}</span>
                  <span fg="#888888">│</span>
                </text>
              );
              return <box>{content}</box>;
            };

            // Calculate filler rows so grid lines run down to the totals row
            // Target roughly the available body height (terminal height minus surrounding UI).
            const bodyTargetRows = Math.max(entries.length, terminalRows - 12);
            const fillerRows = Math.max(0, bodyTargetRows - entries.length);

            return (
              <box flexDirection="column" width="66%" height="100%">
                {/* Header section */}
                <box><text fg="#888888">{topBorder}</text></box>
                <TableRow c1="Description" c2="IDR" c3="CAD" contentColor="#FFFFFF" bold />
                <box><text fg="#888888">{headerSep}</text></box>

                {/* Content section - flexGrow to fill available space */}
                <box flexDirection="column" flexGrow={1} overflow="hidden">
                  {/* Data rows */}
                  {entries.length === 0 ? (
                    <TableRow c1={searchQuery ? 'No matches' : "Press 'a'"} c2="" c3="" contentColor="#888888" />
                  ) : (
                    entries.map((entry, idx) => {
                      const isSelected = idx === safeSelectedRow;
                      const isEditing = mode === EditorMode.InlineEdit && isSelected;

                      let desc = truncate(entry.description, col1 - 4);
                      let idr = formatIDR(entry.idr);
                      let cad = formatCAD(entry.cad);

                      if (isEditing) {
                        if (selectedCol === EditColumn.Description) {
                          desc = editValue + '_';
                        } else if (selectedCol === EditColumn.IDR) {
                          idr = 'Rp ' + editValue + '_';
                        } else if (selectedCol === EditColumn.CAD) {
                          cad = '$' + editValue + '_';
                        }
                      }

                      return (
                        <TableRow
                          key={entry.id}
                          c1={desc}
                          c2={idr}
                          c3={cad}
                          contentColor={isSelected ? '#FFFFFF' : '#888888'}
                          bold={isSelected}
                          selectedCell={isSelected ? selectedCol : undefined}
                        />
                      );
                    })
                  )}
                  {/* Filler rows with side borders - render enough to fill, overflow hidden */}
                  {Array.from({ length: fillerRows }).map((_, i) => (
                    <TableRow key={`filler-${i}`} c1="" c2="" c3="" contentColor="#888888" />
                  ))}
                </box>

                {/* Total section */}
                <box><text fg="#888888">{headerSep}</text></box>
                <TableRow c1="Total:" c2={formatIDR(idrTotal)} c3={formatCAD(cadTotal)} contentColor="#FFFFFF" bold />
                <box><text fg="#888888">{bottomBorder}</text></box>
              </box>
            );
          })()}

          {/* Journal panel (right) */}
          <box
            flexDirection="column"
            borderStyle="single"
            borderColor="#888888"
            title="Journal"
            titleAlignment="center"
            width="33%"
            height="100%"
          >
            <box paddingLeft={1} paddingRight={1} paddingTop={1} paddingBottom={1}>
              {mode === EditorMode.Journal ? (
                <box flexDirection="column">
                  <text fg="#FFFFFF">{journalValue}_</text>
                  <text fg="#888888">
                    <i>Esc: save</i>
                  </text>
                </box>
              ) : (
                <text fg={day.journal ? '#CCCCCC' : '#888888'}>
                  {day.journal ? day.journal : <i>Press 'j' to add journal</i>}
                </text>
              )}
            </box>
          </box>
        </box>
      </box>
    </StyledBox>
  );
}
