import React from 'react';

interface BoxProps {
  children: React.ReactNode;
  title?: string;
  footer?: React.ReactNode;
  width?: number | 'auto' | `${number}%`;
  height?: number | 'auto' | `${number}%`;
  borderColor?: string;
}

/**
 * A styled box container with optional title, notification, and footer
 * Matches the Go version's RenderBoxWithTitle styling
 */
export function Box({
  children,
  title,
  footer,
  width = '100%',
  height = '100%',
  borderColor = '#444444',
}: BoxProps) {
  return (
    <box
      flexDirection="column"
      width={width}
      height={height}
      borderStyle="rounded"
      borderColor={borderColor}
    >
      {/* Title row (only rendered when a title is provided) */}
      {title && (
        <box
          flexDirection="row"
          justifyContent="flex-start"
          height={1}
          marginBottom={1}
        >
          <text fg="#FFFFFF">
            <b>{title}</b>
          </text>
        </box>
      )}

      {/* Main content */}
      <box flexGrow={1} flexDirection="column">
        {children}
      </box>

      {/* Footer */}
      {footer && (
        <box marginTop={1}>
          {footer}
        </box>
      )}
    </box>
  );
}

interface SimpleBoxProps {
  children: React.ReactNode;
  focused?: boolean;
  width?: number | 'auto' | `${number}%`;
}

/**
 * A simple bordered box, optionally with focus highlighting
 */
export function SimpleBox({ children, focused = false, width }: SimpleBoxProps) {
  return (
    <box
      borderStyle="rounded"
      borderColor={focused ? '#FFFFFF' : '#444444'}
      width={width}
    >
      {children}
    </box>
  );
}
