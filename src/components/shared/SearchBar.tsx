import React from 'react';

interface SearchBarProps {
  value: string;
  matchCount?: number;
  totalCount?: number;
  placeholder?: string;
}

/**
 * A search bar component for filtering entries
 */
export function SearchBar({ value, matchCount, totalCount, placeholder = 'Type to search...' }: SearchBarProps) {
  return (
    <box flexDirection="row" backgroundColor="#222222" paddingLeft={1} paddingRight={1}>
      <text fg="#FFFFFF"><b>/</b></text>
      <text fg="#CCCCCC">
        {value || placeholder}
      </text>
      {matchCount !== undefined && totalCount !== undefined && (
        <text fg="#888888">
          {' '}({matchCount}/{totalCount})
        </text>
      )}
    </box>
  );
}

interface InputFieldProps {
  label?: string;
  value: string;
  placeholder?: string;
  focused?: boolean;
  error?: string;
}

/**
 * A styled input field component
 */
export function InputField({ label, value, placeholder, focused = false, error }: InputFieldProps) {
  return (
    <box flexDirection="column" gap={1}>
      {label && (
        <text fg="#888888">{label}</text>
      )}
      <box
        borderStyle="single"
        borderColor={focused ? '#FFFFFF' : '#888888'}
        paddingLeft={1}
        paddingRight={1}
      >
        <text fg={focused ? '#FFFFFF' : '#CCCCCC'}>
          {value || placeholder || ''}
        </text>
      </box>
      {error && (
        <text fg="#CCCCCC">{error}</text>
      )}
    </box>
  );
}
