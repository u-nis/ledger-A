import React, { useState } from 'react';
import { useKeyboard } from '@opentui/react';
import { Box as StyledBox } from './shared/Box';
import { StatusBar } from './shared/StatusBar';
import { parseDateInput, today, addDays, isSameDay } from '../utils/date';
import { formatDateDisplay } from '../utils/format';

export enum DatePickerAction {
  None = 'None',
  Selected = 'Selected',
  Cancelled = 'Cancelled',
}

interface Preset {
  key: string;
  label: string;
  getRange: () => { start: Date; end: Date };
}

interface DatePickerProps {
  onAction: (action: DatePickerAction, startDate?: Date, endDate?: Date) => void;
}

export function DatePicker({ onAction }: DatePickerProps) {
  const now = today();
  const [cursorDate, setCursorDate] = useState(now);
  const [startDate, setStartDate] = useState<Date | null>(null);
  const [endDate, setEndDate] = useState<Date | null>(null);
  const [inputMode, setInputMode] = useState(false);
  const [inputValue, setInputValue] = useState('');

  const presets: Preset[] = [
    { key: '1', label: 'Today', getRange: () => ({ start: now, end: now }) },
    { key: '2', label: 'Yesterday', getRange: () => { const d = addDays(now, -1); return { start: d, end: d }; } },
    { key: '3', label: 'Last 7 days', getRange: () => ({ start: addDays(now, -6), end: now }) },
    { key: '4', label: 'Last 30 days', getRange: () => ({ start: addDays(now, -29), end: now }) },
    { key: '5', label: 'This month', getRange: () => ({ start: new Date(now.getFullYear(), now.getMonth(), 1), end: now }) },
    { key: '6', label: 'Last month', getRange: () => ({ start: new Date(now.getFullYear(), now.getMonth() - 1, 1), end: new Date(now.getFullYear(), now.getMonth(), 0) }) },
  ];

  const handlePresetSelect = (preset: Preset) => {
    const { start, end } = preset.getRange();
    setStartDate(start);
    setEndDate(end);
    setCursorDate(start);
  };

  const handleDateSelect = () => {
    if (!startDate) {
      setStartDate(cursorDate);
    } else if (!endDate) {
      let start = startDate;
      let end = cursorDate;
      if (end < start) [start, end] = [end, start];
      setStartDate(start);
      setEndDate(end);
    }
  };

  const handleConfirm = () => {
    if (startDate && endDate) {
      onAction(DatePickerAction.Selected, startDate, endDate);
    } else if (startDate) {
      onAction(DatePickerAction.Selected, startDate, startDate);
    } else {
      onAction(DatePickerAction.Selected, cursorDate, cursorDate);
    }
  };

  const handleTextInput = () => {
    const result = parseDateInput(inputValue, now.getFullYear());
    if (result.valid) {
      setStartDate(result.startDate);
      setEndDate(result.endDate);
      setCursorDate(result.startDate);
      setInputMode(false);
      setInputValue('');
    }
  };

  useKeyboard((event) => {
    const key = event.name;
    const char = event.raw || '';

    if (inputMode) {
      if (key === 'escape') {
        setInputMode(false);
        setInputValue('');
      } else if (key === 'return') {
        handleTextInput();
      } else if (key === 'backspace') {
        setInputValue(prev => prev.slice(0, -1));
      } else if (char && char.length === 1) {
        setInputValue(prev => prev + char);
      }
      return;
    }

    if (key === 'escape' || char === 'q') {
      onAction(DatePickerAction.Cancelled);
    } else if (key === 'return') {
      if (startDate && endDate) handleConfirm();
      else handleDateSelect();
    } else if (key === 'space') {
      handleDateSelect();
    } else if (char === '/') {
      setInputMode(true);
      setInputValue('');
    } else if (char === 'c') {
      setStartDate(null);
      setEndDate(null);
    } else if (key === 'up') {
      setCursorDate(prev => addDays(prev, -7));
    } else if (key === 'down') {
      setCursorDate(prev => addDays(prev, 7));
    } else if (key === 'left') {
      setCursorDate(prev => addDays(prev, -1));
    } else if (key === 'right') {
      setCursorDate(prev => addDays(prev, 1));
    } else if (char >= '1' && char <= '6') {
      const preset = presets[parseInt(char) - 1];
      if (preset) handlePresetSelect(preset);
    }
  });

  // Calendar rendering
  const year = cursorDate.getFullYear();
  const month = cursorDate.getMonth();
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const startOffset = firstDay.getDay();

  const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'];

  // Build calendar as text lines for perfect alignment
  const calendarLines: { text: string; colors: { start: number; end: number; fg: string; bg?: string; bold?: boolean }[] }[] = [];

  // Month/Year header
  const monthYear = `${monthNames[month]} ${year}`;
  const headerPad = Math.floor((21 - monthYear.length) / 2);
  calendarLines.push({
    text: ' '.repeat(headerPad) + monthYear + ' '.repeat(21 - headerPad - monthYear.length),
    colors: [{ start: 0, end: 21, fg: '#FFFFFF', bold: true }]
  });

  // Day names
  calendarLines.push({
    text: 'Su Mo Tu We Th Fr Sa',
    colors: [{ start: 0, end: 20, fg: '#888888' }]
  });

  // Calendar grid
  let dayNum = 1;
  for (let week = 0; week < 6; week++) {
    let line = '';
    const colors: { start: number; end: number; fg: string; bg?: string; bold?: boolean }[] = [];

    for (let dow = 0; dow < 7; dow++) {
      const pos = dow * 3;
      if ((week === 0 && dow < startOffset) || dayNum > lastDay.getDate()) {
        line += '   ';
      } else {
        const currentDate = new Date(year, month, dayNum);
        const isToday = isSameDay(currentDate, now);
        const isCursor = isSameDay(currentDate, cursorDate);
        const isStart = startDate && isSameDay(currentDate, startDate);
        const isEnd = endDate && isSameDay(currentDate, endDate);
        const inRange = startDate && endDate && currentDate >= startDate && currentDate <= endDate;

        const dayStr = String(dayNum).padStart(2, ' ');
        line += dayStr + ' ';

        let fg = '#888888';
        let bg: string | undefined;
        let bold = false;

        if (isStart || isEnd) {
          fg = '#000000';
          bg = '#FFFFFF';
          bold = true;
        } else if (inRange) {
          fg = '#FFFFFF';
          bg = '#444444';
        } else if (isCursor) {
          fg = '#000000';
          bg = '#C9637D';
          bold = true;
        } else if (isToday) {
          fg = '#FFFFFF';
          bold = true;
        }

        colors.push({ start: pos, end: pos + 2, fg, bg, bold });
        dayNum++;
      }
    }

    calendarLines.push({ text: line.trimEnd().padEnd(20, ' '), colors });
    if (dayNum > lastDay.getDate()) break;
  }

  // Selection display
  const selectionText = startDate && endDate
    ? `${formatDateDisplay(startDate)} → ${formatDateDisplay(endDate)}`
    : startDate
      ? `${formatDateDisplay(startDate)} → ...`
      : 'No selection';

  const helpItems = inputMode
    ? [{ key: 'Enter', description: 'parse' }, { key: 'Esc', description: 'cancel' }]
    : [{ key: 'Space', description: 'select' }, { key: 'Enter', description: 'confirm' }, { key: '/', description: 'type' }, { key: 'c', description: 'clear' }, { key: 'q', description: 'back' }];

  return (
    <StyledBox
      title="Query"
      footer={<StatusBar helpItems={helpItems} />}
    >
      <box flexDirection="column" alignItems="center" justifyContent="center" height="100%">
        <box flexDirection="row" gap={2}>
          {/* Calendar panel */}
          <box
            flexDirection="column"
            borderStyle="single"
            borderColor="#888888"
            title="Calendar"
            titleAlignment="center"
						paddingRight={2}
						paddingLeft={2}
						paddingBottom={1}
						paddingTop={1}
          >
            {calendarLines.map((line, i) => (
              <box key={i} marginBottom={i < calendarLines.length - 1 ? 1 : 0}>
                <text>
                  {line.colors.length === 0 ? (
                    <span fg="#888888">{line.text}</span>
                  ) : (
                    <>
                      {(() => {
                        const parts: React.ReactNode[] = [];
                        let lastEnd = 0;
                        line.colors.forEach((c, idx) => {
                          if (c.start > lastEnd) {
                            parts.push(<span key={`gap-${1}`} fg="#888888">{line.text.slice(lastEnd, c.start)}</span>);
                          }
                          const segment = line.text.slice(c.start, c.end);
                          parts.push(
                            <span key={idx} fg={c.fg} bg={c.bg}>
                              {c.bold ? <b>{segment}</b> : segment}
                            </span>
                          );
                          lastEnd = c.end;
                        });
                        if (lastEnd < line.text.length) {
                          parts.push(<span key="end" fg="#888888">{line.text.slice(lastEnd)}</span>);
                        }
                        return parts;
                      })()}
                    </>
                  )}
                </text>
              </box>
            ))}
          </box>

          {/* Presets panel */}
          <box
            flexDirection="column"
            borderStyle="single"
            borderColor="#888888"
            title="Quick Select"
            titleAlignment="center"
            padding={1}
          >
            {presets.map((preset, i) => (
              <box key={preset.key} flexDirection="row" marginBottom={i < presets.length - 1 ? 1 : 0}>
                <text>
                  <span fg="#C9637D"><b>{preset.key}</b></span>
                  <span fg="#888888"> {preset.label}</span>
                </text>
              </box>
            ))}
          </box>
        </box>

        {/* Selection display */}
        <box marginTop={2}>
          <text>
            <span fg="#888888">Selection: </span>
            <span fg="#FFFFFF"><b>{selectionText}</b></span>
          </text>
        </box>

        {/* Text input */}
        {inputMode && (
          <box marginTop={1} borderStyle="single" borderColor="#C9637D" paddingLeft={1} paddingRight={1}>
            <text>
              <span fg="#C9637D">/</span>
              <span fg="#FFFFFF">{inputValue || 'type date...'}</span>
              <span fg="#C9637D">_</span>
            </text>
          </box>
        )}
      </box>
    </StyledBox>
  );
}
