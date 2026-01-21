import React, { useState } from 'react';
import { useKeyboard } from '@opentui/react';
import { Box as StyledBox } from './shared/Box';
import { StatusBar } from './shared/StatusBar';

export enum DateInputAction {
  None = 'None',
  Submitted = 'Submitted',
  Cancelled = 'Cancelled',
}

interface DateInputProps {
  title: string;
  onAction: (action: DateInputAction, date?: Date) => void;
}

/**
 * Auto-inserts slashes at the right positions for MM/DD/YYYY format
 */
function autoInsertSlashes(s: string): string {
  const digits = s.replace(/\//g, '');
  let result = '';
  for (let i = 0; i < digits.length && i < 8; i++) {
    if (i === 2 || i === 4) {
      result += '/';
    }
    result += digits[i];
  }
  return result;
}

export function DateInput({ title, onAction }: DateInputProps) {
  const [value, setValue] = useState('');
  const [error, setError] = useState('');

  useKeyboard((event) => {
    const key = event.name;
    const char = event.raw || '';

    if (key === 'escape' || key === 'q' || char === 'q') {
      onAction(DateInputAction.Cancelled);
    } else if (key === 'return') {
      // Validate and submit
      if (value.length !== 10) {
        setError('Please enter complete date (MM/DD/YYYY)');
        return;
      }

      const match = value.match(/^(\d{2})\/(\d{2})\/(\d{4})$/);
      if (!match) {
        setError('Invalid date format');
        return;
      }

      const month = parseInt(match[1], 10);
      const day = parseInt(match[2], 10);
      const year = parseInt(match[3], 10);

      if (month < 1 || month > 12) {
        setError('Invalid month');
        return;
      }

      const date = new Date(year, month - 1, day);
      if (isNaN(date.getTime())) {
        setError('Invalid date');
        return;
      }

      onAction(DateInputAction.Submitted, date);
    } else if (key === 'backspace') {
      setValue(prev => {
        let newVal = prev.slice(0, -1);
        if (newVal.endsWith('/')) {
          newVal = newVal.slice(0, -1);
        }
        return newVal;
      });
      setError('');
    } else if (/^\d$/.test(char)) {
      setValue(prev => {
        const newVal = autoInsertSlashes(prev + char);
        return newVal.slice(0, 10);
      });
      setError('');
    }
  });

  const helpItems = [
    { key: 'Enter', description: 'confirm' },
    { key: 'Backspace', description: 'delete' },
    { key: 'Esc', description: 'back' },
  ];

  return (
    <StyledBox
      title={title}
      footer={<StatusBar helpItems={helpItems} />}
    >
      <box flexDirection="column" alignItems="center" gap={2}>
        <text fg="#888888">Date:</text>

        <box
          borderStyle="single"
          borderColor={error ? '#888888' : '#FFFFFF'}
          paddingLeft={2}
          paddingRight={2}
          paddingTop={1}
          paddingBottom={1}
        >
          <text fg="#FFFFFF">
            {value || 'MM/DD/YYYY'}
            <text fg="#FFFFFF">_</text>
          </text>
        </box>

        <text fg="#888888">
          <i>Type numbers - slashes are added automatically</i>
        </text>

        {error && (
          <text fg="#CCCCCC">
            Error: {error}
          </text>
        )}
      </box>
    </StyledBox>
  );
}
